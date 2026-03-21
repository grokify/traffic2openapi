package postman

import (
	"strings"
	"testing"

	postman "github.com/rbretecher/go-postman-collection"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

func TestConverterBasic(t *testing.T) {
	// Create a minimal collection
	collection := &postman.Collection{
		Info: postman.Info{
			Name: "Test API",
			Description: postman.Description{
				Content: "Test API Description",
			},
			Version: "1.0.0",
		},
		Items: []*postman.Items{
			{
				Name:        "Get Users",
				Description: "Retrieves all users",
				Request: &postman.Request{
					Method: postman.Get,
					URL: &postman.URL{
						Raw:      "https://api.example.com/users",
						Protocol: "https",
						Host:     []string{"api", "example", "com"},
						Path:     []string{"users"},
					},
				},
				Responses: []*postman.Response{
					{
						Name:   "Success",
						Code:   200,
						Status: "OK",
						Body:   `{"users":[]}`,
					},
				},
			},
		},
	}

	converter := NewConverter()
	result, err := converter.Convert(collection)

	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	if len(result.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(result.Records))
	}

	record := result.Records[0]

	// Check request
	if record.Request.Method != ir.RequestMethodGET {
		t.Errorf("expected GET, got %s", record.Request.Method)
	}

	if record.Request.Path != "/users" {
		t.Errorf("expected /users, got %s", record.Request.Path)
	}

	if record.Request.Host == nil || *record.Request.Host != "api.example.com" {
		t.Errorf("expected api.example.com, got %v", record.Request.Host)
	}

	// Check response
	if record.Response.Status != 200 {
		t.Errorf("expected 200, got %d", record.Response.Status)
	}

	// Check metadata
	if record.OperationId == nil || *record.OperationId != "getUsers" {
		t.Errorf("expected getUsers, got %v", record.OperationId)
	}

	if record.Summary == nil || *record.Summary != "Get Users" {
		t.Errorf("expected 'Get Users', got %v", record.Summary)
	}

	if record.Description == nil || !strings.Contains(*record.Description, "Retrieves all users") {
		t.Errorf("expected description containing 'Retrieves all users', got %v", record.Description)
	}

	// Check source
	if record.Source == nil || *record.Source != ir.IRRecordSourcePostman {
		t.Errorf("expected postman source, got %v", record.Source)
	}

	// Check API metadata
	if result.Metadata == nil {
		t.Fatal("expected metadata")
	}

	if result.Metadata.Title == nil || *result.Metadata.Title != "Test API" {
		t.Errorf("expected 'Test API', got %v", result.Metadata.Title)
	}

	if result.Metadata.APIVersion == nil || *result.Metadata.APIVersion != "1.0.0" {
		t.Errorf("expected '1.0.0', got %v", result.Metadata.APIVersion)
	}
}

func TestConverterFolderStructure(t *testing.T) {
	// Create collection with nested folders
	collection := &postman.Collection{
		Info: postman.Info{Name: "Test API"},
		Items: []*postman.Items{
			{
				Name:        "Users",
				Description: "User management endpoints",
				Items: []*postman.Items{
					{
						Name: "Get User",
						Request: &postman.Request{
							Method: postman.Get,
							URL:    &postman.URL{Raw: "https://api.example.com/users/1"},
						},
						Responses: []*postman.Response{{Code: 200}},
					},
					{
						Name: "Create User",
						Request: &postman.Request{
							Method: postman.Post,
							URL:    &postman.URL{Raw: "https://api.example.com/users"},
						},
						Responses: []*postman.Response{{Code: 201}},
					},
				},
			},
			{
				Name: "Products",
				Items: []*postman.Items{
					{
						Name: "List Products",
						Request: &postman.Request{
							Method: postman.Get,
							URL:    &postman.URL{Raw: "https://api.example.com/products"},
						},
						Responses: []*postman.Response{{Code: 200}},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result, err := converter.Convert(collection)

	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	if len(result.Records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(result.Records))
	}

	// Check tags
	userRecords := 0
	productRecords := 0
	for _, r := range result.Records {
		for _, tag := range r.Tags {
			if tag == "Users" {
				userRecords++
			}
			if tag == "Products" {
				productRecords++
			}
		}
	}

	if userRecords != 2 {
		t.Errorf("expected 2 records with Users tag, got %d", userRecords)
	}

	if productRecords != 1 {
		t.Errorf("expected 1 record with Products tag, got %d", productRecords)
	}

	// Check tag definitions
	if len(result.TagDefinitions) != 2 {
		t.Errorf("expected 2 tag definitions, got %d", len(result.TagDefinitions))
	}

	usersTagFound := false
	for _, td := range result.TagDefinitions {
		if td.Name == "Users" {
			usersTagFound = true
			if td.Description == nil || *td.Description != "User management endpoints" {
				t.Errorf("expected Users tag description, got %v", td.Description)
			}
		}
	}
	if !usersTagFound {
		t.Error("Users tag definition not found")
	}
}

func TestConverterVariableResolution(t *testing.T) {
	collection := &postman.Collection{
		Info: postman.Info{Name: "Test API"},
		Variables: []*postman.Variable{
			{Key: "apiPath", Value: "api/v1"},
		},
		Items: []*postman.Items{
			{
				Name: "Get Users",
				Request: &postman.Request{
					Method: postman.Get,
					URL: &postman.URL{
						Raw:      "{{url}}/{{apiPath}}/users",
						Protocol: "https",
						Host:     []string{"{{url}}"},
						Path:     []string{"{{apiPath}}", "users"},
					},
				},
				Responses: []*postman.Response{{Code: 200}},
			},
		},
	}

	converter := NewConverter()
	converter.BaseURL = "api.example.com"

	result, err := converter.Convert(collection)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	record := result.Records[0]

	if record.Request.Host == nil || *record.Request.Host != "api.example.com" {
		t.Errorf("expected api.example.com, got %v", record.Request.Host)
	}

	if record.Request.Path != "/api/v1/users" {
		t.Errorf("expected /api/v1/users, got %s", record.Request.Path)
	}
}

func TestConverterQueryParams(t *testing.T) {
	collection := &postman.Collection{
		Info: postman.Info{Name: "Test API"},
		Items: []*postman.Items{
			{
				Name: "Search",
				Request: &postman.Request{
					Method: postman.Get,
					URL: &postman.URL{
						Raw:      "https://api.example.com/search?q=test&limit=10",
						Protocol: "https",
						Host:     []string{"api", "example", "com"},
						Path:     []string{"search"},
						Query: []*postman.QueryParam{
							{Key: "q", Value: "test"},
							{Key: "limit", Value: "10"},
						},
					},
				},
				Responses: []*postman.Response{{Code: 200}},
			},
		},
	}

	converter := NewConverter()
	result, err := converter.Convert(collection)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	record := result.Records[0]

	if record.Request.Query == nil {
		t.Fatal("expected query params")
	}

	if record.Request.Query["q"] != "test" {
		t.Errorf("expected q=test, got %v", record.Request.Query["q"])
	}

	if record.Request.Query["limit"] != "10" {
		t.Errorf("expected limit=10, got %v", record.Request.Query["limit"])
	}
}

func TestConverterRequestBody(t *testing.T) {
	collection := &postman.Collection{
		Info: postman.Info{Name: "Test API"},
		Items: []*postman.Items{
			{
				Name: "Create User",
				Request: &postman.Request{
					Method: postman.Post,
					URL:    &postman.URL{Raw: "https://api.example.com/users"},
					Body: &postman.Body{
						Mode: "raw",
						Raw:  `{"name":"Alice","email":"alice@example.com"}`,
						Options: &postman.BodyOptions{
							Raw: postman.BodyOptionsRaw{Language: "json"},
						},
					},
				},
				Responses: []*postman.Response{{Code: 201}},
			},
		},
	}

	converter := NewConverter()
	result, err := converter.Convert(collection)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	record := result.Records[0]

	if record.Request.Body == nil {
		t.Fatal("expected request body")
	}

	bodyMap, ok := record.Request.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", record.Request.Body)
	}

	if bodyMap["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %v", bodyMap["name"])
	}

	if record.Request.ContentType == nil || *record.Request.ContentType != "application/json" {
		t.Errorf("expected application/json, got %v", record.Request.ContentType)
	}
}

func TestConverterMultipleResponses(t *testing.T) {
	collection := &postman.Collection{
		Info: postman.Info{Name: "Test API"},
		Items: []*postman.Items{
			{
				Name: "Get User",
				Request: &postman.Request{
					Method: postman.Get,
					URL:    &postman.URL{Raw: "https://api.example.com/users/1"},
				},
				Responses: []*postman.Response{
					{
						Name: "Success",
						Code: 200,
						Body: `{"id":1,"name":"Alice"}`,
					},
					{
						Name: "Not Found",
						Code: 404,
						Body: `{"error":"User not found"}`,
					},
				},
			},
		},
	}

	converter := NewConverter()
	result, err := converter.Convert(collection)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	if len(result.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(result.Records))
	}

	// Check first response
	if result.Records[0].Response.Status != 200 {
		t.Errorf("expected 200, got %d", result.Records[0].Response.Status)
	}

	// Check second response
	if result.Records[1].Response.Status != 404 {
		t.Errorf("expected 404, got %d", result.Records[1].Response.Status)
	}
}

func TestConverterHeaders(t *testing.T) {
	collection := &postman.Collection{
		Info: postman.Info{Name: "Test API"},
		Items: []*postman.Items{
			{
				Name: "Get Users",
				Request: &postman.Request{
					Method: postman.Get,
					URL:    &postman.URL{Raw: "https://api.example.com/users"},
					Header: []*postman.Header{
						{Key: "Accept", Value: "application/json"},
						{Key: "X-Custom-Header", Value: "custom-value"},
					},
				},
				Responses: []*postman.Response{
					{
						Code: 200,
						Headers: &postman.HeaderList{
							Headers: []*postman.Header{
								{Key: "Content-Type", Value: "application/json"},
							},
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result, err := converter.Convert(collection)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	record := result.Records[0]

	// Check request headers
	if record.Request.Headers == nil {
		t.Fatal("expected request headers")
	}

	if record.Request.Headers["accept"] != "application/json" {
		t.Errorf("expected accept header, got %v", record.Request.Headers["accept"])
	}

	if record.Request.Headers["x-custom-header"] != "custom-value" {
		t.Errorf("expected x-custom-header, got %v", record.Request.Headers["x-custom-header"])
	}

	// Check response headers
	if record.Response.Headers == nil {
		t.Fatal("expected response headers")
	}

	if record.Response.Headers["content-type"] != "application/json" {
		t.Errorf("expected content-type header, got %v", record.Response.Headers["content-type"])
	}
}

func TestSanitizeOperationId(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Get Users", "getUsers"},
		{"Get User by ID", "getUserById"},
		{"create-user", "createUser"},
		{"delete_user", "deleteUser"},
		{"GET /users/{id}", "getUsersId"},
		{"123 Start", "_123Start"},
		{"", ""},
	}

	for _, tc := range tests {
		result := sanitizeOperationId(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeOperationId(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestResolveVars(t *testing.T) {
	vars := map[string]string{
		"url":     "api.example.com",
		"version": "v1",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"{{url}}", "api.example.com"},
		{"{{url}}/{{version}}/users", "api.example.com/v1/users"},
		{"no-vars", "no-vars"},
		{"{{unknown}}", "{{unknown}}"}, // unresolved stays as-is
		{"", ""},
	}

	for _, tc := range tests {
		result := resolveVars(tc.input, vars)
		if result != tc.expected {
			t.Errorf("resolveVars(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestConvertToBatch(t *testing.T) {
	collection := &postman.Collection{
		Info: postman.Info{
			Name:    "Test API",
			Version: "2.0.0",
		},
		Items: []*postman.Items{
			{
				Name: "Health Check",
				Request: &postman.Request{
					Method: postman.Get,
					URL:    &postman.URL{Raw: "https://api.example.com/health"},
				},
				Responses: []*postman.Response{{Code: 200}},
			},
		},
	}

	converter := NewConverter()
	batch, err := converter.ConvertToBatch(collection)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	if batch.Version != ir.Version {
		t.Errorf("expected version %s, got %s", ir.Version, batch.Version)
	}

	if len(batch.Records) != 1 {
		t.Errorf("expected 1 record, got %d", len(batch.Records))
	}

	if batch.Metadata == nil {
		t.Fatal("expected metadata")
	}

	if batch.Metadata.Title == nil || *batch.Metadata.Title != "Test API" {
		t.Errorf("expected 'Test API', got %v", batch.Metadata.Title)
	}
}
