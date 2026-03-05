package listviewport

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

type ListItem interface {
	Id() string
	Height() int
	View() string
	Focusable() bool
	Resize(width int)
}

type Model struct {
	width  int
	height int

	focusedItem int

	baseStyle lipgloss.Style

	items []ListItem
}

func New(width, height int, baseStyle lipgloss.Style) (m Model) {
	return Model{
		width:       width,
		height:      height,
		focusedItem: 0,
		baseStyle:   baseStyle,
	}
}

func (m *Model) AtTop() bool {
	return m.focusedItem <= 0
}

func (m *Model) AtBottom() bool {
	return m.focusedItem > len(m.items)-1
}

func (m *Model) SetItems(items []ListItem) {
	m.items = items

	// Ensure the focused item is within acceptable boundaries
	m.Focus(m.focusedItem)
}

func (m *Model) Focus(idx int) {
	if len(m.items) == 0 {
		m.focusedItem = -1
		return
	}

	m.focusedItem = clamp(idx, 0, len(m.items)-1)
}

func (m *Model) GoToTop() {
	for i := 0; i < len(m.items); i++ {
		if m.items[i].Focusable() {
			m.Focus(i)
			return
		}
	}
}

func (m *Model) GoToBottom() {
	for i := len(m.items) - 1; i >= 0; i-- {
		if m.items[i].Focusable() {
			m.Focus(i)
			return
		}
	}
}

func (m *Model) ScrollDown(n int) {
	for i := m.focusedItem + 1; i < len(m.items); i++ {
		if m.items[i].Focusable() {
			m.Focus(i)
			return
		}
	}
}

func (m *Model) ScrollUp(n int) {
	for i := m.focusedItem - 1; i >= 0; i-- {
		if m.items[i].Focusable() {
			m.Focus(i)
			return
		}
	}
}

func (m *Model) PageDown() string {
	yMovement := 0
	for i := m.focusedItem + 1; i < len(m.items); i++ {
		yMovement += m.items[i].Height()
		if yMovement >= m.height && m.items[i].Focusable() {
			m.Focus(i)
			return m.items[i].Id()
		}
	}

	m.GoToBottom()
	return m.items[m.focusedItem].Id()
}

func (m *Model) PageUp() string {
	yMovement := 0
	for i := m.focusedItem - 1; i >= 0; i-- {
		yMovement += m.items[i].Height()
		if yMovement >= m.height && m.items[i].Focusable() {
			m.Focus(i)
			return m.items[i].Id()
		}
	}

	m.GoToTop()
	return m.items[m.focusedItem].Id()
}

func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height

	itemWidth := max(width-m.baseStyle.GetHorizontalFrameSize(), 0)

	for _, item := range m.items {
		item.Resize(itemWidth)
	}
}

func (m *Model) View() string {
	focusedItemContent := m.items[m.focusedItem].View()
	focusedItemHeight := lipgloss.Height(focusedItemContent)

	availableHeight := (m.height - focusedItemHeight - m.baseStyle.GetVerticalFrameSize())

	if availableHeight < 0 || m.width-m.baseStyle.GetHorizontalFrameSize() < 0 {
		// The available space is too small, we return an empty block
		return m.baseStyle.
			Padding(0).
			Margin(0).
			Border(lipgloss.Border{}, false).
			Height(m.height).
			Width(m.width).
			Render("")
	}

	beforeFocusContent := make([]string, 0)
	if m.focusedItem > 0 {
		for i := m.focusedItem - 1; i >= 0; i-- {
			newLines := strings.Split(m.items[i].View(), "\n")
			slices.Reverse(newLines)
			beforeFocusContent = append(beforeFocusContent, newLines...)
			if len(beforeFocusContent) > availableHeight {
				break
			}
		}
	}

	afterFocusContent := make([]string, 0)
	if m.focusedItem < len(m.items)-1 {
		for i := m.focusedItem + 1; i < len(m.items); i++ {
			newLines := strings.Split(m.items[i].View(), "\n")
			afterFocusContent = append(afterFocusContent, newLines...)
			if len(afterFocusContent) >= availableHeight {
				break
			}
		}
	}

	enoughContentBefore := len(beforeFocusContent) >= (availableHeight)/2
	enoughContentAfter := len(afterFocusContent) >= (availableHeight)/2
	linesBefore := 0
	linesAfter := 0
	if enoughContentAfter && enoughContentBefore {
		linesBefore = availableHeight / 2
		linesAfter = min(availableHeight-linesBefore, len(afterFocusContent))
	} else if enoughContentBefore && !enoughContentAfter {
		linesAfter = len(afterFocusContent)
		linesBefore = min(availableHeight-linesAfter, len(beforeFocusContent))
	} else if !enoughContentBefore && enoughContentAfter {
		linesBefore = len(beforeFocusContent)
		linesAfter = min(availableHeight-linesBefore, len(afterFocusContent))
	} else {
		linesBefore = len(beforeFocusContent)
		linesAfter = len(afterFocusContent)
	}

	output := ""

	beforeFocusContent = beforeFocusContent[0:linesBefore]
	slices.Reverse(beforeFocusContent)
	afterFocusContent = afterFocusContent[0:linesAfter]

	if len(beforeFocusContent) > 0 {
		output = lipgloss.JoinVertical(lipgloss.Left, strings.Join(beforeFocusContent, "\n"), focusedItemContent)
	} else {
		output = focusedItemContent
	}

	if len(afterFocusContent) > 0 {
		output = lipgloss.JoinVertical(lipgloss.Left, output, strings.Join(afterFocusContent, "\n"))
	}

	return m.baseStyle.Render(output)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
