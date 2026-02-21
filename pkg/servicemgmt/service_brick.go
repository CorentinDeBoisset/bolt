package servicemgmt

import (
	"image/color"

	"github.com/charmbracelet/lipgloss"
	"github.com/corentindeboisset/bolt/pkg"
)

type serviceState int

const (
	SERVICE_OFF serviceState = iota
	SERVICE_STARTING
	SERVICE_RUNNING
	SERVICE_ERROR
)

const BRICK_MIN_WIDTH = 30

const (
	INDICATOR_OFF      = "⠺ OFF      ⠗"
	INDICATOR_STARTING = "⠺ STARTING ⠗"
	INDICATOR_RUNNING  = "⠺ RUNNING  ⠗"
	INDICATOR_ERROR    = "⠺ ERROR    ⠗"
)

const HPADDING = 2
const NAME_MAX_HEIGHT = 3

type ServiceBrickModel struct {
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
	brickBackgound := pkg.NoticeableBackgroundColor
	switch s.focusLevel {
	case 1:
		brickBackgound = pkg.AccentBackgroundColor
	case 2:
		brickBackgound = pkg.HighlightBackgroundColor
	}

	baseStyle := pkg.BaseAppStyle.Background(brickBackgound)

	s.titleStyle = baseStyle.
		Foreground(pkg.BodyColor).
		PaddingRight(HPADDING).
		MaxHeight(NAME_MAX_HEIGHT).
		Width(s.width - 16 - HPADDING*2)
	s.brickStyle = baseStyle.Padding(1, 2).Margin(0, 2)

	s.offStatusStyle = baseStyle.Foreground(pkg.BodyColor)
	s.startingStatusStyle = baseStyle.Foreground(lipgloss.Color("#d3a825"))
	s.runningStatusStyle = baseStyle.Foreground(lipgloss.Color("#1eaa25"))
	s.errorStatusStyle = baseStyle.Foreground(lipgloss.Color("#d82525"))
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
		return ""
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
