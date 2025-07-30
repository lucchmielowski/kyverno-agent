package cmd

import (
	"context"
	"github.com/lucchmielowski/kyverno-agent/internal/logger"
	"os/exec"
	"time"
)

// ShellExecutor defines the interface for executing shell commands
type ShellExecutor interface {
	Exec(ctx context.Context, command string, args ...string) (output []byte, err error)
}

// DefaultShellExecutor implements ShellExecutor using os/exec
type DefaultShellExecutor struct{}

// Exec executes a command using os/exec.CommandContext
func (e *DefaultShellExecutor) Exec(ctx context.Context, command string, args ...string) ([]byte, error) {
	log := logger.WithContext(ctx)
	startTime := time.Now()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()

	duration := time.Since(startTime)

	if err != nil {
		log.Error("command execution failed",
			"command", command,
			"args", args,
			"output", string(output),
			"error", err,
			"duration_seconds", duration.Seconds(),
		)
	} else {
		log.Info("command execution succeeded",
			"command", command,
			"args", args,
			"output", string(output),
		)
	}

	return output, err
}

// Context key for shell executor injection
type contextKey string

const shellExecutorKey contextKey = "shellExecutor"

// WithShellExecutor returns a context with the given shell executor
func WithShellExecutor(ctx context.Context, executor ShellExecutor) context.Context {
	return context.WithValue(ctx, shellExecutorKey, executor)
}

// GetShellExecutor retrieves the shell executor from context, or returns default
func GetShellExecutor(ctx context.Context) ShellExecutor {
	if executor, ok := ctx.Value(shellExecutorKey).(ShellExecutor); ok {
		return executor
	}
	return &DefaultShellExecutor{}
}
