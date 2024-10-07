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

type TaskStatus struct {
	state TaskState

	Output SafeBuffer
}

type JobStatus struct {
	state TaskState

	BeforeHooksSuccess bool
	AfterHooksSuccess  bool
	MainJobSuccess     bool

	Output SafeBuffer
}

type StepStatus struct {
	BeforeHooks *TaskStatus
	Jobs        []JobStatus
	AfterHooks  *TaskStatus

	state TaskState
	Mtx   sync.Mutex
}

func executeCi(ctx context.Context, config *CiConfig, stepStatuses []StepStatus, readyToDisplay, done chan struct{}) {
	// First, initialize the status structs
	for stepIdx, step := range config.Steps {
		if len(step.RunBefore) > 0 {
			stepStatuses[stepIdx].BeforeHooks = &TaskStatus{}
		}
		if len(step.RunAfter) > 0 {
			stepStatuses[stepIdx].AfterHooks = &TaskStatus{}
		}
		stepStatuses[stepIdx].Jobs = make([]JobStatus, len(step.Jobs))
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
			if !runTask(ctx, hook.Cmd, hook.Path, config.basePath, &stepStatuses[stepIdx].BeforeHooks.Output) {
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

		// Within each step, the jobs are run asynchronously
		var jobsWg sync.WaitGroup
		for jobIdx, job := range step.Jobs {
			jobsWg.Add(1)
			go func() {
				defer jobsWg.Done()
				runJob(ctx, job, config, &stepStatuses[stepIdx].Jobs[jobIdx], &stepStatuses[stepIdx])
			}()
		}
		jobsWg.Wait()

		// Read the status from the jobs
		stepStatuses[stepIdx].Mtx.Lock()
		jobsOk := true
		for jobIdx := range stepStatuses[stepIdx].Jobs {
			js := &stepStatuses[stepIdx].Jobs[jobIdx]
			if js.state == STATE_FAILED {
				jobsOk = false
			}
		}
		if !jobsOk {
			stepStatuses[stepIdx].state = STATE_FAILED
			stepStatuses[stepIdx].Mtx.Unlock()
			return
		}
		if stepStatuses[stepIdx].AfterHooks != nil {
			stepStatuses[stepIdx].AfterHooks.state = STATE_RUNNING
		}
		stepStatuses[stepIdx].Mtx.Unlock()

		for _, hook := range step.RunAfter {
			if !runTask(ctx, hook.Cmd, hook.Path, config.basePath, &stepStatuses[stepIdx].AfterHooks.Output) {
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

func runTask(ctx context.Context, cmd, path, basePath string, output *SafeBuffer) bool {
	// Note: this setup only works on Unix
	// TODO: maybe add support for windows (or not... :shrug:)
	_, _ = output.Write([]byte(fmt.Sprintf("> %s\n", cmd)))
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
		_, _ = output.Write([]byte(fmt.Sprintf("\n\nThe command could not start due to the following error:\n%s", err.Error())))
		return false
	}

	if err := task.Wait(); err != nil {
		_, _ = output.Write([]byte(fmt.Sprintf("\n\nThe command failed with the following error:\n%s", err.Error())))
		return false
	}

	_, _ = output.Write([]byte("\n\n"))

	return true
}

func runJob(ctx context.Context, config JobConfig, globalConfig *CiConfig, jobStatus *JobStatus, globalStatus *StepStatus) bool {
	globalStatus.Mtx.Lock()
	jobStatus.BeforeHooksSuccess = false
	jobStatus.MainJobSuccess = false
	jobStatus.AfterHooksSuccess = false
	jobStatus.state = STATE_RUNNING
	globalStatus.Mtx.Unlock()

	for _, hook := range config.RunBefore {
		if !runTask(ctx, hook.Cmd, hook.Path, globalConfig.basePath, &jobStatus.Output) {
			globalStatus.Mtx.Lock()
			jobStatus.state = STATE_FAILED
			globalStatus.Mtx.Unlock()
			return false
		}
	}

	globalStatus.Mtx.Lock()
	jobStatus.BeforeHooksSuccess = true
	globalStatus.Mtx.Unlock()

	// run config.Cmd
	if !runTask(ctx, config.Cmd, config.Path, globalConfig.basePath, &jobStatus.Output) {
		globalStatus.Mtx.Lock()
		jobStatus.state = STATE_FAILED
		globalStatus.Mtx.Unlock()
		return false
	}

	globalStatus.Mtx.Lock()
	jobStatus.MainJobSuccess = true
	globalStatus.Mtx.Unlock()

	for _, hook := range config.RunAfter {
		if !runTask(ctx, hook.Cmd, hook.Path, globalConfig.basePath, &jobStatus.Output) {
			globalStatus.Mtx.Lock()
			jobStatus.state = STATE_FAILED
			globalStatus.Mtx.Unlock()
			return false
		}
	}

	globalStatus.Mtx.Lock()
	jobStatus.state = STATE_SUCCESSFUL
	jobStatus.AfterHooksSuccess = true
	globalStatus.Mtx.Unlock()

	return true
}
