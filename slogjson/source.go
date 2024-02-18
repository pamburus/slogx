package slogjson

import (
	"log/slog"
	"strconv"
	"strings"
)

// SourceShortString returns a SourceEncodeFunc that encodes source keeping only the file name with package name instead of full path,
// function name, and line number in a form of JSON string.
func SourceShortString() SourceEncodeFunc {
	return func(c slog.Source) slog.Value {
		var bufStorage [64]byte
		buf := bufStorage[:0]

		buf = append(buf, fileWithPackage(c)...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(c.Line), 10)

		return slog.StringValue(string(buf))
	}
}

// SourceShortObject returns a SourceEncodeFunc that encodes source keeping only the file name with package name instead of full path,
// function name, and line number in a form of JSON object.
func SourceShortObject() SourceEncodeFunc {
	return func(c slog.Source) slog.Value {
		return slog.GroupValue(
			slog.String("file", fileWithPackage(c)),
			slog.Int("line", c.Line),
			slog.String("function", c.Function),
		)
	}
}

// SourceLongString returns a SourceEncodeFunc that encodes source keeping the full path, function name, and line number in a form of JSON string.
func SourceLongString() SourceEncodeFunc {
	return func(src slog.Source) slog.Value {
		var bufStorage [64]byte
		buf := bufStorage[:0]

		buf = append(buf, src.File...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(src.Line), 10)

		return slog.StringValue(string(buf))
	}
}

// SourceLongObject returns a SourceEncodeFunc that encodes source keeping the full path, function name, and line number in a form of JSON object.
func SourceLongObject() SourceEncodeFunc {
	return func(c slog.Source) slog.Value {
		return slog.GroupValue(
			slog.String("file", c.File),
			slog.Int("line", c.Line),
			slog.String("function", c.Function),
		)
	}
}

func fileWithPackage(src slog.Source) string {
	found := strings.LastIndexByte(src.File, '/')
	if found == -1 {
		return src.File
	}
	found = strings.LastIndexByte(src.File[:found], '/')
	if found == -1 {
		return src.File
	}

	return src.File[found+1:]
}
