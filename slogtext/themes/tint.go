package themes

// Tint returns a theme with emulation of color scheme of [tint](https://github.com/lmittmann/tint) package.
func Tint() Theme {
	level := [NumLevels]Item{
		{},
		{Prefix: "\x1b[32m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[93m", Suffix: "\x1b[m"},
		{Prefix: "\x1b[91m", Suffix: "\x1b[m"},
	}

	return Theme{
		Time: Item{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[m",
		},
		Logger: Item{
			Prefix: "\x1b[2m",
			Suffix: "",
		},
		LoggerMessageSep: Item{
			Suffix: "\x1b[m",
		},
		Message: StringItem{
			Escape: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
			},
			Elipsis: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
			},
		},
		Level:      level,
		LevelValue: level,
		Key: Item{
			Prefix: "\x1b[2m",
		},
		KeyValueSep: Item{
			Suffix: "\x1b[m",
		},
		Source: Item{
			Prefix: "\x1b[2;3m@ ",
			Suffix: "\x1b[m",
		},
		StringValue: StringItem{
			Escape: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
			},
			Elipsis: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
			},
		},
		ErrorValue: StringItem{
			Whole: Item{
				Prefix: "\x1b[31m",
				Suffix: "\x1b[m",
			},
			Escape: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
			},
			Elipsis: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
			},
		},
		ExpandedMessageSign: Item{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[22m",
		},
		ExpandedKeySign: Item{
			Prefix: "\x1b[2m",
		},
		ExpandedKey: Item{
			Suffix: "\x1b[m",
		},
		Unresolved: UnresolvedItem{
			Begin: Item{
				Prefix: "\x1b[31;2m",
				Suffix: "\x1b[22m",
			},
			End: Item{
				Prefix: "\x1b[31;2m",
				Suffix: "\x1b[m",
			},
		},
	}
}
