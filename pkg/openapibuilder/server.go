package openapibuilder

import (
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// ServerBuilder builds an OpenAPI Server object.
type ServerBuilder struct {
	server  *openapi.Server
	parent  *SpecBuilder
	version Version
}

// newServerBuilder creates a new server builder.
func newServerBuilder(url string, parent *SpecBuilder) *ServerBuilder {
	return &ServerBuilder{
		server: &openapi.Server{
			URL: url,
		},
		parent:  parent,
		version: Version310,
	}
}

// Description sets the server description.
func (b *ServerBuilder) Description(d string) *ServerBuilder {
	b.server.Description = d
	return b
}

// Done returns to the parent SpecBuilder.
func (b *ServerBuilder) Done() *SpecBuilder {
	if b.parent != nil {
		b.parent.addServer(b.server)
	}
	return b.parent
}

// Build returns the constructed server.
func (b *ServerBuilder) Build() *openapi.Server {
	return b.server
}

// StandaloneServerBuilder builds servers without a parent.
type StandaloneServerBuilder struct {
	server  *openapi.Server
	version Version
}

// NewServer creates a standalone server builder.
func NewServer(url string) *StandaloneServerBuilder {
	return &StandaloneServerBuilder{
		server: &openapi.Server{
			URL: url,
		},
		version: Version310,
	}
}

// Description sets the server description.
func (b *StandaloneServerBuilder) Description(d string) *StandaloneServerBuilder {
	b.server.Description = d
	return b
}

// Build returns the constructed server.
func (b *StandaloneServerBuilder) Build() *openapi.Server {
	return b.server
}
