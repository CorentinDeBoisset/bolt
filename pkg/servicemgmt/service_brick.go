package servicemgmt

import (
	"image/color"

	"github.com/charmbracelet/lipgloss"
	"github.com/corentindeboisset/bolt/pkg"
	"github.com/google/uuid"
)

type serviceState int

const (
	SERVICE_OFF serviceState = iota
	SERVICE_STARTING
	SERVICE_RUNNING
	SERVICE_ERROR
)

const BRICK_MIN_WIDTH = 25

// It is important that all indicators have the same len
const INDICATOR_LEN = 12
const (
	INDICATOR_OFF      = "⠺ OFF      ⠗"
	INDICATOR_STARTING = "⠺ STARTING ⠗"
	INDICATOR_RUNNING  = "⠺ RUNNING  ⠗"
	INDICATOR_ERROR    = "⠺ ERROR    ⠗"
)

const HPADDING = 2

type ServiceBrickModel struct {
	id string

	background   color.Color
	width        int
	focusLevel   int
	cachedHeight int

	ServiceName string
	State       serviceState

	brickStyle lipgloss.Style
	titleStyle lipgloss.Style

	// Status indicator styles
	offStatusStyle      lipgloss.Style
	startingStatusStyle lipgloss.Style
	runningStatusStyle  lipgloss.Style
	errorStatusStyle    lipgloss.Style
}

func NewServiceBrick(name string, width int, background color.Color) *ServiceBrickModel {
	model := ServiceBrickModel{
		id:          uuid.NewString(),
		ServiceName: name,
		State:       SERVICE_OFF,

		background:   background,
		width:        width,
		focusLevel:   0,
		cachedHeight: 0,
	}
	model.refreshStyles()
	model.refreshCachedHeight()

	return &model
}

func (s *ServiceBrickModel) Resize(width int) {
	s.width = width
	s.refreshStyles()
	s.refreshCachedHeight()
}

func (s *ServiceBrickModel) SetFocusLevel(level int) {
	s.focusLevel = max(min(level, 2), 0)
	s.refreshStyles()
}

func (s *ServiceBrickModel) refreshStyles() {
	brickStyle := pkg.NoticeableSurfaceStyle
	switch s.focusLevel {
	case 1:
		brickStyle = pkg.HighlightSurfaceStyle
	case 2:
		brickStyle = pkg.AccentSurfaceStyle
	}

	s.titleStyle = brickStyle.
		PaddingRight(HPADDING).
		MaxHeight(1).
		Width(s.width - INDICATOR_LEN - HPADDING*2)

	s.brickStyle = brickStyle.Padding(1, 2)

	s.offStatusStyle = brickStyle
	s.startingStatusStyle = brickStyle.Foreground(lipgloss.Color("#d3a825"))
	s.runningStatusStyle = brickStyle.Foreground(lipgloss.Color("#1eaa25"))
	s.errorStatusStyle = brickStyle.Foreground(lipgloss.Color("#d82525"))
}

func (s *ServiceBrickModel) refreshCachedHeight() {
	content := s.View()
	s.cachedHeight = lipgloss.Height(content) + 1
}

func (s *ServiceBrickModel) Height() int {
	return s.cachedHeight
}

func (s *ServiceBrickModel) Focusable() bool {
	return true
}

func (s *ServiceBrickModel) View() string {
	if s.width < BRICK_MIN_WIDTH {
		return s.brickStyle.
			Width(s.width).
			Render("")
	}

	title := s.titleStyle.Render(s.ServiceName)
	var indicator string
	switch s.State {
	case SERVICE_OFF:
		indicator = s.offStatusStyle.Render(INDICATOR_OFF)
	case SERVICE_STARTING:
		indicator = s.startingStatusStyle.Render(INDICATOR_STARTING)
	case SERVICE_RUNNING:
		indicator = s.runningStatusStyle.Render(INDICATOR_RUNNING)
	case SERVICE_ERROR:
		indicator = s.errorStatusStyle.Render(INDICATOR_ERROR)
	}
	content := lipgloss.JoinHorizontal(lipgloss.Top, title, indicator)

	return s.brickStyle.Render(content)
}

func (s *ServiceBrickModel) Id() string {
	return s.id
}
