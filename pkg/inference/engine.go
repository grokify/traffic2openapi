package inference

import (
	"io"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

// Engine orchestrates the inference process.
type Engine struct {
	clusterer *EndpointClusterer
	options   EngineOptions
}

// EngineOptions configures the inference engine.
type EngineOptions struct {
	// IncludeErrorResponses includes 4xx/5xx responses in the spec
	IncludeErrorResponses bool

	// MinStatusCode is the minimum status code to include (default: 100)
	MinStatusCode int

	// MaxStatusCode is the maximum status code to include (default: 599)
	MaxStatusCode int

	// SkipEmptyBodies skips recording empty request/response bodies
	SkipEmptyBodies bool
}

// DefaultEngineOptions returns the default engine options.
func DefaultEngineOptions() EngineOptions {
	return EngineOptions{
		IncludeErrorResponses: true,
		MinStatusCode:         100,
		MaxStatusCode:         599,
		SkipEmptyBodies:       false,
	}
}

// NewEngine creates a new inference engine.
func NewEngine(options EngineOptions) *Engine {
	return &Engine{
		clusterer: NewEndpointClusterer(),
		options:   options,
	}
}

// ProcessRecords processes a slice of IR records.
func (e *Engine) ProcessRecords(records []ir.IRRecord) {
	for i := range records {
		e.ProcessRecord(&records[i])
	}
}

// ProcessReader processes all records from an IRReader.
// Reads until io.EOF is returned.
func (e *Engine) ProcessReader(reader ir.IRReader) error {
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		e.ProcessRecord(record)
	}
}

// ProcessRecord processes a single IR record.
func (e *Engine) ProcessRecord(record *ir.IRRecord) {
	// Skip if status code out of range
	status := record.Response.Status
	if status < e.options.MinStatusCode || status > e.options.MaxStatusCode {
		return
	}

	// Skip error responses if configured
	if !e.options.IncludeErrorResponses && status >= 400 {
		return
	}

	// Extract fields from record
	method := string(record.Request.Method)
	path := record.Request.Path

	var pathTemplate string
	if record.Request.PathTemplate != nil {
		pathTemplate = *record.Request.PathTemplate
	}

	pathParams := record.Request.PathParams

	// Convert query params
	query := make(map[string]any)
	for k, v := range record.Request.Query {
		query[k] = v
	}

	// Get headers
	headers := record.Request.Headers

	// Get request body
	var requestBody any
	if !e.options.SkipEmptyBodies || record.Request.Body != nil {
		requestBody = record.Request.Body
	}

	var requestContentType string
	if record.Request.ContentType != nil {
		requestContentType = *record.Request.ContentType
	}

	// Get response body
	var responseBody any
	if !e.options.SkipEmptyBodies || record.Response.Body != nil {
		responseBody = record.Response.Body
	}

	var responseContentType string
	if record.Response.ContentType != nil {
		responseContentType = *record.Response.ContentType
	}

	// Get response headers
	responseHeaders := record.Response.Headers

	// Get host and scheme
	var host string
	if record.Request.Host != nil {
		host = *record.Request.Host
	}

	scheme := string(record.Request.Scheme)
	if scheme == "" {
		scheme = "https"
	}

	// Add to clusterer
	e.clusterer.AddRecord(
		method,
		path,
		pathTemplate,
		pathParams,
		query,
		headers,
		requestBody,
		requestContentType,
		status,
		responseBody,
		responseContentType,
		responseHeaders,
		host,
		scheme,
	)
}

// Finalize completes the inference process.
func (e *Engine) Finalize() *InferenceResult {
	e.clusterer.Finalize()
	return e.clusterer.GetResult()
}

// InferFromRecords is a convenience function that processes records and returns results.
func InferFromRecords(records []ir.IRRecord) *InferenceResult {
	engine := NewEngine(DefaultEngineOptions())
	engine.ProcessRecords(records)
	return engine.Finalize()
}

// InferFromReader is a convenience function that processes records from an IRReader.
func InferFromReader(reader ir.IRReader) (*InferenceResult, error) {
	engine := NewEngine(DefaultEngineOptions())
	if err := engine.ProcessReader(reader); err != nil {
		return nil, err
	}
	return engine.Finalize(), nil
}

// InferFromFile reads an IR file and returns inference results.
func InferFromFile(path string) (*InferenceResult, error) {
	records, err := ir.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return InferFromRecords(records), nil
}

// InferFromDir reads all IR files from a directory and returns inference results.
func InferFromDir(dir string) (*InferenceResult, error) {
	records, err := ir.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	return InferFromRecords(records), nil
}
