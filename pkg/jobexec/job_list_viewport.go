package jobexec

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ListViewportLine struct {
	Padding int
	Content string
}

func (l *ListViewportLine) Render(width int) []string {
	return strings.Split(
		lipgloss.NewStyle().
			Padding(0, 0, 0, l.Padding).
			Width(width).
			Render(l.Content),
		"\n",
	)
}

func NewListViewportModel(width, height int) (m ListViewportModel) {
	m.width = width
	m.height = height
	m.focusedLine = 0
	m.yOffset = 0

	return m
}

type ListViewportModel struct {
	width  int
	height int

	yOffset     int
	focusedLine int

	Style lipgloss.Style

	sourceLines []ListViewportLine

	visibleLines [][]string
}

func (m *ListViewportModel) AtTop() bool {
	return m.focusedLine <= 0
}

func (m *ListViewportModel) AtBottom() bool {
	return m.focusedLine >= len(m.sourceLines)-1
}

func (m *ListViewportModel) SetContent(lines []ListViewportLine) {
	m.sourceLines = lines
	m.recomputeVisibleLines()

	if len(lines) == 0 || m.focusedLine == -1 {
		// Reset the focus to the top
		m.Focus(0)
		return
	}

	if m.focusedLine >= len(m.sourceLines) {
		m.Focus(len(m.sourceLines) - 1)
	} else {
		// Recompute y offset
		m.Focus(m.focusedLine)
	}
}

func (m *ListViewportModel) recomputeVisibleLines() {
	visibleLines := make([][]string, 0)
	for _, line := range m.sourceLines {
		line.Content = strings.ReplaceAll(line.Content, "\r\n", "\n")
		visibleLines = append(visibleLines, line.Render(m.width))
	}
	m.visibleLines = visibleLines
}

// Get the y offset of the line at the requested index
func (m *ListViewportModel) getSourceLineOffset(idx int) int {
	if idx <= 0 || idx > len(m.visibleLines) {
		return 0
	}

	ret := 0
	for i := range idx {
		ret += len(m.visibleLines[i])
	}
	return ret
}

func (m *ListViewportModel) ContentHeight() int {
	ret := 0
	for _, visibleLine := range m.visibleLines {
		ret += len(visibleLine)
	}
	return ret
}

func (m *ListViewportModel) Focus(idx int) {
	if len(m.sourceLines) == 0 {
		m.focusedLine = -1
		m.SetYOffset(0)
		return
	}

	m.focusedLine = clamp(idx, 0, len(m.sourceLines)-1)
	m.SetYOffset(m.getSourceLineOffset(m.focusedLine) - (m.height / 2)) // FIXME: this is not the right height
}

func (m *ListViewportModel) maxYOffset() int {
	return max(0, (m.ContentHeight() - m.height + m.Style.GetVerticalFrameSize())) // FIXME: same
}

func (m *ListViewportModel) SetYOffset(n int) {
	m.yOffset = clamp(n, 0, m.maxYOffset())
}

func (m *ListViewportModel) PageDown() {
	// FIXME: this should not use m.height
	m.ScrollDown(m.height)
}

func (m *ListViewportModel) PageUp() {
	// FIXME: this should not use m.height
	m.ScrollUp(m.height)
}

func (m *ListViewportModel) ScrollDown(n int) {
	m.Focus(m.focusedLine + n)
}

func (m *ListViewportModel) ScrollUp(n int) {
	m.Focus(m.focusedLine - n)
}

func (m *ListViewportModel) GoToTop() {
	m.Focus(0)
}

func (m *ListViewportModel) GoToBottom() {
	m.Focus(len(m.sourceLines) - 1)
}

func (m *ListViewportModel) Resize(width, height int) {
	m.width = width
	m.height = height

	m.recomputeVisibleLines()
	m.Focus(m.focusedLine)
}

func (m *ListViewportModel) View() string {
	w, h := m.width, m.height
	if sw := m.Style.GetWidth(); sw != 0 {
		w = min(w, sw)
	}
	if sh := m.Style.GetHeight(); sh != 0 {
		h = min(h, sh)
	}

	innerWidth := w - m.Style.GetHorizontalFrameSize()
	innerHeight := h - m.Style.GetVerticalFrameSize()

	displayedLines := make([]string, 0)
	for _, visibleLine := range m.visibleLines {
		displayedLines = append(displayedLines, visibleLine...)
	}
	displayedLines = displayedLines[m.yOffset:min(m.yOffset+innerHeight, len(displayedLines))]

	contents := lipgloss.NewStyle().
		Width(innerWidth).      // pad to width
		Height(innerHeight).    // pad to height
		MaxHeight(innerHeight). // truncate height if taller
		MaxWidth(innerWidth).   // truncate width if wider
		Render(strings.Join(displayedLines, "\n"))

	// Style size already applied
	return m.Style.UnsetWidth().UnsetHeight().Render(contents)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
