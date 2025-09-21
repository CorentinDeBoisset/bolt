package pkg

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func setupLogs(path string) {
	if len(path) > 0 {
		logFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err == nil {
			log.SetOutput(logFile)
		}
		// No need to close the logfile, it will be closed as the program terminates
	} else {
		log.SetOutput(io.Discard)
	}
}

func RunAutocomplete(confPath string) []string {
	config, err := findAndParseConfig(confPath)
	if err != nil {
		return nil
	}

	results := make([]string, 0, len(config.Jobs))
	for _, job := range config.Jobs {
		results = append(results, job.Name)
	}

	return results
}

func RunCmd(confPath string, jobToRun string) error {
	config, err := findAndParseConfig(confPath)
	if err != nil {
		return err
	}

	setupLogs(config.LogFilePath)

	var pickedJob *JobConfig
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
	ciDone := make(chan struct{})

	ctx, cancelCtx := context.WithCancel(context.Background())

	// Run the ci in a goroutine. The synchronisation is handled by the channels
	go executeCi(ctx, config.basePath, pickedJob, stepStatuses, readyToDisplay, ciDone)
	<-readyToDisplay
	if _, err := tea.NewProgram(newModel(pickedJob, stepStatuses), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		// TODO: Handle error
		cancelCtx()
		<-ciDone
		return err
	}
	cancelCtx()
	<-ciDone

	return nil
}
