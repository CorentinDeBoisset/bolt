package servicemgmt

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/corentindeboisset/bolt/pkg"
)

func StartServiceManagement(confPath string) error {
	config, err := pkg.FindAndParseConfig(confPath)
	if err != nil {
		return err
	}

	pkg.SetupLogs(config.LogFilePath)

	orchestrator, err := NewOrchestrator(config.Services)
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(newModel(orchestrator), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	return err
}
