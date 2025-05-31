package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	globalLogger *zap.Logger
	sugar        *zap.SugaredLogger
	isStdOutput  bool // Flag to track if output is stdout/stderr
)

// Config represents the logging configuration
type Config struct {
	Level      string `toml:"level"`       // debug, info, warn, error
	Format     string `toml:"format"`      // json, console
	AddSource  bool   `toml:"add_source"`  // add caller information
	OutputPath string `toml:"output_path"` // file path or "stdout", "stderr"
}

// Initialize sets up the global logger with the given configuration
func Initialize(cfg Config) error {
	level, err := parseLogLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	var encoder zapcore.Encoder
	encoderConfig := getEncoderConfig()

	switch strings.ToLower(cfg.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console", "": // Default to console
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return fmt.Errorf("invalid log format: %s", cfg.Format)
	}

	// Setup output
	var writeSyncer zapcore.WriteSyncer
	normalizedOutputPath := strings.ToLower(cfg.OutputPath)

	if normalizedOutputPath == "" || normalizedOutputPath == "stdout" || normalizedOutputPath == "/dev/stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
		isStdOutput = true
	} else if normalizedOutputPath == "stderr" || normalizedOutputPath == "/dev/stderr" {
		writeSyncer = zapcore.AddSync(os.Stderr)
		isStdOutput = true
	} else {
		isStdOutput = false
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file '%s': %w", cfg.OutputPath, err)
		}
		writeSyncer = zapcore.AddSync(file)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Add caller information if requested
	var options []zap.Option
	if cfg.AddSource {
		options = append(options, zap.AddCaller(), zap.AddCallerSkip(1)) // Adjust skip if needed
	}

	globalLogger = zap.New(core, options...)
	sugar = globalLogger.Sugar()

	// Initial log message to confirm logger is working
	sugar.Debugw("Logger initialized",
		"level", cfg.Level,
		"format", cfg.Format,
		"add_source", cfg.AddSource,
		"output_path", cfg.OutputPath,
		"is_std_output", isStdOutput,
	)

	return nil
}

// getEncoderConfig returns a sensible encoder configuration
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,              // Or zapcore.LowercaseLevelEncoder for JSON
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339Nano), // More precision
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// parseLogLevel converts string level to zapcore.Level
func parseLogLevel(level string) (zapcore.Level, error) {
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		// Fallback for common variations if UnmarshalText fails or for ""
		switch strings.ToLower(level) {
		case "warn", "warning":
			return zapcore.WarnLevel, nil
		case "": // Default to Info if empty
			return zapcore.InfoLevel, nil
		default:
			return zapcore.InfoLevel, fmt.Errorf("unknown log level: '%s'", level)
		}
	}
	return l, nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		// Fallback to a basic logger if not initialized (should not happen in normal flow)
		fmt.Fprintln(os.Stderr, "Warning: Global logger accessed before initialization. Using development logger.")
		globalLogger, _ = zap.NewDevelopment()
	}
	return globalLogger
}

// GetSugar returns the global sugared logger instance
func GetSugar() *zap.SugaredLogger {
	if sugar == nil {
		sugar = GetLogger().Sugar()
	}
	return sugar
}

// NewLogger creates a new logger with a specific name/component
func NewLogger(name string) *zap.Logger {
	return GetLogger().Named(name)
}

// NewSugar creates a new sugared logger with a specific name/component
func NewSugar(name string) *zap.SugaredLogger {
	return GetSugar().Named(name)
}

// Sync flushes any buffered log entries.
// For stdout/stderr, this is often a no-op or can cause issues in certain environments.
func Sync() error {
	if globalLogger != nil {
		if isStdOutput {
			// For stdout/stderr, Sync() can sometimes cause "invalid argument" or "bad file descriptor".
			// It's generally safe to skip actual syncing as the OS handles flushing.
			// We can still attempt it and ignore common errors.
			err := globalLogger.Sync()
			if err != nil {
				// Check for common, ignorable errors on stdout/stderr
				errMsg := err.Error()
				if strings.Contains(errMsg, "invalid argument") ||
					strings.Contains(errMsg, "bad file descriptor") ||
					strings.Contains(errMsg, "sync /dev/stdout") || // Be more specific
					strings.Contains(errMsg, "sync /dev/stderr") {
					// These errors are common and usually benign for stdout/stderr in some environments.
					return nil
				}
				// Log other unexpected errors from Sync, but not the common ones for stdout.
				// Using fmt.Printf here to avoid recursive logging if logger itself is broken.
				fmt.Fprintf(os.Stderr, "Warning: Error during logger.Sync() for std_output: %v\n", err)
				return err // Propagate other errors
			}
			return nil
		}
		// For file outputs, Sync is important.
		return globalLogger.Sync()
	}
	return nil
}

// Convenience functions for quick logging (structured)
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// ErrorField is used to avoid conflict with the Error function name
func Error(msg string, fields ...zap.Field) { // Renamed from Error to avoid clash
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Convenience functions for sugared logging (printf-style)
func Debugf(template string, args ...interface{}) {
	GetSugar().Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	GetSugar().Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	GetSugar().Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	GetSugar().Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	GetSugar().Fatalf(template, args...)
}

// WithFields creates a logger with predefined fields
func WithFields(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// WithSugarFields creates a sugared logger with predefined fields
func WithSugarFields(args ...interface{}) *zap.SugaredLogger {
	return GetSugar().With(args...)
}
