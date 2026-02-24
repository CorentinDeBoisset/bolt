package outputviewer

import (
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchBarModel struct {
	currentSearch []rune
	cursorIdx     int
	showCursor    bool
	virtualCursor cursor.Model
}

func sanitize(input []rune) []rune {
	sanitized := make([]rune, 0, 2*len(input)) // worst case scenario: all characters must be escaped

	for idx, r := range input {
		switch r {
		case '\t':
			sanitized = append(sanitized, '\\', 't')
		case '\r':
			// If followed by \n, we just skip the \r
			if idx < len(input)-1 && input[idx+1] == '\n' {
				continue
			}

			sanitized = append(sanitized, '\\', 'n')
		case '\n':
			sanitized = append(sanitized, '\\', 'n')
		default:
			if unicode.IsControl(r) {
				continue
			}
			sanitized = append(sanitized, r)
		}
	}

	return sanitized
}

func (m *SearchBarModel) insertRunes(input []rune) {
	m.currentSearch = slices.Concat(m.currentSearch[:m.cursorIdx], input, m.currentSearch[m.cursorIdx:])
	m.moveCursor(len(input))
}

// Move the cursor to the absolute index `idx`
func (m *SearchBarModel) moveCursorAbs(idx int) {
	newIndex := min(max(idx, 0), len(m.currentSearch))
	m.cursorIdx = newIndex
	if m.cursorIdx == len(m.currentSearch) {
		m.virtualCursor.SetChar(" ")
	} else {
		m.virtualCursor.SetChar(string(m.currentSearch[m.cursorIdx]))
	}
}

// Move the cursor of `n` relative to the current position
func (m *SearchBarModel) moveCursor(n int) {
	newIndex := min(max(m.cursorIdx+n, 0), len(m.currentSearch))
	m.cursorIdx = newIndex
	if m.cursorIdx == len(m.currentSearch) {
		m.virtualCursor.SetChar(" ")
	} else {
		m.virtualCursor.SetChar(string(m.currentSearch[m.cursorIdx]))
	}
}

func (m *SearchBarModel) Submit() *regexp.Regexp {
	if len(m.currentSearch) == 0 {
		return nil
	}

	reg, err := regexp.Compile(string(m.currentSearch))
	if err != nil {
		log.Printf("The requested regexp is invalid: %s", err)
		return nil
	}

	return reg
}

func (m *SearchBarModel) HandleKeyMsg(msg tea.KeyMsg) {
	switch msg.String() {
	case "left":
		m.moveCursor(-1)
	case "right":
		m.moveCursor(1)
	case "home":
		m.moveCursorAbs(0)
	case "end":
		m.moveCursorAbs(len(m.currentSearch))
	case "backspace":
		if len(m.currentSearch) > 0 && m.cursorIdx > 0 {
			m.currentSearch = slices.Concat(m.currentSearch[:max(0, m.cursorIdx-1)], m.currentSearch[m.cursorIdx:])
			m.moveCursor(-1)
		}
	case "delete":
		if len(m.currentSearch) > 0 && m.cursorIdx < len(m.currentSearch) {
			m.currentSearch = slices.Concat(m.currentSearch[:m.cursorIdx], m.currentSearch[m.cursorIdx+1:])
			m.moveCursor(0)
		}
	case " ":
		m.insertRunes(msg.Runes)
	default:
		if msg.Type == tea.KeyRunes {
			// Important to sanitize, in case of a ctrl-v there are a lot of control characters (new lines, tabs...)
			m.insertRunes(sanitize(msg.Runes))
		}
	}
}

func (m *SearchBarModel) ToggleCursor(visible bool) {
	m.showCursor = visible
	m.moveCursorAbs(len(m.currentSearch))
}

func (m *SearchBarModel) SetCursorVisibility(focus bool) {
	m.virtualCursor.Blink = !focus
}

func (m *SearchBarModel) Clear() {
	m.currentSearch = nil
	m.moveCursorAbs(0)
}

func (m *SearchBarModel) View(width int) string {
	var content string
	if !m.showCursor {
		content = string(m.currentSearch)
	} else {
		if m.cursorIdx >= 0 && m.cursorIdx < len(m.currentSearch) {
			parts := []string{string(m.currentSearch[:m.cursorIdx]), m.virtualCursor.View(), string(m.currentSearch[m.cursorIdx+1:])}
			content = strings.Join(parts, "")
		} else {
			content = string(m.currentSearch) + m.virtualCursor.View()
		}
	}

	return lipgloss.NewStyle().
		Width(width).
		MaxWidth(width).
		Height(1).
		MaxHeight(1).
		Render(fmt.Sprintf("Search: %s", content))
}
