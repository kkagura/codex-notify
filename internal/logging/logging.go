package logging

import (
	"io"
	"os"
	"time"

	"codex-notify/internal/config"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type App struct {
	Logger zerolog.Logger
	Paths  config.Paths
}

func NewApp(verbose bool) (*App, error) {
	paths := config.ResolvePaths()
	if err := os.MkdirAll(paths.LogDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(paths.StateDir, 0o755); err != nil {
		return nil, err
	}

	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	}

	rotatingFile := &lumberjack.Logger{
		Filename:   paths.LogFile,
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
	}

	writer := io.MultiWriter(os.Stderr, rotatingFile)
	logger := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("service", "codex-notify").
		Logger()

	zerolog.TimeFieldFormat = time.RFC3339

	return &App{
		Logger: logger,
		Paths:  paths,
	}, nil
}
