// Package ir provides types and utilities for the Traffic2OpenAPI Intermediate Representation.
package ir

import (
	"time"
)

// Version is the current IR schema version.
const Version = "ir.v1"

// Batch represents a collection of IR records with metadata.
// This is the wrapper format for JSON batch files.
type Batch struct {
	Version  string         `json:"version"`
	Metadata *BatchMetadata `json:"metadata,omitempty"`
	Records  []IRRecord     `json:"records"`
}

// BatchMetadata contains optional metadata about a batch of records.
type BatchMetadata struct {
	GeneratedAt *time.Time `json:"generatedAt,omitempty"`
	Source      string     `json:"source,omitempty"`
	RecordCount int        `json:"recordCount,omitempty"`
}

// NewBatch creates a new batch with the current version.
func NewBatch(records []IRRecord) *Batch {
	now := time.Now().UTC()
	return &Batch{
		Version: Version,
		Metadata: &BatchMetadata{
			GeneratedAt: &now,
			RecordCount: len(records),
		},
		Records: records,
	}
}

// NewRecord creates a new IR record with required fields.
func NewRecord(method RequestMethod, path string, status int) *IRRecord {
	return &IRRecord{
		Request: Request{
			Method: method,
			Path:   path,
		},
		Response: Response{
			Status: status,
		},
	}
}

// SetID sets the record ID and returns the record for chaining.
func (r *IRRecord) SetID(id string) *IRRecord {
	r.Id = &id
	return r
}

// SetTimestamp sets the timestamp and returns the record for chaining.
func (r *IRRecord) SetTimestamp(t time.Time) *IRRecord {
	r.Timestamp = &t
	return r
}

// SetSource sets the source adapter type and returns the record for chaining.
func (r *IRRecord) SetSource(source IRRecordSource) *IRRecord {
	r.Source = &source
	return r
}

// SetHost sets the request host and returns the record for chaining.
func (r *IRRecord) SetHost(host string) *IRRecord {
	r.Request.Host = &host
	return r
}

// SetRequestBody sets the request body and returns the record for chaining.
func (r *IRRecord) SetRequestBody(body interface{}) *IRRecord {
	r.Request.Body = body
	return r
}

// SetResponseBody sets the response body and returns the record for chaining.
func (r *IRRecord) SetResponseBody(body interface{}) *IRRecord {
	r.Response.Body = body
	return r
}

// SetQuery sets query parameters and returns the record for chaining.
func (r *IRRecord) SetQuery(query map[string]interface{}) *IRRecord {
	r.Request.Query = query
	return r
}

// SetPathTemplate sets the path template and parameters.
func (r *IRRecord) SetPathTemplate(template string, params map[string]string) *IRRecord {
	r.Request.PathTemplate = &template
	r.Request.PathParams = params
	return r
}

// SetDuration sets the duration in milliseconds.
func (r *IRRecord) SetDuration(ms float64) *IRRecord {
	r.DurationMs = &ms
	return r
}

// EffectivePathTemplate returns pathTemplate if set, otherwise returns path.
func (r *IRRecord) EffectivePathTemplate() string {
	if r.Request.PathTemplate != nil && *r.Request.PathTemplate != "" {
		return *r.Request.PathTemplate
	}
	return r.Request.Path
}

// MethodString returns the method as a string.
func (r *IRRecord) MethodString() string {
	return string(r.Request.Method)
}
