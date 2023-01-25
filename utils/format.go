package utils

import (
	"strings"

	"github.com/stablecog/go-apps/shared"
)

// RemoveRedundantSpaces removes all redundant spaces from a string
// e.g. "  hello   world  " -> " hello world "
func RemoveRedundantSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// RemoveLineBreaks removes all line breaks from a string
// e.g. "hello\nworld" -> "hello world"
func RemoveLineBreaks(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

// FormatPrompt applies formatting to a prompt string
// e.g. "  hello   world  " -> "hello world"
func FormatPrompt(s string) string {
	cleanStr := RemoveRedundantSpaces(RemoveLineBreaks(s))
	if len(cleanStr) > shared.MAX_PROMPT_LENGTH {
		cleanStr = cleanStr[:shared.MAX_PROMPT_LENGTH]
	}
	return cleanStr
}
