package stat

import "log/slog"

type Stat struct {
	LinesTotal   int64
	LinesInvalid int64
	ErrorsTotal  int64
}

func (s Stat) IsZero() bool {
	return s == Stat{}
}

func (s Stat) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Group("lines",
			slog.Int64("total", s.LinesTotal),
			slog.Int64("invalid", s.LinesInvalid),
		),
		slog.Group("errors",
			slog.Int64("total", s.ErrorsTotal),
		),
	)
}
