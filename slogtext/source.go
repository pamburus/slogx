package slogtext

import (
	"log/slog"
	"strconv"
	"strings"
)

// CallerShort returns a CallerEncodeFunc that encodes caller
// keeping only package name, base filename and line number.
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

// CallerLong returns a CallerEncodeFunc that encodes caller keeping full file path and line number.
func SourceLong() SourceEncodeFunc {
	return func(buf []byte, src slog.Source) []byte {
		buf = append(buf, src.File...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(src.Line), 10)

		return buf
	}
}
