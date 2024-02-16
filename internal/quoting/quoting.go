package quoting

import "unicode/utf8"

// IsNeeded returns true if the string needs to be quoted.
func IsNeeded(s string) bool {
	looksLikeNumber := true
	nDots := 0

	for _, r := range s {
		switch r {
		case '.':
			nDots++
		case '=', '"', ' ', utf8.RuneError:
			return true
		default:
			if r < ' ' {
				return true
			}
			if !isDigit(r) {
				looksLikeNumber = false
			}
		}
	}

	return looksLikeNumber && nDots <= 1
}

// IsNeededForBytes returns true if the string needs to be quoted.
func IsNeededForBytes(s []byte) bool {
	looksLikeNumber := true
	nDots := 0

	for i := 0; i < len(s); {
		r, width := utf8.DecodeRune(s[i:])
		i += width
		switch r {
		case '.':
			nDots++
		case '=', '"', ' ', utf8.RuneError:
			return true
		default:
			if r < ' ' {
				return true
			}
			if !isDigit(r) {
				looksLikeNumber = false
			}
		}
	}

	return looksLikeNumber && nDots <= 1
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
