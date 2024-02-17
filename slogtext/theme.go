package slogtext

import (
	"strings"

	"github.com/pamburus/slogx/internal/stripansi"
)

// NumLevels is the number of levels.
const NumLevels = 4

// ThemeDefault returns a theme with default colors.
func ThemeDefault() Theme {
	level := [NumLevels]VariableThemeItem{
		{Prefix: "\x1b[2m|\x1b[0;35m", Suffix: "\x1b[0;2m|\x1b[m"},
		{Prefix: "\x1b[2m|\x1b[0;36m", Suffix: "\x1b[0;2m|\x1b[m"},
		{Prefix: "\x1b[2m|\x1b[0;93m", Suffix: "\x1b[0;2m|\x1b[m"},
		{Prefix: "\x1b[2m|\x1b[0;91m", Suffix: "\x1b[0;2m|\x1b[m"},
	}
	levelValue := [NumLevels]VariableThemeItem{
		{Prefix: "\x1b[35m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[36m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[93m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[91m", Suffix: "\x1b[m"},
	}

	return Theme{
		Timestamp:   VariableThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[m"},
		Level:       level,
		LevelValue:  levelValue,
		Message:     VariableThemeItem{Prefix: "\x1b[1m", Suffix: "\x1b[m"},
		Key:         VariableThemeItem{Prefix: "\x1b[32m"},
		EqualSign:   VariableThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[m"},
		Source:      VariableThemeItem{Prefix: "\x1b[2;3m@ ", Suffix: "\x1b[m"},
		Quote:       VariableThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[22m"},
		Escape:      VariableThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[22m"},
		Bool:        VariableThemeItem{Prefix: "\x1b[36m", Suffix: "\x1b[m"},
		Number:      VariableThemeItem{Prefix: "\x1b[94m", Suffix: "\x1b[m"},
		Null:        VariableThemeItem{Prefix: "\x1b[31m", Suffix: "\x1b[m"},
		Error:       VariableThemeItem{Prefix: "\x1b[31m", Suffix: "\x1b[m"},
		Duration:    VariableThemeItem{Prefix: "\x1b[94m", Suffix: "\x1b[m"},
		ArrayBegin:  VariableThemeItem{Prefix: "\x1b[95m", Suffix: "\x1b[m"},
		ArrayEnd:    VariableThemeItem{Prefix: "\x1b[95m", Suffix: "\x1b[m"},
		ArraySep:    VariableThemeItem{Prefix: "\x1b[95m", Suffix: "\x1b[m"},
		EncodeError: VariableThemeItem{Prefix: "\x1b[31;2m$!(ERROR: \x1b[22m", Suffix: "\x1b[2m)\x1b[m"},
		EncodePanic: VariableThemeItem{Prefix: "\x1b[31;2m$!(PANIC: \x1b[22m", Suffix: "\x1b[2m)\x1b[m"},
	}
}

// ThemeFancy returns a variant of default theme with fancy characters.
func ThemeFancy() Theme {
	theme := ThemeDefault()

	tuneLevel := func(s *string) {
		*s = strings.ReplaceAll(*s, "|", "│")
	}

	for i := range theme.Level {
		tuneLevel(&theme.Level[i].Prefix)
		tuneLevel(&theme.Level[i].Suffix)
	}

	theme.Source.Prefix = strings.ReplaceAll(theme.Source.Prefix, "@", "→")

	return theme
}

// ThemeTint returns a theme with emulation of color scheme of [tint](https://github.com/lmittmann/tint) package.
func ThemeTint() Theme {
	level := [NumLevels]VariableThemeItem{
		{},
		{Prefix: "\x1b[32m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[93m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[91m", Suffix: "\x1b[m"},
	}

	return Theme{
		Timestamp:   VariableThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[m"},
		Level:       level,
		LevelValue:  level,
		Key:         VariableThemeItem{Prefix: "\x1b[2m"},
		EqualSign:   VariableThemeItem{Suffix: "\x1b[m"},
		Source:      VariableThemeItem{Prefix: "\x1b[2;3m@ ", Suffix: "\x1b[m"},
		Escape:      VariableThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[22m"},
		Error:       VariableThemeItem{Prefix: "\x1b[31m", Suffix: "\x1b[m"},
		EncodeError: VariableThemeItem{Prefix: "\x1b[31;2m$!(ERROR: \x1b[22m", Suffix: "\x1b[2m)\x1b[m"},
		EncodePanic: VariableThemeItem{Prefix: "\x1b[31;2m$!(PANIC: \x1b[22m", Suffix: "\x1b[2m)\x1b[m"},
	}
}

// Theme is a theme for the Handler.
type Theme struct {
	Timestamp   VariableThemeItem
	Level       [NumLevels]VariableThemeItem
	LevelValue  [NumLevels]VariableThemeItem
	Message     VariableThemeItem
	Key         VariableThemeItem
	EqualSign   VariableThemeItem
	Source      VariableThemeItem
	String      VariableThemeItem
	Quote       VariableThemeItem
	Escape      VariableThemeItem
	Bool        VariableThemeItem
	Number      VariableThemeItem
	Null        VariableThemeItem
	Error       VariableThemeItem
	Duration    VariableThemeItem
	Time        VariableThemeItem
	ArrayBegin  VariableThemeItem
	ArrayEnd    VariableThemeItem
	ArraySep    VariableThemeItem
	EncodeError VariableThemeItem
	EncodePanic VariableThemeItem
}

// Plain returns a theme with no color.
func (t Theme) Plain() Theme {
	return Theme{
		Timestamp:   t.Timestamp.Plain(),
		Level:       [NumLevels]VariableThemeItem{t.Level[0].Plain(), t.Level[1].Plain(), t.Level[2].Plain(), t.Level[3].Plain()},
		Message:     t.Message.Plain(),
		Key:         t.Key.Plain(),
		EqualSign:   t.EqualSign.Plain(),
		Source:      t.Source.Plain(),
		String:      t.String.Plain(),
		Quote:       t.Quote.Plain(),
		Escape:      t.Escape.Plain(),
		Bool:        t.Bool.Plain(),
		Number:      t.Number.Plain(),
		Null:        t.Null.Plain(),
		Error:       t.Error.Plain(),
		Duration:    t.Duration.Plain(),
		Time:        t.Time.Plain(),
		ArrayBegin:  t.ArrayBegin.Plain(),
		ArrayEnd:    t.ArrayEnd.Plain(),
		ArraySep:    t.ArraySep.Plain(),
		EncodeError: t.EncodeError.Plain(),
		EncodePanic: t.EncodePanic.Plain(),
	}
}

// ---

// FixedThemeItem is a theme item with a static content.
type FixedThemeItem struct {
	Text string
}

// Plain returns a theme item with no color.
func (i FixedThemeItem) Plain() FixedThemeItem {
	return FixedThemeItem{
		Text: stripansi.Strip(i.Text),
	}
}

// ---

// VariableThemeItem is a theme item that can have a prefix and a suffix.
type VariableThemeItem struct {
	Prefix string
	Suffix string
}

// IsEmpty returns true if the theme item is empty.
func (i VariableThemeItem) IsEmpty() bool {
	return i.Prefix == "" && i.Suffix == ""
}

// Plain returns a theme item with no color.
func (i VariableThemeItem) Plain() VariableThemeItem {
	return VariableThemeItem{
		Prefix: stripansi.Strip(i.Prefix),
		Suffix: stripansi.Strip(i.Suffix),
	}
}
