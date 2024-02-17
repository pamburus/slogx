package stylecache

// ---

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

func (s Style) render(inner string) string {
	return s.Prefix + inner + s.Suffix
}

// ---

type StringStyle struct {
	Unquoted Style
	Quoted   Style
	Empty    string
	Null     string
	Escape   Escape
}

func (s StringStyle) append(suffix string) StringStyle {
	return StringStyle{
		Unquoted: s.Unquoted.append(suffix),
		Quoted:   s.Quoted.append(suffix),
		Empty:    s.Empty + suffix,
		Null:     s.Null + suffix,
		Escape:   s.Escape,
	}
}

func (s StringStyle) ws() StringStyle {
	return StringStyle{
		Unquoted: s.Unquoted.ws(),
		Quoted:   s.Quoted.ws(),
		Empty:    s.Empty + " ",
		Null:     s.Null + " ",
		Escape:   s.Escape,
	}
}

// ---

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