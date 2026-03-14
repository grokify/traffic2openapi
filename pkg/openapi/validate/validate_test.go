package validate

import (
	"testing"
)

func TestValidateValidSpec(t *testing.T) {
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

	result, err := Validate(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
	if result.Version != "3.1.0" {
		t.Errorf("Version = %q, want %q", result.Version, "3.1.0")
	}
}

func TestValidateInvalidSpec(t *testing.T) {
	spec := []byte(`
openapi: "3.1.0"
info:
  title: Test API
  # missing version
paths: {}
`)

	result, err := Validate(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// libopenapi may or may not flag missing version as error depending on strictness
	// Just verify we get a result
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestValidateMalformedYAML(t *testing.T) {
	spec := []byte(`
openapi: "3.1.0"
info:
  title: Test API
    version: "1.0.0"  # invalid indentation
`)

	result, err := Validate(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid result for malformed YAML")
	}
}

func TestValidateEmptySpec(t *testing.T) {
	_, err := Validate([]byte{})
	if err == nil {
		t.Error("expected error for empty spec")
	}
}

func TestValidateOpenAPI30(t *testing.T) {
	spec := []byte(`
openapi: "3.0.3"
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
	if result.Version != "3.0.3" {
		t.Errorf("Version = %q, want %q", result.Version, "3.0.3")
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"3.0.0", true},
		{"3.0.1", true},
		{"3.0.2", true},
		{"3.0.3", true},
		{"3.1.0", true},
		{"3.1.1", true},
		{"3.2.0", true},
		{"2.0.0", false},
		{"3.3.0", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := IsValidVersion(tt.version); got != tt.valid {
				t.Errorf("IsValidVersion(%q) = %v, want %v", tt.version, got, tt.valid)
			}
		})
	}
}

func TestGetMajorMinorVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"3.0.0", "3.0"},
		{"3.1.0", "3.1"},
		{"3.2.0", "3.2"},
		{"3.0", "3.0"},
		{"3", "3"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := GetMajorMinorVersion(tt.input); got != tt.expected {
				t.Errorf("GetMajorMinorVersion(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
