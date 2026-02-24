package outputviewer

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/corentindeboisset/bolt/pkg/cmdrunr"
	"github.com/corentindeboisset/bolt/pkg/iface"
)

var StackedVerticalBorderFirst = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "╭",
	TopRight:    "╮",
	BottomLeft:  "├",
	BottomRight: "┤",
}

var StackedVerticalBorderLast = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "├",
	TopRight:    "┤",
	BottomLeft:  "╰",
	BottomRight: "╯",
}

type Model struct {
	width  int
	height int

	buffer *cmdrunr.SafeBuffer

	theme            iface.Theme
	frameStyle       lipgloss.Style
	rawOutput        []byte
	displayedContent []string
	offset           int

	showSearch        bool
	searchHasFocus    bool
	currentSearch     []byte
	searchRegexp      *regexp.Regexp
	searchResultLines []int
	highlightedMatch  int
}

func New(width, height int, theme iface.Theme, borderColor lipgloss.TerminalColor, buffer *cmdrunr.SafeBuffer) (m Model) {
	return Model{
		width:  width,
		height: height,

		buffer: buffer,

		theme: theme,
		frameStyle: lipgloss.NewStyle().
			Padding(0, 2).
			BorderForeground(borderColor),
		rawOutput:        nil,
		displayedContent: nil,
		offset:           0,

		showSearch:        false,
		currentSearch:     nil,
		searchRegexp:      nil,
		searchResultLines: nil,
		highlightedMatch:  -1,
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

func (m *Model) SetBuffer(b *cmdrunr.SafeBuffer, goToBottom bool) {
	m.buffer = b

	m.clearSearch()

	if goToBottom {
		m.GoToBottom()
	} else {
		m.GoToTop()
	}
}

func (m *Model) SetBorderColor(color lipgloss.TerminalColor) {
	m.frameStyle = m.frameStyle.BorderForeground(color)
}

func (m *Model) RefreshContent() {
	if m.buffer == nil {
		return
	}

	m.rawOutput = bytes.ReplaceAll(m.buffer.Bytes(), []byte("\r\n"), []byte("\n")) // Normalize line endings
	m.refreshDisplayedContent()
}

func (m *Model) refreshDisplayedContent() {
	wasAtBottom := m.AtBottom()

	widthStyle := lipgloss.NewStyle().Width(m.InnerFrameWidth())

	if m.searchRegexp != nil {
		newLineNb := 0
		var decoratedOutput []byte
		decoratedOutput, m.searchResultLines = cmdrunr.DecorateCmdOutput(m.searchRegexp, m.rawOutput, m.highlightedMatch, m.theme.NoticeableSurfaceStyle, m.theme.AccentSurfaceStyle)
		splitDecoratedOutput := bytes.Split(decoratedOutput, []byte("\n"))
		resultLines := make([]string, 0)
		for lineIdx, line := range splitDecoratedOutput {
			// Recalculate the new line number for every match
			for matchIdx, match := range m.searchResultLines {
				if match == lineIdx {
					m.searchResultLines[matchIdx] = newLineNb
				}
			}

			// Append the rendered line
			renderedLines := strings.Split(widthStyle.Render(string(line)), "\n")
			newLineNb += len(renderedLines)
			resultLines = append(resultLines, renderedLines...)
		}

		m.displayedContent = resultLines
	} else {
		m.searchResultLines = nil
		m.displayedContent = strings.Split(widthStyle.Render(string(m.rawOutput)), "\n")
	}

	if wasAtBottom || m.offset > m.maxOffset() {
		m.GoToBottom()
	}
}

func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height

	m.refreshDisplayedContent()
}

func (m *Model) maxOffset() int {
	// Manually add 2 to account for the borders
	return max(0, m.ContentHeight()-m.height+m.frameStyle.GetVerticalFrameSize()+2)
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
	return len(m.displayedContent)
}

func (m *Model) InnerFrameWidth() int {
	// Manually add 2 to account for the borders
	return max(m.width-m.frameStyle.GetHorizontalFrameSize()-2, 0)
}

func (m *Model) InnerFrameHeight() int {
	// Manually add 2 to account for the borders
	return max(m.height-m.SearchHeight()-m.frameStyle.GetVerticalFrameSize()-2, 0)
}

func (m *Model) SearchHeight() int {
	if !m.showSearch {
		return 0
	}

	return 2
}

func (m *Model) getVisibleLines(offset, height int) []string {
	if offset > len(m.displayedContent) || height <= 0 {
		return nil
	}
	if offset+height > len(m.displayedContent) {
		return m.displayedContent[offset:]
	}
	return m.displayedContent[offset : offset+height]
}

func (m *Model) setSearchVisibility(visible bool) {
	wasAtBottom := m.AtBottom()
	m.showSearch = visible
	m.searchHasFocus = false

	if wasAtBottom || m.offset > m.maxOffset() {
		m.GoToBottom()
	}
}

func (m *Model) updateSearchRegexp() {
	if len(m.currentSearch) == 0 {
		return
	}

	reg, err := regexp.Compile(string(m.currentSearch))
	if err != nil {
		m.searchRegexp = nil
		return
	}

	m.searchRegexp = reg
}

func (m *Model) executeSearch() {
	m.updateSearchRegexp()
	m.searchHasFocus = false

	// We have to do a double refresh: once to calculate the results,
	// then once to put the highlight at the right position
	m.refreshDisplayedContent()

	if len(m.searchResultLines) > 0 {
		// TODO: select the result closest to the current offset
		m.highlightedMatch = len(m.searchResultLines) - 1
		m.refreshDisplayedContent()
		m.scrollToSearchResult()
	}
}

func (m *Model) nextSearchResult() {
	if len(m.searchResultLines) == 0 {
		return
	}
	if m.highlightedMatch >= len(m.searchResultLines)-1 {
		m.highlightedMatch = 0
	} else {
		m.highlightedMatch++
	}

	m.refreshDisplayedContent()
	m.scrollToSearchResult()
}

func (m *Model) prevSearchResult() {
	if len(m.searchResultLines) == 0 {
		return
	}
	if m.highlightedMatch <= 0 {
		m.highlightedMatch = len(m.searchResultLines) - 1
	} else {
		m.highlightedMatch--
	}

	m.refreshDisplayedContent()
	m.scrollToSearchResult()
}

func (m *Model) scrollToSearchResult() {
	if m.searchResultLines[m.highlightedMatch] < m.maxOffset() {
		m.offset = m.searchResultLines[m.highlightedMatch]
	} else {
		m.GoToBottom()
	}
}

func (m *Model) clearSearchResults() {
	m.searchRegexp = nil
	m.searchResultLines = nil
	m.highlightedMatch = -1
}

func (m *Model) clearSearch() {
	m.currentSearch = nil
	m.clearSearchResults()
	m.setSearchVisibility(false)
}

func (m *Model) View() string {
	contentWidth := m.InnerFrameWidth()
	contentHeight := m.InnerFrameHeight()

	// Render the content
	content := lipgloss.NewStyle().
		Height(contentHeight).
		MaxHeight(contentHeight).
		Width(contentWidth).
		Render(strings.Join(m.getVisibleLines(m.offset, contentHeight), "\n"))

	if !m.showSearch {
		return m.frameStyle.Border(lipgloss.RoundedBorder(), true).Render(content)
	}

	contentBlock := m.frameStyle.Border(StackedVerticalBorderFirst, true, true, false, true).Render(content)

	resultBlock := ""
	if m.searchRegexp != nil {
		resultBlock = fmt.Sprintf(" %d / %d", len(m.searchResultLines)-m.highlightedMatch, len(m.searchResultLines))
	}

	// Render the search block
	searchBlockContent := lipgloss.NewStyle().
		Width(contentWidth - len(resultBlock)).
		Height(1).
		MaxHeight(3).
		Render(fmt.Sprintf("Search: %s", m.currentSearch))

	searchBlock := m.frameStyle.Border(StackedVerticalBorderLast, true).Render(searchBlockContent + resultBlock)

	return lipgloss.JoinVertical(lipgloss.Left, contentBlock, searchBlock)
}

func (m *Model) HandleKeyMsg(msg tea.KeyMsg) {
	if m.showSearch && m.searchHasFocus {
		// Search input manipulation
		switch msg.String() {
		case "enter":
			m.executeSearch()
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
		case "esc":
			m.clearSearch()
		default:
			if msg.Type == tea.KeyRunes {
				for _, r := range msg.Runes {
					m.currentSearch = utf8.AppendRune(m.currentSearch, r)
				}
			}
		}
	} else {
		switch msg.String() {
		case "up", "k":
			m.ScrollUp(3)

		case "down", "j":
			m.ScrollDown(3)

		case "pgup":
			m.PageUp()

		case "pgdown":
		case " ":
			m.PageDown()

		case "home":
			m.GoToTop()

		case "end":
			m.GoToBottom()

		case "enter", "n":
			if m.showSearch {
				m.nextSearchResult()
			}
		case "N":
			if m.showSearch {
				m.prevSearchResult()
			}
		case "/":
			if m.showSearch {
				m.clearSearchResults()
			} else {
				m.setSearchVisibility(true)
			}
			m.searchHasFocus = true
		case "esc":
			if m.showSearch {
				m.clearSearch()
			}
		}
	}
}
