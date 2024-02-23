package parsing

import (
	"bytes"
	"errors"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/valyala/fastjson"
)

func NewParser() *Parser {
	return &Parser{}
}

type Parser struct {
	scanner fastjson.Scanner
	buf     strings.Builder
}

func (p *Parser) Parse(input []byte) (*Chunk, error) {
	p.scanner.InitBytes(input)

	var err error
	chunk := pool.Get().(*Chunk)
	defer func() {
		if err != nil {
			chunk.Free()
		}
		if p := recover(); p != nil {
			chunk.Free()
			panic(p)
		}
	}()

	err = p.parse(chunk)
	if err != nil {
		return nil, err
	}

	return chunk, nil
}

func (p *Parser) parse(chunk *Chunk) error {
	for p.scanner.Next() {
		line := p.scanner.Value()
		if line.Type() != fastjson.TypeObject {
			continue
		}

		object, err := line.Object()
		if err != nil {
			return err
		}

		record, err := p.parseLine(chunk, object)
		if err != nil {
			return err
		}

		chunk.Records = append(chunk.Records, record)
	}

	return p.scanner.Error()
}

func (p *Parser) parseLine(chunk *Chunk, object *fastjson.Object) (slog.Record, error) {
	var record slog.Record
	attrs := make([]slog.Attr, 0, 32)

	object.Visit(func(key []byte, value *fastjson.Value) {
		switch {
		case bytes.Equal(key, timeKey):
			fallthrough
		case bytes.Equal(key, timeKey2):
			record.Time, _ = p.parseTime(value)
		case bytes.Equal(key, levelKey):
			record.Level, _ = p.parseLevel(value)
		case bytes.Equal(key, messageKey):
			record.Message, _ = p.str(value.StringBytes())
		case bytes.Equal(key, callerKey):
			value, _ := p.parseCaller(value)
			attrs = append(attrs, slog.Attr{Key: slog.SourceKey, Value: value})
		default:
			value, _ := p.parseValue(value)
			attrs = append(attrs, slog.Attr{Key: string(key), Value: value})
		}
	})

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

		switch {
		case bytes.EqualFold(b, levelDebug):
			return slog.LevelDebug, nil
		case bytes.EqualFold(b, levelInfo):
			return slog.LevelInfo, nil
		case bytes.EqualFold(b, levelWarn):
			return slog.LevelWarn, nil
		case bytes.EqualFold(b, levelError):
			return slog.LevelError, nil
		default:
			return slog.Level(0), errUnknownLevel
		}
	default:
		return slog.Level(0), errUnexpectedLevelType
	}
}

func (p *Parser) parseCaller(value *fastjson.Value) (slog.Value, error) {
	switch value.Type() {
	case fastjson.TypeString:
		s, _ := p.str(value.StringBytes())

		var source slog.Source
		i := strings.IndexByte(s, ':')
		if i > 0 {
			source.File = s[:i]
			source.Line, _ = strconv.Atoi(s[i+1:])
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
		s, err := p.str(value.StringBytes())
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

func (p *Parser) str(b []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if len(b) > p.buf.Cap()-p.buf.Len() {
		p.buf.Reset()
		p.buf.Grow(64 << 10)
	}

	p.buf.Write(b)

	return p.buf.String()[p.buf.Len()-len(b):], nil
}

// ---

type Chunk struct {
	Records []slog.Record
}

func (p *Chunk) Free() {
	if len(p.Records) <= 8192 {
		p.Records = p.Records[:0]
		pool.Put(p)
	}
}

// ---

// ---

var pool = sync.Pool{
	New: func() any {
		return &Chunk{
			Records: make([]slog.Record, 0, 2048),
		}
	},
}

// ---

var (
	timeKey    = []byte(slog.TimeKey)
	timeKey2   = []byte("ts")
	levelKey   = []byte(slog.LevelKey)
	messageKey = []byte(slog.MessageKey)
	callerKey  = []byte("caller")
)

var (
	levelDebug = []byte("debug")
	levelInfo  = []byte("info")
	levelWarn  = []byte("warn")
	levelError = []byte("error")
)

var (
	errUnexpectedTimeType   = errors.New("unexpected time type")
	errUnexpectedLevelType  = errors.New("unexpected level type")
	errUnknownLevel         = errors.New("unknown level")
	errUnexpectedCallerType = errors.New("unexpected caller type")
)
