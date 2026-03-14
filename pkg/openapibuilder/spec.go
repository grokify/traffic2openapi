package openapibuilder

import (
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// SpecBuilder builds an OpenAPI Specification.
type SpecBuilder struct {
	spec    *openapi.Spec
	version Version
	errors  *ValidationErrors
}

// NewSpec creates a new spec builder with the given OpenAPI version.
func NewSpec(version Version) *SpecBuilder {
	return &SpecBuilder{
		spec: &openapi.Spec{
			OpenAPI: version.String(),
			Paths:   make(map[string]*openapi.PathItem),
		},
		version: version,
		errors:  &ValidationErrors{},
	}
}

// Title sets the API title.
func (b *SpecBuilder) Title(title string) *SpecBuilder {
	b.spec.Info.Title = title
	return b
}

// Description sets the API description.
func (b *SpecBuilder) Description(desc string) *SpecBuilder {
	b.spec.Info.Description = desc
	return b
}

// Version sets the API version.
func (b *SpecBuilder) Version(v string) *SpecBuilder {
	b.spec.Info.Version = v
	return b
}

// Contact sets the contact information.
func (b *SpecBuilder) Contact(name, url, email string) *SpecBuilder {
	b.spec.Info.Contact = &openapi.Contact{
		Name:  name,
		URL:   url,
		Email: email,
	}
	return b
}

// License sets the license information.
func (b *SpecBuilder) License(name, url string) *SpecBuilder {
	b.spec.Info.License = &openapi.License{
		Name: name,
		URL:  url,
	}
	return b
}

// Server adds a server with a simple URL.
func (b *SpecBuilder) Server(url string) *SpecBuilder {
	b.spec.Servers = append(b.spec.Servers, openapi.Server{URL: url})
	return b
}

// ServerWithDescription adds a server with a URL and description.
func (b *SpecBuilder) ServerWithDescription(url, description string) *SpecBuilder {
	b.spec.Servers = append(b.spec.Servers, openapi.Server{
		URL:         url,
		Description: description,
	})
	return b
}

// ServerBuilder starts building a server with more options.
func (b *SpecBuilder) ServerBuilder(url string) *ServerBuilder {
	return newServerBuilder(url, b)
}

// Path starts building a path item.
func (b *SpecBuilder) Path(path string) *PathItemBuilder {
	return newPathItemBuilder(path, b)
}

// AddPath adds a pre-built path item.
func (b *SpecBuilder) AddPath(path string, pathItem *openapi.PathItem) *SpecBuilder {
	b.spec.Paths[path] = pathItem
	return b
}

// Components starts building the components section.
func (b *SpecBuilder) Components() *ComponentsBuilder {
	return newComponentsBuilder(b)
}

// Security adds a global security requirement.
func (b *SpecBuilder) Security(name string, scopes ...string) *SpecBuilder {
	// Note: The current Spec type doesn't have a Security field.
	// This would need to be added to the types if global security is needed.
	return b
}

// Build validates and returns the constructed spec.
func (b *SpecBuilder) Build() (*openapi.Spec, error) {
	b.validate()
	if b.errors.HasErrors() {
		return nil, b.errors
	}
	return b.spec, nil
}

// MustBuild builds the spec and panics on error.
func (b *SpecBuilder) MustBuild() *openapi.Spec {
	spec, err := b.Build()
	if err != nil {
		panic(err)
	}
	return spec
}

// BuildUnchecked returns the spec without validation.
func (b *SpecBuilder) BuildUnchecked() *openapi.Spec {
	return b.spec
}

// validate checks for required fields.
func (b *SpecBuilder) validate() {
	if b.spec.Info.Title == "" {
		b.errors.Add(ErrMissingTitle)
	}
	if b.spec.Info.Version == "" {
		b.errors.Add(ErrMissingVersion)
	}

	// Validate responses have descriptions
	for path, pathItem := range b.spec.Paths {
		b.validatePathItem(path, pathItem)
	}
}

// validatePathItem validates a path item.
func (b *SpecBuilder) validatePathItem(path string, pi *openapi.PathItem) {
	operations := map[string]*openapi.Operation{
		"get":     pi.Get,
		"post":    pi.Post,
		"put":     pi.Put,
		"delete":  pi.Delete,
		"patch":   pi.Patch,
		"options": pi.Options,
		"head":    pi.Head,
		"trace":   pi.Trace,
	}

	for method, op := range operations {
		if op != nil {
			b.validateOperation(path, method, op)
		}
	}
}

// validateOperation validates an operation.
func (b *SpecBuilder) validateOperation(path, method string, op *openapi.Operation) {
	for statusCode, resp := range op.Responses {
		if resp.Description == "" {
			b.errors.AddField(
				path+"."+method+".responses."+statusCode,
				"response description is required",
			)
		}
	}
}

// addPath adds a path item to the spec (called by PathItemBuilder).
func (b *SpecBuilder) addPath(path string, pathItem *openapi.PathItem) {
	b.spec.Paths[path] = pathItem
}

// addServer adds a server to the spec (called by ServerBuilder).
func (b *SpecBuilder) addServer(server *openapi.Server) {
	b.spec.Servers = append(b.spec.Servers, *server)
}

// setComponents sets the components (called by ComponentsBuilder).
func (b *SpecBuilder) setComponents(components *openapi.Components) {
	b.spec.Components = components
}

// GetVersion returns the OpenAPI version.
func (b *SpecBuilder) GetVersion() Version {
	return b.version
}

// Spec returns the underlying spec (for advanced use).
func (b *SpecBuilder) Spec() *openapi.Spec {
	return b.spec
}
