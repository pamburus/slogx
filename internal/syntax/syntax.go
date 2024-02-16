package syntax

// ExpansionMessageSuffix is the suffix for expansion messages.
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

// LevelLabelOverflow is the offset overflow indicator for logging levels.
const LevelLabelOffsetOverflow = '>'

// LevelLabelUnderflow is the offset underflow indicator for logging levels.
const LevelLabelOffsetUnderflow = '<'
