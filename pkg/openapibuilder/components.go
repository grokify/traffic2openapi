package openapibuilder

import (
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// ComponentsBuilder builds an OpenAPI Components object.
type ComponentsBuilder struct {
	components *openapi.Components
	parent     *SpecBuilder
	version    Version
}

// newComponentsBuilder creates a new components builder.
func newComponentsBuilder(parent *SpecBuilder) *ComponentsBuilder {
	return &ComponentsBuilder{
		components: &openapi.Components{},
		parent:     parent,
		version:    Version310,
	}
}

// Schema adds a schema to the components.
func (b *ComponentsBuilder) Schema(name string, schema *SchemaBuilder) *ComponentsBuilder {
	if b.components.Schemas == nil {
		b.components.Schemas = make(map[string]*openapi.Schema)
	}
	if schema != nil {
		b.components.Schemas[name] = schema.Build()
	}
	return b
}

// Response adds a response to the components.
func (b *ComponentsBuilder) Response(name string, response *StandaloneResponseBuilder) *ComponentsBuilder {
	if b.components.Responses == nil {
		b.components.Responses = make(map[string]*openapi.Response)
	}
	if response != nil {
		b.components.Responses[name] = response.Build()
	}
	return b
}

// Parameter adds a parameter to the components.
func (b *ComponentsBuilder) Parameter(name string, param *StandaloneParameterBuilder) *ComponentsBuilder {
	if b.components.Parameters == nil {
		b.components.Parameters = make(map[string]*openapi.Parameter)
	}
	if param != nil {
		b.components.Parameters[name] = param.Build()
	}
	return b
}

// RequestBody adds a request body to the components.
func (b *ComponentsBuilder) RequestBody(name string, body *StandaloneRequestBodyBuilder) *ComponentsBuilder {
	if b.components.RequestBodies == nil {
		b.components.RequestBodies = make(map[string]*openapi.RequestBody)
	}
	if body != nil {
		b.components.RequestBodies[name] = body.Build()
	}
	return b
}

// Header adds a header to the components.
func (b *ComponentsBuilder) Header(name string, header *HeaderBuilder) *ComponentsBuilder {
	if b.components.Headers == nil {
		b.components.Headers = make(map[string]*openapi.Header)
	}
	if header != nil {
		b.components.Headers[name] = header.Build()
	}
	return b
}

// SecurityScheme starts building a security scheme.
func (b *ComponentsBuilder) SecurityScheme(name string) *SecuritySchemeBuilder {
	return newSecuritySchemeBuilder(name, b)
}

// AddSecurityScheme adds a pre-built security scheme.
func (b *ComponentsBuilder) AddSecurityScheme(name string, scheme *StandaloneSecuritySchemeBuilder) *ComponentsBuilder {
	if b.components.SecuritySchemes == nil {
		b.components.SecuritySchemes = make(map[string]*openapi.SecurityScheme)
	}
	if scheme != nil {
		b.components.SecuritySchemes[name] = scheme.Build()
	}
	return b
}

// Done returns to the parent SpecBuilder.
func (b *ComponentsBuilder) Done() *SpecBuilder {
	if b.parent != nil {
		b.parent.setComponents(b.components)
	}
	return b.parent
}

// Build returns the constructed components.
func (b *ComponentsBuilder) Build() *openapi.Components {
	return b.components
}

// addSecurityScheme adds a security scheme (called by SecuritySchemeBuilder).
func (b *ComponentsBuilder) addSecurityScheme(name string, scheme *openapi.SecurityScheme) {
	if b.components.SecuritySchemes == nil {
		b.components.SecuritySchemes = make(map[string]*openapi.SecurityScheme)
	}
	b.components.SecuritySchemes[name] = scheme
}
