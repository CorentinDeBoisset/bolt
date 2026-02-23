package cmdrunr

import (
	"regexp"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func TestPrepareSequence(t *testing.T) {
	t.Parallel()

	// empty content
	input := []byte("")
	assert.Len(t, prepareSequence(input), 0)

	// simple string
	input = []byte("abcd")
	expected := []sequence{
		{[]byte("a"), 1, true},
		{[]byte("b"), 1, true},
		{[]byte("c"), 1, true},
		{[]byte("d"), 1, true},
	}
	assert.Equal(t, prepareSequence(input), expected)

	// String with escape codes and emojis
	input = []byte("\x1b[31mHel\x1b[31mlo, world 🎉\x1b[0m")
	expected = []sequence{
		{[]byte("\x1b[31m"), 5, false},
		{[]byte("H"), 1, true},
		{[]byte("e"), 1, true},
		{[]byte("l"), 1, true},
		{[]byte("\x1b[31m"), 5, false},
		{[]byte("l"), 1, true},
		{[]byte("o"), 1, true},
		{[]byte(","), 1, true},
		{[]byte(" "), 1, true},
		{[]byte("w"), 1, true},
		{[]byte("o"), 1, true},
		{[]byte("r"), 1, true},
		{[]byte("l"), 1, true},
		{[]byte("d"), 1, true},
		{[]byte(" "), 1, true},
		{[]byte("🎉"), 4, true},
		{[]byte("\x1b[0m"), 4, false},
	}
	assert.Equal(t, prepareSequence(input), expected)
}

func TestFindOutput(t *testing.T) {
	t.Parallel()

	// Input with escape sequences
	input := prepareSequence([]byte("\x1b[31mHel\x1b[31mlo, bellissimo mundo! 🎉\x1b[0m"))

	re := regexp.MustCompile("ll.")
	assert.Equal(t, findAllInSequence(re, input), [][]int{{7, 15}, {19, 22}})

	// Input without any escape sequence, but with emojis
	input = prepareSequence([]byte("Hello 💥 bellissimo mundo!"))

	re = regexp.MustCompile("ll.")
	assert.Equal(t, findAllInSequence(re, input), [][]int{{2, 5}, {13, 16}})

	// non-matching regexp
	input = prepareSequence([]byte("Hello world!"))

	re = regexp.MustCompile("xyz")
	assert.Nil(t, findAllInSequence(re, input))
}

func TestDecorateOutput(t *testing.T) {
	t.Parallel()

	// Force colors
	lipgloss.SetColorProfile(termenv.TrueColor)

	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("1"))

	// Non-matching input
	input := []byte("super")
	re := regexp.MustCompile("xyz")
	expected := "super"
	assert.Equal(t, string(DecorateCmdOutput(re, input, highlightStyle)), expected)

	// Simple input
	input = []byte("Super content")
	re = regexp.MustCompile("e.")
	expected = "Sup\x1b[41me\x1b[0m\x1b[41mr\x1b[0m cont\x1b[41me\x1b[0m\x1b[41mn\x1b[0mt"
	assert.Equal(t, string(DecorateCmdOutput(re, input, highlightStyle)), string(expected))

	// Complex input
	input = []byte("\x1b[31mHel\x1b[31mlo, bellissimo mundo! 🎉\x1b[0m")
	re = regexp.MustCompile("ll.")
	expected = "\x1b[31mHe\x1b[41ml\x1b[0m\x1b[31m\x1b[41ml\x1b[0m\x1b[41mo\x1b[0m, be\x1b[41ml\x1b[0m\x1b[41ml\x1b[0m\x1b[41mi\x1b[0mssimo mundo! 🎉\x1b[0m"
	assert.Equal(t, string(DecorateCmdOutput(re, input, highlightStyle)), expected)
}
