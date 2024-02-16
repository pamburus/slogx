package slogtext

import "github.com/pamburus/slogx/internal/stripansi"

// ThemeDefault returns a theme with default colors.
func ThemeDefault() Theme {
	level := [4]FixedThemeItem{
		{Text: "DBG"},
		{Text: "\x1b[92mINF\x1b[m"},
		{Text: "\x1b[93mWRN\x1b[m"},
		{Text: "\x1b[91mERR\x1b[m"},
	}

	return Theme{
		Timestamp: VariableThemeItem{Prefix: "\x1b[2m", Suffix: "\x1b[m"},
		Level:     level,
		Message:   VariableThemeItem{},
		Key:       VariableThemeItem{Prefix: "\x1b[2m"},
		EqualSign: VariableThemeItem{Suffix: "\x1b[m"},
		Source:    VariableThemeItem{Prefix: "\x1b[2;3m@ ", Suffix: "\x1b[m"},
		String:    VariableThemeItem{},
		Bool:      VariableThemeItem{},
		Number:    VariableThemeItem{},
		Null:      VariableThemeItem{},
	}
}

// Theme is a theme for the Handler.
type Theme struct {
	Timestamp VariableThemeItem
	Level     [4]FixedThemeItem
	Message   VariableThemeItem
	Key       VariableThemeItem
	EqualSign VariableThemeItem
	Source    VariableThemeItem
	String    VariableThemeItem
	Bool      VariableThemeItem
	Number    VariableThemeItem
	Null      VariableThemeItem
}

// Plain returns a theme with no color.
func (t Theme) Plain() Theme {
	return Theme{
		Timestamp: t.Timestamp.Plain(),
		Level:     [4]FixedThemeItem{t.Level[0].Plain(), t.Level[1].Plain(), t.Level[2].Plain(), t.Level[3].Plain()},
		Message:   t.Message.Plain(),
		Key:       t.Key.Plain(),
		Source:    t.Source.Plain(),
		String:    t.String.Plain(),
		Bool:      t.Bool.Plain(),
		Number:    t.Number.Plain(),
		Null:      t.Null.Plain(),
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

// Plain returns a theme item with no color.
func (i VariableThemeItem) Plain() VariableThemeItem {
	return VariableThemeItem{
		Prefix: stripansi.Strip(i.Prefix),
		Suffix: stripansi.Strip(i.Suffix),
	}
}
