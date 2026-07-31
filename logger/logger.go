package logger

import (
	"log/slog"
	"os"
)

// Logger defines the interface for our structured logging wrapper.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
}

type slogLogger struct {
	logger *slog.Logger
}

// New creates a new structured JSON logger.
// If isProd is true, it only logs Info and above. Otherwise, it logs Debug too.
func New(isProd bool) Logger {
	logLevel := slog.LevelDebug
	if isProd {
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	return &slogLogger{
		logger: slog.New(handler),
	}
}

func (l *slogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *slogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Fatal logs the message at Error level and terminates the process.
// WARNING: os.Exit(1) bypasses all deferred calls. Only use during
// startup phase (before server begins accepting requests).
func (l *slogLogger) Fatal(msg string, args ...any) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}
