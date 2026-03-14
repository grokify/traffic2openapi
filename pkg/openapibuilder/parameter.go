package openapibuilder

import (
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// ParameterBuilder builds an OpenAPI Parameter object.
type ParameterBuilder struct {
	param   *openapi.Parameter
	parent  *OperationBuilder
	version Version
}

// newParameterBuilder creates a new parameter builder.
func newParameterBuilder(name, in string, parent *OperationBuilder) *ParameterBuilder {
	b := &ParameterBuilder{
		param: &openapi.Parameter{
			Name: name,
			In:   in,
		},
		parent:  parent,
		version: Version310,
	}
	// Path parameters are always required
	if in == "path" {
		b.param.Required = true
	}
	return b
}

// Description sets the parameter description.
func (b *ParameterBuilder) Description(d string) *ParameterBuilder {
	b.param.Description = d
	return b
}

// Required marks the parameter as required.
func (b *ParameterBuilder) Required() *ParameterBuilder {
	b.param.Required = true
	return b
}

// Deprecated marks the parameter as deprecated.
func (b *ParameterBuilder) Deprecated() *ParameterBuilder {
	b.param.Deprecated = true
	return b
}

// AllowEmptyValue allows empty values for the parameter.
func (b *ParameterBuilder) AllowEmptyValue() *ParameterBuilder {
	b.param.AllowEmptyValue = true
	return b
}

// Schema sets the parameter schema.
func (b *ParameterBuilder) Schema(schema *SchemaBuilder) *ParameterBuilder {
	if schema != nil {
		b.param.Schema = schema.Build()
	}
	return b
}

// Type sets the parameter type (creates a simple schema).
func (b *ParameterBuilder) Type(t string) *ParameterBuilder {
	if b.param.Schema == nil {
		b.param.Schema = &openapi.Schema{}
	}
	b.param.Schema.Type = t
	return b
}

// Format sets the parameter format.
func (b *ParameterBuilder) Format(f string) *ParameterBuilder {
	if b.param.Schema == nil {
		b.param.Schema = &openapi.Schema{}
	}
	b.param.Schema.Format = f
	return b
}

// Enum sets the allowed values.
func (b *ParameterBuilder) Enum(values ...any) *ParameterBuilder {
	if b.param.Schema == nil {
		b.param.Schema = &openapi.Schema{}
	}
	b.param.Schema.Enum = values
	return b
}

// Example sets an example value for the parameter.
func (b *ParameterBuilder) Example(v any) *ParameterBuilder {
	b.param.Example = v
	return b
}

// DoneOp returns to the parent OperationBuilder.
func (b *ParameterBuilder) DoneOp() *OperationBuilder {
	if b.parent != nil {
		b.parent.addParameter(b.param)
	}
	return b.parent
}

// Build returns the constructed parameter.
func (b *ParameterBuilder) Build() *openapi.Parameter {
	return b.param
}

// StandaloneParameterBuilder builds parameters without a parent operation.
type StandaloneParameterBuilder struct {
	param   *openapi.Parameter
	version Version
}

// PathParam creates a path parameter builder (standalone).
func PathParam(name string) *StandaloneParameterBuilder {
	return &StandaloneParameterBuilder{
		param: &openapi.Parameter{
			Name:     name,
			In:       "path",
			Required: true,
		},
		version: Version310,
	}
}

// QueryParam creates a query parameter builder (standalone).
func QueryParam(name string) *StandaloneParameterBuilder {
	return &StandaloneParameterBuilder{
		param: &openapi.Parameter{
			Name: name,
			In:   "query",
		},
		version: Version310,
	}
}

// HeaderParam creates a header parameter builder (standalone).
func HeaderParam(name string) *StandaloneParameterBuilder {
	return &StandaloneParameterBuilder{
		param: &openapi.Parameter{
			Name: name,
			In:   "header",
		},
		version: Version310,
	}
}

// CookieParam creates a cookie parameter builder (standalone).
func CookieParam(name string) *StandaloneParameterBuilder {
	return &StandaloneParameterBuilder{
		param: &openapi.Parameter{
			Name: name,
			In:   "cookie",
		},
		version: Version310,
	}
}

// Description sets the parameter description.
func (b *StandaloneParameterBuilder) Description(d string) *StandaloneParameterBuilder {
	b.param.Description = d
	return b
}

// Required marks the parameter as required.
func (b *StandaloneParameterBuilder) Required() *StandaloneParameterBuilder {
	b.param.Required = true
	return b
}

// Deprecated marks the parameter as deprecated.
func (b *StandaloneParameterBuilder) Deprecated() *StandaloneParameterBuilder {
	b.param.Deprecated = true
	return b
}

// AllowEmptyValue allows empty values for the parameter.
func (b *StandaloneParameterBuilder) AllowEmptyValue() *StandaloneParameterBuilder {
	b.param.AllowEmptyValue = true
	return b
}

// Schema sets the parameter schema.
func (b *StandaloneParameterBuilder) Schema(schema *SchemaBuilder) *StandaloneParameterBuilder {
	if schema != nil {
		b.param.Schema = schema.Build()
	}
	return b
}

// Type sets the parameter type (creates a simple schema).
func (b *StandaloneParameterBuilder) Type(t string) *StandaloneParameterBuilder {
	if b.param.Schema == nil {
		b.param.Schema = &openapi.Schema{}
	}
	b.param.Schema.Type = t
	return b
}

// Format sets the parameter format.
func (b *StandaloneParameterBuilder) Format(f string) *StandaloneParameterBuilder {
	if b.param.Schema == nil {
		b.param.Schema = &openapi.Schema{}
	}
	b.param.Schema.Format = f
	return b
}

// Enum sets the allowed values.
func (b *StandaloneParameterBuilder) Enum(values ...any) *StandaloneParameterBuilder {
	if b.param.Schema == nil {
		b.param.Schema = &openapi.Schema{}
	}
	b.param.Schema.Enum = values
	return b
}

// Example sets an example value for the parameter.
func (b *StandaloneParameterBuilder) Example(v any) *StandaloneParameterBuilder {
	b.param.Example = v
	return b
}

// Build returns the constructed parameter.
func (b *StandaloneParameterBuilder) Build() *openapi.Parameter {
	return b.param
}
