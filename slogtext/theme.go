package slogtext

import (
	"strings"

	"github.com/pamburus/slogx/internal/stripansi"
)

// NumLevels is the number of levels.
const NumLevels = 4

// ThemeDefault returns a theme with default colors.
func ThemeDefault() Theme {
	level := [NumLevels]ThemeItem{
		{Prefix: "\x1b[2m|\x1b[0;35m", Suffix: "\x1b[0;2m|\x1b[m"},
		{Prefix: "\x1b[2m|\x1b[0;36m", Suffix: "\x1b[0;2m|\x1b[m"},
		{Prefix: "\x1b[2m|\x1b[0;93m", Suffix: "\x1b[0;2m|\x1b[m"},
		{Prefix: "\x1b[2m|\x1b[0;91m", Suffix: "\x1b[0;2m|\x1b[m"},
	}
	levelValue := [NumLevels]ThemeItem{
		{Prefix: "\x1b[35m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[36m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[93m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[91m", Suffix: "\x1b[m"},
	}
	composite := func(i ThemeItem) ThemeCompositeItem {
		return ThemeCompositeItem{Begin: i, Sep1: i, Sep2: i, End: i}
	}

	return Theme{
		Timestamp:  ThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[m"},
		Level:      level,
		LevelValue: levelValue,
		Message:    ThemeItem{Prefix: "\x1b[1m", Suffix: "\x1b[m"},
		Key:        ThemeItem{Prefix: "\x1b[32m"},
		EqualSign:  ThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[m"},
		Source:     ThemeItem{Prefix: "\x1b[2;3m@ ", Suffix: "\x1b[m"},
		Quote:      ThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[22m"},
		Escape:     ThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[22m"},
		Bool:       ThemeItem{Prefix: "\x1b[36m", Suffix: "\x1b[m"},
		Number:     ThemeItem{Prefix: "\x1b[94m", Suffix: "\x1b[m"},
		Null:       ThemeItem{Prefix: "\x1b[31m", Suffix: "\x1b[m"},
		Error:      ThemeItem{Prefix: "\x1b[31m", Suffix: "\x1b[m"},
		Duration:   ThemeItem{Prefix: "\x1b[94m", Suffix: "\x1b[m"},
		Array:      composite(ThemeItem{Prefix: "\x1b[95m", Suffix: "\x1b[m"}),
		Map:        composite(ThemeItem{Prefix: "\x1b[95m", Suffix: "\x1b[m"}),
		EvalError:  ThemeItem{Prefix: "\x1b[31;2m$!(ERROR: \x1b[22m", Suffix: "\x1b[2m)\x1b[m"},
		EvalPanic:  ThemeItem{Prefix: "\x1b[31;2m$!(PANIC: \x1b[22m", Suffix: "\x1b[2m)\x1b[m"},
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
	level := [NumLevels]ThemeItem{
		{},
		{Prefix: "\x1b[32m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[93m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[91m", Suffix: "\x1b[m"},
	}

	return Theme{
		Timestamp:  ThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[m"},
		Level:      level,
		LevelValue: level,
		Key:        ThemeItem{Prefix: "\x1b[2m"},
		EqualSign:  ThemeItem{Suffix: "\x1b[m"},
		Source:     ThemeItem{Prefix: "\x1b[2;3m@ ", Suffix: "\x1b[m"},
		Escape:     ThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[22m"},
		Error:      ThemeItem{Prefix: "\x1b[31m", Suffix: "\x1b[m"},
		EvalError:  ThemeItem{Prefix: "\x1b[31;2m$!(ERROR: \x1b[22m", Suffix: "\x1b[2m)\x1b[m"},
		EvalPanic:  ThemeItem{Prefix: "\x1b[31;2m$!(PANIC: \x1b[22m", Suffix: "\x1b[2m)\x1b[m"},
	}
}

// Theme is a theme for the Handler.
type Theme struct {
	Timestamp  ThemeItem
	Level      [NumLevels]ThemeItem
	LevelValue [NumLevels]ThemeItem
	Message    ThemeItem
	Key        ThemeItem
	EqualSign  ThemeItem
	Source     ThemeItem
	String     ThemeItem
	Quote      ThemeItem
	Escape     ThemeItem
	Bool       ThemeItem
	Number     ThemeItem
	Null       ThemeItem
	Error      ThemeItem
	Duration   ThemeItem
	Time       ThemeItem
	Array      ThemeCompositeItem
	Map        ThemeCompositeItem
	EvalError  ThemeItem
	EvalPanic  ThemeItem
}

// Plain returns a theme with no color.
func (t Theme) Plain() Theme {
	return Theme{
		Timestamp: t.Timestamp.Plain(),
		Level:     [NumLevels]ThemeItem{t.Level[0].Plain(), t.Level[1].Plain(), t.Level[2].Plain(), t.Level[3].Plain()},
		Message:   t.Message.Plain(),
		Key:       t.Key.Plain(),
		EqualSign: t.EqualSign.Plain(),
		Source:    t.Source.Plain(),
		String:    t.String.Plain(),
		Quote:     t.Quote.Plain(),
		Escape:    t.Escape.Plain(),
		Bool:      t.Bool.Plain(),
		Number:    t.Number.Plain(),
		Null:      t.Null.Plain(),
		Error:     t.Error.Plain(),
		Duration:  t.Duration.Plain(),
		Time:      t.Time.Plain(),
		Array:     t.Array.Plain(),
		Map:       t.Map.Plain(),
		EvalError: t.EvalError.Plain(),
		EvalPanic: t.EvalPanic.Plain(),
	}
}

// ---

// ThemeCompositeItem is a theme item for a composite type representation that can have a begin, 1 or 2 separators, and an end.
type ThemeCompositeItem struct {
	Begin ThemeItem
	Sep1  ThemeItem
	Sep2  ThemeItem
	End   ThemeItem
}

// Plain returns a theme composite item with no color.
func (i ThemeCompositeItem) Plain() ThemeCompositeItem {
	return ThemeCompositeItem{
		Begin: i.Begin.Plain(),
		Sep1:  i.Sep1.Plain(),
		Sep2:  i.Sep2.Plain(),
		End:   i.End.Plain(),
	}
}

// ---

// ThemeItem is a theme item that can have a prefix and a suffix.
type ThemeItem struct {
	Prefix string
	Suffix string
}

// IsEmpty returns true if the theme item is empty.
func (i ThemeItem) IsEmpty() bool {
	return i.Prefix == "" && i.Suffix == ""
}

// Plain returns a theme item with no color.
func (i ThemeItem) Plain() ThemeItem {
	return ThemeItem{
		Prefix: stripansi.Strip(i.Prefix),
		Suffix: stripansi.Strip(i.Suffix),
	}
}
