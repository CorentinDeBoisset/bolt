package listviewport

import (
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ListItem interface {
	Height() int
	View() string
	Focusable() bool
	Resize(width int)
}

type Model struct {
	width  int
	height int

	yOffset     int
	focusedItem int

	Style lipgloss.Style

	items []ListItem
}

func New(width, height int) (m Model) {
	m.width = width
	m.height = height
	m.focusedItem = 0

	return m
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

	// TODO: set focusLevel of the children
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

func (m *Model) PageDown() {
	// TODO: return the idx of the focused item to the parent
	yMovement := 0
	for i := m.focusedItem + 1; i < len(m.items); i++ {
		yMovement += m.items[i].Height()
		if yMovement >= m.height && m.items[i].Focusable() {
			m.Focus(i)
			return
		}
	}
	m.GoToBottom()
}

func (m *Model) PageUp() {
	// TODO: return the idx of the focused item to the parent
	yMovement := 0
	for i := m.focusedItem - 1; i >= 0; i-- {
		yMovement += m.items[i].Height()
		if yMovement >= m.height && m.items[i].Focusable() {
			m.Focus(i)
			return
		}
	}
	m.GoToTop()
}

func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height

	for _, item := range m.items {
		item.Resize(width)
	}
}

func (m *Model) View() string {
	renderedItems := make([]string, 0)

	focusedItemHeight := m.items[m.focusedItem].Height()

	beforeFocusHeight := 0
	if m.focusedItem > 0 {
		for i := m.focusedItem - 1; i >= 0; i++ {
			candidateBeforeFocusHeight := beforeFocusHeight + m.items[i].Height()
			if candidateBeforeFocusHeight >= ((m.height - focusedItemHeight) / 2) {
				break
			}
			renderedItems = append(renderedItems, m.items[i].View())
			beforeFocusHeight = candidateBeforeFocusHeight
		}
		slices.Reverse(renderedItems)
	}

	renderedItems = append(renderedItems, m.items[m.focusedItem].View())

	afterFocusHeight := 0
	if m.focusedItem < len(m.items)-1 {
		for i := m.focusedItem + 1; i < len(m.items); i++ {
			candidateAfterFocusHeight := afterFocusHeight + m.items[i].Height()
			remainingHeight := m.height - focusedItemHeight - beforeFocusHeight
			if candidateAfterFocusHeight >= remainingHeight {
				// Only show a part of the last visible item
				lastRenderedItem := m.items[i].View()
				lastRenderedItem = strings.ReplaceAll(lastRenderedItem, "\r\n", "\n")
				lastRenderedItemLines := strings.Split(lastRenderedItem, "\n")
				renderedItems = append(renderedItems, strings.Join(lastRenderedItemLines[:remainingHeight], "\n"))
				break
			}
			renderedItems = append(renderedItems, m.items[i].View())
			afterFocusHeight = candidateAfterFocusHeight
		}
	}

	return strings.Join(renderedItems, "\n")
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
