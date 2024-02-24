// Package stylecache contanis helpers for preparing and using cache of styles and styled elements.
package stylecache

// Config represents the configuration of the style cache.
type Config struct {
	LoggerMessageSep string
	KeyValueSep      string
	ArrayBegin       string
	ArraySep         string
	ArrayEnd         string
	MapBegin         string
	MapPairSep       string
	MapKeyValueSep   string
	MapEnd           string
	EvalErrorPrefix  string
	EvalErrorSuffix  string
	EvalPanicPrefix  string
	EvalPanicSuffix  string
	LevelLabels      [NumLevels]string
	LevelNames       [NumLevels]string
}

// DefaultConfig returns the default configuration of the style cache.
func DefaultConfig() *Config {
	return &Config{
		LoggerMessageSep: ":",
		KeyValueSep:      "=",
		ArrayBegin:       "[",
		ArraySep:         ",",
		ArrayEnd:         "]",
		MapBegin:         "{",
		MapPairSep:       ",",
		MapKeyValueSep:   ":",
		MapEnd:           "}",
		EvalErrorPrefix:  "![ERROR:",
		EvalErrorSuffix:  "]",
		EvalPanicPrefix:  "![PANIC:",
		EvalPanicSuffix:  "]",
		LevelLabels:      [NumLevels]string{"DBG", "INF", "WRN", "ERR"},
		LevelNames:       [NumLevels]string{"DEBUG", "INFO", "WARN", "ERROR"},
	}
}
