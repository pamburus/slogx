// Package syntax provides syntax-related constants and types.
package syntax

// ExpandedMessageSuffix is the suffix for expansion messages.
const ExpandedMessageSuffix = " >>"

// ExpandedKeyPrefix is the prefix for expanded keys.
const ExpandedKeyPrefix = ">- "

// ExpandedValuePrefix is the prefix for expanded values.
const ExpandedValuePrefix = "  \t"

// LevelLabel is a logging level label.
type LevelLabel string

// LevelLabel constants.
const (
	LevelLabelDebug LevelLabel = "DBG"
	LevelLabelInfo  LevelLabel = "INF"
	LevelLabelWarn  LevelLabel = "WRN"
	LevelLabelError LevelLabel = "ERR"
)

// LevelLabelOffsetOverflow is the offset overflow indicator for logging levels.
const LevelLabelOffsetOverflow = '>'

// LevelLabelOffsetUnderflow is the offset underflow indicator for logging levels.
const LevelLabelOffsetUnderflow = '<'
