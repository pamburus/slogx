package slogtext

// ThemePlain returns a theme with no color.
func ThemePlain() Theme {
	return Theme{
		Key:    VariableThemeItem{Suffix: "="},
		Source: VariableThemeItem{Prefix: "@ "},
	}
}

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
		Key:       VariableThemeItem{Prefix: "\x1b[2m", Suffix: "=\x1b[m"},
		Source:    VariableThemeItem{Prefix: "\x1b[2;3m@ ", Suffix: "\x1b[m"},
		String:    VariableThemeItem{},
		Bool:      VariableThemeItem{},
		Number:    VariableThemeItem{},
		Null:      VariableThemeItem{},
	}
}

type Theme struct {
	Timestamp VariableThemeItem
	Level     [4]FixedThemeItem
	Message   VariableThemeItem
	Key       VariableThemeItem
	Source    VariableThemeItem
	String    VariableThemeItem
	Bool      VariableThemeItem
	Number    VariableThemeItem
	Null      VariableThemeItem
}

type FixedThemeItem struct {
	Text string
}

type VariableThemeItem struct {
	Prefix string
	Suffix string
}
