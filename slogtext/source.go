package slogtext

import (
	"log/slog"
	"strconv"
	"strings"
)

// SourceShort returns a SourceEncodeFunc that encodes source keeping only the file name,
// it's parent folder name and line number.
func SourceShort() SourceEncodeFunc {
	fileWithPackage := func(src slog.Source) string {
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

	return func(buf []byte, c slog.Source) []byte {
		buf = append(buf, fileWithPackage(c)...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(c.Line), 10)

		return buf
	}
}

// SourceLong returns a SourceEncodeFunc that encodes source keeping the full file path and line number.
func SourceLong() SourceEncodeFunc {
	return func(buf []byte, src slog.Source) []byte {
		buf = append(buf, src.File...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(src.Line), 10)

		return buf
	}
}
