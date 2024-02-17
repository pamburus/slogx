package themes

// Default returns a theme with default colors and decorations.
func Default() Theme {
	level := [NumLevels]Item{
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
	levelValue := [NumLevels]Item{
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

	brace := Item{Prefix: "\x1b[95m", Suffix: "\x1b[m"}
	sep := Item{Prefix: "\x1b[2m", Suffix: "\x1b[22m"}

	return Theme{
		Time: Item{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[m",
		},
		Level:      level,
		LevelValue: levelValue,
		Message: StringItem{
			Content: Item{
				Prefix: "\x1b[1m",
				Suffix: "\x1b[m",
			},
			Escape: Item{
				Prefix: "\x1b[m",
				Suffix: "\x1b[1m",
			},
			Elipsis: Item{
				Prefix: "\x1b[m",
			},
		},
		Key: Item{
			Prefix: "\x1b[32m",
			Suffix: "\x1b[m",
		},
		KeyValueSep: Item{
			Prefix: "\x1b[2m",
			Suffix: "\x1b[m",
		},
		ExpandedKey: Item{
			Prefix: "> \x1b[32m",
			Suffix: "\x1b[m",
		},
		Source: Item{
			Prefix: "\x1b[2;3m@ ",
			Suffix: "\x1b[m",
		},
		StringValue: StringItem{
			Quote: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
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
		BoolValue: Item{
			Prefix: "\x1b[36m",
			Suffix: "\x1b[m",
		},
		NumberValue: Item{
			Prefix: "\x1b[94m",
			Suffix: "\x1b[m",
		},
		NullValue: Item{
			Prefix: "\x1b[31m",
			Suffix: "\x1b[m",
		},
		ErrorValue: StringItem{
			Whole: Item{
				Prefix: "\x1b[31m",
				Suffix: "\x1b[m",
			},
			Quote: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
			},
			Escape: Item{
				Prefix: "\x1b[2m",
				Suffix: "\x1b[22m",
			},
			Elipsis: Item{
				Prefix: "\x1b[31;2m",
				Suffix: "\x1b[m",
			},
		},
		DurationValue: Item{
			Prefix: "\x1b[94m",
			Suffix: "\x1b[m",
		},
		Array: ArrayItem{
			Begin: brace,
			Sep:   sep,
			End:   brace,
		},
		Map: MapItem{
			Begin:       brace,
			PairSep:     sep,
			KeyValueSep: sep,
			End:         brace,
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
