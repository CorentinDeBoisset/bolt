package servicemgmt

import (
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/corentindeboisset/tera/pkg/iface"
	"github.com/corentindeboisset/tera/pkg/listviewport"
	"github.com/corentindeboisset/tera/pkg/outputviewer"
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
	theme                 iface.Theme
	width                 int
	height                int
	keymap                keymap
	help                  help.Model
	serviceListPanelWidth int
	hideOutputPanel       bool
	hideHelp              bool

	focusOutput bool
	focusedTask int
	outputPanel outputviewer.Model

	serviceBricks    []*ServiceBrickModel
	serviceListPanel listviewport.Model

	orchestrator *Orchestrator
}

func tickReadOutputsMsg() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return refreshStatusMsg(t)
	})
}

func newModel(orchestrator *Orchestrator, theme iface.Theme) ifaceModel {
	m := ifaceModel{
		theme: theme,
		help:  help.New(),
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
				key.WithHelp("↵/Enter", "Start (and open) the service"),
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
		serviceListPanel:      listviewport.New(30, 10, lipgloss.NewStyle().Padding(1, 2)),
		outputPanel:           outputviewer.New(30, 10, theme, nil),
	}

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
	m.outputPanel.Resize(m.width-m.serviceListPanelWidth, panelsHeight)
}

func (m *ifaceModel) initializeServiceList() {
	serviceList := m.orchestrator.SortedServices()

	m.serviceBricks = make([]*ServiceBrickModel, len(serviceList))
	panelItems := make([]listviewport.ListItem, 0)

	idx := 0
	for _, service := range serviceList {
		brick := NewServiceBrick(service.Id, service, m.theme, m.width)
		m.serviceBricks[idx] = brick
		panelItems = append(panelItems, brick)
		if idx < len(serviceList)-1 {
			panelItems = append(panelItems, NewSeparator(m.width, m.theme))
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
	m.outputPanel.SetBuffer(&(m.serviceBricks[m.focusedTask].service.Output), true)
	m.serviceBricks[m.focusedTask].SetFocusLevel(2)
}

func (m ifaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Global (independent of the panel with focus)
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if !m.hideOutputPanel {
				m.focusOutput = !m.focusOutput
				// TODO: m.updateKeyBindings()
				m.outputPanel.SetFocus(m.focusOutput)
				if m.focusOutput {
					m.serviceBricks[m.focusedTask].SetFocusLevel(1)
				} else {
					m.serviceBricks[m.focusedTask].SetFocusLevel(2)
				}
				return m, nil
			}
		}

		if m.focusOutput {
			return m, m.outputPanel.Update(msg)
		}

		switch msg.String() {
		case "up", "k":
			if m.focusedTask <= 0 {
				m.updateFocusedTask(len(m.serviceBricks) - 1)
				m.serviceListPanel.GoToBottom()
			} else {
				m.updateFocusedTask(m.focusedTask - 1)
				m.serviceListPanel.ScrollUp(1)
			}

		case "down", "j":
			if m.focusedTask >= len(m.serviceBricks)-1 {
				m.updateFocusedTask(0)
				m.serviceListPanel.GoToTop()
			} else {
				m.updateFocusedTask(m.focusedTask + 1)
				m.serviceListPanel.ScrollDown(1)
			}

		case "pgup":
			itemId := m.serviceListPanel.PageUp()
			m.focusBrickById(itemId)

		case "pgdown":
			itemId := m.serviceListPanel.PageDown()
			m.focusBrickById(itemId)

		case "home":
			m.updateFocusedTask(0)
			m.serviceListPanel.GoToTop()

		case "end":
			m.updateFocusedTask(len(m.serviceBricks) - 1)
			m.serviceListPanel.GoToBottom()

		case "Q":
			go m.orchestrator.KillService(m.serviceBricks[m.focusedTask].id, true)

		case "R":
			go m.orchestrator.RestartService(
				m.serviceBricks[m.focusedTask].id,
				true,
				m.outputPanel.InnerFrameWidth(),
				m.outputPanel.InnerFrameHeight(),
			)

		case "q":
			go m.orchestrator.KillService(m.serviceBricks[m.focusedTask].id, false)

		case "r":
			go m.orchestrator.RestartService(
				m.serviceBricks[m.focusedTask].id,
				false,
				m.outputPanel.InnerFrameWidth(),
				m.outputPanel.InnerFrameHeight(),
			)

		case "enter":
			go func() {
				m.orchestrator.StartService(
					m.serviceBricks[m.focusedTask].id,
					m.outputPanel.InnerFrameWidth(),
					m.outputPanel.InnerFrameHeight(),
				)
				m.orchestrator.OpenService(m.serviceBricks[m.focusedTask].id)
			}()

		case "o":
			go m.orchestrator.OpenService(m.serviceBricks[m.focusedTask].id)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.refreshLayoutSizes()

		// Clear the screen to avoid artifacts
		return m, tea.ClearScreen

	case refreshStatusMsg:
		if !m.hideOutputPanel {
			m.outputPanel.RefreshContent()
		}
		return m, tickReadOutputsMsg()

	case tea.PasteMsg:
		if m.focusOutput {
			return m, m.outputPanel.Update(msg)
		}

	case tea.MouseWheelMsg:
		if m.focusOutput {
			return m, m.outputPanel.Update(msg)
		}

		switch msg.Button {
		case tea.MouseWheelUp:
			m.updateFocusedTask(m.focusedTask - 1)
			m.serviceListPanel.ScrollUp(1)
		case tea.MouseWheelDown:
			m.updateFocusedTask(m.focusedTask + 1)
			m.serviceListPanel.ScrollDown(1)
		}
	}

	return m, nil
}

func (m ifaceModel) View() tea.View {
	var view tea.View
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.KeyboardEnhancements.ReportEventTypes = true

	panelsContent := lipgloss.NewStyle().Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.serviceListPanel.View(),
		m.outputPanel.View()),
	)

	if m.hideHelp {
		view.SetContent(panelsContent)
		return view
	}

	help := m.help.FullHelpView([][]key.Binding{
		{m.keymap.up, m.keymap.down, m.keymap.tab, m.keymap.quit},
		{m.keymap.standardStart, m.keymap.standardKill, m.keymap.standardRestart},
	})

	view.SetContent(panelsContent + "\n\n" + help)
	return view
}

func (m *ifaceModel) focusBrickById(id string) {
	for brickIdx, brick := range m.serviceBricks {
		if id == brick.Id() {
			m.updateFocusedTask(brickIdx)
			return
		}
	}
}
