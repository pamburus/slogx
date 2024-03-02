package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/pamburus/ansitty"
	"github.com/pamburus/slogx/cmd/slogxfmt/parsing/levelparser"
	"github.com/pamburus/slogx/cmd/slogxfmt/pipeline"
	"github.com/pamburus/slogx/slogjson"
	"github.com/pamburus/slogx/slogtext"
	"github.com/pamburus/slogx/slogtext/themes"
)

type args struct {
	Concurrency  int      `arg:"--concurrency" help:"Number of concurrent workers."`
	Color        string   `arg:"--color" help:"Color output control [auto|always|never]." default:"auto"`
	C            bool     `arg:"-c" help:"Force color output."`
	Level        string   `arg:"-l,--level" help:"Log level filter [debug|info|warn|error]." default:"debug"`
	Output       string   `arg:"-o" help:"Output file."`
	OutputFormat string   `arg:"--output-format" help:"Output format."`
	Expansion    string   `arg:"-x,--expansion" help:"Attribute expansion control [auto|always|never|low|medium|high]." default:"auto"`
	Inputs       []string `arg:"positional" help:"Input files to process."`
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := args{}
	arg.MustParse(&args)

	input, closeInputs, err := openInputs(args.Inputs)
	if err != nil {
		return err
	}
	defer closeInputs()

	output, closeOutput, err := openOutput(args.Output)
	if err != nil {
		return err
	}
	defer closeOutput()

	if args.C {
		args.Color = "always"
	}

	color := slogtext.ColorNever
	switch args.Color {
	case "auto":
		if ansitty.Enable(output) {
			color = slogtext.ColorAlways
		}
	case "always":
		color = slogtext.ColorAlways
	}

	expansion := slogtext.ExpansionAuto
	switch args.Expansion {
	case "auto":
		expansion = slogtext.ExpansionAuto
	case "always":
		expansion = slogtext.ExpansionAlways
	case "never":
		expansion = slogtext.ExpansionNever
	case "low":
		expansion = slogtext.ExpansionLow
	case "medium":
		expansion = slogtext.ExpansionMedium
	case "high":
		expansion = slogtext.ExpansionHigh
	default:
		return fmt.Errorf("invalid expansion setting: %s", args.Expansion)
	}

	level, err := levelparser.ParseLevel(args.Level)
	if err != nil {
		return err
	}

	handler := func(level slog.Level, color slogtext.ColorSetting) pipeline.HandlerFactory {
		if args.OutputFormat == "json" {
			return func(w io.Writer) slog.Handler {
				return slogjson.NewHandler(w,
					slogjson.WithLevel(level),
					slogjson.WithSource(true),
				)
			}
		}

		if args.OutputFormat == "logfmt" {
			return func(w io.Writer) slog.Handler {
				return slog.NewTextHandler(w, &slog.HandlerOptions{
					Level:     level,
					AddSource: true,
				})
			}
		}

		return func(w io.Writer) slog.Handler {
			return slogtext.NewHandler(w,
				slogtext.WithLevel(level),
				slogtext.WithColor(color),
				slogtext.WithExpansion(expansion),
				slogtext.WithSource(true),
				slogtext.WithTheme(themes.Fancy()),
				slogtext.WithLoggerKey("logger"),
			)
		}
	}

	var internalLevel slog.Level
	if os.Getenv("SLOGXFMT_DEBUG") != "" {
		internalLevel = slog.LevelDebug
	}
	slog.SetDefault(
		slog.New(handler(internalLevel, ansitty.Enable)(os.Stderr)).
			With(slog.String("logger", "slogxfmt")),
	)

	pipeline := pipeline.New(handler(level, color)).WithConcurrency(args.Concurrency)
	err = pipeline.Run(context.Background(), input, output)
	if err != nil {
		return err
	}

	return nil
}

func openOutput(output string) (io.Writer, func(), error) {
	if output == "" || output == "-" {
		return os.Stdout, func() {}, nil
	}

	file, err := os.Create(output)
	if err != nil {
		return nil, nil, err
	}

	return file, closeFunc(file), nil
}

func openInputs(inputs []string) (io.Reader, func(), error) {
	if len(inputs) == 0 {
		return os.Stdin, func() {}, nil
	}

	closeFuncs := make([]func(), 0, len(inputs))
	closeAll := func() {
		for _, closeFunc := range closeFuncs {
			closeFunc()
		}
	}

	readers := make([]io.Reader, 0, len(inputs))
	for _, input := range inputs {
		if input == "-" {
			readers = append(readers, os.Stdin)
			continue
		}

		file, err := os.Open(input)
		if err != nil {
			closeAll()
			return nil, nil, err
		}

		readers = append(readers, file)
		closeFuncs = append(closeFuncs, closeFunc(file))
	}

	return io.MultiReader(readers...), closeAll, nil
}

func closeFunc(closer io.Closer) func() {
	return func() {
		err := closer.Close()
		if err != nil {
			slog.Error("failed to close file", slog.Any("error", err))
		}
	}
}
