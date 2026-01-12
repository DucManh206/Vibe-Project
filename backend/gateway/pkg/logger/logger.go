package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional convenience methods
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	config := zap.NewProductionConfig()

	// Set log level from environment
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch logLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn", "warning":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Use JSON format in production, console in development
	env := os.Getenv("GATEWAY_ENV")
	if env != "production" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Build the logger
	zapLogger, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		// Fallback to a basic logger
		zapLogger = zap.NewNop()
	}

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}
}

// NewTestLogger creates a logger for testing
func NewTestLogger() *Logger {
	return &Logger{
		SugaredLogger: zap.NewNop().Sugar(),
	}
}

// WithFields returns a new logger with the given fields
func (l *Logger) WithFields(fields ...interface{}) *Logger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With(fields...),
	}
}

// Info logs an info message with optional key-value pairs
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Infow(msg, keysAndValues...)
}

// Debug logs a debug message with optional key-value pairs
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Debugw(msg, keysAndValues...)
}

// Warn logs a warning message with optional key-value pairs
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Warnw(msg, keysAndValues...)
}

// Error logs an error message with optional key-value pairs
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Errorw(msg, keysAndValues...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Fatalw(msg, keysAndValues...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.SugaredLogger.Sync()
}