package jobexec

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/corentindeboisset/bolt/pkg/cfg"
)

func GetJobList(confPath string) []string {
	config, err := cfg.FindAndParseConfig(confPath)
	if err != nil {
		return nil
	}

	results := make([]string, 0, len(config.Jobs))
	for _, job := range config.Jobs {
		results = append(results, job.Name)
	}

	return results
}

func ExecuteJob(confPath string, jobToRun string) error {
	config, err := cfg.FindAndParseConfig(confPath)
	if err != nil {
		return err
	}

	cfg.SetupLogs(config.LogFilePath)

	var pickedJob *cfg.JobConfig
	for _, job := range config.Jobs {
		if (len(jobToRun) > 0 && job.Name == jobToRun) || job.Name == "default" {
			pickedJob = &job
		}
	}

	if pickedJob == nil {
		if len(jobToRun) > 0 {
			return fmt.Errorf("there is no job named \"%s\"", jobToRun)
		}
		return fmt.Errorf("there is no job named \"default\"")
	}

	stepStatuses := make([]StepStatus, len(pickedJob.Steps))

	readyToDisplay := make(chan struct{})
	jobDone := make(chan struct{})

	ctx, cancelJob := context.WithCancel(context.Background())

	// Run the job in a goroutine. The synchronisation is handled by the channels
	go executeJob(ctx, config.BasePath, pickedJob, stepStatuses, readyToDisplay, jobDone)
	<-readyToDisplay
	program := tea.NewProgram(newModel(pickedJob, stepStatuses), tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithoutSignalHandler())

	programErr := make(chan error)
	go func() {
		_, err := program.Run()
		if err != tea.ErrProgramKilled {
			programErr <- err
		}
		close(programErr)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(sigChan)

	select {
	case <-sigChan:
		program.Quit()
	case err = <-programErr:
		// nothing to do, the program already quit
	}

	// Run the cleanup, and wait for all tasks to be done
	cancelJob()
	<-jobDone

	// Ensure the program is finished before proceeding
	for err = range programErr {
	}

	return err
}
