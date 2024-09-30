package pkg

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
)

var (
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("111"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("7"))

	successFlag  = lipgloss.NewStyle().SetString("✓").Bold(true).Foreground(lipgloss.Color("082"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	failureFlag  = lipgloss.NewStyle().SetString("✗").Bold(true).Foreground(lipgloss.Color("196"))
)

type RefreshStatusMsg time.Time

type keymap = struct {
	up, down, tab, enter, quit key.Binding
}

type ifaceModel struct {
	width           int
	height          int
	keymap          keymap
	help            help.Model
	stepPanelWidth  int
	hideOutputPanel bool

	focusOutput    bool
	selectedTaskId string
	spinner        spinner.Model
	stepPanel      viewport.Model
	outputPanel    viewport.Model

	globalConfig *CiConfig
	statuses     []StepStatus
	taskOutputs  map[string]*SafeBuffer
}

func tickReadOutputsMsg() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return RefreshStatusMsg(t)
	})
}

func newModel(config *CiConfig, statuses []StepStatus) ifaceModel {
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
				key.WithHelp("↵/Enter", "Select a job to display"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("Ctrl+C ", "Exit"),
			),
		},
		focusOutput:     false,
		hideOutputPanel: false,
		globalConfig:    config,
		stepPanelWidth:  15,
		statuses:        statuses,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(spinnerStyle),
		),
		stepPanel:   viewport.New(15, 10),
		outputPanel: viewport.New(30, 10),
	}

	m.updateKeyBindings()
	m.calculateMinPanelSize()
	m.stepPanel.Width = m.stepPanelWidth
	m.stepPanel.Style = focusedBorderStyle
	m.outputPanel.Style = blurredBorderStyle

	m.initializeTaskOutputs()

	return m
}

func (m *ifaceModel) initializeTaskOutputs() {
	m.taskOutputs = make(map[string]*SafeBuffer)
	for stepIdx := range m.statuses {
		m.statuses[stepIdx].Mtx.Lock()
		if m.statuses[stepIdx].BeforeHooks != nil {
			id := fmt.Sprintf("step#%d__bh", stepIdx)
			m.taskOutputs[id] = &m.statuses[stepIdx].BeforeHooks.Output
		}
		if m.statuses[stepIdx].AfterHooks != nil {
			id := fmt.Sprintf("step#%d__ah", stepIdx)
			m.taskOutputs[id] = &m.statuses[stepIdx].AfterHooks.Output
		}
		for jobIdx := range m.statuses[stepIdx].Jobs {
			id := fmt.Sprintf("step#%d__job#%d", stepIdx, jobIdx)
			m.taskOutputs[id] = &m.statuses[stepIdx].Jobs[jobIdx].Output
			if len(m.selectedTaskId) == 0 {
				m.selectedTaskId = id
			}
		}
		m.statuses[stepIdx].Mtx.Unlock()
	}
}

func (m *ifaceModel) calculateMinPanelSize() {
	maxLen := 10
	for _, step := range m.globalConfig.Steps {
		if maxLen < utf8.RuneCountInString(step.Name) {
			maxLen = utf8.RuneCountInString(step.Name)
		}
		for _, job := range step.Jobs {
			if maxLen < utf8.RuneCountInString(job.Name)+2 {
				maxLen = utf8.RuneCountInString(job.Name) + 2
			}
		}
		if len(step.RunBefore) > 0 && maxLen < 19 {
			maxLen = 19
		}
		if len(step.RunAfter) > 0 && maxLen < 18 {
			maxLen = 18
		}
	}

	m.stepPanelWidth = maxLen + 8
}

func (m *ifaceModel) updateSizes() {
	// The height is fixed
	m.stepPanel.Height = m.height - 5
	m.outputPanel.Height = m.height - 5

	outputWidth := m.width - m.stepPanelWidth
	if outputWidth < 10 {
		m.hideOutputPanel = true
		m.stepPanel.Width = m.width
		return
	}
	m.outputPanel.Width = outputWidth
	m.stepPanel.Width = m.stepPanelWidth
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

func (m *ifaceModel) formatTask(padding int, state TaskState, name string) string {
	paddingString := strings.Repeat(" ", padding)
	switch state {
	case STATE_NOT_STARTED:
		return fmt.Sprintf("%s   %s", paddingString, name)
	case STATE_RUNNING:
		return fmt.Sprintf("%s%s %s", paddingString, m.spinner.View(), name)
	case STATE_SUCCESSFUL:
		return fmt.Sprintf("%s%s  %s", paddingString, successFlag, name)
	case STATE_FAILED:
		return fmt.Sprintf("%s%s  %s", paddingString, failureFlag, name)
	}

	return ""
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
				// Scroll up in the logs
			} else {
				// Move up in the list of jobs
			}
		case "down", "j":
			// Scroll down the focused panel
			if m.focusOutput {
				// Scroll down in the logs
			} else {
				// Move down in the list of jobs
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
			if m.focusOutput {
				// Select a new job
				// Edit m.selectedTaskId
				// Scroll to bottom
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
	case RefreshStatusMsg:
		if !m.hideOutputPanel {
			isAtBottom := m.outputPanel.AtBottom()
			m.outputPanel.SetContent(m.taskOutputs[m.selectedTaskId].String())
			if isAtBottom {
				m.outputPanel.GotoBottom()
			}
		}
		return m, tickReadOutputsMsg()
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		stepPanelLines := make([]string, 0, len(m.statuses)+len(m.taskOutputs))
		for stepIdx, stepConfig := range m.globalConfig.Steps {
			m.statuses[stepIdx].Mtx.Lock()
			stepPanelLines = append(stepPanelLines, m.formatTask(1, m.statuses[stepIdx].state, stepConfig.Name))
			if m.statuses[stepIdx].BeforeHooks != nil {
				stepPanelLines = append(stepPanelLines, m.formatTask(3, m.statuses[stepIdx].BeforeHooks.state, "Run-Before hooks"))
			}
			for jobIdx, jobConfig := range stepConfig.Jobs {
				stepPanelLines = append(stepPanelLines, m.formatTask(3, m.statuses[stepIdx].Jobs[jobIdx].state, jobConfig.Name))
			}
			if m.statuses[stepIdx].AfterHooks != nil {
				stepPanelLines = append(stepPanelLines, m.formatTask(3, m.statuses[stepIdx].AfterHooks.state, "Run-After hooks"))
			}
			stepPanelLines = append(stepPanelLines, "")
			m.statuses[stepIdx].Mtx.Unlock()
		}

		m.stepPanel.SetContent(strings.Join(stepPanelLines, "\n"))

		// Build the side content
		return m, cmd
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
