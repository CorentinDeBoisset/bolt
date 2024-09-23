package pkg

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
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

type TaskStatus struct {
	Running  bool
	ExitCode int

	Output SafeBuffer
}

type JobStatus struct {
	BeforeHooks []TaskStatus
	MainJob     TaskStatus
	AfterHooks  []TaskStatus

	Running bool
	Success bool
}

type StepStatus struct {
	BeforeHooks []TaskStatus
	Jobs        []JobStatus
	AfterHooks  []TaskStatus

	Running bool
	Success bool
	Mtx     sync.Mutex
}

func executeCi(config *CiConfig, readyToDisplay, done chan struct{}) {
	stepStatuses := make([]StepStatus, len(config.Steps))

	// First, initialize the status structs
	for stepIdx, step := range config.Steps {
		if len(step.RunBefore) > 0 {
			stepStatuses[stepIdx].BeforeHooks = make([]TaskStatus, len(step.RunBefore))
		}
		if len(step.RunAfter) > 0 {
			stepStatuses[stepIdx].AfterHooks = make([]TaskStatus, len(step.RunAfter))
		}
		stepStatuses[stepIdx].Jobs = make([]JobStatus, len(step.Jobs))

		// Within each step, the jobs are run asynchronously
		for jobIdx, job := range step.Jobs {
			if len(job.RunBefore) > 0 {
				stepStatuses[stepIdx].Jobs[jobIdx].BeforeHooks = make([]TaskStatus, len(job.RunBefore))
			}
			if len(job.RunAfter) > 0 {
				stepStatuses[stepIdx].Jobs[jobIdx].AfterHooks = make([]TaskStatus, len(job.RunAfter))
			}
		}
	}

	// Notify that the statuses are ready to be displayed
	close(readyToDisplay)
	defer close(done)

	for stepIdx, step := range config.Steps {
		for hookIdx, hook := range step.RunBefore {
			if err := runTask(hook.Cmd, hook.Path, config.basePath, &stepStatuses[stepIdx].BeforeHooks[hookIdx], &stepStatuses[stepIdx]); err != nil {
				fmt.Println("an error occured with a run_before hook")
				return
			}
		}

		// Within each step, the jobs are run asynchronously
		var jobsWg sync.WaitGroup
		for jobIdx, job := range step.Jobs {
			jobsWg.Add(1)
			go func() {
				defer jobsWg.Done()
				_ = runJob(job, config, &stepStatuses[stepIdx].Jobs[jobIdx], &stepStatuses[stepIdx])
			}()
		}
		jobsWg.Wait()

		// Read the status from the jobs
		stepStatuses[stepIdx].Mtx.Lock()
		stepStatuses[stepIdx].Running = false
		jobsOk := true
		for jobIdx := range stepStatuses[stepIdx].Jobs {
			if !stepStatuses[stepIdx].Jobs[jobIdx].Success {
				jobsOk = false
			}
		}
		stepStatuses[stepIdx].Success = jobsOk
		stepStatuses[stepIdx].Mtx.Unlock()

		if !jobsOk {
			fmt.Println("an error occured with the jobs")
			return
		}

		for hookIdx, hook := range step.RunAfter {
			if err := runTask(hook.Cmd, hook.Path, config.basePath, &stepStatuses[stepIdx].AfterHooks[hookIdx], &stepStatuses[stepIdx]); err != nil {
				fmt.Println("an error occured with a run_after hook")
				return
			}
		}
	}
}

func getCmdPath(basePath, cmdPath string) string {
	if filepath.IsAbs(cmdPath) {
		return cmdPath
	} else {
		return filepath.Join(basePath, cmdPath)
	}
}

func runTask(cmd, path, basePath string, status *TaskStatus, globalStatus *StepStatus) error {
	task := exec.Command("/bin/sh", "-c", cmd)
	task.Dir = getCmdPath(basePath, path)
	task.Stdout = &status.Output
	task.Stderr = &status.Output

	defer func() {
		fmt.Printf("Task \"%s\" in %s (exit code %d) : %s\n", cmd, task.Dir, task.ProcessState.ExitCode(), status.Output.String())
	}()

	if err := task.Run(); err != nil {
		globalStatus.Mtx.Lock()
		status.Running = false
		status.ExitCode = task.ProcessState.ExitCode()
		globalStatus.Mtx.Unlock()
		return err
	}

	globalStatus.Mtx.Lock()
	status.Running = false
	status.ExitCode = task.ProcessState.ExitCode()
	globalStatus.Mtx.Unlock()

	return nil
}

func runJob(config JobConfig, globalConfig *CiConfig, status *JobStatus, globalStatus *StepStatus) error {
	for hookIdx, hook := range config.RunBefore {
		if err := runTask(hook.Cmd, hook.Path, globalConfig.basePath, &status.BeforeHooks[hookIdx], globalStatus); err != nil {
			globalStatus.Mtx.Lock()
			status.Success = false
			status.Running = false
			globalStatus.Mtx.Unlock()
			return err
		}
	}

	// run config.Cmd
	if err := runTask(config.Cmd, config.Path, globalConfig.basePath, &status.MainJob, globalStatus); err != nil {
		globalStatus.Mtx.Lock()
		status.Success = false
		status.Running = false
		globalStatus.Mtx.Unlock()
		return err
	}

	for hookIdx, hook := range config.RunAfter {
		if err := runTask(hook.Cmd, hook.Path, globalConfig.basePath, &status.AfterHooks[hookIdx], globalStatus); err != nil {
			globalStatus.Mtx.Lock()
			status.Success = false
			status.Running = false
			globalStatus.Mtx.Unlock()
			return err
		}
	}

	globalStatus.Mtx.Lock()
	status.Success = true
	status.Running = false
	globalStatus.Mtx.Unlock()

	return nil
}
