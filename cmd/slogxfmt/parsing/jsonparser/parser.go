package jsonparser

import (
	"bytes"
	"errors"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/pamburus/slogx/cmd/slogxfmt/model"
	"github.com/valyala/fastjson"
)

func New() *Parser {
	return WithConfig(Config{})
}

func WithConfig(cfg Config) *Parser {
	return &Parser{cfg: cfg.withDefaults().optimized()}
}

type Parser struct {
	cfg  config
	p    fastjson.Parser
	buf  strings.Builder
	stat Stat
}

func (p *Parser) Parse(input []byte) *Chunk {
	chunk := model.NewChunk()

	for len(input) > 0 {
		i := bytes.IndexByte(input, '\n')
		if i < 0 {
			i = len(input)
		}

		lineBytes := input[:i]
		input = input[i+1:]
		p.stat.LinesTotal++

		line, err := p.p.ParseBytes(lineBytes)
		if err != nil {
			p.stat.LinesInvalid++
			continue
		}

		if line.Type() != fastjson.TypeObject {
			p.stat.LinesInvalid++
			continue
		}

		record, err := p.parseLine(chunk, line)
		if err != nil {
			p.stat.LinesInvalid++
			continue
		}

		chunk.Records = append(chunk.Records, record)
	}

	return chunk
}

func (p *Parser) Stat() Stat {
	return p.stat
}

func (p *Parser) parseLine(chunk *Chunk, line *fastjson.Value) (slog.Record, error) {
	object, err := line.Object()
	if err != nil {
		return slog.Record{}, err
	}

	return p.parseLineObject(chunk, object)
}

func (p *Parser) parseLineObject(chunk *Chunk, object *fastjson.Object) (slog.Record, error) {
	var record slog.Record
	attrs := make([]slog.Attr, 0, 32)

	priorities := prioritiesTemplate
	var timeValue *fastjson.Value
	var levelValue *fastjson.Value
	var messageValue *fastjson.Value
	var callerValue *fastjson.Value

	object.Visit(func(key []byte, value *fastjson.Value) {
		field := p.cfg.fields[string(key)]
		switch field.kind {
		case fieldTime:
			if priorities[fieldTime] > field.priority {
				priorities[fieldTime] = field.priority
				timeValue = value
			}
		case fieldLevel:
			if priorities[fieldLevel] > field.priority {
				priorities[fieldLevel] = field.priority
				levelValue = value
			}
		case fieldMessage:
			if priorities[fieldMessage] > field.priority {
				priorities[fieldMessage] = field.priority
				messageValue = value
			}
		case fieldCaller:
			if priorities[fieldCaller] > field.priority {
				priorities[fieldCaller] = field.priority
				callerValue = value
			}
		case fieldError:
			value, err := p.parseValue(value)
			if err != nil {
				p.stat.ErrorsTotal++
			} else {
				if value.Kind() == slog.KindString {
					value = slog.AnyValue(errorValue{value.String()})
				}
				attrs = append(attrs, slog.Attr{Key: p.str(key), Value: value})
			}
		default:
			value, err := p.parseValue(value)
			if err != nil {
				p.stat.ErrorsTotal++
			} else {
				attrs = append(attrs, slog.Attr{Key: p.str(key), Value: value})
			}
		}
	})

	if timeValue != nil {
		var err error
		record.Time, err = p.parseTime(timeValue)
		if err != nil {
			p.stat.ErrorsTotal++
		}
	}
	if levelValue != nil {
		var err error
		record.Level, err = p.parseLevel(levelValue)
		if err != nil {
			p.stat.ErrorsTotal++
		}
	}
	if messageValue != nil {
		var err error
		record.Message, err = p.stre(messageValue.StringBytes())
		if err != nil {
			p.stat.ErrorsTotal++
		}
	}
	if callerValue != nil {
		value, err := p.parseCaller(callerValue)
		if err != nil {
			p.stat.ErrorsTotal++
		} else {
			attrs = append(attrs, slog.Attr{Key: slog.SourceKey, Value: value})
		}
	}

	record.AddAttrs(attrs...)

	return record, nil
}

func (p *Parser) parseTime(value *fastjson.Value) (time.Time, error) {
	switch value.Type() {
	case fastjson.TypeString:
		b, err := value.StringBytes()
		if err != nil {
			return time.Time{}, err
		}

		s := unsafe.String(&b[0], len(b))
		result, err := time.ParseInLocation(time.RFC3339Nano, s, time.UTC)
		runtime.KeepAlive(b)

		return result, err

	default:
		return time.Time{}, errUnexpectedTimeType
	}
}

func (p *Parser) parseLevel(value *fastjson.Value) (slog.Level, error) {
	switch value.Type() {
	case fastjson.TypeString:
		b, err := value.StringBytes()
		if err != nil {
			return slog.Level(0), err
		}

		s := unsafe.String(&b[0], len(b))
		result, err := p.cfg.parseLevel(s)
		runtime.KeepAlive(b)

		return result, err

	default:
		return slog.Level(0), errUnexpectedLevelType
	}
}

func (p *Parser) parseCaller(value *fastjson.Value) (slog.Value, error) {
	switch value.Type() {
	case fastjson.TypeString:
		s, err := p.stre(value.StringBytes())
		if err != nil {
			return slog.Value{}, err
		}

		var source slog.Source
		i := strings.IndexByte(s, ':')
		if i > 0 {
			source.File = s[:i]
			source.Line, err = strconv.Atoi(s[i+1:])
			if err != nil {
				p.stat.ErrorsTotal++
			}
		} else {
			source.Function = s
		}

		return slog.AnyValue(source), nil

	default:
		return slog.Value{}, errUnexpectedCallerType
	}
}

func (p *Parser) parseValue(value *fastjson.Value) (slog.Value, error) {
	switch value.Type() {
	case fastjson.TypeString:
		s, err := p.stre(value.StringBytes())
		if err != nil {
			return slog.Value{}, err
		}

		return slog.StringValue(s), nil

	case fastjson.TypeNumber:
		vu, err := value.Uint64()
		if err == nil {
			return slog.Uint64Value(vu), nil
		}
		vi, err := value.Int64()
		if err == nil {
			return slog.Int64Value(vi), nil
		}
		vf, err := value.Float64()
		if err == nil {
			return slog.Float64Value(vf), nil
		}

		return slog.StringValue(value.String()), nil

	case fastjson.TypeTrue, fastjson.TypeFalse:
		v, err := value.Bool()
		if err != nil {
			return slog.Value{}, err
		}

		return slog.BoolValue(v), nil

	case fastjson.TypeNull:
		return slog.AnyValue(nil), nil

	case fastjson.TypeArray:
		array, err := value.Array()
		if err != nil {
			return slog.Value{}, err
		}

		values := make([]slog.Value, 0, len(array))
		for _, item := range array {
			v, err := p.parseValue(item)
			if err == nil {
				values = append(values, v)
			}
		}

		return slog.AnyValue(values), nil

	case fastjson.TypeObject:
		object, err := value.Object()
		if err != nil {
			return slog.Value{}, err
		}

		attrs := make([]slog.Attr, 0, object.Len())
		object.Visit(func(key []byte, value *fastjson.Value) {
			v, err := p.parseValue(value)
			if err == nil {
				attrs = append(attrs, slog.Attr{Key: string(key), Value: v})
			}
		})

		return slog.GroupValue(attrs...), nil

	default:
		return slog.Value{}, errUnexpectedCallerType
	}
}

func (p *Parser) stre(b []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}

	return p.str(b), nil
}

func (p *Parser) str(b []byte) string {
	if len(b) > p.buf.Cap()-p.buf.Len() {
		p.buf.Reset()
		p.buf.Grow(64 << 10)
	}

	p.buf.Write(b)

	return p.buf.String()[p.buf.Len()-len(b):]
}

// ---

type errorValue struct {
	message string
}

func (e errorValue) Error() string {
	return e.message
}

// ---

var (
	errUnexpectedTimeType   = errors.New("unexpected time type")
	errUnexpectedLevelType  = errors.New("unexpected level type")
	errUnknownLevel         = errors.New("unknown level")
	errUnexpectedCallerType = errors.New("unexpected caller type")
)
