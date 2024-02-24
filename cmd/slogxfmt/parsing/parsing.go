package parsing

import "github.com/pamburus/slogx/cmd/slogxfmt/parsing/json"

type Parser interface {
	Parse(input []byte) (*Chunk, error)
}

func NewJSONParser(config JSONParserConfig) func() Parser {
	return func() Parser {
		return json.NewParser(config)
	}
}
