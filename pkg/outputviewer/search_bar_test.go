package outputviewer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	t.Parallel()

	// No-op
	assert.Equal(t, sanitize([]rune("abcd")), []rune("abcd"))

	// Clear random control sequence
	assert.Equal(t, sanitize([]rune("ab\x06cd")), []rune("abcd"))

	// Sanitize newlines
	assert.Equal(t, sanitize([]rune("\rab\r\r\ncd\r\n")), []rune("\\nab\\n\\ncd\\n"))

	// Sanitize tabs
	assert.Equal(t, sanitize([]rune("ab\tcd")), []rune("ab\\tcd"))

	// Manage emojis
	assert.Equal(t, sanitize([]rune("ab 🥹 c 👯‍♀️ d")), []rune("ab 🥹 c 👯‍♀️ d"))
}
