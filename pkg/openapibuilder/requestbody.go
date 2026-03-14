package openapibuilder

import (
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// RequestBodyBuilder builds an OpenAPI RequestBody object.
type RequestBodyBuilder struct {
	body    *openapi.RequestBody
	parent  *OperationBuilder
	version Version
}

// newRequestBodyBuilder creates a new request body builder.
func newRequestBodyBuilder(parent *OperationBuilder) *RequestBodyBuilder {
	return &RequestBodyBuilder{
		body: &openapi.RequestBody{
			Content: make(map[string]openapi.MediaType),
		},
		parent:  parent,
		version: Version310,
	}
}

// Description sets the request body description.
func (b *RequestBodyBuilder) Description(d string) *RequestBodyBuilder {
	b.body.Description = d
	return b
}

// Required marks the request body as required.
func (b *RequestBodyBuilder) Required() *RequestBodyBuilder {
	b.body.Required = true
	return b
}

// Content adds a media type to the request body.
func (b *RequestBodyBuilder) Content(mediaType string, schema *SchemaBuilder) *RequestBodyBuilder {
	mt := openapi.MediaType{}
	if schema != nil {
		mt.Schema = schema.Build()
	}
	b.body.Content[mediaType] = mt
	return b
}

// JSON adds application/json content.
func (b *RequestBodyBuilder) JSON(schema *SchemaBuilder) *RequestBodyBuilder {
	return b.Content("application/json", schema)
}

// XML adds application/xml content.
func (b *RequestBodyBuilder) XML(schema *SchemaBuilder) *RequestBodyBuilder {
	return b.Content("application/xml", schema)
}

// FormURLEncoded adds application/x-www-form-urlencoded content.
func (b *RequestBodyBuilder) FormURLEncoded(schema *SchemaBuilder) *RequestBodyBuilder {
	return b.Content("application/x-www-form-urlencoded", schema)
}

// MultipartFormData adds multipart/form-data content.
func (b *RequestBodyBuilder) MultipartFormData(schema *SchemaBuilder) *RequestBodyBuilder {
	return b.Content("multipart/form-data", schema)
}

// TextPlain adds text/plain content.
func (b *RequestBodyBuilder) TextPlain(schema *SchemaBuilder) *RequestBodyBuilder {
	return b.Content("text/plain", schema)
}

// OctetStream adds application/octet-stream content.
func (b *RequestBodyBuilder) OctetStream(schema *SchemaBuilder) *RequestBodyBuilder {
	return b.Content("application/octet-stream", schema)
}

// WithExample adds an example to the most recently added content type.
func (b *RequestBodyBuilder) WithExample(example any) *RequestBodyBuilder {
	// Add to the last added content type
	for mediaType, mt := range b.body.Content {
		mt.Example = example
		b.body.Content[mediaType] = mt
		break
	}
	return b
}

// Done returns to the parent OperationBuilder.
func (b *RequestBodyBuilder) Done() *OperationBuilder {
	if b.parent != nil {
		b.parent.setRequestBody(b.body)
	}
	return b.parent
}

// Build returns the constructed request body.
func (b *RequestBodyBuilder) Build() *openapi.RequestBody {
	return b.body
}

// StandaloneRequestBodyBuilder builds request bodies without a parent.
type StandaloneRequestBodyBuilder struct {
	body    *openapi.RequestBody
	version Version
}

// NewRequestBody creates a standalone request body builder.
func NewRequestBody() *StandaloneRequestBodyBuilder {
	return &StandaloneRequestBodyBuilder{
		body: &openapi.RequestBody{
			Content: make(map[string]openapi.MediaType),
		},
		version: Version310,
	}
}

// Description sets the request body description.
func (b *StandaloneRequestBodyBuilder) Description(d string) *StandaloneRequestBodyBuilder {
	b.body.Description = d
	return b
}

// Required marks the request body as required.
func (b *StandaloneRequestBodyBuilder) Required() *StandaloneRequestBodyBuilder {
	b.body.Required = true
	return b
}

// Content adds a media type to the request body.
func (b *StandaloneRequestBodyBuilder) Content(mediaType string, schema *SchemaBuilder) *StandaloneRequestBodyBuilder {
	mt := openapi.MediaType{}
	if schema != nil {
		mt.Schema = schema.Build()
	}
	b.body.Content[mediaType] = mt
	return b
}

// JSON adds application/json content.
func (b *StandaloneRequestBodyBuilder) JSON(schema *SchemaBuilder) *StandaloneRequestBodyBuilder {
	return b.Content("application/json", schema)
}

// Build returns the constructed request body.
func (b *StandaloneRequestBodyBuilder) Build() *openapi.RequestBody {
	return b.body
}
