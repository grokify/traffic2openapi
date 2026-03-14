package convert

import (
	"testing"

	"github.com/grokify/traffic2openapi/pkg/openapi"
)

func TestToVersionBasic(t *testing.T) {
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{
			"/test": {
				Get: &openapi.Operation{
					Responses: map[string]openapi.Response{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	// Convert to 3.0.3
	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if converted.OpenAPI != "3.0.3" {
		t.Errorf("OpenAPI = %q, want %q", converted.OpenAPI, "3.0.3")
	}

	// Original should be unchanged
	if spec.OpenAPI != "3.1.0" {
		t.Errorf("original spec modified: OpenAPI = %q", spec.OpenAPI)
	}
}

func TestToVersionNullableConversion(t *testing.T) {
	// Create a 3.1 spec with nullable using type array
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"NullableString": {
					Type: []string{"string", "null"},
				},
			},
		},
	}

	// Convert to 3.0.3
	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := converted.Components.Schemas["NullableString"]
	if schema.Type != "string" {
		t.Errorf("Type = %v, want %q", schema.Type, "string")
	}
	if !schema.Nullable {
		t.Error("expected Nullable = true")
	}
}

func TestToVersionNullableConversionReverse(t *testing.T) {
	// Create a 3.0 spec with nullable: true
	spec := &openapi.Spec{
		OpenAPI: "3.0.3",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"NullableString": {
					Type:     "string",
					Nullable: true,
				},
			},
		},
	}

	// Convert to 3.1.0
	converted, err := ToVersion(spec, Version310)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := converted.Components.Schemas["NullableString"]
	typeArr, ok := schema.Type.([]string)
	if !ok {
		t.Fatalf("Type = %T, want []string", schema.Type)
	}
	if len(typeArr) != 2 || typeArr[0] != "string" || typeArr[1] != "null" {
		t.Errorf("Type = %v, want [string, null]", typeArr)
	}
	if schema.Nullable {
		t.Error("expected Nullable = false after 3.1 conversion")
	}
}

func TestToVersionExamplesConversion(t *testing.T) {
	// Create a 3.1 spec with examples array
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"StringWithExamples": {
					Type:     "string",
					Examples: []any{"example1", "example2"},
				},
			},
		},
	}

	// Convert to 3.0.3
	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := converted.Components.Schemas["StringWithExamples"]
	if schema.Example != "example1" {
		t.Errorf("Example = %v, want %q", schema.Example, "example1")
	}
	if len(schema.Examples) != 0 {
		t.Errorf("Examples should be empty after 3.0 conversion, got %v", schema.Examples)
	}
}

func TestToVersionExamplesConversionReverse(t *testing.T) {
	// Create a 3.0 spec with singular example
	spec := &openapi.Spec{
		OpenAPI: "3.0.3",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"StringWithExample": {
					Type:    "string",
					Example: "example1",
				},
			},
		},
	}

	// Convert to 3.1.0
	converted, err := ToVersion(spec, Version310)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := converted.Components.Schemas["StringWithExample"]
	if len(schema.Examples) != 1 || schema.Examples[0] != "example1" {
		t.Errorf("Examples = %v, want [example1]", schema.Examples)
	}
	if schema.Example != nil {
		t.Errorf("Example should be nil after 3.1 conversion, got %v", schema.Example)
	}
}

func TestToMultipleVersionsExtended(t *testing.T) {
	// Start with 3.0 spec
	spec := &openapi.Spec{
		OpenAPI: "3.0.3",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"Test": {Type: "string", Nullable: true},
			},
		},
	}

	results, err := ToMultipleVersions(spec, Version303, Version310, Version320)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// 3.0 should still have nullable
	if !results["3.0.3"].Components.Schemas["Test"].Nullable {
		t.Error("3.0.3 should have nullable=true")
	}

	// 3.1 should have type array (converted from nullable)
	if typeArr, ok := results["3.1.0"].Components.Schemas["Test"].Type.([]string); !ok {
		t.Errorf("3.1.0 should have type array, got %T", results["3.1.0"].Components.Schemas["Test"].Type)
	} else if len(typeArr) != 2 {
		t.Errorf("3.1.0 type array should have 2 elements, got %d", len(typeArr))
	}
}

func TestConvertNestedSchemas(t *testing.T) {
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"NestedObject": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"nested": {
							Type: []string{"string", "null"},
						},
						"array": {
							Type: "array",
							Items: &openapi.Schema{
								Type: []string{"integer", "null"},
							},
						},
					},
				},
			},
		},
	}

	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nestedSchema := converted.Components.Schemas["NestedObject"]
	if nestedSchema.Properties["nested"].Type != "string" {
		t.Errorf("nested.Type = %v, want %q", nestedSchema.Properties["nested"].Type, "string")
	}
	if !nestedSchema.Properties["nested"].Nullable {
		t.Error("nested.Nullable should be true")
	}

	arraySchema := nestedSchema.Properties["array"]
	if arraySchema.Items.Type != "integer" {
		t.Errorf("array.Items.Type = %v, want %q", arraySchema.Items.Type, "integer")
	}
	if !arraySchema.Items.Nullable {
		t.Error("array.Items.Nullable should be true")
	}
}

func TestConvertCompositionSchemas(t *testing.T) {
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"ComposedSchema": {
					AllOf: []*openapi.Schema{
						{Type: []string{"string", "null"}},
					},
					OneOf: []*openapi.Schema{
						{Type: []string{"integer", "null"}},
					},
					AnyOf: []*openapi.Schema{
						{Type: []string{"boolean", "null"}},
					},
				},
			},
		},
	}

	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := converted.Components.Schemas["ComposedSchema"]
	if schema.AllOf[0].Type != "string" || !schema.AllOf[0].Nullable {
		t.Errorf("AllOf[0] not converted correctly")
	}
	if schema.OneOf[0].Type != "integer" || !schema.OneOf[0].Nullable {
		t.Errorf("OneOf[0] not converted correctly")
	}
	if schema.AnyOf[0].Type != "boolean" || !schema.AnyOf[0].Nullable {
		t.Errorf("AnyOf[0] not converted correctly")
	}
}

func TestConvertPathOperations(t *testing.T) {
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{
			"/test": {
				Get: &openapi.Operation{
					Parameters: []openapi.Parameter{
						{
							Name: "id",
							In:   "path",
							Schema: &openapi.Schema{
								Type: []string{"string", "null"},
							},
						},
					},
					Responses: map[string]openapi.Response{
						"200": {
							Description: "Success",
							Content: map[string]openapi.MediaType{
								"application/json": {
									Schema: &openapi.Schema{
										Type: []string{"object", "null"},
									},
								},
							},
							Headers: map[string]openapi.Header{
								"X-Count": {
									Schema: &openapi.Schema{
										Type: []string{"integer", "null"},
									},
								},
							},
						},
					},
				},
				Post: &openapi.Operation{
					RequestBody: &openapi.RequestBody{
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Type: []string{"object", "null"},
								},
							},
						},
					},
					Responses: map[string]openapi.Response{
						"201": {Description: "Created"},
					},
				},
			},
		},
	}

	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathItem := converted.Paths["/test"]

	// Check GET operation
	getOp := pathItem.Get
	if getOp.Parameters[0].Schema.Type != "string" {
		t.Errorf("GET parameter schema type = %v", getOp.Parameters[0].Schema.Type)
	}
	if !getOp.Parameters[0].Schema.Nullable {
		t.Error("GET parameter should be nullable")
	}

	respSchema := getOp.Responses["200"].Content["application/json"].Schema
	if respSchema.Type != "object" || !respSchema.Nullable {
		t.Error("GET response schema not converted correctly")
	}

	headerSchema := getOp.Responses["200"].Headers["X-Count"].Schema
	if headerSchema.Type != "integer" || !headerSchema.Nullable {
		t.Error("GET response header schema not converted correctly")
	}

	// Check POST operation
	postOp := pathItem.Post
	reqSchema := postOp.RequestBody.Content["application/json"].Schema
	if reqSchema.Type != "object" || !reqSchema.Nullable {
		t.Error("POST request body schema not converted correctly")
	}
}

func TestConvertPathItemParameters(t *testing.T) {
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{
			"/test/{id}": {
				Parameters: []openapi.Parameter{
					{
						Name: "id",
						In:   "path",
						Schema: &openapi.Schema{
							Type: []string{"string", "null"},
						},
					},
				},
				Get: &openapi.Operation{
					Responses: map[string]openapi.Response{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathItem := converted.Paths["/test/{id}"]
	if pathItem.Parameters[0].Schema.Type != "string" {
		t.Errorf("path-level parameter schema type = %v", pathItem.Parameters[0].Schema.Type)
	}
	if !pathItem.Parameters[0].Schema.Nullable {
		t.Error("path-level parameter should be nullable")
	}
}

func TestVersionedFilenameNoExtension(t *testing.T) {
	got := VersionedFilename("api", Version310)
	want := "api-3.1.0"
	if got != want {
		t.Errorf("VersionedFilename(%q, %q) = %q, want %q", "api", Version310, got, want)
	}
}

func TestAllOperationTypes(t *testing.T) {
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]*openapi.PathItem{
			"/test": {
				Get:     &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "OK"}}},
				Put:     &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "OK"}}},
				Post:    &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "OK"}}},
				Delete:  &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "OK"}}},
				Options: &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "OK"}}},
				Head:    &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "OK"}}},
				Patch:   &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "OK"}}},
				Trace:   &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "OK"}}},
			},
		},
	}

	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathItem := converted.Paths["/test"]
	ops := []*openapi.Operation{
		pathItem.Get, pathItem.Put, pathItem.Post, pathItem.Delete,
		pathItem.Options, pathItem.Head, pathItem.Patch, pathItem.Trace,
	}
	for i, op := range ops {
		if op == nil {
			t.Errorf("operation %d is nil", i)
		}
	}
}
