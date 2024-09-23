package pkg

import (
	"fmt"
)

func RunCmd(confPath string, selectedStep string) error {
	config, err := findAndParseConfig(confPath)
	if err != nil {
		return err
	}

	readyToDisplay := make(chan struct{})
	ciDone := make(chan struct{})

	// TODO: add a cancellable context, to stop the commmands

	// Run the ci in a goroutine. The synchronisation is handled by the channels
	go executeCi(config, readyToDisplay, ciDone)
	<-readyToDisplay
	fmt.Println("ready to display")
	<-ciDone
	fmt.Println("done running the CI")

	return nil
}
