package themes

import "strings"

// Fancy returns a variant of default theme with fancy unicode decorations.
func Fancy() Theme {
	theme := Default()

	tuneLevel := func(s *string) {
		*s = strings.ReplaceAll(*s, "|", "│")
	}

	for i := range theme.Level {
		tuneLevel(&theme.Level[i].Prefix)
		tuneLevel(&theme.Level[i].Suffix)
	}

	theme.Source.Prefix = strings.ReplaceAll(theme.Source.Prefix, "@", "→")
	theme.ExpandedKey.Prefix = strings.ReplaceAll(theme.ExpandedKey.Prefix, ">", "•")

	return theme
}
