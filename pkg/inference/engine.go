package inference

import (
	"io"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

// Engine orchestrates the inference process.
type Engine struct {
	clusterer   *EndpointClusterer
	options     EngineOptions
	apiMetadata *APIMetadataData
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

	// Extract documentation fields
	var docs *RecordDocumentation
	if record.OperationId != nil || record.Summary != nil || record.Description != nil ||
		len(record.Tags) > 0 || (record.Deprecated != nil && *record.Deprecated) || record.ExternalDocs != nil {
		docs = &RecordDocumentation{}
		if record.OperationId != nil {
			docs.OperationID = *record.OperationId
		}
		if record.Summary != nil {
			docs.Summary = *record.Summary
		}
		if record.Description != nil {
			docs.Description = *record.Description
		}
		if len(record.Tags) > 0 {
			docs.Tags = record.Tags
		}
		if record.Deprecated != nil && *record.Deprecated {
			docs.Deprecated = true
		}
		if record.ExternalDocs != nil {
			docs.ExternalDocs = &ExternalDocsData{
				URL:         record.ExternalDocs.URL,
				Description: "",
			}
			if record.ExternalDocs.Description != nil {
				docs.ExternalDocs.Description = *record.ExternalDocs.Description
			}
		}
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
		docs,
	)
}

// SetAPIMetadata sets API-level metadata from IR batch metadata.
func (e *Engine) SetAPIMetadata(metadata *APIMetadataData) {
	e.apiMetadata = metadata
}

// SetAPIMetadataFromIR extracts API metadata from an IR APIMetadata struct.
func (e *Engine) SetAPIMetadataFromIR(irMeta *ir.APIMetadata) {
	if irMeta == nil {
		return
	}

	meta := &APIMetadataData{}

	if irMeta.Title != nil {
		meta.Title = *irMeta.Title
	}
	if irMeta.Description != nil {
		meta.Description = *irMeta.Description
	}
	if irMeta.APIVersion != nil {
		meta.APIVersion = *irMeta.APIVersion
	}
	if irMeta.TermsOfService != nil {
		meta.TermsOfService = *irMeta.TermsOfService
	}
	if irMeta.Contact != nil {
		if irMeta.Contact.Name != nil {
			meta.ContactName = *irMeta.Contact.Name
		}
		if irMeta.Contact.Email != nil {
			meta.ContactEmail = *irMeta.Contact.Email
		}
		if irMeta.Contact.URL != nil {
			meta.ContactURL = *irMeta.Contact.URL
		}
	}
	if irMeta.License != nil {
		meta.LicenseName = irMeta.License.Name
		if irMeta.License.URL != nil {
			meta.LicenseURL = *irMeta.License.URL
		}
	}
	if irMeta.ExternalDocs != nil {
		meta.ExternalDocs = &ExternalDocsData{
			URL: irMeta.ExternalDocs.URL,
		}
		if irMeta.ExternalDocs.Description != nil {
			meta.ExternalDocs.Description = *irMeta.ExternalDocs.Description
		}
	}

	// Convert tag definitions
	for _, td := range irMeta.TagDefinitions {
		tagDef := TagDefinitionData{
			Name: td.Name,
		}
		if td.Description != nil {
			tagDef.Description = *td.Description
		}
		if td.ExternalDocs != nil {
			tagDef.ExternalDocs = &ExternalDocsData{
				URL: td.ExternalDocs.URL,
			}
			if td.ExternalDocs.Description != nil {
				tagDef.ExternalDocs.Description = *td.ExternalDocs.Description
			}
		}
		meta.TagDefinitions = append(meta.TagDefinitions, tagDef)
	}

	e.apiMetadata = meta
}

// Finalize completes the inference process.
func (e *Engine) Finalize() *InferenceResult {
	e.clusterer.Finalize()
	result := e.clusterer.GetResult()
	result.APIMetadata = e.apiMetadata
	return result
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
