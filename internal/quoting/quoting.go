package quoting

import (
	"unicode/utf8"
)

// MessageContext returns the quoting context for log messages.
func MessageContext() Context {
	return Context{
		needed:           &messageQuoteSetV,
		neededForNumbers: false,
	}
}

// StringValueContext returns the quoting context for string values.
func StringValueContext() Context {
	return Context{
		needed:           &stringValueQuoteSetV,
		neededForNumbers: true,
	}
}

// Context represents the context of a string occurrence that is considered to be quoted.
type Context struct {
	needed           *[utf8.RuneSelf]bool
	neededForNumbers bool
}

// IsNeeded returns true if the string needs to be quoted.
func (c Context) IsNeeded(s string) bool {
	if !c.neededForNumbers {
		for _, r := range s {
			if c.isNeededForRune(r) {
				return true
			}
		}

		return false
	}

	looksLikeNumber := true
	nDots := 0

	for _, r := range s {
		if c.isNeededForRune(r) {
			return true
		}
		if r == '.' {
			nDots++
		} else if !isDigit(r) {
			looksLikeNumber = false
		}
	}

	return looksLikeNumber && nDots <= 1
}

// IsNeededBytes returns true if the string needs to be quoted.
func (c Context) IsNeededBytes(s []byte) bool {
	if !c.neededForNumbers {
		for i := 0; i < len(s); {
			r, width := utf8.DecodeRune(s[i:])
			i += width
			if c.isNeededForRune(r) {
				return true
			}
		}

		return false
	}

	looksLikeNumber := true
	nDots := 0

	for i := 0; i < len(s); {
		r, width := utf8.DecodeRune(s[i:])
		i += width
		if c.isNeededForRune(r) {
			return true
		}
		if r == '.' {
			nDots++
		} else if !isDigit(r) {
			looksLikeNumber = false
		}
	}

	return looksLikeNumber && nDots <= 1
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
		',': true,
	}
	for r := 0; r <= ' '; r++ {
		needed[r] = true
	}

	return needed
}

var (
	messageQuoteSetV     = messageQuoteSet()
	stringValueQuoteSetV = stringValueQuoteSet()
)
