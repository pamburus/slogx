// Package quotation provides quoting helpers for log messages and string values.
package quotation

import (
	"strings"
	"unicode/utf8"

	"github.com/pamburus/slogx/slogtext/internal/syntax"
)

// MessageContext returns the quoting context for log messages.
func MessageContext() Context {
	return Context{
		needed:           &messageQuoteSetV,
		neededForNumbers: false,
		extraCheck:       messageExtraCheck,
	}
}

// StringValueContext returns the quoting context for string values.
func StringValueContext() Context {
	return Context{
		needed:           &stringValueQuoteSetV,
		neededForNumbers: true,
		extraCheck:       stringValueExtraCheck,
	}
}

// Context represents the context of a string occurrence that is considered to be quoted.
type Context struct {
	needed           *[utf8.RuneSelf]bool
	neededForNumbers bool
	extraCheck       func(string) bool
}

// IsNeeded returns true if the string needs to be quoted.
func (c Context) IsNeeded(s string) bool {
	if !c.neededForNumbers {
		for _, r := range s {
			if c.isNeededForRune(r) {
				return true
			}
		}
	} else {
		looksLikeNumber := true
		nDots := 0

		for _, r := range s {
			switch {
			case c.isNeededForRune(r):
				return true
			case r == '.':
				nDots++
			case !isDigit(r):
				looksLikeNumber = false
			}
		}

		if looksLikeNumber && nDots <= 1 {
			return true
		}
	}

	return c.extraCheck(s)
}

func (c Context) isNeededForRune(r rune) bool {
	switch {
	case r == utf8.RuneError:
		return true
	case r >= utf8.RuneSelf:
		return false
	case c.needed[r]:
		return true
	default:
		return false
	}
}

// ---

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func messageQuoteSet() [utf8.RuneSelf]bool {
	needed := defaultQuoteSet()
	needed['='] = true
	needed[' '] = false

	return needed
}

func stringValueQuoteSet() [utf8.RuneSelf]bool {
	needed := defaultQuoteSet()
	needed['='] = true

	return needed
}

func defaultQuoteSet() [utf8.RuneSelf]bool {
	needed := [utf8.RuneSelf]bool{
		'"': true,
		'[': true,
		']': true,
		'{': true,
		'}': true,
	}
	for r := 0; r <= ' '; r++ {
		needed[r] = true
	}

	return needed
}

func stringValueExtraCheck(s string) bool {
	switch s {
	case "true", "false", "null", "...":
		return true
	}

	return false
}

func messageExtraCheck(s string) bool {
	n := len(s)

	if n < 3 {
		return false
	}

	switch s[0] {
	case ' ':
		return strings.HasPrefix(s, syntax.ExpandedKeyPrefix) || strings.HasPrefix(s, syntax.ExpandedValuePrefix)
	case '@':
		return s[1] == ' '
	}

	return false
}

var (
	messageQuoteSetV     = messageQuoteSet()
	stringValueQuoteSetV = stringValueQuoteSet()
)