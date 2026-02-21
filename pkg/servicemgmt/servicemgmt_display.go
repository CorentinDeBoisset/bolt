package servicemgmt

import (
	"image/color"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/corentindeboisset/bolt/pkg/iface"
	"github.com/corentindeboisset/bolt/pkg/listviewport"
)

var (
	focusedBorderStyle = iface.BaseSurfaceStyle.
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("111")).
				Padding(0, 2)

	blurredBorderStyle = iface.BaseSurfaceStyle.
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("7")).
				Padding(0, 2)
)

type refreshStatusMsg time.Time

type keymap = struct {
	up   key.Binding
	down key.Binding
	tab  key.Binding
	quit key.Binding

	appOnlyKill     key.Binding
	appOnlyRestart  key.Binding
	standardKill    key.Binding
	standardRestart key.Binding
	standardStart   key.Binding
	open            key.Binding
}

type ifaceModel struct {
	width                 int
	height                int
	keymap                keymap
	help                  help.Model
	serviceListPanelWidth int
	hideOutputPanel       bool
	hideHelp              bool

	focusOutput bool
	focusedTask int
	outputPanel viewport.Model

	serviceBricks    []*ServiceBrickModel
	serviceListPanel listviewport.Model

	orchestrator *Orchestrator
}

func tickReadOutputsMsg() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return refreshStatusMsg(t)
	})
}

func newModel(orchestrator *Orchestrator) ifaceModel {
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

			appOnlyKill: key.NewBinding(
				key.WithKeys("Q"),
				key.WithHelp("Q", "Kill only the service"),
			),
			appOnlyRestart: key.NewBinding(
				key.WithKeys("R"),
				key.WithHelp("R", "Restart the service"),
			),
			standardKill: key.NewBinding(
				key.WithKeys("q"),
				key.WithHelp("q", "Kill the service"),
			),
			standardRestart: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "Restart the service"),
			),
			standardStart: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("↵/Enter", "Start the service"),
			),
			open: key.NewBinding(
				key.WithKeys("o"),
				key.WithHelp("o", "Open the app"),
			),
		},
		focusOutput:           false,
		hideOutputPanel:       false,
		orchestrator:          orchestrator,
		serviceListPanelWidth: BRICK_MIN_WIDTH,
		serviceListPanel:      listviewport.New(30, 10, iface.BaseSurfaceStyle.Padding(1, 2)),
		outputPanel:           viewport.New(30, 10),
	}

	m.outputPanel.Style = blurredBorderStyle

	m.initializeServiceList()

	return m
}

func (m *ifaceModel) refreshLayoutSizes() {
	panelsHeight := m.height - 6

	if m.height <= 10 {
		panelsHeight = m.height
		m.hideHelp = true
	} else {
		m.hideHelp = false
	}

	if m.width > 75 {
		m.hideOutputPanel = false
		m.serviceListPanelWidth = min(max(m.width*40/100, BRICK_MIN_WIDTH), 100)
	} else {
		m.hideOutputPanel = true
		m.serviceListPanelWidth = m.width
	}

	m.serviceListPanel.Resize(m.serviceListPanelWidth, panelsHeight)
	m.outputPanel.Height = panelsHeight
	m.outputPanel.Width = m.width - m.serviceListPanelWidth
}

func (m *ifaceModel) initializeServiceList() {
	serviceList := m.orchestrator.SortedServices()

	m.serviceBricks = make([]*ServiceBrickModel, len(serviceList))
	panelItems := make([]listviewport.ListItem, 0)

	idx := 0
	for _, service := range serviceList {
		brick := NewServiceBrick(service.Id, service.Name, m.width, color.RGBA{0, 0, 0, 0})
		m.serviceBricks[idx] = brick
		panelItems = append(panelItems, brick)
		if idx < len(serviceList)-1 {
			panelItems = append(panelItems, NewSeparator(m.width))
		}

		idx++
	}

	m.serviceListPanel.SetItems(panelItems)

	m.updateFocusedTask(0)
}

func (m ifaceModel) Init() tea.Cmd {
	return tickReadOutputsMsg()
}

func (m *ifaceModel) updateFocusedTask(newTaskId int) {
	m.serviceBricks[m.focusedTask].SetFocusLevel(0)
	m.focusedTask = max(min(newTaskId, len(m.serviceBricks)-1), 0)
	m.serviceBricks[m.focusedTask].SetFocusLevel(1)
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
				if m.focusedTask <= 0 {
					m.updateFocusedTask(len(m.serviceBricks) - 1)
					m.serviceListPanel.GoToBottom()
				} else {
					m.updateFocusedTask(m.focusedTask - 1)
					m.serviceListPanel.ScrollUp(1)
				}
			}

		case "down", "j":
			if m.focusOutput {
				m.outputPanel.ScrollDown(3)
			} else {
				if m.focusedTask >= len(m.serviceBricks)-1 {
					m.updateFocusedTask(0)
					m.serviceListPanel.GoToTop()
				} else {
					m.updateFocusedTask(m.focusedTask + 1)
					m.serviceListPanel.ScrollDown(1)
				}
			}

		case "pgup":
			if m.focusOutput {
				m.outputPanel.PageUp()
			} else {
				itemId := m.serviceListPanel.PageUp()
				m.focusBrickById(itemId)
			}

		case "pgdown":
			if m.focusOutput {
				m.outputPanel.PageDown()
			} else {
				itemId := m.serviceListPanel.PageDown()
				m.focusBrickById(itemId)
			}

		case "home":
			if m.focusOutput {
				m.outputPanel.GotoTop()
			} else {
				m.updateFocusedTask(0)
				m.serviceListPanel.GoToTop()
			}

		case "end":
			if m.focusOutput {
				m.outputPanel.GotoBottom()
			} else {
				m.updateFocusedTask(len(m.serviceBricks) - 1)
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

		case "Q":
			if !m.focusOutput {
				// TODO: appOnlyKill
			}

		case "R":
			if !m.focusOutput {
				// TODO: appOnlyRestart
			}

		case "q":
			if !m.focusOutput {
				// TODO: standardKill
			}

		case "r":
			if !m.focusOutput {
				// TODO: standardRestart
			}

		case "enter":
			if !m.focusOutput {
				// TODO: standardStart
			}

		case "o":
			if !m.focusOutput {
				// TODO: open
			}
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
					m.updateFocusedTask(m.focusedTask - 1)
					m.serviceListPanel.ScrollUp(1)
				}
			case tea.MouseButtonWheelDown:
				if m.focusOutput {
					m.outputPanel.ScrollDown(3)
				} else {
					m.updateFocusedTask(m.focusedTask + 1)
					m.serviceListPanel.ScrollDown(1)
				}
			}
		}
	}

	return m, nil
}

func (m ifaceModel) View() string {
	panelsContent := iface.BaseSurfaceStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.serviceListPanel.View(),
		m.outputPanel.View()),
	)

	if m.hideHelp {
		return panelsContent
	}

	help := m.help.FullHelpView([][]key.Binding{
		{m.keymap.up, m.keymap.down, m.keymap.tab, m.keymap.quit},
		{},
		{m.keymap.standardStart, m.keymap.standardKill, m.keymap.open, m.keymap.standardRestart},
	})

	return panelsContent + "\n\n" + help
}

func (m *ifaceModel) focusBrickById(id string) {
	for brickIdx, brick := range m.serviceBricks {
		if id == brick.Id() {
			m.updateFocusedTask(brickIdx)
			return
		}
	}
}
