package openapibuilder

import (
	"strconv"

	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// ResponseBuilder builds an OpenAPI Response object.
type ResponseBuilder struct {
	response   *openapi.Response
	statusCode string
	parent     *OperationBuilder
	version    Version
}

// newResponseBuilder creates a new response builder.
func newResponseBuilder(statusCode int, parent *OperationBuilder) *ResponseBuilder {
	return &ResponseBuilder{
		response:   &openapi.Response{},
		statusCode: strconv.Itoa(statusCode),
		parent:     parent,
		version:    Version310,
	}
}

// newResponseBuilderString creates a new response builder with string status.
func newResponseBuilderString(status string, parent *OperationBuilder) *ResponseBuilder {
	return &ResponseBuilder{
		response:   &openapi.Response{},
		statusCode: status,
		parent:     parent,
		version:    Version310,
	}
}

// Description sets the response description (required).
func (b *ResponseBuilder) Description(d string) *ResponseBuilder {
	b.response.Description = d
	return b
}

// Content adds a media type to the response.
func (b *ResponseBuilder) Content(mediaType string, schema *SchemaBuilder) *ResponseBuilder {
	if b.response.Content == nil {
		b.response.Content = make(map[string]openapi.MediaType)
	}
	mt := openapi.MediaType{}
	if schema != nil {
		mt.Schema = schema.Build()
	}
	b.response.Content[mediaType] = mt
	return b
}

// JSON adds application/json content.
func (b *ResponseBuilder) JSON(schema *SchemaBuilder) *ResponseBuilder {
	return b.Content("application/json", schema)
}

// XML adds application/xml content.
func (b *ResponseBuilder) XML(schema *SchemaBuilder) *ResponseBuilder {
	return b.Content("application/xml", schema)
}

// TextPlain adds text/plain content.
func (b *ResponseBuilder) TextPlain(schema *SchemaBuilder) *ResponseBuilder {
	return b.Content("text/plain", schema)
}

// OctetStream adds application/octet-stream content.
func (b *ResponseBuilder) OctetStream(schema *SchemaBuilder) *ResponseBuilder {
	return b.Content("application/octet-stream", schema)
}

// HTML adds text/html content.
func (b *ResponseBuilder) HTML(schema *SchemaBuilder) *ResponseBuilder {
	return b.Content("text/html", schema)
}

// Header adds a header to the response.
func (b *ResponseBuilder) Header(name string) *HeaderBuilder {
	return newHeaderBuilder(name, b)
}

// Done returns to the parent OperationBuilder.
func (b *ResponseBuilder) Done() *OperationBuilder {
	if b.parent != nil {
		b.parent.addResponse(b.statusCode, b.response)
	}
	return b.parent
}

// Build returns the constructed response.
func (b *ResponseBuilder) Build() *openapi.Response {
	return b.response
}

// addHeader adds a header to the response.
func (b *ResponseBuilder) addHeader(name string, header *openapi.Header) {
	if b.response.Headers == nil {
		b.response.Headers = make(map[string]openapi.Header)
	}
	b.response.Headers[name] = *header
}

// HeaderBuilder builds an OpenAPI Header object.
type HeaderBuilder struct {
	header *openapi.Header
	name   string
	parent *ResponseBuilder
}

// newHeaderBuilder creates a new header builder.
func newHeaderBuilder(name string, parent *ResponseBuilder) *HeaderBuilder {
	return &HeaderBuilder{
		header: &openapi.Header{},
		name:   name,
		parent: parent,
	}
}

// Description sets the header description.
func (b *HeaderBuilder) Description(d string) *HeaderBuilder {
	b.header.Description = d
	return b
}

// Required marks the header as required.
func (b *HeaderBuilder) Required() *HeaderBuilder {
	b.header.Required = true
	return b
}

// Schema sets the header schema.
func (b *HeaderBuilder) Schema(schema *SchemaBuilder) *HeaderBuilder {
	if schema != nil {
		b.header.Schema = schema.Build()
	}
	return b
}

// Type sets the header type (creates a simple schema).
func (b *HeaderBuilder) Type(t string) *HeaderBuilder {
	if b.header.Schema == nil {
		b.header.Schema = &openapi.Schema{}
	}
	b.header.Schema.Type = t
	return b
}

// Format sets the header format.
func (b *HeaderBuilder) Format(f string) *HeaderBuilder {
	if b.header.Schema == nil {
		b.header.Schema = &openapi.Schema{}
	}
	b.header.Schema.Format = f
	return b
}

// Done returns to the parent ResponseBuilder.
func (b *HeaderBuilder) Done() *ResponseBuilder {
	if b.parent != nil {
		b.parent.addHeader(b.name, b.header)
	}
	return b.parent
}

// Build returns the constructed header.
func (b *HeaderBuilder) Build() *openapi.Header {
	return b.header
}

// StandaloneResponseBuilder builds responses without a parent.
type StandaloneResponseBuilder struct {
	response *openapi.Response
	version  Version
}

// NewResponse creates a standalone response builder.
func NewResponse() *StandaloneResponseBuilder {
	return &StandaloneResponseBuilder{
		response: &openapi.Response{},
		version:  Version310,
	}
}

// Description sets the response description (required).
func (b *StandaloneResponseBuilder) Description(d string) *StandaloneResponseBuilder {
	b.response.Description = d
	return b
}

// Content adds a media type to the response.
func (b *StandaloneResponseBuilder) Content(mediaType string, schema *SchemaBuilder) *StandaloneResponseBuilder {
	if b.response.Content == nil {
		b.response.Content = make(map[string]openapi.MediaType)
	}
	mt := openapi.MediaType{}
	if schema != nil {
		mt.Schema = schema.Build()
	}
	b.response.Content[mediaType] = mt
	return b
}

// JSON adds application/json content.
func (b *StandaloneResponseBuilder) JSON(schema *SchemaBuilder) *StandaloneResponseBuilder {
	return b.Content("application/json", schema)
}

// Build returns the constructed response.
func (b *StandaloneResponseBuilder) Build() *openapi.Response {
	return b.response
}
