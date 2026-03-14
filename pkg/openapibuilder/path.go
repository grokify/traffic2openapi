package openapibuilder

import (
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// PathItemBuilder builds an OpenAPI PathItem object.
type PathItemBuilder struct {
	pathItem *openapi.PathItem
	path     string
	parent   *SpecBuilder
	version  Version
}

// newPathItemBuilder creates a new path item builder.
func newPathItemBuilder(path string, parent *SpecBuilder) *PathItemBuilder {
	return &PathItemBuilder{
		pathItem: &openapi.PathItem{},
		path:     path,
		parent:   parent,
		version:  Version310,
	}
}

// Summary sets the path item summary.
func (b *PathItemBuilder) Summary(s string) *PathItemBuilder {
	b.pathItem.Summary = s
	return b
}

// Description sets the path item description.
func (b *PathItemBuilder) Description(d string) *PathItemBuilder {
	b.pathItem.Description = d
	return b
}

// Get starts building a GET operation.
func (b *PathItemBuilder) Get() *OperationBuilder {
	return newOperationBuilder("get", b)
}

// Post starts building a POST operation.
func (b *PathItemBuilder) Post() *OperationBuilder {
	return newOperationBuilder("post", b)
}

// Put starts building a PUT operation.
func (b *PathItemBuilder) Put() *OperationBuilder {
	return newOperationBuilder("put", b)
}

// Delete starts building a DELETE operation.
func (b *PathItemBuilder) Delete() *OperationBuilder {
	return newOperationBuilder("delete", b)
}

// Patch starts building a PATCH operation.
func (b *PathItemBuilder) Patch() *OperationBuilder {
	return newOperationBuilder("patch", b)
}

// Options starts building an OPTIONS operation.
func (b *PathItemBuilder) Options() *OperationBuilder {
	return newOperationBuilder("options", b)
}

// Head starts building a HEAD operation.
func (b *PathItemBuilder) Head() *OperationBuilder {
	return newOperationBuilder("head", b)
}

// Trace starts building a TRACE operation.
func (b *PathItemBuilder) Trace() *OperationBuilder {
	return newOperationBuilder("trace", b)
}

// Parameter adds a shared parameter to the path item.
func (b *PathItemBuilder) Parameter(param *openapi.Parameter) *PathItemBuilder {
	b.pathItem.Parameters = append(b.pathItem.Parameters, *param)
	return b
}

// Done returns to the parent SpecBuilder.
func (b *PathItemBuilder) Done() *SpecBuilder {
	if b.parent != nil {
		b.parent.addPath(b.path, b.pathItem)
	}
	return b.parent
}

// Build returns the constructed path item.
func (b *PathItemBuilder) Build() *openapi.PathItem {
	return b.pathItem
}

// setOperation sets an operation on the path item.
func (b *PathItemBuilder) setOperation(method string, op *openapi.Operation) {
	switch method {
	case "get":
		b.pathItem.Get = op
	case "post":
		b.pathItem.Post = op
	case "put":
		b.pathItem.Put = op
	case "delete":
		b.pathItem.Delete = op
	case "patch":
		b.pathItem.Patch = op
	case "options":
		b.pathItem.Options = op
	case "head":
		b.pathItem.Head = op
	case "trace":
		b.pathItem.Trace = op
	}
}
