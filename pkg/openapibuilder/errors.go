package openapibuilder

import (
	"errors"
	"fmt"
	"strings"
)

// Common errors.
var (
	ErrMissingTitle       = errors.New("info.title is required")
	ErrMissingVersion     = errors.New("info.version is required")
	ErrMissingDescription = errors.New("response description is required")
	ErrInvalidRef         = errors.New("invalid $ref format")
	ErrEmptyPath          = errors.New("path cannot be empty")
)

// BuildError represents an error that occurred during spec building.
type BuildError struct {
	Field   string
	Message string
	Err     error
}

// Error implements the error interface.
func (e *BuildError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Unwrap returns the underlying error.
func (e *BuildError) Unwrap() error {
	return e.Err
}

// ValidationErrors collects multiple validation errors.
type ValidationErrors struct {
	Errors []error
}

// Error implements the error interface.
func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d validation errors:\n", len(e.Errors))
	for _, err := range e.Errors {
		sb.WriteString("  - ")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

// Add adds an error to the collection.
func (e *ValidationErrors) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// AddField adds a field-specific error.
func (e *ValidationErrors) AddField(field, message string) {
	e.Errors = append(e.Errors, &BuildError{Field: field, Message: message})
}

// HasErrors returns true if there are any errors.
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ErrorOrNil returns the error if there are any, or nil.
func (e *ValidationErrors) ErrorOrNil() error {
	if e.HasErrors() {
		return e
	}
	return nil
}
