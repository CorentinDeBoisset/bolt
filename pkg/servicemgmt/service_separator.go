package servicemgmt

import "github.com/charmbracelet/lipgloss"

type SeparatorModel struct {
	currentStyle lipgloss.Style
}

func NewSeparator(width int) *SeparatorModel {
	return &SeparatorModel{
		lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("2")).
			Width(width),
	}
}

func (m *SeparatorModel) Resize(width int) {
	m.currentStyle = m.currentStyle.Width(width)
}

func (m *SeparatorModel) Focusable() bool {
	return false
}

func (m *SeparatorModel) Height() int {
	return 1
}

func (m *SeparatorModel) View() string {
	return m.currentStyle.Render("")
}
