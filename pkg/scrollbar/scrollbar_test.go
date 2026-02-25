package scrollbar

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestScrollbar(t *testing.T) {
	t.Parallel()

	frameStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)

	// Standard scrollbar
	bar := RenderScrollbar(100, 10, 0.25, frameStyle)
	expected := `╮
│
│
╦
║
╩
│
│
│
│
│
╯`
	assert.Equal(t, expected, bar)

	// Standard scrollbar at the bottom
	bar = RenderScrollbar(40, 10, 1., frameStyle)
	expected = `╮
│
│
│
│
│
│
│
╦
║
╩
╯`
	assert.Equal(t, expected, bar)

	// Standard scrollbar at the top
	bar = RenderScrollbar(40, 10, 0., frameStyle)
	expected = `╮
╦
║
╩
│
│
│
│
│
│
│
╯`
	assert.Equal(t, expected, bar)

	// Too small window
	bar = RenderScrollbar(40, 3, 0.5, frameStyle)
	expected = `╮
│
│
│
╯`
	assert.Equal(t, expected, bar)

	// Too little scroll
	bar = RenderScrollbar(7, 7, 0., frameStyle)
	expected = `╮
│
│
│
│
│
│
│
╯`
	assert.Equal(t, expected, bar)
}
