package themes

import "strings"

// Fancy returns a variant of default theme with fancy unicode decorations.
func Fancy() Theme {
	theme := Default()

	theme.Level = LevelItem{
		{
			Prefix: "\x1b[2m│\x1b[0;35m",
			Suffix: "\x1b[0;2m│\x1b[m",
		},
		{
			Prefix: "\x1b[2m│\x1b[0;36m",
			Suffix: "\x1b[0;2m│\x1b[m",
		},
		{
			Prefix: "\x1b[7;93m│",
			Suffix: "│\x1b[m",
		},
		{
			Prefix: "\x1b[7;90;31m│",
			Suffix: "│\x1b[m",
		},
	}

	theme.Source.Prefix = strings.ReplaceAll(theme.Source.Prefix, "@", "→")

	return theme
}
