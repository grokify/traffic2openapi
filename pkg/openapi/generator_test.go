package openapi

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/traffic2openapi/pkg/inference"
)

func TestGenerateFromExamples(t *testing.T) {
	// Load example IR files
	dir := filepath.Join("..", "..", "examples")
	result, err := inference.InferFromDir(dir)
	if err != nil {
		t.Fatalf("InferFromDir failed: %v", err)
	}

	// Generate OpenAPI spec
	options := DefaultGeneratorOptions()
	options.Title = "Example API"
	options.Description = "Generated from example traffic"

	spec := GenerateFromInference(result, options)

	// Verify basic structure
	if spec.OpenAPI != "3.1.0" {
		t.Errorf("expected OpenAPI 3.1.0, got %s", spec.OpenAPI)
	}
	if spec.Info.Title != "Example API" {
		t.Errorf("expected title 'Example API', got %s", spec.Info.Title)
	}

	// Verify paths exist
	if len(spec.Paths) == 0 {
		t.Error("expected at least one path")
	}

	// Check /users path
	usersPath, ok := spec.Paths["/users"]
	if !ok {
		t.Error("expected /users path")
	} else {
		if usersPath.Get == nil {
			t.Error("expected GET operation on /users")
		}
		if usersPath.Post == nil {
			t.Error("expected POST operation on /users")
		}
	}

	// Check /users/{id} path (sample data uses {id}, not {userId})
	userPath, ok := spec.Paths["/users/{id}"]
	if !ok {
		t.Error("expected /users/{id} path")
	} else {
		if userPath.Get == nil {
			t.Error("expected GET operation on /users/{id}")
		}
		if userPath.Delete == nil {
			t.Error("expected DELETE operation on /users/{id}")
		}

		// Check path parameter
		if userPath.Get != nil {
			hasPathParam := false
			for _, param := range userPath.Get.Parameters {
				if param.In == "path" && param.Name == "id" {
					hasPathParam = true
					if !param.Required {
						t.Error("path parameter should be required")
					}
				}
			}
			if !hasPathParam {
				t.Error("expected id path parameter")
			}
		}
	}
}

func TestGenerateOpenAPI30(t *testing.T) {
	result := &inference.InferenceResult{
		Endpoints: map[string]*inference.EndpointData{
			"GET /test": {
				Method:       "GET",
				PathTemplate: "/test",
				Responses: map[int]*inference.ResponseData{
					200: inference.NewResponseData(200),
				},
			},
		},
	}

	options := DefaultGeneratorOptions()
	options.Version = Version30

	spec := GenerateFromInference(result, options)

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("expected OpenAPI 3.0.3, got %s", spec.OpenAPI)
	}
}

func TestGenerateWithServers(t *testing.T) {
	result := &inference.InferenceResult{
		Endpoints: map[string]*inference.EndpointData{
			"GET /test": {
				Method:       "GET",
				PathTemplate: "/test",
				Responses: map[int]*inference.ResponseData{
					200: inference.NewResponseData(200),
				},
			},
		},
		Hosts:   []string{"api.example.com"},
		Schemes: []string{"https"},
	}

	options := DefaultGeneratorOptions()
	spec := GenerateFromInference(result, options)

	if len(spec.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(spec.Servers))
	}
	if spec.Servers[0].URL != "https://api.example.com" {
		t.Errorf("expected https://api.example.com, got %s", spec.Servers[0].URL)
	}
}

func TestGenerateOperationID(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/users", "getUsers"},
		{"POST", "/users", "postUsers"},
		{"GET", "/users/{userId}", "getUsersByUserId"},
		{"DELETE", "/users/{userId}/posts/{postId}", "deleteUsersByUserIdPostsByPostId"},
		{"GET", "/", "get"},
	}

	for _, tt := range tests {
		result := generateOperationID(tt.method, tt.path)
		if result != tt.expected {
			t.Errorf("generateOperationID(%q, %q) = %q, want %q",
				tt.method, tt.path, result, tt.expected)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	spec := &Spec{
		OpenAPI: "3.1.0",
		Info: Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]*PathItem{
			"/test": {
				Get: &Operation{
					Summary:   "Test endpoint",
					Responses: map[string]Response{"200": {Description: "OK"}},
				},
			},
		},
	}

	jsonStr, err := ToString(spec, FormatJSON)
	if err != nil {
		t.Fatalf("ToString failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check content
	if !strings.Contains(jsonStr, `"openapi": "3.1.0"`) {
		t.Error("expected openapi version in JSON")
	}
	if !strings.Contains(jsonStr, `"title": "Test API"`) {
		t.Error("expected title in JSON")
	}
}

func TestWriteYAML(t *testing.T) {
	spec := &Spec{
		OpenAPI: "3.1.0",
		Info: Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]*PathItem{
			"/test": {
				Get: &Operation{
					Summary:   "Test endpoint",
					Responses: map[string]Response{"200": {Description: "OK"}},
				},
			},
		},
	}

	yamlStr, err := ToString(spec, FormatYAML)
	if err != nil {
		t.Fatalf("ToString failed: %v", err)
	}

	// Check YAML content
	if !strings.Contains(yamlStr, "openapi: 3.1.0") && !strings.Contains(yamlStr, `openapi: "3.1.0"`) {
		t.Error("expected openapi version in YAML")
	}
	if !strings.Contains(yamlStr, "title: Test API") {
		t.Error("expected title in YAML")
	}
}

func TestSchemaConversion(t *testing.T) {
	// Create a schema store with various types
	store := inference.NewSchemaStore()
	store.AddObservation()
	store.AddValue("name", "test")
	store.AddValue("count", float64(42))
	store.AddValue("active", true)
	store.AddValue("email", "test@example.com")

	node := inference.BuildSchemaTree(store)

	gen := NewGenerator(DefaultGeneratorOptions())
	schema := gen.convertSchemaNode(node)

	// Verify properties
	if schema.Properties == nil {
		t.Fatal("expected properties")
	}

	if schema.Properties["name"] == nil || schema.Properties["name"].Type != "string" {
		t.Error("expected name to be string")
	}
	if schema.Properties["count"] == nil || schema.Properties["count"].Type != "integer" {
		t.Error("expected count to be integer")
	}
	if schema.Properties["active"] == nil || schema.Properties["active"].Type != "boolean" {
		t.Error("expected active to be boolean")
	}
	if schema.Properties["email"] == nil || schema.Properties["email"].Format != "email" {
		t.Error("expected email to have email format")
	}
}
