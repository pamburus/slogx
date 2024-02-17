// Package themes provides themes for the Handler.
package themes

import (
	"github.com/pamburus/slogx/internal/stripansi"
)

// NumLevels is the number of levels.
const NumLevels = 4

// Theme is a theme for the Handler.
type Theme struct {
	Time          Item
	Level         LevelItem
	Message       StringItem
	Key           Item
	KeyValueSep   Item
	Source        Item
	TimeValue     Item
	LevelValue    LevelItem
	StringValue   StringItem
	BoolValue     Item
	NumberValue   Item
	NullValue     Item
	ErrorValue    StringItem
	DurationValue Item
	Array         ArrayItem
	Map           MapItem
	Unresolved    UnresolvedItem
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

// Item is a theme item that can have a prefix and a suffix.
type Item struct {
	Prefix string
	Suffix string
}

// Plain returns a theme item with no color.
func (i Item) Plain() Item {
	return Item{
		Prefix: stripansi.Strip(i.Prefix),
		Suffix: stripansi.Strip(i.Suffix),
	}
}

// ---

// StringItem is a theme item for a string type.
type StringItem struct {
	Whole   Item
	Content Item
	Quote   Item
	Elipsis Item
	Escape  Item
}

// Plain returns a copy of i with no color.
func (i StringItem) Plain() StringItem {
	return StringItem{
		Whole:   i.Whole.Plain(),
		Content: i.Content.Plain(),
		Quote:   i.Quote.Plain(),
		Elipsis: i.Elipsis.Plain(),
		Escape:  i.Escape.Plain(),
	}
}

// ---

// LevelItem is a theme item set per logging level.
type LevelItem [NumLevels]Item

// Plain returns a theme item with no color.
func (i LevelItem) Plain() LevelItem {
	return LevelItem{
		i[0].Plain(),
		i[1].Plain(),
		i[2].Plain(),
		i[3].Plain(),
	}
}

// ---

// ArrayItem is a theme item for an array type.
type ArrayItem struct {
	Begin Item
	Sep   Item
	End   Item
}

// Plain returns a copy of i with no color.
func (i ArrayItem) Plain() ArrayItem {
	return ArrayItem{
		Begin: i.Begin.Plain(),
		Sep:   i.Sep.Plain(),
		End:   i.End.Plain(),
	}
}

// ---

// MapItem is a theme item for an map type.
type MapItem struct {
	Begin       Item
	PairSep     Item
	KeyValueSep Item
	End         Item
}

// Plain returns a copy of i with no color.
func (i MapItem) Plain() MapItem {
	return MapItem{
		Begin:       i.Begin.Plain(),
		PairSep:     i.PairSep.Plain(),
		KeyValueSep: i.KeyValueSep.Plain(),
		End:         i.End.Plain(),
	}
}

// ---

// UnresolvedItem is a theme item for a composite type representation that can have a begin, 1 or 2 separators, and an end.
type UnresolvedItem struct {
	Begin Item
	End   Item
}

// Plain returns a copy of i with no color.
func (i UnresolvedItem) Plain() UnresolvedItem {
	return UnresolvedItem{
		Begin: i.Begin.Plain(),
		End:   i.End.Plain(),
	}
}
