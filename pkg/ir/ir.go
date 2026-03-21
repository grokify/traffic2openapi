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
	Version  string       `json:"version"`
	Metadata *APIMetadata `json:"metadata,omitempty"`
	Records  []IRRecord   `json:"records"`
}

// BatchMetadata is an alias for APIMetadata for backward compatibility.
// Deprecated: Use APIMetadata instead.
type BatchMetadata = APIMetadata

// NewBatch creates a new batch with the current version.
func NewBatch(records []IRRecord) *Batch {
	now := time.Now().UTC()
	count := len(records)
	return &Batch{
		Version: Version,
		Metadata: &APIMetadata{
			GeneratedAt: &now,
			RecordCount: &count,
		},
		Records: records,
	}
}

// NewBatchWithMetadata creates a new batch with custom metadata.
func NewBatchWithMetadata(records []IRRecord, metadata *APIMetadata) *Batch {
	if metadata == nil {
		return NewBatch(records)
	}
	now := time.Now().UTC()
	count := len(records)
	if metadata.GeneratedAt == nil {
		metadata.GeneratedAt = &now
	}
	if metadata.RecordCount == nil {
		metadata.RecordCount = &count
	}
	return &Batch{
		Version:  Version,
		Metadata: metadata,
		Records:  records,
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

// SetScheme sets the request scheme and returns the record for chaining.
func (r *IRRecord) SetScheme(scheme RequestScheme) *IRRecord {
	r.Request.Scheme = scheme
	return r
}

// SetRequestBody sets the request body and returns the record for chaining.
func (r *IRRecord) SetRequestBody(body interface{}) *IRRecord {
	r.Request.Body = body
	return r
}

// SetRequestContentType sets the request content type and returns the record for chaining.
func (r *IRRecord) SetRequestContentType(contentType string) *IRRecord {
	r.Request.ContentType = &contentType
	return r
}

// SetResponseBody sets the response body and returns the record for chaining.
func (r *IRRecord) SetResponseBody(body interface{}) *IRRecord {
	r.Response.Body = body
	return r
}

// SetResponseContentType sets the response content type and returns the record for chaining.
func (r *IRRecord) SetResponseContentType(contentType string) *IRRecord {
	r.Response.ContentType = &contentType
	return r
}

// SetQuery sets query parameters and returns the record for chaining.
func (r *IRRecord) SetQuery(query map[string]interface{}) *IRRecord {
	r.Request.Query = query
	return r
}

// SetRequestHeaders sets request headers and returns the record for chaining.
func (r *IRRecord) SetRequestHeaders(headers map[string]string) *IRRecord {
	r.Request.Headers = headers
	return r
}

// SetResponseHeaders sets response headers and returns the record for chaining.
func (r *IRRecord) SetResponseHeaders(headers map[string]string) *IRRecord {
	r.Response.Headers = headers
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

// SetOperationId sets the operation identifier and returns the record for chaining.
func (r *IRRecord) SetOperationId(operationId string) *IRRecord {
	r.OperationId = &operationId
	return r
}

// SetSummary sets the operation summary and returns the record for chaining.
func (r *IRRecord) SetSummary(summary string) *IRRecord {
	r.Summary = &summary
	return r
}

// SetDescription sets the operation description and returns the record for chaining.
func (r *IRRecord) SetDescription(description string) *IRRecord {
	r.Description = &description
	return r
}

// SetTags sets the operation tags and returns the record for chaining.
func (r *IRRecord) SetTags(tags ...string) *IRRecord {
	r.Tags = tags
	return r
}

// AddTag adds a tag to the operation and returns the record for chaining.
func (r *IRRecord) AddTag(tag string) *IRRecord {
	r.Tags = append(r.Tags, tag)
	return r
}

// SetDeprecated sets whether the operation is deprecated.
func (r *IRRecord) SetDeprecated(deprecated bool) *IRRecord {
	r.Deprecated = &deprecated
	return r
}

// SetExternalDocs sets the external documentation reference.
func (r *IRRecord) SetExternalDocs(url string, description string) *IRRecord {
	r.ExternalDocs = &ExternalDocs{
		URL: url,
	}
	if description != "" {
		r.ExternalDocs.Description = &description
	}
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

// IsDeprecated returns true if the operation is marked as deprecated.
func (r *IRRecord) IsDeprecated() bool {
	return r.Deprecated != nil && *r.Deprecated
}

// APIMetadataBuilder provides a fluent interface for building APIMetadata.
type APIMetadataBuilder struct {
	metadata *APIMetadata
}

// NewAPIMetadata creates a new APIMetadataBuilder.
func NewAPIMetadata() *APIMetadataBuilder {
	return &APIMetadataBuilder{
		metadata: &APIMetadata{},
	}
}

// Title sets the API title.
func (b *APIMetadataBuilder) Title(title string) *APIMetadataBuilder {
	b.metadata.Title = &title
	return b
}

// Description sets the API description.
func (b *APIMetadataBuilder) Description(description string) *APIMetadataBuilder {
	b.metadata.Description = &description
	return b
}

// APIVersion sets the API version.
func (b *APIMetadataBuilder) APIVersion(version string) *APIMetadataBuilder {
	b.metadata.APIVersion = &version
	return b
}

// TermsOfService sets the terms of service URL.
func (b *APIMetadataBuilder) TermsOfService(url string) *APIMetadataBuilder {
	b.metadata.TermsOfService = &url
	return b
}

// Contact sets the contact information.
func (b *APIMetadataBuilder) Contact(name, email, url string) *APIMetadataBuilder {
	b.metadata.Contact = &Contact{}
	if name != "" {
		b.metadata.Contact.Name = &name
	}
	if email != "" {
		b.metadata.Contact.Email = &email
	}
	if url != "" {
		b.metadata.Contact.URL = &url
	}
	return b
}

// License sets the license information.
func (b *APIMetadataBuilder) License(name, url string) *APIMetadataBuilder {
	b.metadata.License = &License{Name: name}
	if url != "" {
		b.metadata.License.URL = &url
	}
	return b
}

// AddServer adds a server endpoint.
func (b *APIMetadataBuilder) AddServer(url, description string) *APIMetadataBuilder {
	server := Server{URL: url}
	if description != "" {
		server.Description = &description
	}
	b.metadata.Servers = append(b.metadata.Servers, server)
	return b
}

// AddTag adds a tag definition.
func (b *APIMetadataBuilder) AddTag(name, description string) *APIMetadataBuilder {
	tag := TagDefinition{Name: name}
	if description != "" {
		tag.Description = &description
	}
	b.metadata.TagDefinitions = append(b.metadata.TagDefinitions, tag)
	return b
}

// ExternalDocs sets the external documentation reference.
func (b *APIMetadataBuilder) ExternalDocs(url, description string) *APIMetadataBuilder {
	b.metadata.ExternalDocs = &ExternalDocs{URL: url}
	if description != "" {
		b.metadata.ExternalDocs.Description = &description
	}
	return b
}

// Build returns the constructed APIMetadata.
func (b *APIMetadataBuilder) Build() *APIMetadata {
	return b.metadata
}
