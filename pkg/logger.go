package pkg

import (
	"io"
	"log"
	"os"
)

func SetupLogs(path string) {
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
