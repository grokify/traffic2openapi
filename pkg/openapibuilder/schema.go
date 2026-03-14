package openapibuilder

import (
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// SchemaBuilder builds an OpenAPI Schema object.
type SchemaBuilder struct {
	schema  *openapi.Schema
	version Version
}

// StringSchema creates a new string schema builder.
func StringSchema() *SchemaBuilder {
	return &SchemaBuilder{
		schema:  &openapi.Schema{Type: "string"},
		version: Version310,
	}
}

// IntegerSchema creates a new integer schema builder.
func IntegerSchema() *SchemaBuilder {
	return &SchemaBuilder{
		schema:  &openapi.Schema{Type: "integer"},
		version: Version310,
	}
}

// NumberSchema creates a new number schema builder.
func NumberSchema() *SchemaBuilder {
	return &SchemaBuilder{
		schema:  &openapi.Schema{Type: "number"},
		version: Version310,
	}
}

// BooleanSchema creates a new boolean schema builder.
func BooleanSchema() *SchemaBuilder {
	return &SchemaBuilder{
		schema:  &openapi.Schema{Type: "boolean"},
		version: Version310,
	}
}

// ObjectSchema creates a new object schema builder.
func ObjectSchema() *SchemaBuilder {
	return &SchemaBuilder{
		schema:  &openapi.Schema{Type: "object"},
		version: Version310,
	}
}

// ArraySchema creates a new array schema builder with the given items schema.
func ArraySchema(items *SchemaBuilder) *SchemaBuilder {
	var itemsSchema *openapi.Schema
	if items != nil {
		itemsSchema = items.Build()
	}
	return &SchemaBuilder{
		schema: &openapi.Schema{
			Type:  "array",
			Items: itemsSchema,
		},
		version: Version310,
	}
}

// RefSchema creates a schema reference to a component schema.
// The name is automatically prefixed with #/components/schemas/.
func RefSchema(name string) *SchemaBuilder {
	return &SchemaBuilder{
		schema:  &openapi.Schema{Ref: "#/components/schemas/" + name},
		version: Version310,
	}
}

// Ref creates an explicit schema reference.
func Ref(ref string) *SchemaBuilder {
	return &SchemaBuilder{
		schema:  &openapi.Schema{Ref: ref},
		version: Version310,
	}
}

// NewSchema creates an empty schema builder.
func NewSchema() *SchemaBuilder {
	return &SchemaBuilder{
		schema:  &openapi.Schema{},
		version: Version310,
	}
}

// WithVersion sets the OpenAPI version for version-aware behavior.
func (b *SchemaBuilder) WithVersion(v Version) *SchemaBuilder {
	b.version = v
	return b
}

// Type sets the schema type.
func (b *SchemaBuilder) Type(t string) *SchemaBuilder {
	b.schema.Type = t
	return b
}

// Format sets the schema format.
func (b *SchemaBuilder) Format(f string) *SchemaBuilder {
	b.schema.Format = f
	return b
}

// Title sets the schema title.
func (b *SchemaBuilder) Title(t string) *SchemaBuilder {
	b.schema.Title = t
	return b
}

// Description sets the schema description.
func (b *SchemaBuilder) Description(d string) *SchemaBuilder {
	b.schema.Description = d
	return b
}

// Default sets the default value.
func (b *SchemaBuilder) Default(v any) *SchemaBuilder {
	b.schema.Default = v
	return b
}

// Enum sets the allowed values.
func (b *SchemaBuilder) Enum(values ...any) *SchemaBuilder {
	b.schema.Enum = values
	return b
}

// Const sets the const value.
func (b *SchemaBuilder) Const(v any) *SchemaBuilder {
	b.schema.Const = v
	return b
}

// Nullable marks the schema as nullable.
// For OpenAPI 3.1+, this adds "null" to the type array.
// For OpenAPI 3.0, this sets the nullable keyword.
func (b *SchemaBuilder) Nullable() *SchemaBuilder {
	if b.version.Is31x() || b.version.Is32x() {
		// OpenAPI 3.1+: type becomes an array including "null"
		currentType, ok := b.schema.Type.(string)
		if ok && currentType != "" {
			b.schema.Type = []string{currentType, "null"}
		}
	} else {
		// OpenAPI 3.0: use nullable keyword
		b.schema.Nullable = true
	}
	return b
}

// Minimum sets the minimum value.
func (b *SchemaBuilder) Minimum(v float64) *SchemaBuilder {
	b.schema.Minimum = &v
	return b
}

// Maximum sets the maximum value.
func (b *SchemaBuilder) Maximum(v float64) *SchemaBuilder {
	b.schema.Maximum = &v
	return b
}

// ExclusiveMinimum sets the exclusive minimum value.
func (b *SchemaBuilder) ExclusiveMinimum(v float64) *SchemaBuilder {
	b.schema.ExclusiveMinimum = &v
	return b
}

// ExclusiveMaximum sets the exclusive maximum value.
func (b *SchemaBuilder) ExclusiveMaximum(v float64) *SchemaBuilder {
	b.schema.ExclusiveMaximum = &v
	return b
}

// MultipleOf sets the multipleOf constraint.
func (b *SchemaBuilder) MultipleOf(v float64) *SchemaBuilder {
	b.schema.MultipleOf = &v
	return b
}

// MinLength sets the minimum string length.
func (b *SchemaBuilder) MinLength(v int) *SchemaBuilder {
	b.schema.MinLength = &v
	return b
}

// MaxLength sets the maximum string length.
func (b *SchemaBuilder) MaxLength(v int) *SchemaBuilder {
	b.schema.MaxLength = &v
	return b
}

// Pattern sets the regex pattern for strings.
func (b *SchemaBuilder) Pattern(p string) *SchemaBuilder {
	b.schema.Pattern = p
	return b
}

// Items sets the items schema for arrays.
func (b *SchemaBuilder) Items(items *SchemaBuilder) *SchemaBuilder {
	if items != nil {
		b.schema.Items = items.Build()
	}
	return b
}

// MinItems sets the minimum array length.
func (b *SchemaBuilder) MinItems(v int) *SchemaBuilder {
	b.schema.MinItems = &v
	return b
}

// MaxItems sets the maximum array length.
func (b *SchemaBuilder) MaxItems(v int) *SchemaBuilder {
	b.schema.MaxItems = &v
	return b
}

// UniqueItems requires array items to be unique.
func (b *SchemaBuilder) UniqueItems() *SchemaBuilder {
	b.schema.UniqueItems = true
	return b
}

// Property adds a property to an object schema.
func (b *SchemaBuilder) Property(name string, schema *SchemaBuilder) *SchemaBuilder {
	if b.schema.Properties == nil {
		b.schema.Properties = make(map[string]*openapi.Schema)
	}
	if schema != nil {
		b.schema.Properties[name] = schema.Build()
	}
	return b
}

// Required marks properties as required.
func (b *SchemaBuilder) Required(names ...string) *SchemaBuilder {
	b.schema.Required = append(b.schema.Required, names...)
	return b
}

// AdditionalProperties sets the additionalProperties schema.
// Pass nil to disallow additional properties (false).
// Pass a SchemaBuilder to allow additional properties of that type.
func (b *SchemaBuilder) AdditionalProperties(schema *SchemaBuilder) *SchemaBuilder {
	if schema == nil {
		b.schema.AdditionalProperties = false
	} else {
		b.schema.AdditionalProperties = schema.Build()
	}
	return b
}

// AdditionalPropertiesAllowed allows any additional properties.
func (b *SchemaBuilder) AdditionalPropertiesAllowed() *SchemaBuilder {
	b.schema.AdditionalProperties = true
	return b
}

// MinProperties sets the minimum number of properties.
func (b *SchemaBuilder) MinProperties(v int) *SchemaBuilder {
	b.schema.MinProperties = &v
	return b
}

// MaxProperties sets the maximum number of properties.
func (b *SchemaBuilder) MaxProperties(v int) *SchemaBuilder {
	b.schema.MaxProperties = &v
	return b
}

// AllOf adds schemas to the allOf composition.
func (b *SchemaBuilder) AllOf(schemas ...*SchemaBuilder) *SchemaBuilder {
	for _, s := range schemas {
		if s != nil {
			b.schema.AllOf = append(b.schema.AllOf, s.Build())
		}
	}
	return b
}

// OneOf adds schemas to the oneOf composition.
func (b *SchemaBuilder) OneOf(schemas ...*SchemaBuilder) *SchemaBuilder {
	for _, s := range schemas {
		if s != nil {
			b.schema.OneOf = append(b.schema.OneOf, s.Build())
		}
	}
	return b
}

// AnyOf adds schemas to the anyOf composition.
func (b *SchemaBuilder) AnyOf(schemas ...*SchemaBuilder) *SchemaBuilder {
	for _, s := range schemas {
		if s != nil {
			b.schema.AnyOf = append(b.schema.AnyOf, s.Build())
		}
	}
	return b
}

// Not sets the not schema.
func (b *SchemaBuilder) Not(schema *SchemaBuilder) *SchemaBuilder {
	if schema != nil {
		b.schema.Not = schema.Build()
	}
	return b
}

// Examples adds examples (OpenAPI 3.1).
func (b *SchemaBuilder) Examples(examples ...any) *SchemaBuilder {
	b.schema.Examples = append(b.schema.Examples, examples...)
	return b
}

// Deprecated marks the schema as deprecated.
func (b *SchemaBuilder) Deprecated() *SchemaBuilder {
	b.schema.Deprecated = true
	return b
}

// ReadOnly marks the schema as read-only.
func (b *SchemaBuilder) ReadOnly() *SchemaBuilder {
	b.schema.ReadOnly = true
	return b
}

// WriteOnly marks the schema as write-only.
func (b *SchemaBuilder) WriteOnly() *SchemaBuilder {
	b.schema.WriteOnly = true
	return b
}

// Build returns the constructed schema.
func (b *SchemaBuilder) Build() *openapi.Schema {
	return b.schema
}
