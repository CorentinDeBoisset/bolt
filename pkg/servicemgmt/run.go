package servicemgmt

import (
	"context"
	"errors"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()

	_, err = tea.NewProgram(newModel(orchestrator, theme), tea.WithContext(ctx), tea.WithoutSignalHandler()).Run()
	if errors.Is(err, tea.ErrProgramKilled) {
		err = nil
	}

	// Shutdown all services
	orchestrator.Shutdown()

	return err
}
