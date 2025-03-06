package logger

import (
	"log/slog"
	"os"
)

// logger is a struct that encapsulates a slog.Logger to provide structured logging functionality.
type logger struct {
	log *slog.Logger
}

// globalLogger is a package-level variable that provides access to a pre-configured logger for structured logging.
var (
	globalLogger logger
)

// Init initializes the global logger with the specified logging level and a JSON handler for structured logging.
func Init(lvl slog.Level) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
	globalLogger.log = log
}

// Logger defines an interface for logging messages with varying levels of severity: Debug, Info, Warn, and Error.
// Debug logs are typically used for fine-grained information, useful during development or troubleshooting.
// Info logs offer general information about the application's normal operations.
// Warn logs to indicate situations that are unusual or may require attention but are not errors.
// Error logs report issues or problems that have occurred during application execution.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

func Debug(msg string, args ...interface{}) {
	globalLogger.log.Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	globalLogger.log.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	globalLogger.log.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	globalLogger.log.Error(msg, args...)
}
