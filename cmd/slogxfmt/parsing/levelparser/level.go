package levelparser

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
)

func ParseLevel(input string) (slog.Level, error) {
	if len(input) == 0 {
		return slog.Level(0), errInvalidLevel
	}

	c := input[len(input)-1]
	if isDigit(rune(c)) {
		return parseLevelWithOffset(input)
	}

	return parseLevel(input)
}

// ---

func parseLevelWithOffset(input string) (slog.Level, error) {
	i := strings.LastIndexFunc(input, isNotDigit)
	if i == -1 {
		return slog.Level(0), errInvalidLevel
	}

	level, ok := levels[input[:i]]
	if !ok {
		return slog.Level(0), errInvalidLevel
	}

	offset, err := strconv.Atoi(input[i:])
	if err != nil {
		return slog.Level(0), errInvalidLevel
	}

	return level + slog.Level(offset), nil
}

func parseLevel(input string) (slog.Level, error) {
	if level, ok := levels[input]; ok {
		return level, nil
	}

	return slog.Level(0), errInvalidLevel
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isNotDigit(r rune) bool {
	return !isDigit(r)
}

// ---

var levels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
	"Debug": slog.LevelDebug,
	"Info":  slog.LevelInfo,
	"Warn":  slog.LevelWarn,
	"Error": slog.LevelError,
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
	"DBG":   slog.LevelDebug,
	"INF":   slog.LevelInfo,
	"WRN":   slog.LevelWarn,
	"ERR":   slog.LevelError,
	"D":     slog.LevelDebug,
	"I":     slog.LevelInfo,
	"W":     slog.LevelWarn,
	"E":     slog.LevelError,
	"d":     slog.LevelDebug,
	"i":     slog.LevelInfo,
	"w":     slog.LevelWarn,
	"e":     slog.LevelError,
}

var errInvalidLevel = errors.New("invalid level")
