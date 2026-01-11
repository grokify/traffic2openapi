package inference

import (
	"path/filepath"
	"testing"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

func TestInferFromFile(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "sample-batch.json")
	result, err := InferFromFile(path)
	if err != nil {
		t.Fatalf("InferFromFile failed: %v", err)
	}

	// Should have endpoints
	if len(result.Endpoints) == 0 {
		t.Error("expected at least one endpoint")
	}

	// Check for expected endpoints
	// Note: sample-batch.json has pathTemplate: "/users/{id}" pre-defined
	expectedEndpoints := []string{
		"GET /users",
		"GET /users/{id}",
		"POST /users",
		"DELETE /users/{id}",
	}

	for _, expected := range expectedEndpoints {
		if _, ok := result.Endpoints[expected]; !ok {
			t.Errorf("expected endpoint %q not found", expected)
		}
	}
}

func TestInferFromDir(t *testing.T) {
	dir := filepath.Join("..", "..", "examples")
	result, err := InferFromDir(dir)
	if err != nil {
		t.Fatalf("InferFromDir failed: %v", err)
	}

	// Should have endpoints from both files
	if len(result.Endpoints) == 0 {
		t.Error("expected at least one endpoint")
	}

	// Check hosts
	if len(result.Hosts) == 0 {
		t.Error("expected at least one host")
	}

	found := false
	for _, host := range result.Hosts {
		if host == "api.example.com" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected api.example.com in hosts")
	}
}

func TestPathTemplateInference(t *testing.T) {
	inferrer := NewPathInferrer()

	tests := []struct {
		path     string
		template string
	}{
		{"/users", "/users"},
		{"/users/123", "/users/{userId}"},
		{"/users/550e8400-e29b-41d4-a716-446655440000", "/users/{userId}"},
		{"/posts/456/comments/789", "/posts/{postId}/comments/{commentId}"},
		{"/api/v1/users", "/api/v1/users"},
		{"/orders/12345/items/67890", "/orders/{orderId}/items/{itemId}"},
	}

	for _, tt := range tests {
		template, _ := inferrer.InferTemplate(tt.path)
		if template != tt.template {
			t.Errorf("InferTemplate(%q) = %q, want %q", tt.path, template, tt.template)
		}
	}
}

func TestSchemaInference(t *testing.T) {
	store := NewSchemaStore()

	// Simulate multiple observations
	body1 := map[string]any{
		"id":    "123",
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   float64(30),
	}
	body2 := map[string]any{
		"id":     "456",
		"name":   "Bob",
		"email":  "bob@example.com",
		"active": true,
	}

	ProcessBody(store, body1)
	ProcessBody(store, body2)
	store.FinalizeOptional()

	// Check types
	if store.Types["id"] != TypeString {
		t.Errorf("expected id type string, got %s", store.Types["id"])
	}
	if store.Types["age"] != TypeInteger {
		t.Errorf("expected age type integer, got %s", store.Types["age"])
	}
	if store.Types["active"] != TypeBoolean {
		t.Errorf("expected active type boolean, got %s", store.Types["active"])
	}

	// Check optionality (age only in body1, active only in body2)
	if !store.Optional["age"] {
		t.Error("expected age to be optional")
	}
	if !store.Optional["active"] {
		t.Error("expected active to be optional")
	}

	// id and name should be required (in both)
	if store.Optional["id"] {
		t.Error("expected id to be required")
	}
	if store.Optional["name"] {
		t.Error("expected name to be required")
	}

	// Check format detection
	if store.Formats["email"] != FormatEmail {
		t.Errorf("expected email format, got %s", store.Formats["email"])
	}
}

func TestArraySchemaInference(t *testing.T) {
	store := NewSchemaStore()

	body := map[string]any{
		"items": []any{
			map[string]any{"id": "1", "name": "Item 1"},
			map[string]any{"id": "2", "name": "Item 2"},
		},
		"total": float64(2),
	}

	ProcessBody(store, body)

	// Check array items
	if store.Types["items[].id"] != TypeString {
		t.Errorf("expected items[].id type string, got %s", store.Types["items[].id"])
	}
	if store.Types["items[].name"] != TypeString {
		t.Errorf("expected items[].name type string, got %s", store.Types["items[].name"])
	}
	if store.Types["total"] != TypeInteger {
		t.Errorf("expected total type integer, got %s", store.Types["total"])
	}
}

func TestEndToEndInference(t *testing.T) {
	// Create some IR records
	records := []ir.IRRecord{
		{
			Request: ir.Request{
				Method: ir.RequestMethodGET,
				Path:   "/users",
			},
			Response: ir.Response{
				Status: 200,
				Body: map[string]any{
					"users": []any{
						map[string]any{"id": "1", "name": "Alice"},
					},
				},
			},
		},
		{
			Request: ir.Request{
				Method: ir.RequestMethodGET,
				Path:   "/users/1",
			},
			Response: ir.Response{
				Status: 200,
				Body: map[string]any{
					"id":   "1",
					"name": "Alice",
				},
			},
		},
		{
			Request: ir.Request{
				Method: ir.RequestMethodPOST,
				Path:   "/users",
				Body: map[string]any{
					"name":  "Bob",
					"email": "bob@example.com",
				},
			},
			Response: ir.Response{
				Status: 201,
				Body: map[string]any{
					"id":    "2",
					"name":  "Bob",
					"email": "bob@example.com",
				},
			},
		},
	}

	result := InferFromRecords(records)

	// Check endpoints
	if len(result.Endpoints) != 3 {
		t.Errorf("expected 3 endpoints, got %d", len(result.Endpoints))
	}

	// Check GET /users
	getUsers := result.Endpoints["GET /users"]
	if getUsers == nil {
		t.Fatal("GET /users endpoint not found")
	}
	if getUsers.RequestCount != 1 {
		t.Errorf("expected 1 request, got %d", getUsers.RequestCount)
	}

	// Check GET /users/{userId}
	getUser := result.Endpoints["GET /users/{userId}"]
	if getUser == nil {
		t.Fatal("GET /users/{userId} endpoint not found")
	}

	// Check POST /users has request body
	postUsers := result.Endpoints["POST /users"]
	if postUsers == nil {
		t.Fatal("POST /users endpoint not found")
	}
	if postUsers.RequestBody == nil {
		t.Error("POST /users should have request body")
	}
}
