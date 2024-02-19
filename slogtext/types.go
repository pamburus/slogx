package slogtext

import (
	"github.com/pamburus/slogx/internal/encbuf"
	"github.com/pamburus/slogx/slogtext/themes"
)

// Theme is a theme for the Handler.
type Theme = themes.Theme

// ThemeItem is a theme item that can have a prefix and a suffix.
type ThemeItem = themes.Item

// ---

type (
	buffer = encbuf.Buffer
)
