package stylecache

import "github.com/pamburus/slogx/slogtext/themes"

// NumLevels is the number of log levels.
const NumLevels = themes.NumLevels

// Theme represents a set of styles for log levels and other elements.
type Theme = themes.Theme

// ThemeItem represents a style for a log level or other element.
type ThemeItem = themes.Item
