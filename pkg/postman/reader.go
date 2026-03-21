package postman

import (
	"fmt"
	"io"
	"os"

	postman "github.com/rbretecher/go-postman-collection"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

// ReadFile reads a Postman collection from a file path.
func ReadFile(path string) (*postman.Collection, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return Read(f)
}

// Read reads a Postman collection from an io.Reader.
func Read(r io.Reader) (*postman.Collection, error) {
	collection, err := postman.ParseCollection(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse collection: %w", err)
	}
	return collection, nil
}

// ConvertFile is a convenience function that reads and converts a Postman collection file.
func ConvertFile(path string, opts ...ConverterOption) (*ConvertResult, error) {
	collection, err := ReadFile(path)
	if err != nil {
		return nil, err
	}

	converter := NewConverter()
	for _, opt := range opts {
		opt(converter)
	}

	return converter.Convert(collection)
}

// ConvertFileToRecords is a convenience function that reads and converts to IR records.
func ConvertFileToRecords(path string, opts ...ConverterOption) ([]ir.IRRecord, error) {
	result, err := ConvertFile(path, opts...)
	if err != nil {
		return nil, err
	}
	return result.Records, nil
}

// ConvertFileToBatch is a convenience function that reads and converts to an IR batch.
func ConvertFileToBatch(path string, opts ...ConverterOption) (*ir.Batch, error) {
	collection, err := ReadFile(path)
	if err != nil {
		return nil, err
	}

	converter := NewConverter()
	for _, opt := range opts {
		opt(converter)
	}

	return converter.ConvertToBatch(collection)
}

// ConverterOption is a functional option for configuring the converter.
type ConverterOption func(*Converter)

// WithBaseURL sets the base URL for variable resolution.
func WithBaseURL(url string) ConverterOption {
	return func(c *Converter) {
		c.BaseURL = url
	}
}

// WithVariables sets additional variables for resolution.
func WithVariables(vars map[string]string) ConverterOption {
	return func(c *Converter) {
		for k, v := range vars {
			c.Variables[k] = v
		}
	}
}

// WithVariable sets a single variable for resolution.
func WithVariable(key, value string) ConverterOption {
	return func(c *Converter) {
		c.Variables[key] = value
	}
}

// WithHeaderFilter adds headers to filter out.
func WithHeaderFilter(headers ...string) ConverterOption {
	return func(c *Converter) {
		c.FilterHeaders = append(c.FilterHeaders, headers...)
	}
}

// WithoutHeaders disables header inclusion.
func WithoutHeaders() ConverterOption {
	return func(c *Converter) {
		c.IncludeHeaders = false
	}
}

// WithDisabledItems includes disabled headers and query params.
func WithDisabledItems() ConverterOption {
	return func(c *Converter) {
		c.IncludeDisabled = true
	}
}

// WithoutAuth disables auth-to-header conversion.
func WithoutAuth() ConverterOption {
	return func(c *Converter) {
		c.PreserveAuth = false
	}
}

// WithoutIDs disables automatic ID generation.
func WithoutIDs() ConverterOption {
	return func(c *Converter) {
		c.GenerateIDs = false
	}
}
