package servicemgmt

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/corentindeboisset/tera/pkg/iface"
	"github.com/google/uuid"
)

type SeparatorModel struct {
	id           string
	currentStyle lipgloss.Style
	width        int
}

func NewSeparator(width int, theme iface.Theme) *SeparatorModel {
	return &SeparatorModel{
		id:           uuid.NewString(),
		currentStyle: lipgloss.NewStyle().Foreground(theme.SeparatorColor).AlignHorizontal(lipgloss.Center).Width(width),
		width:        width,
	}
}

func (m *SeparatorModel) Resize(width int) {
	m.width = width
	m.currentStyle = m.currentStyle.Width(width)
}

func (m *SeparatorModel) Focusable() bool {
	return false
}

func (m *SeparatorModel) Height() int {
	return 1
}

func (m *SeparatorModel) View() string {
	var contentWidth int
	if m.width < 10 {
		return strings.Repeat("─", m.width)
	} else if m.width < 30 {
		contentWidth = m.width * 60 / 100
	} else {
		contentWidth = m.width * 35 / 100
	}

	contentHalfWidth := (contentWidth - 3) / 2

	content := strings.Repeat("─", contentHalfWidth) + " ⟡ " + strings.Repeat("─", contentHalfWidth)

	return m.currentStyle.Render(content)
}

func (s *SeparatorModel) Id() string {
	return s.id
}
