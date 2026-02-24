package outputviewer

import (
	"fmt"
	"log"
	"regexp"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchBarModel struct {
	currentSearch []byte
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
	case "backspace":
		if len(m.currentSearch) > 0 {
			_, runeLen := utf8.DecodeLastRune(m.currentSearch)
			m.currentSearch = m.currentSearch[0 : len(m.currentSearch)-runeLen]
		}
	case " ":
		m.currentSearch = utf8.AppendRune(m.currentSearch, msg.Runes[0])
	case "home":
		// TODO: when cursor position is done
	case "end":
		// TODO: when cursor position is done
	case "delete":
		// TODO: when cursor position is done
	default:
		if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				m.currentSearch = utf8.AppendRune(m.currentSearch, r)
			}
		}
	}
}

func (m *SearchBarModel) Clear() {
	m.currentSearch = nil
}

func (m *SearchBarModel) View(width int) string {
	return lipgloss.NewStyle().
		Inline(true).
		Width(width).
		MaxWidth(width).
		MaxHeight(1).
		Render(fmt.Sprintf("Search: %s", m.currentSearch))
}
