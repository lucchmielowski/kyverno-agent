package cmd

import (
	"context"
	"fmt"
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
	startTime := time.Now()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()

	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("command execution failed: %v\n", err)
	} else {
		fmt.Printf("command execution took %v\n", duration)
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
