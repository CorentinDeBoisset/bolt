package jobexec

import (
	"context"
	"sync"

	"github.com/corentindeboisset/bolt/pkg/cfg"
	"github.com/corentindeboisset/bolt/pkg/cmdrunr"
)

type TaskState int

const (
	STATE_NOT_STARTED TaskState = iota
	STATE_RUNNING
	STATE_SUCCESSFUL
	STATE_FAILED
)

type CmdStatus struct {
	state TaskState

	Output cmdrunr.SafeBuffer
}

type TaskStatus struct {
	state TaskState

	BeforeHooksSuccess bool
	AfterHooksSuccess  bool
	MainTaskStatus     bool

	Output cmdrunr.SafeBuffer
}

type StepStatus struct {
	BeforeHooks *CmdStatus
	Tasks       []TaskStatus
	AfterHooks  *CmdStatus

	state TaskState
	Mtx   sync.Mutex
}

func executeJob(ctx context.Context, basePath string, config *cfg.JobConfig, stepStatuses []StepStatus, readyToDisplay, done chan struct{}) {
	// First, initialize the status structs
	for stepIdx, step := range config.Steps {
		if len(step.RunBefore) > 0 {
			stepStatuses[stepIdx].BeforeHooks = &CmdStatus{}
		}
		if len(step.RunAfter) > 0 {
			stepStatuses[stepIdx].AfterHooks = &CmdStatus{}
		}
		stepStatuses[stepIdx].Tasks = make([]TaskStatus, len(step.Tasks))
	}

	// Notify that the statuses are ready to be displayed
	close(readyToDisplay)
	defer close(done)

	for stepIdx, step := range config.Steps {

		stepStatuses[stepIdx].Mtx.Lock()
		stepStatuses[stepIdx].state = STATE_RUNNING
		if stepStatuses[stepIdx].BeforeHooks != nil {
			stepStatuses[stepIdx].BeforeHooks.state = STATE_RUNNING
		}
		stepStatuses[stepIdx].Mtx.Unlock()

		for _, hook := range step.RunBefore {
			if !cmdrunr.RunCommand(ctx, basePath, hook.Path, hook.Cmd, &stepStatuses[stepIdx].BeforeHooks.Output) {
				stepStatuses[stepIdx].Mtx.Lock()
				stepStatuses[stepIdx].BeforeHooks.state = STATE_FAILED
				stepStatuses[stepIdx].state = STATE_FAILED
				stepStatuses[stepIdx].Mtx.Unlock()
				return
			}
		}

		stepStatuses[stepIdx].Mtx.Lock()
		if stepStatuses[stepIdx].BeforeHooks != nil {
			stepStatuses[stepIdx].BeforeHooks.state = STATE_SUCCESSFUL
		}
		stepStatuses[stepIdx].Mtx.Unlock()

		// Within each step, the tasks are run asynchronously
		var taskWg sync.WaitGroup
		for taskIdx, task := range step.Tasks {
			taskWg.Go(func() {
				runTask(ctx, basePath, task, &stepStatuses[stepIdx].Tasks[taskIdx], &stepStatuses[stepIdx])
			})
		}
		taskWg.Wait()

		// Read the status from the tasks
		stepStatuses[stepIdx].Mtx.Lock()
		tasksOk := true
		for taskIdx := range stepStatuses[stepIdx].Tasks {
			js := &stepStatuses[stepIdx].Tasks[taskIdx]
			if js.state == STATE_FAILED {
				tasksOk = false
			}
		}
		if !tasksOk {
			stepStatuses[stepIdx].state = STATE_FAILED
			stepStatuses[stepIdx].Mtx.Unlock()
			return
		}
		if stepStatuses[stepIdx].AfterHooks != nil {
			stepStatuses[stepIdx].AfterHooks.state = STATE_RUNNING
		}
		stepStatuses[stepIdx].Mtx.Unlock()

		for _, hook := range step.RunAfter {
			if !cmdrunr.RunCommand(ctx, basePath, hook.Path, hook.Cmd, &stepStatuses[stepIdx].AfterHooks.Output) {
				stepStatuses[stepIdx].Mtx.Lock()
				stepStatuses[stepIdx].AfterHooks.state = STATE_FAILED
				stepStatuses[stepIdx].state = STATE_FAILED
				stepStatuses[stepIdx].Mtx.Unlock()
				return
			}
		}

		stepStatuses[stepIdx].Mtx.Lock()
		if stepStatuses[stepIdx].AfterHooks != nil {
			stepStatuses[stepIdx].AfterHooks.state = STATE_SUCCESSFUL
		}
		stepStatuses[stepIdx].state = STATE_SUCCESSFUL
		stepStatuses[stepIdx].Mtx.Unlock()
	}
}

func runTask(ctx context.Context, basePath string, config cfg.TaskConfig, taskStatus *TaskStatus, globalStatus *StepStatus) bool {
	globalStatus.Mtx.Lock()
	taskStatus.BeforeHooksSuccess = false
	taskStatus.MainTaskStatus = false
	taskStatus.AfterHooksSuccess = false
	taskStatus.state = STATE_RUNNING
	globalStatus.Mtx.Unlock()

	for _, hook := range config.RunBefore {
		if !cmdrunr.RunCommand(ctx, basePath, hook.Path, hook.Cmd, &taskStatus.Output) {
			globalStatus.Mtx.Lock()
			taskStatus.state = STATE_FAILED
			globalStatus.Mtx.Unlock()
			return false
		}
	}

	globalStatus.Mtx.Lock()
	taskStatus.BeforeHooksSuccess = true
	globalStatus.Mtx.Unlock()

	// run config.Cmd
	if !cmdrunr.RunCommand(ctx, basePath, config.Path, config.Cmd, &taskStatus.Output) {
		globalStatus.Mtx.Lock()
		taskStatus.state = STATE_FAILED
		globalStatus.Mtx.Unlock()
		return false
	}

	globalStatus.Mtx.Lock()
	taskStatus.MainTaskStatus = true
	globalStatus.Mtx.Unlock()

	for _, hook := range config.RunAfter {
		if !cmdrunr.RunCommand(ctx, basePath, hook.Path, hook.Cmd, &taskStatus.Output) {
			globalStatus.Mtx.Lock()
			taskStatus.state = STATE_FAILED
			globalStatus.Mtx.Unlock()
			return false
		}
	}

	globalStatus.Mtx.Lock()
	taskStatus.state = STATE_SUCCESSFUL
	taskStatus.AfterHooksSuccess = true
	globalStatus.Mtx.Unlock()

	return true
}
