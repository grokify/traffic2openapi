// Package convert provides OpenAPI specification version conversion.
package convert

import (
	"encoding/json"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// TargetVersion represents the target OpenAPI version for conversion.
type TargetVersion string

const (
	// OpenAPI 3.0.x versions
	Version300 TargetVersion = "3.0.0"
	Version301 TargetVersion = "3.0.1"
	Version302 TargetVersion = "3.0.2"
	Version303 TargetVersion = "3.0.3"

	// OpenAPI 3.1.x versions
	Version310 TargetVersion = "3.1.0"
	Version311 TargetVersion = "3.1.1"

	// OpenAPI 3.2.x versions
	Version320 TargetVersion = "3.2.0"
)

// Is30x returns true if this is an OpenAPI 3.0.x version.
func (v TargetVersion) Is30x() bool {
	return strings.HasPrefix(string(v), "3.0")
}

// Is31x returns true if this is an OpenAPI 3.1.x version.
func (v TargetVersion) Is31x() bool {
	return strings.HasPrefix(string(v), "3.1")
}

// Is32x returns true if this is an OpenAPI 3.2.x version.
func (v TargetVersion) Is32x() bool {
	return strings.HasPrefix(string(v), "3.2")
}

// ToVersion converts an OpenAPI spec to the specified version.
// This creates a deep copy and modifies version-specific elements.
func ToVersion(spec *openapi.Spec, target TargetVersion) (*openapi.Spec, error) {
	// Deep copy the spec
	copied, err := deepCopy(spec)
	if err != nil {
		return nil, err
	}

	// Update the version
	copied.OpenAPI = string(target)

	// Convert schemas based on target version
	if target.Is30x() {
		convertTo30(copied)
	} else {
		convertTo31Plus(copied)
	}

	return copied, nil
}

// ToMultipleVersions converts an OpenAPI spec to multiple versions.
// Returns a map of version string to converted spec.
func ToMultipleVersions(spec *openapi.Spec, targets ...TargetVersion) (map[string]*openapi.Spec, error) {
	results := make(map[string]*openapi.Spec, len(targets))
	for _, target := range targets {
		converted, err := ToVersion(spec, target)
		if err != nil {
			return nil, err
		}
		results[string(target)] = converted
	}
	return results, nil
}

// convertTo30 converts a spec to OpenAPI 3.0.x format.
func convertTo30(spec *openapi.Spec) {
	// Convert component schemas
	if spec.Components != nil {
		for _, schema := range spec.Components.Schemas {
			convertSchemaTo30(schema)
		}
	}

	// Convert path schemas
	for _, pathItem := range spec.Paths {
		convertPathItemTo30(pathItem)
	}
}

// convertTo31Plus converts a spec to OpenAPI 3.1+ format.
func convertTo31Plus(spec *openapi.Spec) {
	// Convert component schemas
	if spec.Components != nil {
		for _, schema := range spec.Components.Schemas {
			convertSchemaTo31Plus(schema)
		}
	}

	// Convert path schemas
	for _, pathItem := range spec.Paths {
		convertPathItemTo31Plus(pathItem)
	}
}

// convertSchemaTo30 converts a schema to OpenAPI 3.0 format.
func convertSchemaTo30(schema *openapi.Schema) {
	if schema == nil {
		return
	}

	// Convert type array with "null" to nullable: true
	if typeArr, ok := schema.Type.([]any); ok {
		var nonNullTypes []string
		hasNull := false
		for _, t := range typeArr {
			if ts, ok := t.(string); ok {
				if ts == "null" {
					hasNull = true
				} else {
					nonNullTypes = append(nonNullTypes, ts)
				}
			}
		}
		if hasNull {
			schema.Nullable = true
		}
		if len(nonNullTypes) == 1 {
			schema.Type = nonNullTypes[0]
		} else if len(nonNullTypes) > 1 {
			// Keep as array but without null (3.0 doesn't support type arrays)
			// This is a limitation - we use the first type
			schema.Type = nonNullTypes[0]
		}
	}

	// Also handle []string type
	if typeArr, ok := schema.Type.([]string); ok {
		var nonNullTypes []string
		hasNull := false
		for _, t := range typeArr {
			if t == "null" {
				hasNull = true
			} else {
				nonNullTypes = append(nonNullTypes, t)
			}
		}
		if hasNull {
			schema.Nullable = true
		}
		if len(nonNullTypes) == 1 {
			schema.Type = nonNullTypes[0]
		} else if len(nonNullTypes) > 1 {
			schema.Type = nonNullTypes[0]
		}
	}

	// Convert examples array to singular example
	if len(schema.Examples) > 0 && schema.Example == nil {
		schema.Example = schema.Examples[0]
		schema.Examples = nil
	}

	// Recursively convert nested schemas
	if schema.Items != nil {
		convertSchemaTo30(schema.Items)
	}
	for _, prop := range schema.Properties {
		convertSchemaTo30(prop)
	}
	for _, s := range schema.AllOf {
		convertSchemaTo30(s)
	}
	for _, s := range schema.OneOf {
		convertSchemaTo30(s)
	}
	for _, s := range schema.AnyOf {
		convertSchemaTo30(s)
	}
	if schema.Not != nil {
		convertSchemaTo30(schema.Not)
	}
	if addProps, ok := schema.AdditionalProperties.(*openapi.Schema); ok {
		convertSchemaTo30(addProps)
	}
}

// convertSchemaTo31Plus converts a schema to OpenAPI 3.1+ format.
func convertSchemaTo31Plus(schema *openapi.Schema) {
	if schema == nil {
		return
	}

	// Convert nullable: true to type array with "null"
	if schema.Nullable {
		if typeStr, ok := schema.Type.(string); ok {
			schema.Type = []string{typeStr, "null"}
		}
		schema.Nullable = false
	}

	// Convert singular example to examples array
	if schema.Example != nil && len(schema.Examples) == 0 {
		schema.Examples = []any{schema.Example}
		schema.Example = nil
	}

	// Recursively convert nested schemas
	if schema.Items != nil {
		convertSchemaTo31Plus(schema.Items)
	}
	for _, prop := range schema.Properties {
		convertSchemaTo31Plus(prop)
	}
	for _, s := range schema.AllOf {
		convertSchemaTo31Plus(s)
	}
	for _, s := range schema.OneOf {
		convertSchemaTo31Plus(s)
	}
	for _, s := range schema.AnyOf {
		convertSchemaTo31Plus(s)
	}
	if schema.Not != nil {
		convertSchemaTo31Plus(schema.Not)
	}
	if addProps, ok := schema.AdditionalProperties.(*openapi.Schema); ok {
		convertSchemaTo31Plus(addProps)
	}
}

// convertPathItemTo30 converts a path item's schemas to 3.0 format.
func convertPathItemTo30(pathItem *openapi.PathItem) {
	if pathItem == nil {
		return
	}

	for i := range pathItem.Parameters {
		if pathItem.Parameters[i].Schema != nil {
			convertSchemaTo30(pathItem.Parameters[i].Schema)
		}
	}

	operations := []*openapi.Operation{
		pathItem.Get, pathItem.Put, pathItem.Post, pathItem.Delete,
		pathItem.Options, pathItem.Head, pathItem.Patch, pathItem.Trace,
	}

	for _, op := range operations {
		convertOperationTo30(op)
	}
}

// convertPathItemTo31Plus converts a path item's schemas to 3.1+ format.
func convertPathItemTo31Plus(pathItem *openapi.PathItem) {
	if pathItem == nil {
		return
	}

	for i := range pathItem.Parameters {
		if pathItem.Parameters[i].Schema != nil {
			convertSchemaTo31Plus(pathItem.Parameters[i].Schema)
		}
	}

	operations := []*openapi.Operation{
		pathItem.Get, pathItem.Put, pathItem.Post, pathItem.Delete,
		pathItem.Options, pathItem.Head, pathItem.Patch, pathItem.Trace,
	}

	for _, op := range operations {
		convertOperationTo31Plus(op)
	}
}

// convertOperationTo30 converts an operation's schemas to 3.0 format.
func convertOperationTo30(op *openapi.Operation) {
	if op == nil {
		return
	}

	for i := range op.Parameters {
		if op.Parameters[i].Schema != nil {
			convertSchemaTo30(op.Parameters[i].Schema)
		}
	}

	if op.RequestBody != nil {
		for _, mt := range op.RequestBody.Content {
			convertSchemaTo30(mt.Schema)
		}
	}

	for _, resp := range op.Responses {
		for _, mt := range resp.Content {
			convertSchemaTo30(mt.Schema)
		}
		for _, h := range resp.Headers {
			convertSchemaTo30(h.Schema)
		}
	}
}

// convertOperationTo31Plus converts an operation's schemas to 3.1+ format.
func convertOperationTo31Plus(op *openapi.Operation) {
	if op == nil {
		return
	}

	for i := range op.Parameters {
		if op.Parameters[i].Schema != nil {
			convertSchemaTo31Plus(op.Parameters[i].Schema)
		}
	}

	if op.RequestBody != nil {
		for _, mt := range op.RequestBody.Content {
			convertSchemaTo31Plus(mt.Schema)
		}
	}

	for _, resp := range op.Responses {
		for _, mt := range resp.Content {
			convertSchemaTo31Plus(mt.Schema)
		}
		for _, h := range resp.Headers {
			convertSchemaTo31Plus(h.Schema)
		}
	}
}

// deepCopy creates a deep copy of the spec using JSON marshaling.
func deepCopy(spec *openapi.Spec) (*openapi.Spec, error) {
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	var copied openapi.Spec
	if err := json.Unmarshal(data, &copied); err != nil {
		return nil, err
	}
	return &copied, nil
}
