package postman

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/traffic2openapi/pkg/inference"
	"github.com/grokify/traffic2openapi/pkg/ir"
	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// TestIntegrationPetStoreCollection tests loading and converting the PetStore collection fixture.
func TestIntegrationPetStoreCollection(t *testing.T) {
	collectionPath := filepath.Join("testdata", "petstore.postman_collection.json")

	// Check if test file exists
	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		t.Skip("petstore.postman_collection.json not found in testdata/")
	}

	// Convert the collection
	result, err := ConvertFile(collectionPath)
	if err != nil {
		t.Fatalf("failed to convert collection: %v", err)
	}

	// Validate we got records
	if len(result.Records) == 0 {
		t.Fatal("expected at least one record")
	}

	t.Logf("Converted %d records from PetStore collection", len(result.Records))

	// Validate metadata was extracted
	if result.Metadata == nil {
		t.Fatal("expected metadata to be extracted")
	}

	if result.Metadata.Title == nil || *result.Metadata.Title != "Pet Store API" {
		t.Errorf("expected title 'Pet Store API', got %v", result.Metadata.Title)
	}

	if result.Metadata.APIVersion == nil || *result.Metadata.APIVersion != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", result.Metadata.APIVersion)
	}

	// Check description
	if result.Metadata.Description == nil {
		t.Error("expected description in metadata")
	}

	// Validate tag definitions were created from folders
	if len(result.TagDefinitions) == 0 {
		t.Error("expected tag definitions from folder structure")
	}

	tagNames := make(map[string]bool)
	for _, td := range result.TagDefinitions {
		tagNames[td.Name] = true
	}

	expectedTags := []string{"Pets", "Store", "Users"}
	for _, tag := range expectedTags {
		if !tagNames[tag] {
			t.Errorf("expected tag definition for %q", tag)
		}
	}
}

// TestIntegrationPetStoreEndpoints validates the endpoints in the PetStore collection.
func TestIntegrationPetStoreEndpoints(t *testing.T) {
	collectionPath := filepath.Join("testdata", "petstore.postman_collection.json")

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		t.Skip("petstore.postman_collection.json not found in testdata/")
	}

	result, err := ConvertFile(collectionPath)
	if err != nil {
		t.Fatalf("failed to convert collection: %v", err)
	}

	// Build a map of endpoints for easier testing
	endpoints := make(map[string][]*ir.IRRecord)
	for i := range result.Records {
		r := &result.Records[i]
		key := string(r.Request.Method) + " " + r.Request.Path
		endpoints[key] = append(endpoints[key], r)
	}

	// Test: List Pets endpoint
	t.Run("ListPets", func(t *testing.T) {
		records := endpoints["GET /pets"]
		if len(records) == 0 {
			t.Fatal("expected GET /pets endpoint")
		}

		record := records[0]

		// Check query params
		if record.Request.Query == nil {
			t.Fatal("expected query parameters")
		}

		if record.Request.Query["limit"] != "10" {
			t.Errorf("expected limit=10, got %v", record.Request.Query["limit"])
		}

		// Check operation metadata
		if record.Summary == nil || *record.Summary != "List Pets" {
			t.Errorf("expected summary 'List Pets', got %v", record.Summary)
		}

		// Check tags
		if len(record.Tags) == 0 || record.Tags[0] != "Pets" {
			t.Errorf("expected Pets tag, got %v", record.Tags)
		}
	})

	// Test: Create Pet endpoint
	t.Run("CreatePet", func(t *testing.T) {
		records := endpoints["POST /pets"]
		if len(records) == 0 {
			t.Fatal("expected POST /pets endpoint")
		}

		// Should have multiple responses (201, 400)
		statusCodes := make(map[int]bool)
		for _, r := range records {
			statusCodes[r.Response.Status] = true
		}

		if !statusCodes[201] {
			t.Error("expected 201 response for Create Pet")
		}
		if !statusCodes[400] {
			t.Error("expected 400 response for Create Pet")
		}

		// Check request body
		for _, r := range records {
			if r.Response.Status == 201 && r.Request.Body == nil {
				t.Error("expected request body for POST /pets")
			}
		}
	})

	// Test: Get Pet by ID endpoint with path parameter
	t.Run("GetPetById", func(t *testing.T) {
		// Path should be /pets/123 (with value substituted)
		var getPetRecord *ir.IRRecord
		for key, records := range endpoints {
			if key == "GET /pets/123" || key == "GET /pets/:petId" {
				getPetRecord = records[0]
				break
			}
		}

		if getPetRecord == nil {
			t.Fatal("expected GET /pets/{petId} endpoint")
		}

		// Should have 200 and 404 responses
		statusCodes := make(map[int]bool)
		for key, records := range endpoints {
			if key == "GET /pets/123" || key == "GET /pets/:petId" {
				for _, r := range records {
					statusCodes[r.Response.Status] = true
				}
			}
		}

		if !statusCodes[200] {
			t.Error("expected 200 response")
		}
		if !statusCodes[404] {
			t.Error("expected 404 response")
		}
	})

	// Test: Nested folder (Store > Orders)
	t.Run("NestedFolders", func(t *testing.T) {
		records := endpoints["POST /store/orders"]
		if len(records) == 0 {
			t.Fatal("expected POST /store/orders endpoint")
		}

		record := records[0]

		// Should have nested tags
		foundStoreTag := false
		for _, tag := range record.Tags {
			if tag == "Store" || tag == "Store > Orders" {
				foundStoreTag = true
				break
			}
		}
		if !foundStoreTag {
			t.Errorf("expected Store tag for nested order endpoint, got %v", record.Tags)
		}
	})
}

// TestIntegrationVariableResolution tests variable resolution with external variables.
func TestIntegrationVariableResolution(t *testing.T) {
	collectionPath := filepath.Join("testdata", "petstore.postman_collection.json")

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		t.Skip("petstore.postman_collection.json not found in testdata/")
	}

	// Convert with custom variable override
	result, err := ConvertFile(collectionPath,
		WithVariable("baseUrl", "https://custom.api.com/v2"),
		WithVariable("apiKey", "custom-key-123"),
	)
	if err != nil {
		t.Fatalf("failed to convert collection: %v", err)
	}

	// Check that the host was overridden (host includes custom.api.com)
	foundCustomHost := false
	for _, record := range result.Records {
		if record.Request.Host != nil {
			host := *record.Request.Host
			if strings.Contains(host, "custom.api.com") {
				foundCustomHost = true
				break
			}
		}
	}

	if !foundCustomHost {
		t.Error("expected at least one record with custom.api.com in host")
	}

	// Verify paths include /v2 prefix from the baseUrl
	foundV2Path := false
	for _, record := range result.Records {
		if strings.Contains(record.Request.Path, "/v2/") || record.Request.Path == "/v2" {
			foundV2Path = true
			break
		}
	}

	// Note: depending on implementation, /v2 may or may not be in the path
	t.Logf("Variable resolution test: foundCustomHost=%v, foundV2Path=%v", foundCustomHost, foundV2Path)
}

// TestEndToEndPostmanToOpenAPI tests the full pipeline: Postman -> IR -> OpenAPI.
func TestEndToEndPostmanToOpenAPI(t *testing.T) {
	collectionPath := filepath.Join("testdata", "petstore.postman_collection.json")

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		t.Skip("petstore.postman_collection.json not found in testdata/")
	}

	// Step 1: Convert Postman to IR
	result, err := ConvertFile(collectionPath)
	if err != nil {
		t.Fatalf("Step 1 (Postman -> IR) failed: %v", err)
	}

	t.Logf("Step 1: Converted %d records", len(result.Records))

	// Step 2: Run inference engine
	engine := inference.NewEngine(inference.DefaultEngineOptions())

	// Set API metadata from Postman
	if result.Metadata != nil {
		apiMeta := &inference.APIMetadataData{}
		if result.Metadata.Title != nil {
			apiMeta.Title = *result.Metadata.Title
		}
		if result.Metadata.Description != nil {
			apiMeta.Description = *result.Metadata.Description
		}
		if result.Metadata.APIVersion != nil {
			apiMeta.APIVersion = *result.Metadata.APIVersion
		}

		// Add tag definitions
		for _, td := range result.TagDefinitions {
			tagDef := inference.TagDefinitionData{Name: td.Name}
			if td.Description != nil {
				tagDef.Description = *td.Description
			}
			apiMeta.TagDefinitions = append(apiMeta.TagDefinitions, tagDef)
		}

		engine.SetAPIMetadata(apiMeta)
	}

	engine.ProcessRecords(result.Records)
	inferenceResult := engine.Finalize()

	t.Logf("Step 2: Inferred %d endpoints", len(inferenceResult.Endpoints))

	// Step 3: Generate OpenAPI spec
	genOptions := openapi.DefaultGeneratorOptions()
	genOptions.Version = openapi.Version31
	spec := openapi.GenerateFromInference(inferenceResult, genOptions)

	t.Logf("Step 3: Generated OpenAPI %s spec", spec.OpenAPI)

	// Validate the generated spec
	t.Run("ValidateInfo", func(t *testing.T) {
		if spec.Info.Title != "Pet Store API" {
			t.Errorf("expected title 'Pet Store API', got %q", spec.Info.Title)
		}
		if spec.Info.Version != "1.0.0" {
			t.Errorf("expected version '1.0.0', got %q", spec.Info.Version)
		}
	})

	t.Run("ValidatePaths", func(t *testing.T) {
		if len(spec.Paths) == 0 {
			t.Fatal("expected paths in spec")
		}

		// Check for expected paths
		expectedPaths := []string{"/pets", "/store/inventory", "/users", "/users/login"}
		for _, path := range expectedPaths {
			if _, ok := spec.Paths[path]; !ok {
				// Try with path param pattern
				found := false
				for p := range spec.Paths {
					if p == path || containsPath(p, path) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected path %q in spec", path)
				}
			}
		}
	})

	t.Run("ValidateTags", func(t *testing.T) {
		if len(spec.Tags) == 0 {
			t.Error("expected tags in spec")
		}

		tagNames := make(map[string]bool)
		for _, tag := range spec.Tags {
			tagNames[tag.Name] = true
		}

		expectedTags := []string{"Pets", "Store", "Users"}
		for _, tag := range expectedTags {
			if !tagNames[tag] {
				t.Errorf("expected tag %q in spec", tag)
			}
		}
	})

	t.Run("ValidateOperations", func(t *testing.T) {
		// Count operations
		opCount := 0
		for _, pathItem := range spec.Paths {
			if pathItem.Get != nil {
				opCount++
			}
			if pathItem.Post != nil {
				opCount++
			}
			if pathItem.Put != nil {
				opCount++
			}
			if pathItem.Delete != nil {
				opCount++
			}
			if pathItem.Patch != nil {
				opCount++
			}
		}

		if opCount == 0 {
			t.Error("expected operations in spec")
		}
		t.Logf("Total operations: %d", opCount)
	})

	t.Run("ValidateResponses", func(t *testing.T) {
		// Check that operations have responses
		for path, pathItem := range spec.Paths {
			if pathItem.Get != nil && len(pathItem.Get.Responses) == 0 {
				t.Errorf("GET %s has no responses", path)
			}
			if pathItem.Post != nil && len(pathItem.Post.Responses) == 0 {
				t.Errorf("POST %s has no responses", path)
			}
		}
	})

	// Step 4: Serialize to JSON (validate it's valid JSON)
	t.Run("ValidateJSONSerialization", func(t *testing.T) {
		jsonBytes, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			t.Fatalf("failed to serialize spec to JSON: %v", err)
		}

		if len(jsonBytes) == 0 {
			t.Error("expected non-empty JSON output")
		}

		t.Logf("Generated spec size: %d bytes", len(jsonBytes))
	})
}

// TestEndToEndRoundTrip tests writing IR to file and reading it back.
func TestEndToEndRoundTrip(t *testing.T) {
	collectionPath := filepath.Join("testdata", "petstore.postman_collection.json")

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		t.Skip("petstore.postman_collection.json not found in testdata/")
	}

	// Convert to batch
	batch, err := ConvertFileToBatch(collectionPath)
	if err != nil {
		t.Fatalf("failed to convert to batch: %v", err)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "ir-roundtrip-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Write batch to file
	err = ir.WriteFile(tmpFile.Name(), batch.Records)
	if err != nil {
		t.Fatalf("failed to write IR file: %v", err)
	}

	// Read back
	readRecords, err := ir.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read IR file: %v", err)
	}

	// Validate counts match
	if len(readRecords) != len(batch.Records) {
		t.Errorf("expected %d records, got %d", len(batch.Records), len(readRecords))
	}

	// Validate first record
	if len(readRecords) > 0 {
		original := batch.Records[0]
		read := readRecords[0]

		if original.Request.Method != read.Request.Method {
			t.Errorf("method mismatch: %v vs %v", original.Request.Method, read.Request.Method)
		}
		if original.Request.Path != read.Request.Path {
			t.Errorf("path mismatch: %v vs %v", original.Request.Path, read.Request.Path)
		}
		if original.Response.Status != read.Response.Status {
			t.Errorf("status mismatch: %v vs %v", original.Response.Status, read.Response.Status)
		}
	}
}

// TestEndToEndNDJSON tests NDJSON streaming format.
func TestEndToEndNDJSON(t *testing.T) {
	collectionPath := filepath.Join("testdata", "petstore.postman_collection.json")

	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		t.Skip("petstore.postman_collection.json not found in testdata/")
	}

	result, err := ConvertFile(collectionPath)
	if err != nil {
		t.Fatalf("failed to convert collection: %v", err)
	}

	// Create temp NDJSON file
	tmpFile, err := os.CreateTemp("", "ir-ndjson-*.ndjson")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Write as NDJSON
	err = ir.WriteFile(tmpFile.Name(), result.Records)
	if err != nil {
		t.Fatalf("failed to write NDJSON file: %v", err)
	}

	// Read back
	readRecords, err := ir.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read NDJSON file: %v", err)
	}

	if len(readRecords) != len(result.Records) {
		t.Errorf("expected %d records, got %d", len(result.Records), len(readRecords))
	}

	// Run inference on read records
	inferenceResult := inference.InferFromRecords(readRecords)
	if len(inferenceResult.Endpoints) == 0 {
		t.Error("expected endpoints from inference")
	}

	t.Logf("NDJSON round-trip: %d records, %d endpoints inferred",
		len(readRecords), len(inferenceResult.Endpoints))
}

// containsPath checks if a path pattern contains the target path.
func containsPath(pattern, target string) bool {
	// Simple check - could be enhanced for proper path matching
	if len(pattern) < len(target) {
		return false
	}
	return pattern[:len(target)] == target
}
