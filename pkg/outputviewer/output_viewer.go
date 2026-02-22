package outputviewer

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int

	buffer fmt.Stringer

	style         lipgloss.Style
	cachedContent []string
	offset        int
}

func New(width, height int, borderColor lipgloss.TerminalColor, buffer fmt.Stringer) (m Model) {
	return Model{
		width:  width,
		height: height,

		buffer: buffer,

		style: lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(borderColor),
		cachedContent: nil,
		offset:        0,
	}
}

func (m *Model) AtTop() bool {
	return m.offset <= 0
}

func (m *Model) AtBottom() bool {
	return m.offset >= m.maxOffset()
}

func (m *Model) ScrollPercent() float64 {
	if m.height >= m.ContentHeight() {
		return 1.0
	}
	y := float64(m.offset)
	h := float64(m.height)
	t := float64(m.ContentHeight())
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

func (m *Model) SetBuffer(b fmt.Stringer, goToBottom bool) {
	m.buffer = b
	if goToBottom {
		m.GoToBottom()
	} else {
		m.GoToTop()
	}
}

func (m *Model) SetBorderColor(color lipgloss.TerminalColor) {
	m.style = m.style.BorderForeground(color)
}

func (m *Model) RefreshContent() {
	if m.buffer == nil {
		return
	}

	wasAtBottom := m.AtBottom()
	output := m.buffer.String()
	output = strings.ReplaceAll(output, "\r\n", "\n") // normalize line endings

	m.cachedContent = strings.Split(output, "\n")

	if wasAtBottom || m.offset > m.maxOffset() {
		m.GoToBottom()
	}
}

func (m *Model) Resize(width, height int) {
	wasAtBottom := m.AtBottom()
	m.width = width
	m.height = height

	if wasAtBottom || m.offset > m.maxOffset() {
		m.GoToBottom()
	}
}

func (m *Model) maxOffset() int {
	return max(0, m.ContentHeight()-m.height+m.style.GetVerticalFrameSize())
}

func (m *Model) SetOffset(n int) {
	m.offset = min(max(n, 0), m.maxOffset())
}

func (m *Model) PageDown() {
	if m.AtBottom() {
		return
	}

	m.ScrollDown(m.height)
}

func (m *Model) PageUp() {
	if m.AtTop() {
		return
	}

	m.ScrollUp(m.height)
}

func (m *Model) ScrollDown(n int) {
	if m.AtBottom() || n <= 0 || m.ContentHeight() == 0 {
		return
	}

	m.SetOffset(m.offset + n)
}

func (m *Model) ScrollUp(n int) {
	if m.AtTop() || n <= 0 || m.ContentHeight() == 0 {
		return
	}

	m.SetOffset(m.offset - n)
}

func (m *Model) GoToBottom() {
	if m.AtBottom() {
		return
	}

	m.SetOffset(m.maxOffset())
}

func (m *Model) GoToTop() {
	if m.AtTop() {
		return
	}

	m.SetOffset(0)
}

func (m *Model) ContentHeight() int {
	if m.cachedContent == nil {
		return 0
	}
	return len(m.cachedContent)
}

func (m *Model) InnerFrameWidth() int {
	return max(m.width-m.style.GetHorizontalFrameSize(), 0)
}

func (m *Model) InnerFrameHeight() int {
	return max(m.height-m.style.GetVerticalFrameSize(), 0)
}

func (m *Model) View() string {
	contentWidth := m.InnerFrameWidth()
	contentHeight := m.InnerFrameHeight()

	var content string
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		MaxHeight(contentHeight)
	if m.ContentHeight() == 0 {
		content = contentStyle.Render("")
	} else {
		content = contentStyle.Render(strings.Join(m.cachedContent, "\n"))
	}

	return m.style.Render(content)
}
