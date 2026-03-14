// Package validate provides OpenAPI specification validation using libopenapi.
package validate

import (
	"errors"
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"
)

// ValidationResult contains the results of validating an OpenAPI spec.
type ValidationResult struct {
	Valid    bool
	Version  string
	Errors   []ValidationError
	Warnings []ValidationError
}

// ValidationError represents a single validation error or warning.
type ValidationError struct {
	Message  string
	Line     int
	Column   int
	Path     string
	RuleID   string
	Severity string // "error" or "warning"
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", e.Path, e.Line, e.Column, e.Message)
	}
	return e.Message
}

// Validate validates an OpenAPI specification from YAML or JSON bytes.
// It uses libopenapi to parse and validate the spec.
func Validate(specBytes []byte) (*ValidationResult, error) {
	if len(specBytes) == 0 {
		return nil, errors.New("empty specification")
	}

	// Create a new document from the spec bytes
	doc, err := libopenapi.NewDocument(specBytes)
	if err != nil {
		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{
				{Message: fmt.Sprintf("failed to parse specification: %v", err), Severity: "error"},
			},
		}, nil
	}

	result := &ValidationResult{
		Valid:   true,
		Version: doc.GetVersion(),
	}

	// Build the model to trigger validation
	v3Model, modelErr := doc.BuildV3Model()
	if modelErr != nil {
		result.Errors = append(result.Errors, ValidationError{
			Message:  modelErr.Error(),
			Severity: "error",
		})
		result.Valid = false
	}

	// If we got a model, we can do additional validation
	if v3Model != nil && v3Model.Index != nil {
		// Check for circular references and other issues
		circularRefs := v3Model.Index.GetCircularReferences()
		for _, ref := range circularRefs {
			result.Warnings = append(result.Warnings, ValidationError{
				Message:  fmt.Sprintf("circular reference detected: %s", ref.Start.Definition),
				Severity: "warning",
			})
		}
	}

	return result, nil
}

// ValidateFile validates an OpenAPI specification from a file path.
func ValidateFile(path string) (*ValidationResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return Validate(data)
}

// IsValidVersion checks if the given version string is a valid OpenAPI version.
func IsValidVersion(version string) bool {
	switch version {
	case "3.0.0", "3.0.1", "3.0.2", "3.0.3",
		"3.1.0", "3.1.1",
		"3.2.0":
		return true
	default:
		return false
	}
}

// GetMajorMinorVersion extracts the major.minor portion of a version string.
func GetMajorMinorVersion(version string) string {
	if len(version) >= 3 {
		return version[:3]
	}
	return version
}

// ParseAndValidate parses an OpenAPI spec and returns the libopenapi document
// along with validation results. This is useful when you need both the parsed
// document for further processing and validation results.
func ParseAndValidate(specBytes []byte) (*libopenapi.Document, *ValidationResult, error) {
	if len(specBytes) == 0 {
		return nil, nil, errors.New("empty specification")
	}

	doc, err := libopenapi.NewDocument(specBytes)
	if err != nil {
		return nil, &ValidationResult{
			Valid: false,
			Errors: []ValidationError{
				{Message: fmt.Sprintf("failed to parse specification: %v", err), Severity: "error"},
			},
		}, nil
	}

	result := &ValidationResult{
		Valid:   true,
		Version: doc.GetVersion(),
	}

	// Build the model to trigger validation
	v3Model, modelErr := doc.BuildV3Model()
	if modelErr != nil {
		result.Errors = append(result.Errors, ValidationError{
			Message:  modelErr.Error(),
			Severity: "error",
		})
		result.Valid = false
	}

	// Check for circular references
	if v3Model != nil && v3Model.Index != nil {
		circularRefs := v3Model.Index.GetCircularReferences()
		for _, ref := range circularRefs {
			result.Warnings = append(result.Warnings, ValidationError{
				Message:  fmt.Sprintf("circular reference detected: %s", ref.Start.Definition),
				Severity: "warning",
			})
		}
	}

	return &doc, result, nil
}
