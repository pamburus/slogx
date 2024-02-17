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
		{
			Prefix: "\x1b[2m|\x1b[0;35m",
			Suffix: "\x1b[0;2m|\x1b[m",
		},
		{
			Prefix: "\x1b[2m|\x1b[0;36m",
			Suffix: "\x1b[0;2m|\x1b[m",
		},
		{
			Prefix: "\x1b[7;93m|",
			Suffix: "|\x1b[m",
		},
		{
			Prefix: "\x1b[7;91m|",
			Suffix: "|\x1b[m",
		},
	}
	levelValue := [NumLevels]ThemeItem{
		{
			Prefix: "\x1b[35m",
			Suffix: "\x1b[m",
		},
		{
			Prefix: "\x1b[36m",
			Suffix: "\x1b[m",
		},
		{
			Prefix: "\x1b[93m",
			Suffix: "\x1b[m",
		},
		{
			Prefix: "\x1b[91m",
			Suffix: "\x1b[m",
		},
	}

	brace := ThemeItem{Prefix: "\x1b[95m", Suffix: "\x1b[m"}
	sep := ThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[22m"}

	return Theme{
		Time: ThemeItem{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[m",
		},
		Level:      level,
		LevelValue: levelValue,
		Message: ThemeItem{
			Prefix: "\x1b[1m",
			Suffix: "\x1b[m",
		},
		Key: ThemeItem{
			Prefix: "\x1b[32m",
		},
		KeyValueSep: ThemeItem{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[m",
		},
		Source: ThemeItem{
			Prefix: "\x1b[2;3m@ ",
			Suffix: "\x1b[m",
		},
		StringQuote: ThemeItem{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[22m",
		},
		StringEscape: ThemeItem{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[22m",
		},
		BoolValue: ThemeItem{
			Prefix: "\x1b[36m",
			Suffix: "\x1b[m",
		},
		NumberValue: ThemeItem{
			Prefix: "\x1b[94m",
			Suffix: "\x1b[m",
		},
		NullValue: ThemeItem{
			Prefix: "\x1b[31m",
			Suffix: "\x1b[m",
		},
		ErrorValue: ThemeItem{
			Prefix: "\x1b[31m",
			Suffix: "\x1b[m",
		},
		DurationValue: ThemeItem{
			Prefix: "\x1b[94m",
			Suffix: "\x1b[m",
		},
		Array: ArrayThemeItem{
			Begin: brace,
			Sep:   sep,
			End:   brace,
		},
		Map: MapThemeItem{
			Begin:       brace,
			PairSep:     sep,
			KeyValueSep: sep,
			End:         brace,
		},
		Unresolved: UnresolvedThemeItem{
			Begin: ThemeItem{
				Prefix: "\x1b[31;2m",
				Suffix: "\x1b[22m",
			},
			End: ThemeItem{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[m",
			},
		},
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
		Time: ThemeItem{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[m",
		},
		Level:      level,
		LevelValue: level,
		Key: ThemeItem{
			Prefix: "\x1b[2m",
		},
		KeyValueSep: ThemeItem{
			Suffix: "\x1b[m",
		},
		Source: ThemeItem{
			Prefix: "\x1b[2;3m@ ",
			Suffix: "\x1b[m",
		},
		StringEscape: ThemeItem{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[22m",
		},
		ErrorValue: ThemeItem{
			Prefix: "\x1b[31m",
			Suffix: "\x1b[m",
		},
		Unresolved: UnresolvedThemeItem{
			Begin: ThemeItem{
				Prefix: "\x1b[31;2m",
				Suffix: "\x1b[22m",
			},
			End: ThemeItem{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[m",
			},
		},
	}
}

// Theme is a theme for the Handler.
type Theme struct {
	Time          ThemeItem
	Level         LevelThemeItem
	Message       ThemeItem
	Key           ThemeItem
	KeyValueSep   ThemeItem
	Source        ThemeItem
	TimeValue     ThemeItem
	LevelValue    LevelThemeItem
	StringValue   ThemeItem
	StringQuote   ThemeItem
	StringEscape  ThemeItem
	BoolValue     ThemeItem
	NumberValue   ThemeItem
	NullValue     ThemeItem
	ErrorValue    ThemeItem
	DurationValue ThemeItem
	Array         ArrayThemeItem
	Map           MapThemeItem
	Unresolved    UnresolvedThemeItem
}

// Plain returns a theme with no color.
func (t Theme) Plain() Theme {
	return Theme{
		Time:          t.Time.Plain(),
		Level:         t.Level.Plain(),
		Message:       t.Message.Plain(),
		Key:           t.Key.Plain(),
		KeyValueSep:   t.KeyValueSep.Plain(),
		Source:        t.Source.Plain(),
		TimeValue:     t.TimeValue.Plain(),
		LevelValue:    t.LevelValue.Plain(),
		StringValue:   t.StringValue.Plain(),
		StringQuote:   t.StringQuote.Plain(),
		StringEscape:  t.StringEscape.Plain(),
		BoolValue:     t.BoolValue.Plain(),
		NumberValue:   t.NumberValue.Plain(),
		NullValue:     t.NullValue.Plain(),
		ErrorValue:    t.ErrorValue.Plain(),
		DurationValue: t.DurationValue.Plain(),
		Array:         t.Array.Plain(),
		Map:           t.Map.Plain(),
		Unresolved:    t.Unresolved.Plain(),
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

// ---

// LevelThemeItem is a theme item set per logging level.
type LevelThemeItem [NumLevels]ThemeItem

// Plain returns a theme item with no color.
func (i LevelThemeItem) Plain() LevelThemeItem {
	return LevelThemeItem{
		i[0].Plain(),
		i[1].Plain(),
		i[2].Plain(),
		i[3].Plain(),
	}
}

// ---

// ArrayThemeItem is a theme item for an array type.
type ArrayThemeItem struct {
	Begin ThemeItem
	Sep   ThemeItem
	End   ThemeItem
}

// Plain returns a copy of i with no color.
func (i ArrayThemeItem) Plain() ArrayThemeItem {
	return ArrayThemeItem{
		Begin: i.Begin.Plain(),
		Sep:   i.Sep.Plain(),
		End:   i.End.Plain(),
	}
}

// ---

// ThemeItem is a theme item for an map type.
type MapThemeItem struct {
	Begin       ThemeItem
	PairSep     ThemeItem
	KeyValueSep ThemeItem
	End         ThemeItem
}

// Plain returns a copy of i with no color.
func (i MapThemeItem) Plain() MapThemeItem {
	return MapThemeItem{
		Begin:       i.Begin.Plain(),
		PairSep:     i.PairSep.Plain(),
		KeyValueSep: i.KeyValueSep.Plain(),
		End:         i.End.Plain(),
	}
}

// ---

// UnresolvedThemeItem is a theme item for a composite type representation that can have a begin, 1 or 2 separators, and an end.
type UnresolvedThemeItem struct {
	Begin ThemeItem
	End   ThemeItem
}

// Plain returns a copy of i with no color.
func (i UnresolvedThemeItem) Plain() UnresolvedThemeItem {
	return UnresolvedThemeItem{
		Begin: i.Begin.Plain(),
		End:   i.End.Plain(),
	}
}
