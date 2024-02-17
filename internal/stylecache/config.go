package stylecache

type Config struct {
	KeyValueSep     string
	ArrayBegin      string
	ArraySep        string
	ArrayEnd        string
	MapBegin        string
	MapPairSep      string
	MapKeyValueSep  string
	MapEnd          string
	EvalErrorPrefix string
	EvalErrorSuffix string
	EvalPanicPrefix string
	EvalPanicSuffix string
	LevelLabels     [NumLevels]string
	LevelNames      [NumLevels]string
}

func DefaultConfig() *Config {
	return &Config{
		KeyValueSep:     "=",
		ArrayBegin:      "[",
		ArraySep:        ",",
		ArrayEnd:        "]",
		MapBegin:        "{",
		MapPairSep:      ",",
		MapKeyValueSep:  ":",
		MapEnd:          "}",
		EvalErrorPrefix: "![ERROR:",
		EvalErrorSuffix: "]",
		EvalPanicPrefix: "![PANIC:",
		EvalPanicSuffix: "]",
		LevelLabels:     [NumLevels]string{"DBG", "INF", "WRN", "ERR"},
		LevelNames:      [NumLevels]string{"DEBUG", "INFO", "WARN", "ERROR"},
	}
}
