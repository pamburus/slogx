package stylecache

import "strings"

// ---

// Style represents a pair of strings that can be used to style text.
type Style struct {
	Prefix string
	Suffix string
}

func (s *Style) set(prefix, suffix string) {
	s.Prefix = prefix
	s.Suffix = suffix
}

func (s Style) ws() Style {
	return s.append(" ")
}

func (s Style) append(suffix string) Style {
	s.set(s.Prefix, s.Suffix+suffix)

	return s
}

func (s Style) prepend(prefix string) Style {
	s.set(prefix+s.Prefix, s.Suffix)

	return s
}

func (s Style) render(inner string) string {
	if strings.TrimSpace(inner) == "" {
		return inner
	}

	return s.Prefix + inner + s.Suffix
}

// ---

// StringStyle represents a set of styles for string values.
type StringStyle struct {
	Unquoted Style
	Quoted   Style
	Empty    string
	Null     string
	Elipsis  string
	Escape   Escape
}

func (s StringStyle) append(suffix string) StringStyle {
	return StringStyle{
		Unquoted: s.Unquoted.append(suffix),
		Quoted:   s.Quoted.append(suffix),
		Empty:    s.Empty + suffix,
		Null:     s.Null + suffix,
		Elipsis:  s.Elipsis + suffix,
		Escape:   s.Escape,
	}
}

func (s StringStyle) ws() StringStyle {
	return s.append(" ")
}

// ---

// Escape represents a set of styled escape sequences for string value escaping.
type Escape struct {
	Style     Style
	Tab       string
	CR        string
	LF        string
	Backslash string
	Quote     string
}

// ---

func sti(prefix, suffix string) Style {
	return Style{prefix, suffix}
}

func st(item ThemeItem) Style {
	return Style{item.Prefix, item.Suffix}
}
