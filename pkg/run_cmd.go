package pkg

import (
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

func RunCmd(confPath string, selectedStep string) error {
	config, err := findAndParseConfig(confPath)
	if err != nil {
		return err
	}

	setupLogs(config.LogFilePath)

	stepStatuses := make([]StepStatus, len(config.Steps))

	readyToDisplay := make(chan struct{})
	ciDone := make(chan struct{})

	// TODO: add a cancellable context, to stop the commmands

	// Run the ci in a goroutine. The synchronisation is handled by the channels
	go executeCi(config, stepStatuses, readyToDisplay, ciDone)
	<-readyToDisplay
	if _, err := tea.NewProgram(newModel(config, stepStatuses), tea.WithAltScreen()).Run(); err != nil {
		// TODO: Handle error
		<-ciDone
		return err
	}
	<-ciDone

	return nil
}
