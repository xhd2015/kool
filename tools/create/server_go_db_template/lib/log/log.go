package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *slog.Logger

// Config holds log configuration
type Config struct {
	Path       string // log file path, empty means stderr only
	MaxSize    int    // max size in megabytes before rotation
	MaxBackups int    // max number of old log files to keep
	MaxAge     int    // max days to retain old log files
	Compress   bool   // whether to compress rotated files
	Level      string // log level: debug, info, warn, error
}

func init() {
	// Default to stderr with zap backend
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stderr),
		zapcore.InfoLevel,
	)
	logger = slog.New(zapslog.NewHandler(core))
}

// Init initializes the logger with the given configuration
// If path is empty, logs to stderr only
// If path is set, logs to both stderr and the file with rotation
func Init(cfg Config) {
	// Set defaults
	if cfg.MaxSize == 0 {
		cfg.MaxSize = 100 // 100MB
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 3
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 28 // 28 days
	}

	// Parse log level
	level := parseLevel(cfg.Level)

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create writers
	var writers []io.Writer
	writers = append(writers, os.Stderr)

	if cfg.Path != "" {
		// Create lumberjack logger for file rotation
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writers = append(writers, fileWriter)
	}

	// Create zap core with multi-writer
	multiWriter := io.MultiWriter(writers...)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(multiWriter),
		level,
	)

	logger = slog.New(zapslog.NewHandler(core))
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info", "":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// getCaller returns the caller's file:line, skipping the specified number of frames
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	return filepath.Base(file) + ":" + strconv.Itoa(line)
}

// Infof logs at info level with formatting
func Infof(ctx context.Context, msg string, args ...any) {
	logWithTrace(ctx, slog.LevelInfo, msg, args...)
}

// Errorf logs at error level with formatting
func Errorf(ctx context.Context, msg string, args ...any) {
	logWithTrace(ctx, slog.LevelError, msg, args...)
}

// Warnf logs at warn level with formatting
func Warnf(ctx context.Context, msg string, args ...any) {
	logWithTrace(ctx, slog.LevelWarn, msg, args...)
}

// Debugf logs at debug level with formatting
func Debugf(ctx context.Context, msg string, args ...any) {
	logWithTrace(ctx, slog.LevelDebug, msg, args...)
}

// Info logs at info level without formatting
func Info(ctx context.Context, msg string) {
	logWithTrace(ctx, slog.LevelInfo, msg)
}

// Error logs at error level without formatting
func Error(ctx context.Context, msg string) {
	logWithTrace(ctx, slog.LevelError, msg)
}

// Warn logs at warn level without formatting
func Warn(ctx context.Context, msg string) {
	logWithTrace(ctx, slog.LevelWarn, msg)
}

// Debug logs at debug level without formatting
func Debug(ctx context.Context, msg string) {
	logWithTrace(ctx, slog.LevelDebug, msg)
}

func logWithTrace(ctx context.Context, level slog.Level, msg string, args ...any) {
	caller := getCaller(3) // skip logWithTrace -> public func -> caller
	baseArgs := []any{"caller", caller}

	traceID := trace.GetTraceID(ctx)
	if traceID != "" {
		baseArgs = append(baseArgs, "trace_id", traceID.String())
	}

	logger.Log(ctx, level, msg, append(baseArgs, args...)...)
}

// With returns a logger with additional attributes
func With(args ...any) *slog.Logger {
	return logger.With(args...)
}

// Logger returns the underlying slog.Logger
func Logger() *slog.Logger {
	return logger
}
