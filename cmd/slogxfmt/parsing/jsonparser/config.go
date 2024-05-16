package jsonparser

import (
	"encoding/json"
	"log/slog"
	"math"

	"github.com/pamburus/slogx/cmd/slogxfmt/parsing/levelparser"
)

type Config struct {
	TimeKeys    []string
	LevelKeys   []string
	MessageKeys []string
	CallerKeys  []string
	ErrorKeys   []string
	ParseLevel  LevelParseFunc
}

func (c Config) withDefaults() Config {
	if len(c.TimeKeys) == 0 {
		c.TimeKeys = []string{slog.TimeKey, "ts", "timestamp", "@timestamp"}
	}

	if len(c.LevelKeys) == 0 {
		c.LevelKeys = []string{slog.LevelKey}
	}

	if len(c.MessageKeys) == 0 {
		c.MessageKeys = []string{slog.MessageKey}
	}

	if len(c.CallerKeys) == 0 {
		c.CallerKeys = []string{"caller"}
	}

	if len(c.ErrorKeys) == 0 {
		c.ErrorKeys = []string{"error", "err", "error-verbose"}
	}

	if c.ParseLevel == nil {
		c.ParseLevel = levelparser.ParseLevel
	}

	return c
}

func (c Config) optimized() config {
	result := config{
		fields:     make(map[string]fieldConfig),
		parseLevel: c.ParseLevel,
	}

	addField := func(key string, kind fieldKind, priority int) {
		cfg := fieldConfig{kind: kind, priority: priority}
		result.fields[key] = cfg
	}

	for i, key := range c.TimeKeys {
		addField(key, fieldTime, i)
	}

	for i, key := range c.LevelKeys {
		addField(key, fieldLevel, i)
	}

	for i, key := range c.MessageKeys {
		addField(key, fieldMessage, i)
	}

	for i, key := range c.CallerKeys {
		addField(key, fieldCaller, i)
	}

	for i, key := range c.ErrorKeys {
		addField(key, fieldError, i)
	}

	return result
}

// ---

type config struct {
	fields     map[string]fieldConfig
	parseLevel LevelParseFunc
}

type fieldConfig struct {
	kind     fieldKind
	priority int
}

type fieldKind int

const (
	fieldOther fieldKind = iota
	fieldTime
	fieldLevel
	fieldMessage
	fieldCaller
	fieldError
	fieldNUM
)

func marshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// ---

var prioritiesTemplate = [fieldNUM]int{
	math.MaxInt,
	math.MaxInt,
	math.MaxInt,
	math.MaxInt,
	math.MaxInt,
}
