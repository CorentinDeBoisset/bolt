package jobexec

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/corentindeboisset/bolt/pkg"
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

	outputStyle = lipgloss.NewStyle().Width(30)

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
	Output *SafeBuffer
}

type ifaceModel struct {
	width           int
	height          int
	keymap          keymap
	help            help.Model
	stepPanelWidth  int
	hideOutputPanel bool

	focusOutput  bool
	selectedTask int
	focusedTask  int
	spinner      spinner.Model
	stepPanel    ListViewportModel
	outputPanel  viewport.Model

	jobConfig *pkg.JobConfig
	statuses  []StepStatus
	taskIds   []registeredTask
}

func tickReadOutputsMsg() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return RefreshStatusMsg(t)
	})
}

func newModel(config *pkg.JobConfig, statuses []StepStatus) ifaceModel {
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
		outputPanel: viewport.New(30, 10),
	}

	m.updateKeyBindings()
	m.calculateMinPanelSize()
	m.stepPanel.Style = focusedBorderStyle
	m.outputPanel.Style = blurredBorderStyle

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
	m.outputPanel.Height = panelsHeight

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
	m.outputPanel.Width = outputWidth
	outputStyle = lipgloss.NewStyle().Width(outputWidth - m.outputPanel.Style.GetBorderLeftSize() - m.outputPanel.Style.GetBorderRightSize())
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			// Scroll up the focused panel
			if m.focusOutput {
				m.outputPanel.ScrollUp(3)
			} else {
				if m.focusedTask == 0 {
					m.focusedTask = len(m.taskIds) - 1
					m.stepPanel.GoToBottom()
				} else {
					m.focusedTask -= 1
					m.stepPanel.Focus(m.calculateTaskLine(m.focusedTask))
				}
			}

		case "down", "j":
			// Scroll down the focused panel
			if m.focusOutput {
				m.outputPanel.ScrollDown(3)
			} else {
				if m.focusedTask == len(m.taskIds)-1 {
					m.focusedTask = 0
					m.stepPanel.GoToTop()
				} else {
					m.focusedTask += 1
					m.stepPanel.Focus(m.calculateTaskLine(m.focusedTask))
				}
			}

		// Other movement keys, not displayed in the help
		case "pgup":
			if m.focusOutput {
				m.outputPanel.PageUp()
			} else {
				m.stepPanel.PageUp()
			}

		case "pgdown":
			if m.focusOutput {
				m.outputPanel.PageDown()
			} else {
				m.stepPanel.PageDown()
			}

		case "home":
			if m.focusOutput {
				m.outputPanel.GotoTop()
			} else {
				m.focusedTask = 0
				m.stepPanel.GoToTop()
			}

		case "end":
			if m.focusOutput {
				m.outputPanel.GotoBottom()
			} else {
				m.focusedTask = len(m.taskIds) - 1
				m.stepPanel.GoToBottom()
			}

		case "tab":
			if !m.hideOutputPanel {
				m.focusOutput = !m.focusOutput
				m.updateKeyBindings()
				if m.focusOutput {
					m.stepPanel.Style = blurredBorderStyle
					m.outputPanel.Style = focusedBorderStyle
				} else {
					m.stepPanel.Style = focusedBorderStyle
					m.outputPanel.Style = blurredBorderStyle
				}
			}

		case "enter":
			if !m.focusOutput {
				m.selectedTask = m.focusedTask
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
	case RefreshStatusMsg:
		if !m.hideOutputPanel {
			isAtBottom := m.outputPanel.AtBottom()
			outputContent := outputStyle.Render(m.taskIds[m.selectedTask].Output.String())
			m.outputPanel.SetContent(outputContent)
			if isAtBottom {
				m.outputPanel.GotoBottom()
			}
		}
		return m, tickReadOutputsMsg()
	case spinner.TickMsg:
		var cmd tea.Cmd
		// Build the side panel content
		m.spinner, cmd = m.spinner.Update(msg)
		m.stepPanel.SetContent(m.calculateStepPanelContent())

		return m, cmd
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				if m.focusOutput {
					m.outputPanel.ScrollUp(3)
				} else {
					m.focusedTask = max(m.focusedTask-1, 0)
					m.stepPanel.Focus(m.calculateTaskLine(m.focusedTask))
				}
			case tea.MouseButtonWheelDown:
				if m.focusOutput {
					m.outputPanel.ScrollDown(3)
				} else {
					m.focusedTask = min(m.focusedTask+1, len(m.taskIds)-1)
					m.stepPanel.Focus(m.calculateTaskLine(m.focusedTask))
				}
			}
		}
	}

	return m, nil
}

func (m ifaceModel) View() string {
	help := m.help.FullHelpView([][]key.Binding{
		{m.keymap.up, m.keymap.down},
		{m.keymap.tab, m.keymap.enter, m.keymap.quit},
	})

	var views []string
	views = append(views, m.stepPanel.View())
	if !m.hideOutputPanel {
		views = append(views, m.outputPanel.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help

}
