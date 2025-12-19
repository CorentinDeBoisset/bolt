package servicemgmt

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("111")).
				Padding(0, 2)

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("7")).
				Padding(0, 2)
)

type refreshStatusMsg time.Time

type keymap = struct {
	up, down, tab, enter, quit key.Binding
}

type ifaceModel struct {
	width                 int
	height                int
	keymap                keymap
	help                  help.Model
	serviceListPanelWidth int
	hideOutputPanel       bool

	focusOutput   bool
	focusedTask   int
	outputPanel   viewport.Model
	serviceBricks []ServiceBrickModel

	// statuses []ServiceStatus
}

func tickReadOutputsMsg() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return refreshStatusMsg(t)
	})
}

func (m ifaceModel) Init() tea.Cmd {
	return tickReadOutputsMsg()
}

func (m ifaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.focusOutput {
				m.outputPanel.ScrollUp(3)
			} else {
				if (m.focusedTask) == 0 {
					m.focusedTask = len(m.serviceBricks) - 1
				}
			}
		}
	case tea.WindowSizeMsg:
		// pass
	default:
		return m, tickReadOutputsMsg()
	}

	return m, nil
}

func (m ifaceModel) View() string {
	help := m.help.FullHelpView([][]key.Binding{
		{m.keymap.up, m.keymap.down},
		{m.keymap.tab, m.keymap.enter, m.keymap.quit},
	})

	views := make([]string, 0)
	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help
}
