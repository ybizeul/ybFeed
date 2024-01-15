package yblog

import (
	"os"

	"golang.org/x/exp/slog"
)

type YBLogger struct {
	Name     string
	Logger   *slog.Logger
	loglevel *slog.LevelVar
}

func NewYBLogger(name string, env []string) *YBLogger {
	ll := new(slog.LevelVar)
	lg := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: ll})).WithGroup(name)

	for _, e := range env {
		if os.Getenv(e) != "" {
			ll.Set(slog.LevelDebug)
		}
	}

	return &YBLogger{
		Name:     name,
		Logger:   lg,
		loglevel: ll,
	}
}

func (lg *YBLogger) SetLevel(l slog.Level) {
	lg.loglevel.Set(l)
}

func (lg *YBLogger) Level() slog.Level {
	return lg.loglevel.Level()
}
