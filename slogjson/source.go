package slogjson

import (
	"log/slog"
	"strconv"
	"strings"
)

// SourceShortString is a SourceEncodeFunc that encodes source keeping only the file name with package name instead of full path,
// function name, and line number in a form of JSON string.
func SourceShortString(buf []byte, c slog.Source) ([]byte, slog.Value) {
	buf = append(buf, fileWithPackage(c)...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(c.Line), 10)

	return buf, slog.Value{}
}

// SourceShortObject is a SourceEncodeFunc that encodes source keeping only the file name with package name instead of full path,
// function name, and line number in a form of JSON object.
func SourceShortObject(_ []byte, c slog.Source) ([]byte, slog.Value) {
	return nil, slog.GroupValue(
		slog.String("file", fileWithPackage(c)),
		slog.Int("line", c.Line),
		slog.String("function", c.Function),
	)
}

// SourceLongString is a SourceEncodeFunc that encodes source keeping the full path, function name, and line number in a form of JSON string.
func SourceLongString(buf []byte, src slog.Source) ([]byte, slog.Value) {
	buf = append(buf, src.File...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(src.Line), 10)

	return buf, slog.Value{}
}

// SourceLongObject is a SourceEncodeFunc that encodes source keeping the full path, function name, and line number in a form of JSON object.
func SourceLongObject(_ []byte, c slog.Source) ([]byte, slog.Value) {
	return nil, slog.GroupValue(
		slog.String("file", c.File),
		slog.Int("line", c.Line),
		slog.String("function", c.Function),
	)
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
