// Code generated from ir.v1.schema.json. DO NOT EDIT.

package ir

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// IRRecord represents a single HTTP request/response record with optional documentation metadata.
type IRRecord struct {
	// Unique identifier for this record (e.g., UUID).
	Id *string `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id,omitempty"`

	// ISO 8601 timestamp of the request capture.
	Timestamp *time.Time `json:"timestamp,omitempty" yaml:"timestamp,omitempty" mapstructure:"timestamp,omitempty"`

	// Adapter/source that generated this record.
	Source *IRRecordSource `json:"source,omitempty" yaml:"source,omitempty" mapstructure:"source,omitempty"`

	// Request corresponds to the JSON schema field "request".
	Request Request `json:"request" yaml:"request" mapstructure:"request"`

	// Response corresponds to the JSON schema field "response".
	Response Response `json:"response" yaml:"response" mapstructure:"response"`

	// Round-trip time in milliseconds.
	DurationMs *float64 `json:"durationMs,omitempty" yaml:"durationMs,omitempty" mapstructure:"durationMs,omitempty"`

	// Explicit operation identifier (e.g., getUserById). Must be valid identifier.
	OperationId *string `json:"operationId,omitempty" yaml:"operationId,omitempty" mapstructure:"operationId,omitempty"`

	// Short one-line summary of the operation.
	Summary *string `json:"summary,omitempty" yaml:"summary,omitempty" mapstructure:"summary,omitempty"`

	// Full description of the operation. Supports markdown.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// Tags for grouping operations (e.g., from Postman folders).
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty" mapstructure:"tags,omitempty"`

	// Whether this operation is deprecated.
	Deprecated *bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty" mapstructure:"deprecated,omitempty"`

	// Reference to external documentation.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty" mapstructure:"externalDocs,omitempty"`
}

// IRRecordSource represents the adapter/source that generated a record.
type IRRecordSource string

const (
	IRRecordSourceHar              IRRecordSource = "har"
	IRRecordSourcePlaywright       IRRecordSource = "playwright"
	IRRecordSourceLoggingTransport IRRecordSource = "logging-transport"
	IRRecordSourceProxy            IRRecordSource = "proxy"
	IRRecordSourceManual           IRRecordSource = "manual"
	IRRecordSourcePostman          IRRecordSource = "postman"
	IRRecordSourceInsomnia         IRRecordSource = "insomnia"
	IRRecordSourceOpenAPI          IRRecordSource = "openapi"
	IRRecordSourceSwagger          IRRecordSource = "swagger"
)

var enumValues_IRRecordSource = []interface{}{
	"har",
	"playwright",
	"logging-transport",
	"proxy",
	"manual",
	"postman",
	"insomnia",
	"openapi",
	"swagger",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *IRRecordSource) UnmarshalJSON(value []byte) error {
	var v string
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_IRRecordSource {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_IRRecordSource, v)
	}
	*j = IRRecordSource(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *IRRecord) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["request"]; raw != nil && !ok {
		return fmt.Errorf("field request in IRRecord: required")
	}
	if _, ok := raw["response"]; raw != nil && !ok {
		return fmt.Errorf("field response in IRRecord: required")
	}
	type Plain IRRecord
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if plain.DurationMs != nil && 0 > *plain.DurationMs {
		return fmt.Errorf("field %s: must be >= %v", "durationMs", 0)
	}
	*j = IRRecord(plain)
	return nil
}

// Request represents HTTP request details.
type Request struct {
	// HTTP method.
	Method RequestMethod `json:"method" yaml:"method" mapstructure:"method"`

	// URL scheme.
	Scheme RequestScheme `json:"scheme,omitempty" yaml:"scheme,omitempty" mapstructure:"scheme,omitempty"`

	// Request host (e.g., api.example.com).
	Host *string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host,omitempty"`

	// Raw request path without query string (e.g., /users/123).
	Path string `json:"path" yaml:"path" mapstructure:"path"`

	// Normalized path template with parameters (e.g., /users/{id}). Optional - can be inferred.
	PathTemplate *string `json:"pathTemplate,omitempty" yaml:"pathTemplate,omitempty" mapstructure:"pathTemplate,omitempty"`

	// Extracted path parameter values.
	PathParams map[string]string `json:"pathParams,omitempty" yaml:"pathParams,omitempty" mapstructure:"pathParams,omitempty"`

	// Query parameters as key/value pairs.
	Query map[string]interface{} `json:"query,omitempty" yaml:"query,omitempty" mapstructure:"query,omitempty"`

	// Request headers (keys should be lowercase).
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" mapstructure:"headers,omitempty"`

	// Content-Type header value (e.g., application/json).
	ContentType *string `json:"contentType,omitempty" yaml:"contentType,omitempty" mapstructure:"contentType,omitempty"`

	// Parsed request body. Object/array for JSON, string for other content types, null for no body.
	Body interface{} `json:"body,omitempty" yaml:"body,omitempty" mapstructure:"body,omitempty"`
}

// RequestMethod represents the HTTP method.
type RequestMethod string

const (
	RequestMethodGET     RequestMethod = "GET"
	RequestMethodPOST    RequestMethod = "POST"
	RequestMethodPUT     RequestMethod = "PUT"
	RequestMethodPATCH   RequestMethod = "PATCH"
	RequestMethodDELETE  RequestMethod = "DELETE"
	RequestMethodHEAD    RequestMethod = "HEAD"
	RequestMethodOPTIONS RequestMethod = "OPTIONS"
	RequestMethodTRACE   RequestMethod = "TRACE"
)

var enumValues_RequestMethod = []interface{}{
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"HEAD",
	"OPTIONS",
	"TRACE",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *RequestMethod) UnmarshalJSON(value []byte) error {
	var v string
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_RequestMethod {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_RequestMethod, v)
	}
	*j = RequestMethod(v)
	return nil
}

// RequestScheme represents the URL scheme.
type RequestScheme string

const (
	RequestSchemeHTTP  RequestScheme = "http"
	RequestSchemeHTTPS RequestScheme = "https"
)

var enumValues_RequestScheme = []interface{}{
	"http",
	"https",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *RequestScheme) UnmarshalJSON(value []byte) error {
	var v string
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_RequestScheme {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_RequestScheme, v)
	}
	*j = RequestScheme(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Request) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["method"]; raw != nil && !ok {
		return fmt.Errorf("field method in Request: required")
	}
	if _, ok := raw["path"]; raw != nil && !ok {
		return fmt.Errorf("field path in Request: required")
	}
	type Plain Request
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if v, ok := raw["scheme"]; !ok || v == nil {
		plain.Scheme = "https"
	}
	*j = Request(plain)
	return nil
}

// Response represents HTTP response details.
type Response struct {
	// HTTP status code.
	Status int `json:"status" yaml:"status" mapstructure:"status"`

	// Response headers (keys should be lowercase).
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" mapstructure:"headers,omitempty"`

	// Content-Type header value.
	ContentType *string `json:"contentType,omitempty" yaml:"contentType,omitempty" mapstructure:"contentType,omitempty"`

	// Parsed response body. Object/array for JSON, string for other content types, null for no body.
	Body interface{} `json:"body,omitempty" yaml:"body,omitempty" mapstructure:"body,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Response) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["status"]; raw != nil && !ok {
		return fmt.Errorf("field status in Response: required")
	}
	type Plain Response
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 599 < plain.Status {
		return fmt.Errorf("field %s: must be <= %v", "status", 599)
	}
	if 100 > plain.Status {
		return fmt.Errorf("field %s: must be >= %v", "status", 100)
	}
	*j = Response(plain)
	return nil
}

// ExternalDocs represents a reference to external documentation.
type ExternalDocs struct {
	// URL to external documentation.
	URL string `json:"url" yaml:"url" mapstructure:"url"`

	// Description of the external documentation.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ExternalDocs) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["url"]; raw != nil && !ok {
		return fmt.Errorf("field url in ExternalDocs: required")
	}
	type Plain ExternalDocs
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = ExternalDocs(plain)
	return nil
}

// Contact represents contact information for the API.
type Contact struct {
	// Contact name.
	Name *string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`

	// Contact URL.
	URL *string `json:"url,omitempty" yaml:"url,omitempty" mapstructure:"url,omitempty"`

	// Contact email.
	Email *string `json:"email,omitempty" yaml:"email,omitempty" mapstructure:"email,omitempty"`
}

// License represents license information for the API.
type License struct {
	// License name (e.g., MIT, Apache-2.0).
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// URL to the license text.
	URL *string `json:"url,omitempty" yaml:"url,omitempty" mapstructure:"url,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *License) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in License: required")
	}
	type Plain License
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = License(plain)
	return nil
}

// Server represents a server endpoint for the API.
type Server struct {
	// Server URL. May contain variables in {braces}.
	URL string `json:"url" yaml:"url" mapstructure:"url"`

	// Description of the server (e.g., Production, Staging).
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Server) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["url"]; raw != nil && !ok {
		return fmt.Errorf("field url in Server: required")
	}
	type Plain Server
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = Server(plain)
	return nil
}

// TagDefinition represents a tag definition with metadata.
type TagDefinition struct {
	// Tag name.
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Tag description. Supports markdown.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// Reference to external documentation.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty" mapstructure:"externalDocs,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *TagDefinition) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in TagDefinition: required")
	}
	type Plain TagDefinition
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = TagDefinition(plain)
	return nil
}

// APIMetadata represents API-level metadata for documentation generation.
type APIMetadata struct {
	// When this batch was generated.
	GeneratedAt *time.Time `json:"generatedAt,omitempty" yaml:"generatedAt,omitempty" mapstructure:"generatedAt,omitempty"`

	// Source system or adapter name.
	Source *string `json:"source,omitempty" yaml:"source,omitempty" mapstructure:"source,omitempty"`

	// Number of records in this batch.
	RecordCount *int `json:"recordCount,omitempty" yaml:"recordCount,omitempty" mapstructure:"recordCount,omitempty"`

	// API title (e.g., Saviynt EIC API).
	Title *string `json:"title,omitempty" yaml:"title,omitempty" mapstructure:"title,omitempty"`

	// API description. Supports markdown.
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// API version (e.g., 5.0.0). Distinct from IR schema version.
	APIVersion *string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty" mapstructure:"apiVersion,omitempty"`

	// URL to terms of service.
	TermsOfService *string `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty" mapstructure:"termsOfService,omitempty"`

	// Contact information for the API.
	Contact *Contact `json:"contact,omitempty" yaml:"contact,omitempty" mapstructure:"contact,omitempty"`

	// License information for the API.
	License *License `json:"license,omitempty" yaml:"license,omitempty" mapstructure:"license,omitempty"`

	// List of server endpoints.
	Servers []Server `json:"servers,omitempty" yaml:"servers,omitempty" mapstructure:"servers,omitempty"`

	// Tag definitions with descriptions (from Postman folders, etc.).
	TagDefinitions []TagDefinition `json:"tagDefinitions,omitempty" yaml:"tagDefinitions,omitempty" mapstructure:"tagDefinitions,omitempty"`

	// Reference to external documentation.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty" mapstructure:"externalDocs,omitempty"`
}
