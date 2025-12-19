package servicemgmt

import (
	"fmt"
	"image/color"

	"github.com/charmbracelet/lipgloss"
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
	background color.Color
	width      int
	focusLevel int

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

func NewServiceBrick(name string, width int, background color.Color) ServiceBrickModel {
	model := ServiceBrickModel{
		ServiceName: name,
		State:       SERVICE_OFF,

		background: background,
		width:      width,
		focusLevel: 0,
	}
	model.refreshStyles()

	return model
}

func (s *ServiceBrickModel) Resize(width int) {
	s.width = width
	s.refreshStyles()
}

func (s *ServiceBrickModel) SetFocusLevel(level int) {
	s.focusLevel = max(min(level, 2), 0)
	s.refreshStyles()
}

func (s *ServiceBrickModel) refreshStyles() {
	baseBackgroundR, baseBackgroundG, baseBackgroundB, baseBackgroundA := s.background.RGBA()

	var baseStyle lipgloss.Style
	switch s.focusLevel {
	case 1:
		backgroundR := (baseBackgroundR*50 + 255*50) / 100
		backgroundG := (baseBackgroundG*50 + 255*50) / 100
		backgroundB := (baseBackgroundB*50 + 255*50) / 100
		backgroundHex := fmt.Sprintf("#%02x%02x%02x%02x", backgroundR, backgroundG, backgroundB, baseBackgroundA)
		baseStyle = lipgloss.NewStyle().Background(lipgloss.Color(backgroundHex))
	case 2:
		backgroundR := (baseBackgroundR*70 + 255*30) / 100
		backgroundG := (baseBackgroundG*70 + 255*30) / 100
		backgroundB := (baseBackgroundB*70 + 255*30) / 100
		backgroundHex := fmt.Sprintf("#%02x%02x%02x%02x", backgroundR, backgroundG, backgroundB, baseBackgroundA)
		baseStyle = lipgloss.NewStyle().Background(lipgloss.Color(backgroundHex))
	default:
		backgroundHex := fmt.Sprintf("#%02x%02x%02x%02x", baseBackgroundR, baseBackgroundG, baseBackgroundB, baseBackgroundA)
		baseStyle = lipgloss.NewStyle().Background(lipgloss.Color(backgroundHex))
	}

	s.titleStyle = baseStyle.
		Foreground(lipgloss.Color("2")).
		MarginRight(HPADDING).
		MaxHeight(NAME_MAX_HEIGHT).
		Width(s.width - 14 - HPADDING*2)
	s.brickStyle = baseStyle.Padding(1, 2)

	s.offStatusStyle = baseStyle.Foreground(lipgloss.Color("#5e5e5eff"))
	s.startingStatusStyle = baseStyle.Foreground(lipgloss.Color("#d3a825ff"))
	s.runningStatusStyle = baseStyle.Foreground(lipgloss.Color("#1eaa25ff"))
	s.errorStatusStyle = baseStyle.Foreground(lipgloss.Color("#d82525ff"))
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
	content := lipgloss.JoinVertical(lipgloss.Top, title, indicator)

	return s.brickStyle.Render(content)
}
