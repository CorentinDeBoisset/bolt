package pkg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

type SafeBuffer struct {
	buf bytes.Buffer
	mtx sync.Mutex
}

func (s *SafeBuffer) Write(p []byte) (n int, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.buf.Write(p)
}

func (s *SafeBuffer) String() string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.buf.String()
}

type TaskState int

const (
	STATE_NOT_STARTED TaskState = iota
	STATE_RUNNING
	STATE_SUCCESSFUL
	STATE_FAILED
)

type CmdStatus struct {
	state TaskState

	Output SafeBuffer
}

type TaskStatus struct {
	state TaskState

	BeforeHooksSuccess bool
	AfterHooksSuccess  bool
	MainTaskStatus     bool

	Output SafeBuffer
}

type StepStatus struct {
	BeforeHooks *CmdStatus
	Tasks       []TaskStatus
	AfterHooks  *CmdStatus

	state TaskState
	Mtx   sync.Mutex
}

func executeCi(ctx context.Context, basePath string, config *JobConfig, stepStatuses []StepStatus, readyToDisplay, done chan struct{}) {
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
			if !runCommand(ctx, basePath, hook.Path, hook.Cmd, &stepStatuses[stepIdx].BeforeHooks.Output) {
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
			taskWg.Add(1)
			go func() {
				defer taskWg.Done()
				runTask(ctx, basePath, task, &stepStatuses[stepIdx].Tasks[taskIdx], &stepStatuses[stepIdx])
			}()
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
			if !runCommand(ctx, basePath, hook.Path, hook.Cmd, &stepStatuses[stepIdx].AfterHooks.Output) {
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

func getCmdPath(basePath, cmdPath string) string {
	if filepath.IsAbs(cmdPath) {
		return cmdPath
	} else {
		return filepath.Join(basePath, cmdPath)
	}
}

func runCommand(ctx context.Context, basePath, path, cmd string, output *SafeBuffer) bool {
	// Note: this setup only works on Unix
	// TODO: maybe add support for windows (or not... :shrug:)
	_, _ = fmt.Fprintf(output, "> %s\n", cmd)
	task := exec.CommandContext(ctx, "/bin/sh", "-c", cmd)
	task.Dir = getCmdPath(basePath, path)
	task.Env = os.Environ() // Pass the environment to the child processes
	task.Stdout = output
	task.Stderr = output
	task.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	task.Cancel = func() error {
		// Kill the whole process group, killing the subprocesses as well
		return syscall.Kill(-task.Process.Pid, syscall.SIGKILL)
	}

	if err := task.Start(); err != nil {
		_, _ = fmt.Fprintf(output, "\n\nThe command could not start due to the following error:\n%s", err.Error())
		return false
	}

	if err := task.Wait(); err != nil {
		_, _ = fmt.Fprintf(output, "\n\nThe command failed with the following error:\n%s", err.Error())
		return false
	}

	_, _ = fmt.Fprintf(output, "\n\n")

	return true
}

func runTask(ctx context.Context, basePath string, config TaskConfig, taskStatus *TaskStatus, globalStatus *StepStatus) bool {
	globalStatus.Mtx.Lock()
	taskStatus.BeforeHooksSuccess = false
	taskStatus.MainTaskStatus = false
	taskStatus.AfterHooksSuccess = false
	taskStatus.state = STATE_RUNNING
	globalStatus.Mtx.Unlock()

	for _, hook := range config.RunBefore {
		if !runCommand(ctx, basePath, hook.Path, hook.Cmd, &taskStatus.Output) {
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
	if !runCommand(ctx, basePath, config.Path, config.Cmd, &taskStatus.Output) {
		globalStatus.Mtx.Lock()
		taskStatus.state = STATE_FAILED
		globalStatus.Mtx.Unlock()
		return false
	}

	globalStatus.Mtx.Lock()
	taskStatus.MainTaskStatus = true
	globalStatus.Mtx.Unlock()

	for _, hook := range config.RunAfter {
		if !runCommand(ctx, basePath, hook.Path, hook.Cmd, &taskStatus.Output) {
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
