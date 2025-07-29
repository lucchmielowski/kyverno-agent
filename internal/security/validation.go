package security

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Path pattern (check directory traversal)
	pathPattern = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidateFilePath validates a file path for security
func ValidateFilePath(path string) error {
	if len(path) > 4096 {
		return ValidationError{Field: "path", Message: "path too long"}
	}

	if strings.Contains(path, "..") {
		return ValidationError{Field: "path", Message: "path traversal not allowed"}
	}

	if !pathPattern.MatchString(path) {
		return ValidationError{Field: "path", Message: "contains invalid characters"}
	}

	return nil
}
