package jobexec

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/corentindeboisset/tera/pkg/cfg"
	"github.com/corentindeboisset/tera/pkg/cmdrunr"
	"github.com/corentindeboisset/tera/pkg/iface"
	"github.com/corentindeboisset/tera/pkg/outputviewer"
)

var (
	selectedTaskStyle           = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111")).Inline(true)
	focusedTaskStyle            = lipgloss.NewStyle().Underline(true).Inline(true)
	focusedAndSelectedTaskStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111")).Underline(true).Inline(true)

	successFlag  = lipgloss.NewStyle().SetString("✓").Bold(true).Foreground(lipgloss.Color("082"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	failureFlag  = lipgloss.NewStyle().SetString("✗").Bold(true).Foreground(lipgloss.Color("196"))
)

type RefreshStatusMsg time.Time

type keymap = struct {
	up, down, tab, enter, quit key.Binding
}

type registeredTask struct {
	Name   string
	Output *cmdrunr.SafeBuffer
}

type ifaceModel struct {
	width  int
	height int
	theme  iface.Theme

	keymap keymap
	help   help.Model

	jobConfig *cfg.JobConfig
	statuses  []StepStatus
	taskIds   []registeredTask

	stepPanelWidth  int
	focusOutput     bool
	hideOutputPanel bool
	selectedTask    int
	focusedTask     int
	spinner         spinner.Model
	stepPanel       ListViewportModel
	outputPanel     outputviewer.Model
}

func tickReadOutputsMsg() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return RefreshStatusMsg(t)
	})
}

func newModel(config *cfg.JobConfig, statuses []StepStatus, theme iface.Theme) ifaceModel {
	m := ifaceModel{
		help:  help.New(),
		theme: theme,
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
			enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("↵/Enter", "Select a task to display"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("Ctrl+C ", "Exit"),
			),
		},
		focusOutput:     false,
		hideOutputPanel: false,
		jobConfig:       config,
		stepPanelWidth:  15,
		statuses:        statuses,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(spinnerStyle),
		),
		stepPanel:   NewListViewportModel(15, 10),
		outputPanel: outputviewer.New(30, 10, theme, nil),
	}

	m.updateKeyBindings()
	m.calculateMinPanelSize()
	m.stepPanel.Style = lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(theme.FocusedOutputBorderColor)
	m.outputPanel.SetFocus(false)

	m.initializeTaskOutputs()

	return m
}

func (m *ifaceModel) initializeTaskOutputs() {
	m.taskIds = make([]registeredTask, 0)
	for stepIdx := range m.statuses {
		m.statuses[stepIdx].Mtx.Lock()
		if m.statuses[stepIdx].BeforeHooks != nil {
			id := fmt.Sprintf("step#%d__bh", stepIdx)
			m.taskIds = append(m.taskIds, registeredTask{Name: id, Output: &m.statuses[stepIdx].BeforeHooks.Output})
		}
		for taskIdx := range m.statuses[stepIdx].Tasks {
			id := fmt.Sprintf("step#%d__task#%d", stepIdx, taskIdx)
			m.taskIds = append(m.taskIds, registeredTask{Name: id, Output: &m.statuses[stepIdx].Tasks[taskIdx].Output})
		}
		if m.statuses[stepIdx].AfterHooks != nil {
			id := fmt.Sprintf("step#%d__ah", stepIdx)
			m.taskIds = append(m.taskIds, registeredTask{Name: id, Output: &m.statuses[stepIdx].AfterHooks.Output})
		}
		m.statuses[stepIdx].Mtx.Unlock()
	}

	// If the first task is a before_run hook, select the next task
	if len(m.taskIds) > 2 && strings.HasSuffix(m.taskIds[0].Name, "__bh") {
		m.selectedTask += 1
	}
	m.focusedTask = m.selectedTask
}

func (m *ifaceModel) calculateMinPanelSize() {
	maxLen := 10
	for _, step := range m.jobConfig.Steps {
		if maxLen < utf8.RuneCountInString(step.Name) {
			maxLen = utf8.RuneCountInString(step.Name)
		}
		for _, task := range step.Tasks {
			if maxLen < utf8.RuneCountInString(task.Name)+2 {
				maxLen = utf8.RuneCountInString(task.Name) + 2
			}
		}
		if len(step.RunBefore) > 0 && maxLen < 19 {
			maxLen = 19
		}
		if len(step.RunAfter) > 0 && maxLen < 18 {
			maxLen = 18
		}
	}

	m.stepPanelWidth = maxLen + 10
}

func (m *ifaceModel) updateSizes() {
	// The height is fixed
	panelsHeight := m.height - 5

	stepPanelWidth := m.stepPanelWidth
	if stepPanelWidth > (m.width/2 - 6) {
		stepPanelWidth = m.width/2 - 6
	}

	outputWidth := m.width - stepPanelWidth
	if outputWidth < 10 {
		m.hideOutputPanel = true
		m.stepPanel.Resize(m.width, panelsHeight)
		return
	}

	m.outputPanel.Resize(outputWidth, panelsHeight)
	m.stepPanel.Resize(stepPanelWidth, panelsHeight)
}

func (m *ifaceModel) updateKeyBindings() {
	m.keymap.tab.SetEnabled(!m.hideOutputPanel)
	m.keymap.enter.SetEnabled(!m.focusOutput && !m.hideOutputPanel)
}

func (m ifaceModel) Init() tea.Cmd {
	return tea.Batch(
		tickReadOutputsMsg(),
		m.spinner.Tick,
	)
}

func (m *ifaceModel) formatTask(id string, state TaskState, name string) string {
	if id == m.taskIds[m.focusedTask].Name && id == m.taskIds[m.selectedTask].Name {
		name = focusedAndSelectedTaskStyle.Render(name)
	} else if id == m.taskIds[m.selectedTask].Name {
		name = selectedTaskStyle.Render(name)
	} else if id == m.taskIds[m.focusedTask].Name {
		name = focusedTaskStyle.Render(name)
	}

	switch state {
	case STATE_NOT_STARTED:
		return fmt.Sprintf("   %s", name)
	case STATE_RUNNING:
		return fmt.Sprintf("%s %s", m.spinner.View(), name)
	case STATE_SUCCESSFUL:
		return fmt.Sprintf("%s  %s", successFlag, name)
	case STATE_FAILED:
		return fmt.Sprintf("%s  %s", failureFlag, name)
	}

	return ""
}

func (m *ifaceModel) calculateStepPanelContent() []ListViewportLine {
	viewportLines := make([]ListViewportLine, 0, len(m.statuses)*2+len(m.taskIds))
	for stepIdx, stepConfig := range m.jobConfig.Steps {
		m.statuses[stepIdx].Mtx.Lock()
		viewportLines = append(viewportLines, ListViewportLine{
			Padding: 0,
			Content: m.formatTask("", m.statuses[stepIdx].state, stepConfig.Name),
		})
		if m.statuses[stepIdx].BeforeHooks != nil {
			id := fmt.Sprintf("step#%d__bh", stepIdx)
			viewportLines = append(viewportLines, ListViewportLine{
				Padding: 2,
				Content: m.formatTask(id, m.statuses[stepIdx].BeforeHooks.state, "Run-Before hooks"),
			})
		}
		for taskIdx, taskConfig := range stepConfig.Tasks {
			id := fmt.Sprintf("step#%d__task#%d", stepIdx, taskIdx)
			viewportLines = append(viewportLines, ListViewportLine{
				Padding: 2,
				Content: m.formatTask(id, m.statuses[stepIdx].Tasks[taskIdx].state, taskConfig.Name),
			})
		}
		if m.statuses[stepIdx].AfterHooks != nil {
			id := fmt.Sprintf("step#%d__ah", stepIdx)
			viewportLines = append(viewportLines, ListViewportLine{
				Padding: 2,
				Content: m.formatTask(id, m.statuses[stepIdx].AfterHooks.state, "Run-After hooks"),
			})
		}
		viewportLines = append(viewportLines, ListViewportLine{Padding: 0, Content: ""})
		m.statuses[stepIdx].Mtx.Unlock()
	}

	return viewportLines
}

func (m *ifaceModel) calculateTaskLine(taskIdx int) int {
	viewportOffset := 0
	taskOffset := 0
	for stepIdx, stepConfig := range m.jobConfig.Steps {
		m.statuses[stepIdx].Mtx.Lock()
		viewportOffset += 1
		if m.statuses[stepIdx].BeforeHooks != nil {
			if taskOffset == taskIdx {
				m.statuses[stepIdx].Mtx.Unlock()
				return viewportOffset
			}
			viewportOffset += 1
			taskOffset += 1
		}
		for range stepConfig.Tasks {
			if taskOffset == taskIdx {
				m.statuses[stepIdx].Mtx.Unlock()
				return viewportOffset
			}
			viewportOffset += 1
			taskOffset += 1
		}
		if m.statuses[stepIdx].AfterHooks != nil {
			if taskOffset == taskIdx {
				m.statuses[stepIdx].Mtx.Unlock()
				return viewportOffset
			}
			viewportOffset += 1
			taskOffset += 1
		}
		viewportOffset += 1
		m.statuses[stepIdx].Mtx.Unlock()
	}

	return viewportOffset
}

func (m ifaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Global (independant of the panel with focus)
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "tab":
			if !m.hideOutputPanel {
				m.focusOutput = !m.focusOutput
				m.outputPanel.SetFocus(m.focusOutput)
				m.updateKeyBindings()
				if m.focusOutput {
					m.stepPanel.Style = m.stepPanel.Style.BorderForeground(m.theme.BlurredOutputBorderColor)
				} else {
					m.stepPanel.Style = m.stepPanel.Style.BorderForeground(m.theme.FocusedOutputBorderColor)
				}
			}
		}

		if m.focusOutput {
			return m, m.outputPanel.Update(msg)
		}

		switch msg.String() {
		case "up", "k":
			if m.focusedTask == 0 {
				m.focusedTask = len(m.taskIds) - 1
				m.stepPanel.GoToBottom()
			} else {
				m.focusedTask -= 1
				m.stepPanel.Focus(m.calculateTaskLine(m.focusedTask))
			}

		case "down", "j":
			if m.focusedTask == len(m.taskIds)-1 {
				m.focusedTask = 0
				m.stepPanel.GoToTop()
			} else {
				m.focusedTask += 1
				m.stepPanel.Focus(m.calculateTaskLine(m.focusedTask))
			}

		// Other movement keys, not displayed in the help
		case "pgup":
			m.stepPanel.PageUp()

		case "pgdown":
			m.stepPanel.PageDown()

		case "home":
			m.focusedTask = 0
			m.stepPanel.GoToTop()

		case "end":
			m.focusedTask = len(m.taskIds) - 1
			m.stepPanel.GoToBottom()

		case "enter":
			m.selectedTask = m.focusedTask
			m.outputPanel.SetBuffer(m.taskIds[m.selectedTask].Output)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

		// Clear the screen to avoid artifacts
		return m, tea.ClearScreen

	case RefreshStatusMsg:
		if !m.hideOutputPanel {
			m.outputPanel.RefreshContent()
		}
		return m, tickReadOutputsMsg()

	case tea.PasteMsg:
		if m.focusOutput {
			return m, m.outputPanel.Update(msg)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		// Rebuild the side panel content with updated spinners
		m.spinner, cmd = m.spinner.Update(msg)
		m.stepPanel.SetContent(m.calculateStepPanelContent())

		return m, cmd

	case tea.MouseWheelMsg:
		if m.focusOutput {
			return m, m.outputPanel.Update(msg)
		}

		switch msg.Button {
		case tea.MouseWheelUp:
			m.focusedTask = max(m.focusedTask-1, 0)
			m.stepPanel.Focus(m.calculateTaskLine(m.focusedTask))
		case tea.MouseWheelDown:
			m.focusedTask = min(m.focusedTask+1, len(m.taskIds)-1)
			m.stepPanel.Focus(m.calculateTaskLine(m.focusedTask))
		}
	}

	return m, nil
}

func (m ifaceModel) View() tea.View {
	help := m.help.FullHelpView([][]key.Binding{
		{m.keymap.up, m.keymap.down},
		{m.keymap.tab, m.keymap.enter, m.keymap.quit},
	})

	var views []string
	views = append(views, m.stepPanel.View())
	if !m.hideOutputPanel {
		views = append(views, m.outputPanel.View())
	}

	view := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help)
	view.AltScreen = true
	view.KeyboardEnhancements.ReportEventTypes = true
	view.MouseMode = tea.MouseModeCellMotion

	return view
}
