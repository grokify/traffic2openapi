package convert

import (
	"testing"

	"github.com/grokify/traffic2openapi/pkg/openapi"
)

func TestToVersion30(t *testing.T) {
	// Create a 3.1-style spec with type array for nullable
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
					Type:     []string{"string", "null"},
					Examples: []any{"example1", "example2"},
				},
				"SimpleObject": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"name": {Type: "string"},
						"age":  {Type: []string{"integer", "null"}},
					},
				},
			},
		},
	}

	converted, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if converted.OpenAPI != "3.0.3" {
		t.Errorf("OpenAPI = %q, want %q", converted.OpenAPI, "3.0.3")
	}

	// Check nullable conversion
	nullableSchema := converted.Components.Schemas["NullableString"]
	if nullableSchema.Type != "string" {
		t.Errorf("Type = %v, want %q", nullableSchema.Type, "string")
	}
	if !nullableSchema.Nullable {
		t.Error("expected Nullable = true")
	}

	// Check examples conversion
	if nullableSchema.Example != "example1" {
		t.Errorf("Example = %v, want %q", nullableSchema.Example, "example1")
	}
	if len(nullableSchema.Examples) != 0 {
		t.Errorf("Examples should be empty, got %v", nullableSchema.Examples)
	}

	// Check nested nullable
	simpleObj := converted.Components.Schemas["SimpleObject"]
	ageSchema := simpleObj.Properties["age"]
	if ageSchema.Type != "integer" {
		t.Errorf("age Type = %v, want %q", ageSchema.Type, "integer")
	}
	if !ageSchema.Nullable {
		t.Error("expected age Nullable = true")
	}
}

func TestToVersion31(t *testing.T) {
	// Create a 3.0-style spec with nullable keyword
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
					Example:  "example1",
				},
			},
		},
	}

	converted, err := ToVersion(spec, Version310)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if converted.OpenAPI != "3.1.0" {
		t.Errorf("OpenAPI = %q, want %q", converted.OpenAPI, "3.1.0")
	}

	// Check nullable conversion to type array
	nullableSchema := converted.Components.Schemas["NullableString"]
	typeArr, ok := nullableSchema.Type.([]string)
	if !ok {
		t.Fatalf("Type should be []string, got %T", nullableSchema.Type)
	}
	if len(typeArr) != 2 || typeArr[0] != "string" || typeArr[1] != "null" {
		t.Errorf("Type = %v, want [string null]", typeArr)
	}
	if nullableSchema.Nullable {
		t.Error("expected Nullable = false after conversion")
	}

	// Check example conversion to examples array
	if len(nullableSchema.Examples) != 1 || nullableSchema.Examples[0] != "example1" {
		t.Errorf("Examples = %v, want [example1]", nullableSchema.Examples)
	}
	if nullableSchema.Example != nil {
		t.Errorf("Example should be nil, got %v", nullableSchema.Example)
	}
}

func TestToMultipleVersions(t *testing.T) {
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
	}

	results, err := ToMultipleVersions(spec, Version303, Version310, Version320)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	if results["3.0.3"].OpenAPI != "3.0.3" {
		t.Errorf("3.0.3 version = %q", results["3.0.3"].OpenAPI)
	}
	if results["3.1.0"].OpenAPI != "3.1.0" {
		t.Errorf("3.1.0 version = %q", results["3.1.0"].OpenAPI)
	}
	if results["3.2.0"].OpenAPI != "3.2.0" {
		t.Errorf("3.2.0 version = %q", results["3.2.0"].OpenAPI)
	}
}

func TestMultiVersionOutput(t *testing.T) {
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
	}

	output, err := StandardVersions(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	versions := output.Versions()
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}

	// Check we can get YAML for all versions
	yamlOutput, err := output.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML error: %v", err)
	}
	if len(yamlOutput) != 2 {
		t.Errorf("expected 2 YAML outputs, got %d", len(yamlOutput))
	}

	// Check we can get JSON for all versions
	jsonOutput, err := output.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	if len(jsonOutput) != 2 {
		t.Errorf("expected 2 JSON outputs, got %d", len(jsonOutput))
	}
}

func TestVersionedFilename(t *testing.T) {
	tests := []struct {
		filename string
		version  TargetVersion
		expected string
	}{
		{"openapi.yaml", Version310, "openapi-3.1.0.yaml"},
		{"api.json", Version303, "api-3.0.3.json"},
		{"spec.yml", Version320, "spec-3.2.0.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := VersionedFilename(tt.filename, tt.version)
			if got != tt.expected {
				t.Errorf("VersionedFilename(%q, %s) = %q, want %q",
					tt.filename, tt.version, got, tt.expected)
			}
		})
	}
}

func TestTargetVersionMethods(t *testing.T) {
	tests := []struct {
		version TargetVersion
		is30x   bool
		is31x   bool
		is32x   bool
	}{
		{Version300, true, false, false},
		{Version303, true, false, false},
		{Version310, false, true, false},
		{Version311, false, true, false},
		{Version320, false, false, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.version), func(t *testing.T) {
			if got := tt.version.Is30x(); got != tt.is30x {
				t.Errorf("Is30x() = %v, want %v", got, tt.is30x)
			}
			if got := tt.version.Is31x(); got != tt.is31x {
				t.Errorf("Is31x() = %v, want %v", got, tt.is31x)
			}
			if got := tt.version.Is32x(); got != tt.is32x {
				t.Errorf("Is32x() = %v, want %v", got, tt.is32x)
			}
		})
	}
}

func TestDeepCopyIsolation(t *testing.T) {
	// Ensure converting doesn't modify original
	spec := &openapi.Spec{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*openapi.PathItem),
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"Test": {
					Type:     []string{"string", "null"},
					Examples: []any{"test"},
				},
			},
		},
	}

	_, err := ToVersion(spec, Version303)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Original should be unchanged
	if spec.OpenAPI != "3.1.0" {
		t.Errorf("original OpenAPI was modified: %s", spec.OpenAPI)
	}

	origSchema := spec.Components.Schemas["Test"]
	if _, ok := origSchema.Type.([]string); !ok {
		t.Errorf("original Type was modified: %v", origSchema.Type)
	}
	if len(origSchema.Examples) != 1 {
		t.Errorf("original Examples was modified: %v", origSchema.Examples)
	}
}
