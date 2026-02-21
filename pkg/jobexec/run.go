package jobexec

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/corentindeboisset/bolt/pkg"
)

func GetJobList(confPath string) []string {
	config, err := pkg.FindAndParseConfig(confPath)
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
	config, err := pkg.FindAndParseConfig(confPath)
	if err != nil {
		return err
	}

	pkg.SetupLogs(config.LogFilePath)

	var pickedJob *pkg.JobConfig
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
	if _, err := tea.NewProgram(newModel(pickedJob, stepStatuses), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		// TODO: Handle error
		cancelJob()
		<-jobDone
		return err
	}
	cancelJob()
	<-jobDone

	return nil
}
