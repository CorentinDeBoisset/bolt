package outputviewer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	t.Parallel()

	// No-op
	assert.Equal(t, sanitize("abcd"), []rune("abcd"))

	// Clear random control sequence
	assert.Equal(t, sanitize("ab\x06cd"), []rune("abcd"))

	// Sanitize newlines
	assert.Equal(t, sanitize("\rab\r\r\ncd\r\n"), []rune("\\nab\\n\\ncd\\n"))

	// Sanitize tabs
	assert.Equal(t, sanitize("ab\tcd"), []rune("ab\\tcd"))

	// Manage emojis
	assert.Equal(t, sanitize("ab 🥹 c 👯‍♀️ d"), []rune("ab 🥹 c 👯‍♀️ d"))
}
