package stylecache

import (
	"strings"

	"github.com/pamburus/slogx/internal/syntax"
	"github.com/pamburus/slogx/slogtext/themes"
)

// ---

func New(theme *Theme, cfg *Config) *StyleCache {
	c := &StyleCache{
		Config:              *cfg,
		Time:                st(theme.Time).ws(),
		Message:             sst(theme.Message).ws(),
		Key:                 st(theme.Key).append(st(theme.KeyValueSep).render(cfg.KeyValueSep)),
		ExpandedKey:         st(theme.ExpandedKey).prepend(st(theme.ExpandedKeySign).render(syntax.ExpandedKeyPrefix)),
		Source:              st(theme.Source).ws(),
		StringValue:         sst(theme.StringValue),
		NumberValue:         st(theme.NumberValue),
		BoolValue:           st(theme.BoolValue),
		TimeValue:           st(theme.TimeValue),
		DurationValue:       st(theme.DurationValue),
		ErrorValue:          sst(theme.ErrorValue),
		EvalError:           sti(st(theme.Unresolved.Begin).render(cfg.EvalErrorPrefix), st(theme.Unresolved.End).render(cfg.EvalErrorSuffix)),
		EvalPanic:           sti(st(theme.Unresolved.Begin).render(cfg.EvalPanicPrefix), st(theme.Unresolved.End).render(cfg.EvalPanicSuffix)),
		Array:               sti(st(theme.Array.Begin).render(cfg.ArrayBegin), st(theme.Array.End).render(cfg.ArrayEnd)),
		ArraySep:            st(theme.Array.Sep).render(cfg.ArraySep),
		Map:                 sti(st(theme.Map.Begin).render(cfg.MapBegin), st(theme.Map.End).render(cfg.MapEnd)),
		MapPairSep:          st(theme.Map.PairSep).render(cfg.MapPairSep),
		MapKeyValueSep:      st(theme.Map.KeyValueSep).render(cfg.MapKeyValueSep),
		ExpandedMessageSign: st(theme.ExpandedMessageSign).render(syntax.ExpandedMessageSuffix),
	}

	c.EmptyArray = strings.TrimSpace(c.Array.Prefix) + strings.TrimSpace(c.Array.Suffix)
	c.EmptyMap = strings.TrimSpace(c.Map.Prefix) + strings.TrimSpace(c.Map.Suffix)
	c.Null = st(theme.NullValue).render("null")

	for i := 0; i < NumLevels; i++ {
		c.LevelLabel[i] = st(theme.Level[i]).ws().render(cfg.LevelLabels[i])
		c.LevelValue[i] = st(theme.LevelValue[i])
	}

	return c
}

type StyleCache struct {
	Config              Config
	Time                Style
	LevelLabel          [NumLevels]string
	Message             StringStyle
	Key                 Style
	ExpandedKey         Style
	Source              Style
	StringValue         StringStyle
	NumberValue         Style
	BoolValue           Style
	TimeValue           Style
	DurationValue       Style
	ErrorValue          StringStyle
	LevelValue          [NumLevels]Style
	Array               Style
	Map                 Style
	EvalError           Style
	EvalPanic           Style
	EmptyArray          string
	ArraySep            string
	EmptyMap            string
	MapPairSep          string
	MapKeyValueSep      string
	Null                string
	ExpandedMessageSign string
}

// ---

func sst(sti themes.StringItem) StringStyle {
	quote := st(sti.Quote).render(`"`)
	unquoted := Style{
		Prefix: sti.Whole.Prefix + sti.Content.Prefix,
		Suffix: sti.Content.Suffix + sti.Whole.Suffix,
	}
	quoted := Style{
		Prefix: sti.Whole.Prefix + quote + sti.Content.Prefix,
		Suffix: sti.Content.Suffix + quote + sti.Whole.Suffix,
	}
	escape := st(sti.Escape)

	return StringStyle{
		Unquoted: unquoted,
		Quoted:   quoted,
		Empty:    st(sti.Quote).render(`""`),
		Null:     quoted.render("null"),
		Elipsis:  st(sti.Elipsis).render("..."),
		Escape: Escape{
			Style:     escape,
			Tab:       escape.render(`\t`),
			CR:        escape.render(`\r`),
			LF:        escape.render(`\n`),
			Backslash: escape.render(`\`) + `\`,
			Quote:     escape.render(`\`) + `"`,
		},
	}
}
