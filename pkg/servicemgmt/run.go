package servicemgmt

import (
	"os"
	"os/signal"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/corentindeboisset/tera/pkg/cfg"
	"github.com/corentindeboisset/tera/pkg/iface"
)

func StartServiceManagement(confPath string) error {
	config, err := cfg.FindAndParseConfig(confPath)
	if err != nil {
		return err
	}

	cfg.SetupLogs(config.LogFilePath)

	orchestrator, err := NewOrchestrator(config.BasePath, config.Services)
	if err != nil {
		return err
	}

	theme := iface.LoadTheme()

	program := tea.NewProgram(newModel(orchestrator, theme), tea.WithoutSignalHandler())

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

	// Shutdown all services
	orchestrator.Shutdown(nil)

	// Ensure the program is finished before proceeding
	for err = range programErr {
	}

	return err
}
