package servicemgmt

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type SeparatorModel struct {
	currentStyle lipgloss.Style
	width        int
}

func NewSeparator(width int) *SeparatorModel {
	return &SeparatorModel{
		lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		width,
	}
}

func (m *SeparatorModel) Resize(width int) {
	m.width = width
}

func (m *SeparatorModel) Focusable() bool {
	return false
}

func (m *SeparatorModel) Height() int {
	return 1
}

func (m *SeparatorModel) View() string {
	return m.currentStyle.Render(strings.Repeat("─", m.width))
}
