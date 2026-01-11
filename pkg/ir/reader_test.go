package ir

import (
	"path/filepath"
	"testing"
)

func TestReadBatchFile(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "sample-batch.json")
	records, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if len(records) != 4 {
		t.Errorf("expected 4 records, got %d", len(records))
	}

	// Verify first record
	r := records[0]
	if r.Request.Method != RequestMethodGET {
		t.Errorf("expected GET method, got %s", r.Request.Method)
	}
	if r.Request.Path != "/users" {
		t.Errorf("expected /users path, got %s", r.Request.Path)
	}
	if r.Response.Status != 200 {
		t.Errorf("expected status 200, got %d", r.Response.Status)
	}
}

func TestReadNDJSONFile(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "sample-stream.ndjson")
	records, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if len(records) != 5 {
		t.Errorf("expected 5 records, got %d", len(records))
	}

	// Verify record with path template
	r := records[1]
	if r.Request.PathTemplate == nil {
		t.Fatal("expected pathTemplate to be set")
	}
	if *r.Request.PathTemplate != "/users/{id}" {
		t.Errorf("expected /users/{id}, got %s", *r.Request.PathTemplate)
	}

	// Verify effective path template
	if r.EffectivePathTemplate() != "/users/{id}" {
		t.Errorf("EffectivePathTemplate: expected /users/{id}, got %s", r.EffectivePathTemplate())
	}

	// Verify record without path template uses path
	r0 := records[0]
	if r0.EffectivePathTemplate() != "/users" {
		t.Errorf("EffectivePathTemplate: expected /users, got %s", r0.EffectivePathTemplate())
	}
}

func TestReadDir(t *testing.T) {
	dir := filepath.Join("..", "..", "examples")
	records, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	// Should combine both files: 4 + 5 = 9
	if len(records) != 9 {
		t.Errorf("expected 9 records, got %d", len(records))
	}
}

func TestNewRecord(t *testing.T) {
	r := NewRecord(RequestMethodPOST, "/api/items", 201).
		SetID("test-001").
		SetHost("api.example.com").
		SetRequestBody(map[string]string{"name": "test"}).
		SetResponseBody(map[string]interface{}{"id": "123"})

	if r.Request.Method != RequestMethodPOST {
		t.Errorf("expected POST, got %s", r.Request.Method)
	}
	if r.Request.Path != "/api/items" {
		t.Errorf("expected /api/items, got %s", r.Request.Path)
	}
	if r.Response.Status != 201 {
		t.Errorf("expected 201, got %d", r.Response.Status)
	}
	if *r.Id != "test-001" {
		t.Errorf("expected test-001, got %s", *r.Id)
	}
}
