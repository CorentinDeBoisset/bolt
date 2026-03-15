package outputviewer

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/corentindeboisset/tera/pkg/cmdrunr"
	"github.com/corentindeboisset/tera/pkg/iface"
	"github.com/corentindeboisset/tera/pkg/scrollbar"
)

var TopBlockBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "╭",
	TopRight:    "╮",
	BottomLeft:  "├",
	BottomRight: "┤",
}

var BottomBlockBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "├",
	TopRight:    "┤",
	BottomLeft:  "╰",
	BottomRight: "╯",
}

var LeftBottomBlockBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "├",
	TopRight:    "┬",
	BottomLeft:  "╰",
	BottomRight: "┴",
}

var RightBottomBlockBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "┬",
	TopRight:    "┤",
	BottomLeft:  "┴",
	BottomRight: "╯",
}

type Model struct {
	width  int
	height int

	buffer *cmdrunr.SafeBuffer

	theme            iface.Theme
	frameStyle       lipgloss.Style
	rawOutput        *cmdrunr.SafeBufferOut
	displayedContent []string
	offset           int

	searchBar         *SearchBarModel
	searchRegexp      *regexp.Regexp
	showSearch        bool
	searchHasFocus    bool
	searchResultLines []int
	highlightedMatch  int
}

func New(width, height int, theme iface.Theme, buffer *cmdrunr.SafeBuffer) (m Model) {
	return Model{
		width:  width,
		height: height,

		buffer: buffer,

		theme: theme,
		frameStyle: lipgloss.NewStyle().
			Padding(0, 2).
			BorderForeground(theme.BlurredOutputBorderColor),
		rawOutput:        nil,
		displayedContent: nil,
		offset:           0,

		searchBar:         &SearchBarModel{},
		showSearch:        false,
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

func (m *Model) SetBuffer(b *cmdrunr.SafeBuffer) {
	m.buffer = b
	m.rawOutput = nil
	m.displayedContent = nil

	m.clearSearch()
	m.RefreshContent()

	m.GoToBottom()
}

func (m *Model) SetFocus(focused bool) {
	if focused {
		m.frameStyle = m.frameStyle.BorderForeground(m.theme.FocusedOutputBorderColor)
	} else {
		m.frameStyle = m.frameStyle.BorderForeground(m.theme.BlurredOutputBorderColor)
	}

	m.searchBar.SetCursorVisibility(focused)
}

func (m *Model) RefreshContent() {
	if m.buffer == nil {
		return
	}

	var lastRead int64 = 0
	if m.rawOutput != nil {
		lastRead = m.rawOutput.WriteTime
	}

	if output := m.buffer.Content(lastRead); output != nil {
		m.rawOutput = output
		m.rawOutput.Content = bytes.ReplaceAll(m.rawOutput.Content, []byte("\r\n"), []byte("\n")) // Normalize line endings
		m.recomputeDisplayedContent()
	}
}

func (m *Model) recomputeDisplayedContent() {
	if m.rawOutput == nil {
		m.searchResultLines = nil
		m.displayedContent = nil
		return
	}

	wasAtBottom := m.AtBottom()
	widthStyle := lipgloss.NewStyle().Width(m.InnerFrameWidth())

	if m.searchRegexp != nil {
		newLineNb := 0
		decoratedOutput, searchResultLines := iface.DecorateCmdOutput(m.searchRegexp, m.rawOutput.Content, m.highlightedMatch, m.theme)
		splitDecoratedOutput := bytes.Split(decoratedOutput, []byte("\n"))
		resultLines := make([]string, 0)
		searchResultLinesAfterFormat := make([]int, len(searchResultLines))
		for lineIdx, line := range splitDecoratedOutput {
			// Recalculate the new line number for every match
			for matchIdx, match := range searchResultLines {
				if match == lineIdx {
					searchResultLinesAfterFormat[matchIdx] = newLineNb
				}
			}

			// Append the rendered line
			renderedLines := strings.Split(widthStyle.Render(string(line)), "\n")
			newLineNb += len(renderedLines)
			resultLines = append(resultLines, renderedLines...)
		}

		m.searchResultLines = searchResultLinesAfterFormat
		m.displayedContent = resultLines
	} else {
		m.searchResultLines = nil
		m.displayedContent = strings.Split(widthStyle.Render(string(m.rawOutput.Content)), "\n")
	}

	if wasAtBottom || m.offset > m.maxOffset() {
		m.GoToBottom()
	}
}

func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height

	m.recomputeDisplayedContent()
}

func (m *Model) maxOffset() int {
	innerFrameHeight := m.height - m.frameStyle.GetVerticalFrameSize()
	if m.showSearch {
		innerFrameHeight -= 2 // The search always has one line of content plus a border
	}
	return max(0, m.ContentHeight()-innerFrameHeight)
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
	if m.offset != m.maxOffset() {
		m.SetOffset(m.maxOffset())
	}
}

func (m *Model) GoToTop() {
	if m.offset != 0 {
		m.SetOffset(0)
	}
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
	m.searchHasFocus = visible
	m.searchBar.ToggleCursor(visible)

	if wasAtBottom || m.offset > m.maxOffset() {
		m.GoToBottom()
	}
}

func (m *Model) executeSearch() {
	m.searchRegexp = m.searchBar.Submit()
	m.searchHasFocus = false
	m.searchBar.ToggleCursor(false)

	// We have to do a double refresh: once to calculate the results,
	// then once to put the highlight at the right position
	m.recomputeDisplayedContent()

	if len(m.searchResultLines) > 0 {
		highlightCandidate := len(m.searchResultLines) - 1
		if !m.AtBottom() {
			// Find the result that is currently closest to the top of the screen
			for resultIdx, resultLine := range slices.Backward(m.searchResultLines) {
				if resultLine < m.offset {
					break
				}
				highlightCandidate = resultIdx
			}
		}
		m.highlightedMatch = highlightCandidate

		m.recomputeDisplayedContent()
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

	m.recomputeDisplayedContent()
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

	m.recomputeDisplayedContent()
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
	m.searchBar.Clear()
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

	offsetRatio := float64(m.offset) / float64(m.maxOffset())

	if !m.showSearch {
		frameAndBorderStyle := m.frameStyle.Border(lipgloss.RoundedBorder(), true, false, true, true)
		scrollBarContent := scrollbar.RenderScrollbar(len(m.displayedContent), contentHeight, offsetRatio, frameAndBorderStyle)

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			frameAndBorderStyle.Render(content),
			scrollBarContent,
		)
	}

	frameAndBorderStyle := m.frameStyle.Border(TopBlockBorder, true, false, false, true)
	scrollBarContent := scrollbar.RenderScrollbar(len(m.displayedContent), contentHeight, offsetRatio, frameAndBorderStyle)
	contentBlock := lipgloss.JoinHorizontal(
		lipgloss.Top,
		frameAndBorderStyle.Render(content),
		scrollBarContent,
	)

	var searchBlock string
	if m.searchRegexp != nil {
		resultBlock := fmt.Sprintf("%d / %d", m.highlightedMatch+1, len(m.searchResultLines))
		searchBlockContent := m.searchBar.View(contentWidth - len(resultBlock) - m.frameStyle.GetHorizontalPadding() - 1) // the one is to account for the inner border
		searchBlock = lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.frameStyle.Border(LeftBottomBlockBorder, true).Render(searchBlockContent),
			m.frameStyle.Border(RightBottomBlockBorder, true, true, true, false).Render(resultBlock),
		)
	} else {
		searchBlock = m.frameStyle.Border(BottomBlockBorder, true).Render(m.searchBar.View(contentWidth))
	}

	return lipgloss.JoinVertical(lipgloss.Left, contentBlock, searchBlock)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.showSearch && m.searchHasFocus {
			switch msg.String() {
			case "enter", "shift+enter":
				m.executeSearch()
			case "esc":
				m.clearSearch()
			default:
				// pass the input to the searchBar
				m.searchBar.HandleKeyMsg(msg)
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
			case "space":
				m.PageDown()

			case "home":
				m.GoToTop()

			case "end":
				m.GoToBottom()

			case "enter", "n":
				if m.showSearch {
					m.prevSearchResult()
				}
			case "shift+enter", "N":
				if m.showSearch {
					m.nextSearchResult()
				}
			case "/":
				if m.showSearch {
					m.clearSearchResults()
				} else {
					m.setSearchVisibility(true)
				}
				m.searchHasFocus = true
				m.searchBar.ToggleCursor(true)
			case "esc":
				if m.showSearch {
					m.clearSearch()
				}
			}
		}

	case tea.PasteMsg:
		if m.searchHasFocus {
			m.searchBar.InsertText(msg.String())
		}

	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.ScrollUp(3)
		case tea.MouseWheelDown:
			m.ScrollDown(3)
		}
	}

	return nil
}
