package errors

import (
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"strings"
	"time"
)

// ToolError represents a structured error with context and recovery suggestions
type ToolError struct {
	Operation    string                 `json:"operation"`
	Cause        error                  `json:"cause"`
	Suggestions  []string               `json:"suggestions"`
	IsRetryable  bool                   `json:"is_retryable"`
	Timestamp    time.Time              `json:"timestamp"`
	ErrorCode    string                 `json:"error_code"`
	Component    string                 `json:"component"`
	ResourceType string                 `json:"resource_type,omitempty"`
	ResourceName string                 `json:"resource_name,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

func (e *ToolError) Error() string {
	return fmt.Sprintf("[%s] %s failed: %v", e.Component, e.Operation, e.Cause)
}

func (e *ToolError) ToMCPResult() *mcp.CallToolResult {
	var message strings.Builder

	// Format the error message with context
	message.WriteString(fmt.Sprintf("❌ **%s Error**\n\n", e.Component))
	message.WriteString(fmt.Sprintf("**Operation**: %s\n", e.Operation))
	message.WriteString(fmt.Sprintf("**Error**: %s\n", e.Cause.Error()))

	if e.ResourceType != "" {
		message.WriteString(fmt.Sprintf("**Resource Type**: %s\n", e.ResourceType))
	}

	if e.ResourceName != "" {
		message.WriteString(fmt.Sprintf("**Resource Name**: %s\n", e.ResourceName))
	}

	message.WriteString(fmt.Sprintf("**Error Code**: %s\n", e.ErrorCode))
	message.WriteString(fmt.Sprintf("**Timestamp**: %s\n", e.Timestamp.Format(time.RFC3339)))

	if e.IsRetryable {
		message.WriteString("**Retryable**: Yes\n")
	} else {
		message.WriteString("**Retryable**: No\n")
	}

	if len(e.Suggestions) > 0 {
		message.WriteString("\n**💡 Suggestions**:\n")
		for i, suggestion := range e.Suggestions {
			message.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
		}
	}

	if len(e.Context) > 0 {
		message.WriteString("\n**📋 Context**:\n")
		for key, value := range e.Context {
			message.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
	}

	return mcp.NewToolResultError(message.String())
}

func NewToolError(component, operation string, cause error) *ToolError {
	return &ToolError{
		Operation:   operation,
		Cause:       cause,
		Suggestions: []string{},
		IsRetryable: false,
		Timestamp:   time.Now(),
		ErrorCode:   "UNKNOWN",
		Component:   component,
		Context:     make(map[string]interface{}),
	}
}

func (e *ToolError) WithSuggestions(suggestions ...string) *ToolError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

func (e *ToolError) WithRetryable(retryable bool) *ToolError {
	e.IsRetryable = retryable
	return e
}

func (e *ToolError) WithErrorCode(errorCode string) *ToolError {
	e.ErrorCode = errorCode
	return e
}

func (e *ToolError) WithResource(resourceType, resourceName string) *ToolError {
	e.ResourceType = resourceType
	e.ResourceName = resourceName
	return e
}

func (e *ToolError) WithContext(key string, value interface{}) *ToolError {
	e.Context[key] = value
	return e
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *ToolError {
	err := NewToolError("Validation", fmt.Sprintf("validate %s", field), fmt.Errorf("%s", message))

	err = err.WithSuggestions(
		"Check the input format",
		"Refer to the documentation for valid values",
		"Verify the parameter requirements",
	).WithRetryable(false).WithErrorCode("VALIDATION_ERROR")

	return err
}

// NewSecurityError creates a security-related error
func NewSecurityError(operation string, cause error) *ToolError {
	err := NewToolError("Security", operation, cause)

	err = err.WithSuggestions(
		"Review the input for potentially dangerous content",
		"Use only trusted input sources",
		"Contact security team if needed",
	).WithRetryable(false).WithErrorCode("SECURITY_ERROR")

	return err
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, timeout time.Duration) *ToolError {
	cause := fmt.Errorf("operation timed out after %v", timeout)
	err := NewToolError("Timeout", operation, cause)

	err = err.WithSuggestions(
		"Try the operation again",
		"Check network connectivity",
		"Increase timeout if possible",
	).WithRetryable(true).WithErrorCode("TIMEOUT_ERROR")

	return err
}

// NewCommandError creates a command execution error
func NewCommandError(command string, cause error) *ToolError {
	err := NewToolError("Command", fmt.Sprintf("execute %s", command), cause)

	err = err.WithSuggestions(
		"Check if the command exists in PATH",
		"Verify command syntax and arguments",
		"Check system permissions",
	).WithRetryable(true).WithErrorCode("COMMAND_ERROR")

	return err
}

// TODO add kyverno-specific error
