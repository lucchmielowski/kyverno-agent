package logger

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"os"
)

var globalLogger *slog.Logger

func Init(useStderr bool) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	output := os.Stdout
	if useStderr {
		output = os.Stderr
	}

	if os.Getenv("KYVERNO_AGENT_LOG_FORMAT") == "json" {
		globalLogger = slog.New(slog.NewJSONHandler(output, opts))
	} else {
		globalLogger = slog.New(slog.NewTextHandler(output, opts))
	}

	slog.SetDefault(globalLogger)
}

func InitWithEnv() {
	useStderr := os.Getenv("KYVERNO_AGENT_USE_STDERR") == "true"
	Init(useStderr)
}

func Get() *slog.Logger {
	if globalLogger == nil {
		InitWithEnv()
	}
	return globalLogger
}

func WithContext(ctx context.Context) *slog.Logger {
	logger := Get()
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		logger = logger.With(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return logger
}
func LogExecCommand(ctx context.Context, logger *slog.Logger, command string, args []string, caller string) {
	logger.Info("executing command",
		"command", command,
		"args", args,
		"caller", caller,
	)
}

func LogExecCommandResult(ctx context.Context, logger *slog.Logger, command string, args []string, output string, err error, duration float64, caller string) {
	if err != nil {
		logger.Error("command execution failed",
			"command", command,
			"args", args,
			"output", output,
			"error", err,
			"duration_seconds", duration,
			"caller", caller,
		)
	} else {
		logger.Info("command execution succeeded",
			"command", command,
			"args", args,
			"output", output,
			"duration_seconds", duration,
			"caller", caller,
		)
	}
}

func Sync() {
	// No-op for slog, but kept for compatibility
}
