package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFile(t *testing.T) {
	// Create a temporary valid spec file
	tmpDir := t.TempDir()
	validSpec := `openapi: "3.1.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test:
    get:
      responses:
        "200":
          description: Success
`
	validPath := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validPath, []byte(validSpec), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ValidateFile(validPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidateFileNotFound(t *testing.T) {
	_, err := ValidateFile("/nonexistent/path/to/spec.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseAndValidate(t *testing.T) {
	spec := []byte(`
openapi: "3.1.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test:
    get:
      responses:
        "200":
          description: Success
`)

	doc, result, err := ParseAndValidate(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil {
		t.Error("expected document, got nil")
	}
	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestParseAndValidateEmpty(t *testing.T) {
	_, _, err := ParseAndValidate([]byte{})
	if err == nil {
		t.Error("expected error for empty spec")
	}
}

func TestParseAndValidateInvalid(t *testing.T) {
	spec := []byte(`not valid yaml: [`)

	doc, result, err := ParseAndValidate(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Document may be nil for parse errors
	if doc == nil && (result == nil || result.Valid) {
		t.Error("expected invalid result for malformed input")
	}
}

func TestValidationErrorError(t *testing.T) {
	// Test ValidationError.Error() with line number
	e := ValidationError{
		Message:  "test error",
		Line:     10,
		Column:   5,
		Path:     "/paths/test",
		Severity: "error",
	}
	errStr := e.Error()
	if errStr != "/paths/test:10:5: test error" {
		t.Errorf("Error() = %q, want %q", errStr, "/paths/test:10:5: test error")
	}

	// Test ValidationError.Error() without line number
	e2 := ValidationError{
		Message:  "test error without line",
		Severity: "error",
	}
	errStr2 := e2.Error()
	if errStr2 != "test error without line" {
		t.Errorf("Error() = %q, want %q", errStr2, "test error without line")
	}
}

func TestValidateOpenAPI32(t *testing.T) {
	spec := []byte(`
openapi: "3.2.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test:
    get:
      responses:
        "200":
          description: Success
`)

	result, err := Validate(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
	if result.Version != "3.2.0" {
		t.Errorf("Version = %q, want %q", result.Version, "3.2.0")
	}
}

func TestValidateWithComplexSpec(t *testing.T) {
	spec := []byte(`
openapi: "3.1.0"
info:
  title: Complex API
  version: "1.0.0"
  description: A complex test API
  contact:
    name: Support
    email: support@example.com
servers:
  - url: https://api.example.com
    description: Production
paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        "200":
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
    post:
      summary: Create user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        "201":
          description: Created
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
      required:
        - id
        - name
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
`)

	result, err := Validate(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidateJSON(t *testing.T) {
	spec := []byte(`{
  "openapi": "3.1.0",
  "info": {
    "title": "JSON API",
    "version": "1.0.0"
  },
  "paths": {
    "/test": {
      "get": {
        "responses": {
          "200": {
            "description": "Success"
          }
        }
      }
    }
  }
}`)

	result, err := Validate(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}
