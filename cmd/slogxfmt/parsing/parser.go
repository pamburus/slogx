package parsing

import "github.com/pamburus/slogx/cmd/slogxfmt/parsing/jsonparser"

func NewDefaultJSONParser() Parser {
	return jsonparser.New()
}

func NewJSONParser(config jsonparser.Config) ParserFactory {
	return func() Parser {
		return jsonparser.WithConfig(config)
	}
}

// ---

type ParserFactory func() Parser

// ---

type Parser interface {
	Parse(input []byte) *Chunk
	Stat() Stat
}
