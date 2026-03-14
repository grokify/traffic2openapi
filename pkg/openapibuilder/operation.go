package openapibuilder

import (
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// OperationBuilder builds an OpenAPI Operation object.
type OperationBuilder struct {
	operation *openapi.Operation
	method    string
	parent    *PathItemBuilder
	version   Version
}

// newOperationBuilder creates a new operation builder.
func newOperationBuilder(method string, parent *PathItemBuilder) *OperationBuilder {
	return &OperationBuilder{
		operation: &openapi.Operation{
			Responses: make(map[string]openapi.Response),
		},
		method:  method,
		parent:  parent,
		version: Version310,
	}
}

// Summary sets the operation summary.
func (b *OperationBuilder) Summary(s string) *OperationBuilder {
	b.operation.Summary = s
	return b
}

// Description sets the operation description.
func (b *OperationBuilder) Description(d string) *OperationBuilder {
	b.operation.Description = d
	return b
}

// OperationID sets the operation ID.
func (b *OperationBuilder) OperationID(id string) *OperationBuilder {
	b.operation.OperationID = id
	return b
}

// Tags adds tags to the operation.
func (b *OperationBuilder) Tags(tags ...string) *OperationBuilder {
	b.operation.Tags = append(b.operation.Tags, tags...)
	return b
}

// Deprecated marks the operation as deprecated.
func (b *OperationBuilder) Deprecated() *OperationBuilder {
	b.operation.Deprecated = true
	return b
}

// PathParam adds a path parameter to the operation.
func (b *OperationBuilder) PathParam(name string) *ParameterBuilder {
	return newParameterBuilder(name, "path", b)
}

// QueryParam adds a query parameter to the operation.
func (b *OperationBuilder) QueryParam(name string) *ParameterBuilder {
	return newParameterBuilder(name, "query", b)
}

// HeaderParam adds a header parameter to the operation.
func (b *OperationBuilder) HeaderParam(name string) *ParameterBuilder {
	return newParameterBuilder(name, "header", b)
}

// CookieParam adds a cookie parameter to the operation.
func (b *OperationBuilder) CookieParam(name string) *ParameterBuilder {
	return newParameterBuilder(name, "cookie", b)
}

// Parameter adds a pre-built parameter to the operation.
func (b *OperationBuilder) Parameter(param *openapi.Parameter) *OperationBuilder {
	b.operation.Parameters = append(b.operation.Parameters, *param)
	return b
}

// RequestBody starts building a request body.
func (b *OperationBuilder) RequestBody() *RequestBodyBuilder {
	return newRequestBodyBuilder(b)
}

// JSONBody is a shortcut for a required JSON request body.
func (b *OperationBuilder) JSONBody(schema *SchemaBuilder) *OperationBuilder {
	body := &openapi.RequestBody{
		Required: true,
		Content:  make(map[string]openapi.MediaType),
	}
	mt := openapi.MediaType{}
	if schema != nil {
		mt.Schema = schema.Build()
	}
	body.Content["application/json"] = mt
	b.operation.RequestBody = body
	return b
}

// Response adds a response for the given status code.
func (b *OperationBuilder) Response(statusCode int) *ResponseBuilder {
	return newResponseBuilder(statusCode, b)
}

// ResponseDefault adds a default response.
func (b *OperationBuilder) ResponseDefault() *ResponseBuilder {
	return newResponseBuilderString("default", b)
}

// Response2XX adds a 2XX response.
func (b *OperationBuilder) Response2XX() *ResponseBuilder {
	return newResponseBuilderString("2XX", b)
}

// Response4XX adds a 4XX response.
func (b *OperationBuilder) Response4XX() *ResponseBuilder {
	return newResponseBuilderString("4XX", b)
}

// Response5XX adds a 5XX response.
func (b *OperationBuilder) Response5XX() *ResponseBuilder {
	return newResponseBuilderString("5XX", b)
}

// Security adds a security requirement.
func (b *OperationBuilder) Security(name string, scopes ...string) *OperationBuilder {
	req := openapi.SecurityRequirement{name: scopes}
	b.operation.Security = append(b.operation.Security, req)
	return b
}

// NoSecurity explicitly removes security requirements (empty array).
func (b *OperationBuilder) NoSecurity() *OperationBuilder {
	b.operation.Security = []openapi.SecurityRequirement{}
	return b
}

// Done returns to the parent PathItemBuilder.
func (b *OperationBuilder) Done() *PathItemBuilder {
	if b.parent != nil {
		b.parent.setOperation(b.method, b.operation)
	}
	return b.parent
}

// Build returns the constructed operation.
func (b *OperationBuilder) Build() *openapi.Operation {
	return b.operation
}

// addParameter adds a parameter to the operation (called by ParameterBuilder).
func (b *OperationBuilder) addParameter(param *openapi.Parameter) {
	b.operation.Parameters = append(b.operation.Parameters, *param)
}

// setRequestBody sets the request body (called by RequestBodyBuilder).
func (b *OperationBuilder) setRequestBody(body *openapi.RequestBody) {
	b.operation.RequestBody = body
}

// addResponse adds a response (called by ResponseBuilder).
func (b *OperationBuilder) addResponse(statusCode string, response *openapi.Response) {
	b.operation.Responses[statusCode] = *response
}
