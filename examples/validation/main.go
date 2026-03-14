// Example: OpenAPI Specification Validation
//
// This example demonstrates how to validate OpenAPI specifications
// using the libopenapi-based validation package.
package main

import (
	"fmt"
	"log"

	"github.com/grokify/traffic2openapi/pkg/openapi/validate"
)

func main() {
	// Example 1: Validate a valid OpenAPI 3.1 spec
	validSpec := []byte(`
openapi: "3.1.0"
info:
  title: Valid API
  version: "1.0.0"
paths:
  /users:
    get:
      summary: List users
      responses:
        "200":
          description: A list of users
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
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
`)

	fmt.Println("=== Validating Valid Spec ===")
	result, err := validate.Validate(validSpec)
	if err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	if result.Valid {
		fmt.Printf("Spec is valid (OpenAPI %s)\n", result.Version)
	} else {
		fmt.Println("Spec is invalid:")
		for _, e := range result.Errors {
			fmt.Printf("  ERROR: %s\n", e.Message)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("Warnings: %d\n", len(result.Warnings))
		for _, w := range result.Warnings {
			fmt.Printf("  WARN: %s\n", w.Message)
		}
	}

	// Example 2: Validate an invalid spec (missing required fields)
	invalidSpec := []byte(`
openapi: "3.1.0"
info:
  title: Invalid API
  # missing version
paths: {}
`)

	fmt.Println("\n=== Validating Invalid Spec ===")
	result, err = validate.Validate(invalidSpec)
	if err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	if result.Valid {
		fmt.Println("Spec is valid")
	} else {
		fmt.Println("Spec is invalid (as expected):")
		for _, e := range result.Errors {
			fmt.Printf("  ERROR: %s\n", e.Message)
		}
	}

	// Example 3: Parse and validate (get both document and validation result)
	fmt.Println("\n=== Parse and Validate ===")
	doc, result, err := validate.ParseAndValidate(validSpec)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	if doc != nil && result.Valid {
		fmt.Printf("Successfully parsed and validated OpenAPI %s spec\n", result.Version)
		// The document can now be used for further processing
	}

	// Example 4: Check version validity
	fmt.Println("\n=== Version Checking ===")
	versions := []string{"3.0.3", "3.1.0", "3.2.0", "2.0.0", "invalid"}
	for _, v := range versions {
		if validate.IsValidVersion(v) {
			fmt.Printf("  %s: valid OpenAPI version\n", v)
		} else {
			fmt.Printf("  %s: not a valid OpenAPI 3.x version\n", v)
		}
	}
}
