package jobexec

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/corentindeboisset/tera/pkg/cfg"
	"github.com/corentindeboisset/tera/pkg/iface"
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

	ctx, stopCtx := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// Run the job in a goroutine. The synchronisation is handled by the channels
	go executeJob(ctx, config.BasePath, pickedJob, stepStatuses, readyToDisplay, jobDone)
	<-readyToDisplay

	theme := iface.LoadTheme()
	_, err = tea.NewProgram(newModel(pickedJob, stepStatuses, theme), tea.WithContext(ctx), tea.WithoutSignalHandler()).Run()
	if errors.Is(err, tea.ErrProgramKilled) {
		err = nil
	}

	// Ensure the cleanup is called, and wait for all tasks to be done
	stopCtx()
	<-jobDone

	return err
}
