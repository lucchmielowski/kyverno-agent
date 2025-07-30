package commands

import (
	"context"
	"github.com/lucchmielowski/kyverno-agent/internal/cmd"
	"github.com/lucchmielowski/kyverno-agent/internal/errors"
	"github.com/lucchmielowski/kyverno-agent/internal/logger"
	"github.com/lucchmielowski/kyverno-agent/internal/security"
	"time"
)

const (
	DefaultTimeout = 10
)

type CommandBuilder struct {
	command    string
	args       []string
	context    string
	kubeconfig string
	output     string
	useTimeout bool
	timeout    time.Duration
}

func NewCommandBuilder(command string) *CommandBuilder {
	return &CommandBuilder{
		command: command,
		args:    make([]string, 0),
		timeout: DefaultTimeout,
	}
}

func NewKyvernoCommandBuilder() *CommandBuilder {
	return NewCommandBuilder("kyverno")
}

func (cb *CommandBuilder) WithKubeconfig(kubeconfig string) *CommandBuilder {
	if kubeconfig != "" {
		if err := security.ValidateFilePath(kubeconfig); err != nil {
			return cb
		}
		cb.kubeconfig = kubeconfig
	}
	return cb
}

func (cb *CommandBuilder) WithArgs(args ...string) *CommandBuilder {
	cb.args = append(cb.args, args...)
	return cb
}

func (cb *CommandBuilder) WithTimeout(timeout time.Duration) *CommandBuilder {
	cb.useTimeout = true
	cb.timeout = timeout
	return cb
}

// WithOutput sets the output format
func (cb *CommandBuilder) WithOutput(output string) *CommandBuilder {
	validOutputs := []string{"json", "yaml", "wide", "name", "custom-columns", "custom-columns-file", "go-template", "go-template-file", "jsonpath", "jsonpath-file"}

	valid := false
	for _, validOutput := range validOutputs {
		if output == validOutput {
			valid = true
			break
		}
	}

	if !valid {
		logger.Get().Error("invalid output format", "output", output)
		return cb
	}

	cb.output = output
	return cb
}

func (cb *CommandBuilder) Execute(ctx context.Context) (string, error) {
	log := logger.WithContext(ctx)
	command, args, err := cb.Build()
	if err != nil {
		log.Error("failed to build command",
			"command", cb.command,
			"error", err,
		)
		return "", err
	}

	log.Debug("executing command",
		"command", command,
		"args", args,
	)
	result, err := cb.executeCommand(ctx, command, args)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (cb *CommandBuilder) Build() (string, []string, error) {
	args := make([]string, 0, len(cb.args)+20)
	args = append(args, cb.args...)

	if cb.kubeconfig != "" {
		args = append(args, "--kubeconfig", cb.kubeconfig)
	}

	if cb.output != "" {
		args = append(args, "--output", cb.output)
	}

	return cb.command, args, nil
}

func (cb *CommandBuilder) executeCommand(ctx context.Context, command string, args []string) (string, error) {
	executor := cmd.GetShellExecutor(ctx)
	output, err := executor.Exec(ctx, command, args...)
	if err != nil {
		toolError := errors.NewCommandError(command, err)
		return string(output), toolError
	}
	return string(output), nil
}
