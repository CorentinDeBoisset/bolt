package servicemgmt

import (
	"image/color"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/corentindeboisset/bolt/pkg"
	"github.com/corentindeboisset/bolt/pkg/listviewport"
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
	up, down, tab, quit key.Binding
}

type ifaceModel struct {
	width                 int
	height                int
	keymap                keymap
	help                  help.Model
	serviceListPanelWidth int
	hideOutputPanel       bool

	focusOutput bool
	focusedTask int
	outputPanel viewport.Model

	serviceBricks    []*ServiceBrickModel
	serviceListPanel listviewport.Model

	serviceConfigList []*pkg.ServiceConfig
}

func tickReadOutputsMsg() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return refreshStatusMsg(t)
	})
}

func newModel(serviceConfigList []*pkg.ServiceConfig) ifaceModel {
	m := ifaceModel{
		help: help.New(),
		keymap: keymap{
			up: key.NewBinding(
				key.WithKeys("k", "up"),
				key.WithHelp("↑/k", "Move up"),
			),
			down: key.NewBinding(
				key.WithKeys("j", "down"),
				key.WithHelp("↓/j", "Move down"),
			),
			tab: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("⇥/tab  ", "Switch focus"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("Ctrl+C ", "Exit"),
			),
		},
		focusOutput:           false,
		hideOutputPanel:       false,
		serviceConfigList:     serviceConfigList,
		serviceListPanelWidth: BRICK_MIN_WIDTH,
		serviceListPanel:      listviewport.New(30, 10),
		outputPanel:           viewport.New(30, 10),
	}

	m.outputPanel.Style = blurredBorderStyle

	m.initializeServiceList()

	return m
}

func (m *ifaceModel) refreshLayoutSizes() {
	panelsHeight := m.height - 5
	m.outputPanel.Height = panelsHeight

	if m.width > 75 {
		m.hideOutputPanel = false
		m.serviceListPanelWidth = min(max(m.width*40/100, BRICK_MIN_WIDTH), 100)
	} else {
		m.hideOutputPanel = true
		m.serviceListPanelWidth = m.width
	}
	m.serviceListPanel.Resize(m.serviceListPanelWidth, panelsHeight)
}

func (m *ifaceModel) initializeServiceList() {
	m.serviceBricks = make([]*ServiceBrickModel, len(m.serviceConfigList))
	panelItems := make([]listviewport.ListItem, 0)

	for idx, serviceConfig := range m.serviceConfigList {
		brick := NewServiceBrick(serviceConfig.Name, m.width, color.Black)
		m.serviceBricks[idx] = brick
		panelItems = append(panelItems, brick)
		if idx < len(m.serviceConfigList)-1 {
			panelItems = append(panelItems, NewSeparator(m.width))
		}
	}

	m.serviceListPanel.SetItems(panelItems)
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
				if (m.focusedTask) <= 0 {
					m.focusedTask = len(m.serviceBricks) - 1
					m.serviceListPanel.GoToBottom()
				} else {
					m.focusedTask -= 1
					m.serviceListPanel.ScrollUp(1)
				}
			}
		case "down", "j":
			if m.focusOutput {
				m.outputPanel.ScrollDown(3)
			} else {
				if (m.focusedTask) >= len(m.serviceBricks)-1 {
					m.focusedTask = 0
					m.serviceListPanel.GoToTop()
				} else {
					m.focusedTask += 1
					m.serviceListPanel.ScrollDown(1)
				}
			}

		// case "pgup":
		// 	if m.focusOutput {
		// 		m.outputPanel.PageUp()
		// 	} else {
		// 		// FIXME: get the id of the focused service
		// 		m.serviceListPanel.PageUp()
		// 	}

		// case "pgdown":
		// 	if m.focusOutput {
		// 		m.outputPanel.PageDown()
		// 	} else {
		// 		// FIXME: get the id of the focused service
		// 		m.serviceListPanel.PageDown()
		// 	}

		case "home":
			if m.focusOutput {
				m.outputPanel.GotoTop()
			} else {
				m.focusedTask = 0
				m.serviceListPanel.GoToTop()
			}

		case "end":
			if m.focusOutput {
				m.outputPanel.GotoBottom()
			} else {
				m.focusedTask = len(m.serviceBricks) - 1
				m.serviceListPanel.GoToBottom()
			}

		case "tab":
			if !m.hideOutputPanel {
				m.focusOutput = !m.focusOutput
				// TODO: m.updateKeyBindings()
				if m.focusOutput {
					m.outputPanel.Style = focusedBorderStyle
				} else {
					m.outputPanel.Style = blurredBorderStyle
				}
			}

			// TODO: add cases for start/restart/kill/open browser
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.refreshLayoutSizes()

	case refreshStatusMsg:
		return m, tickReadOutputsMsg()

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				if m.focusOutput {
					m.outputPanel.ScrollUp(3)
				} else {
					m.focusedTask = max(m.focusedTask-1, 0)
					m.serviceListPanel.ScrollUp(1)
				}
			case tea.MouseButtonWheelDown:
				if m.focusOutput {
					m.outputPanel.ScrollDown(3)
				} else {
					m.focusedTask = min(m.focusedTask+1, len(m.serviceBricks)-1)
					m.serviceListPanel.ScrollDown(1)
				}
			}
		}
	}

	return m, nil
}

func (m ifaceModel) View() string {
	help := m.help.FullHelpView([][]key.Binding{
		{m.keymap.up, m.keymap.down},
		{m.keymap.tab, m.keymap.quit},
	})

	views := make([]string, 0)
	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help
}
