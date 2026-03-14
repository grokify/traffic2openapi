package openapibuilder

import (
	"encoding/json"
	"testing"
)

func TestVersionMethods(t *testing.T) {
	tests := []struct {
		version Version
		is3x    bool
		is30x   bool
		is31x   bool
	}{
		{Version300, true, true, false},
		{Version303, true, true, false},
		{Version310, true, false, true},
		{Version311, true, false, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.version), func(t *testing.T) {
			if got := tt.version.Is3x(); got != tt.is3x {
				t.Errorf("Is3x() = %v, want %v", got, tt.is3x)
			}
			if got := tt.version.Is30x(); got != tt.is30x {
				t.Errorf("Is30x() = %v, want %v", got, tt.is30x)
			}
			if got := tt.version.Is31x(); got != tt.is31x {
				t.Errorf("Is31x() = %v, want %v", got, tt.is31x)
			}
		})
	}
}

func TestNewSpecMinimal(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.OpenAPI != "3.1.0" {
		t.Errorf("OpenAPI = %q, want %q", spec.OpenAPI, "3.1.0")
	}
	if spec.Info.Title != "Test API" {
		t.Errorf("Title = %q, want %q", spec.Info.Title, "Test API")
	}
	if spec.Info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", spec.Info.Version, "1.0.0")
	}
}

func TestNewSpecValidationErrors(t *testing.T) {
	_, err := NewSpec(Version310).Build()
	if err == nil {
		t.Fatal("expected error for missing title and version")
	}

	_, err = NewSpec(Version310).Title("Test").Build()
	if err == nil {
		t.Fatal("expected error for missing version")
	}

	_, err = NewSpec(Version310).Version("1.0.0").Build()
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestSpecWithServer(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Server("https://api.example.com").
		ServerWithDescription("https://staging.example.com", "Staging server").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spec.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(spec.Servers))
	}
	if spec.Servers[0].URL != "https://api.example.com" {
		t.Errorf("Server[0].URL = %q, want %q", spec.Servers[0].URL, "https://api.example.com")
	}
	if spec.Servers[1].Description != "Staging server" {
		t.Errorf("Server[1].Description = %q, want %q", spec.Servers[1].Description, "Staging server")
	}
}

func TestSpecWithPath(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Pet Store").
		Version("1.0.0").
		Path("/pets").
		Get().
		Summary("List pets").
		OperationID("listPets").
		Tags("pets").
		Response(200).Description("Success").JSON(ArraySchema(RefSchema("Pet"))).Done().
		Done().
		Post().
		Summary("Create pet").
		OperationID("createPet").
		JSONBody(RefSchema("Pet")).
		Response(201).Description("Created").JSON(RefSchema("Pet")).Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathItem, ok := spec.Paths["/pets"]
	if !ok {
		t.Fatal("path /pets not found")
	}

	if pathItem.Get == nil {
		t.Fatal("GET operation not found")
	}
	if pathItem.Get.Summary != "List pets" {
		t.Errorf("GET Summary = %q, want %q", pathItem.Get.Summary, "List pets")
	}
	if pathItem.Get.OperationID != "listPets" {
		t.Errorf("GET OperationID = %q, want %q", pathItem.Get.OperationID, "listPets")
	}

	if pathItem.Post == nil {
		t.Fatal("POST operation not found")
	}
	if pathItem.Post.RequestBody == nil {
		t.Fatal("POST request body not found")
	}
	if !pathItem.Post.RequestBody.Required {
		t.Error("POST request body should be required")
	}
}

func TestPathParameters(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/pets/{petId}").
		Get().
		Summary("Get pet").
		PathParam("petId").Type("string").Format("uuid").Description("Pet ID").DoneOp().
		QueryParam("include").Type("string").Enum("owner", "vet").DoneOp().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathItem := spec.Paths["/pets/{petId}"]
	if pathItem.Get == nil {
		t.Fatal("GET operation not found")
	}

	params := pathItem.Get.Parameters
	if len(params) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(params))
	}

	// Path parameter
	if params[0].Name != "petId" {
		t.Errorf("param[0].Name = %q, want %q", params[0].Name, "petId")
	}
	if params[0].In != "path" {
		t.Errorf("param[0].In = %q, want %q", params[0].In, "path")
	}
	if !params[0].Required {
		t.Error("path parameter should be required")
	}
	if params[0].Schema.Format != "uuid" {
		t.Errorf("param[0].Schema.Format = %q, want %q", params[0].Schema.Format, "uuid")
	}

	// Query parameter
	if params[1].Name != "include" {
		t.Errorf("param[1].Name = %q, want %q", params[1].Name, "include")
	}
	if params[1].In != "query" {
		t.Errorf("param[1].In = %q, want %q", params[1].In, "query")
	}
}

func TestResponseWithHeaders(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/items").
		Get().
		Response(200).
		Description("Success").
		JSON(ArraySchema(RefSchema("Item"))).
		Header("X-Total-Count").Type("integer").Description("Total items").Done().
		Header("X-Page").Type("integer").Done().
		Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := spec.Paths["/items"].Get.Responses["200"]
	if len(resp.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(resp.Headers))
	}
	if resp.Headers["X-Total-Count"].Description != "Total items" {
		t.Errorf("header description = %q, want %q", resp.Headers["X-Total-Count"].Description, "Total items")
	}
}

func TestSchemaBuilders(t *testing.T) {
	t.Run("StringSchema", func(t *testing.T) {
		s := StringSchema().Format("email").MinLength(5).MaxLength(100).Build()
		if s.Type != "string" {
			t.Errorf("Type = %v, want %q", s.Type, "string")
		}
		if s.Format != "email" {
			t.Errorf("Format = %q, want %q", s.Format, "email")
		}
		if *s.MinLength != 5 {
			t.Errorf("MinLength = %d, want %d", *s.MinLength, 5)
		}
	})

	t.Run("IntegerSchema", func(t *testing.T) {
		s := IntegerSchema().Minimum(0).Maximum(100).Build()
		if s.Type != "integer" {
			t.Errorf("Type = %v, want %q", s.Type, "integer")
		}
		if *s.Minimum != 0 {
			t.Errorf("Minimum = %f, want %f", *s.Minimum, 0.0)
		}
	})

	t.Run("ObjectSchema", func(t *testing.T) {
		s := ObjectSchema().
			Property("name", StringSchema()).
			Property("age", IntegerSchema()).
			Required("name").
			Build()

		if s.Type != "object" {
			t.Errorf("Type = %v, want %q", s.Type, "object")
		}
		if len(s.Properties) != 2 {
			t.Errorf("Properties count = %d, want %d", len(s.Properties), 2)
		}
		if len(s.Required) != 1 || s.Required[0] != "name" {
			t.Errorf("Required = %v, want %v", s.Required, []string{"name"})
		}
	})

	t.Run("ArraySchema", func(t *testing.T) {
		s := ArraySchema(StringSchema()).MinItems(1).UniqueItems().Build()
		if s.Type != "array" {
			t.Errorf("Type = %v, want %q", s.Type, "array")
		}
		if s.Items == nil {
			t.Error("Items should not be nil")
		}
		if !s.UniqueItems {
			t.Error("UniqueItems should be true")
		}
	})

	t.Run("RefSchema", func(t *testing.T) {
		s := RefSchema("Pet").Build()
		if s.Ref != "#/components/schemas/Pet" {
			t.Errorf("Ref = %q, want %q", s.Ref, "#/components/schemas/Pet")
		}
	})

	t.Run("NullableSchema31", func(t *testing.T) {
		s := StringSchema().WithVersion(Version310).Nullable().Build()
		typeSlice, ok := s.Type.([]string)
		if !ok {
			t.Fatalf("Type should be []string, got %T", s.Type)
		}
		if len(typeSlice) != 2 || typeSlice[0] != "string" || typeSlice[1] != "null" {
			t.Errorf("Type = %v, want %v", typeSlice, []string{"string", "null"})
		}
	})

	t.Run("EnumSchema", func(t *testing.T) {
		s := StringSchema().Enum("active", "inactive", "pending").Build()
		if len(s.Enum) != 3 {
			t.Errorf("Enum count = %d, want %d", len(s.Enum), 3)
		}
	})

	t.Run("CompositionSchema", func(t *testing.T) {
		s := NewSchema().
			AllOf(RefSchema("Base"), ObjectSchema().Property("extra", StringSchema())).
			Build()

		if len(s.AllOf) != 2 {
			t.Errorf("AllOf count = %d, want %d", len(s.AllOf), 2)
		}
	})
}

func TestComponents(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		Schema("Pet", ObjectSchema().
			Property("id", IntegerSchema()).
			Property("name", StringSchema()).
			Required("id", "name")).
		Schema("Error", ObjectSchema().
			Property("code", IntegerSchema()).
			Property("message", StringSchema())).
		SecurityScheme("bearerAuth").BearerAuth().BearerFormat("JWT").Done().
		SecurityScheme("apiKey").APIKeyHeader("X-API-Key").Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Components == nil {
		t.Fatal("Components should not be nil")
	}

	// Check schemas
	if len(spec.Components.Schemas) != 2 {
		t.Errorf("Schemas count = %d, want %d", len(spec.Components.Schemas), 2)
	}
	if spec.Components.Schemas["Pet"] == nil {
		t.Error("Pet schema not found")
	}

	// Check security schemes
	if len(spec.Components.SecuritySchemes) != 2 {
		t.Errorf("SecuritySchemes count = %d, want %d", len(spec.Components.SecuritySchemes), 2)
	}

	bearer := spec.Components.SecuritySchemes["bearerAuth"]
	if bearer == nil {
		t.Fatal("bearerAuth scheme not found")
	}
	if bearer.Type != "http" {
		t.Errorf("bearerAuth.Type = %q, want %q", bearer.Type, "http")
	}
	if bearer.Scheme != "bearer" {
		t.Errorf("bearerAuth.Scheme = %q, want %q", bearer.Scheme, "bearer")
	}
	if bearer.BearerFormat != "JWT" {
		t.Errorf("bearerAuth.BearerFormat = %q, want %q", bearer.BearerFormat, "JWT")
	}

	apiKey := spec.Components.SecuritySchemes["apiKey"]
	if apiKey == nil {
		t.Fatal("apiKey scheme not found")
	}
	if apiKey.Type != "apiKey" {
		t.Errorf("apiKey.Type = %q, want %q", apiKey.Type, "apiKey")
	}
	if apiKey.In != "header" {
		t.Errorf("apiKey.In = %q, want %q", apiKey.In, "header")
	}
}

func TestOAuth2Flow(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		SecurityScheme("oauth2").
		OAuth2().
		AuthorizationCode("https://auth.example.com/authorize", "https://auth.example.com/token").
		Scope("read:pets", "Read pets").
		Scope("write:pets", "Write pets").
		RefreshURL("https://auth.example.com/refresh").
		Done().
		Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	oauth2 := spec.Components.SecuritySchemes["oauth2"]
	if oauth2 == nil {
		t.Fatal("oauth2 scheme not found")
	}
	if oauth2.Type != "oauth2" {
		t.Errorf("Type = %q, want %q", oauth2.Type, "oauth2")
	}
	if oauth2.Flows == nil {
		t.Fatal("Flows should not be nil")
	}
	if oauth2.Flows.AuthorizationCode == nil {
		t.Fatal("AuthorizationCode flow not found")
	}

	flow := oauth2.Flows.AuthorizationCode
	if flow.AuthorizationUrl != "https://auth.example.com/authorize" {
		t.Errorf("AuthorizationUrl = %q, want %q", flow.AuthorizationUrl, "https://auth.example.com/authorize")
	}
	if flow.TokenUrl != "https://auth.example.com/token" {
		t.Errorf("TokenUrl = %q, want %q", flow.TokenUrl, "https://auth.example.com/token")
	}
	if len(flow.Scopes) != 2 {
		t.Errorf("Scopes count = %d, want %d", len(flow.Scopes), 2)
	}
}

func TestOperationSecurity(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/secure").
		Get().
		Security("bearerAuth").
		Response(200).Description("Success").Done().
		Done().
		Done().
		Path("/public").
		Get().
		NoSecurity().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check secured endpoint
	secureOp := spec.Paths["/secure"].Get
	if len(secureOp.Security) != 1 {
		t.Errorf("Security requirements = %d, want %d", len(secureOp.Security), 1)
	}
	if _, ok := secureOp.Security[0]["bearerAuth"]; !ok {
		t.Error("bearerAuth security requirement not found")
	}

	// Check public endpoint (empty security)
	publicOp := spec.Paths["/public"].Get
	if publicOp.Security == nil {
		t.Error("Security should be empty slice, not nil")
	}
	if len(publicOp.Security) != 0 {
		t.Errorf("Security requirements = %d, want %d", len(publicOp.Security), 0)
	}
}

func TestMustBuildPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustBuild should panic on validation error")
		}
	}()

	NewSpec(Version310).MustBuild() // Missing title and version
}

func TestBuildUnchecked(t *testing.T) {
	spec := NewSpec(Version310).BuildUnchecked()
	if spec == nil {
		t.Error("BuildUnchecked should return spec even without required fields")
	}
}

func TestResponseMissingDescription(t *testing.T) {
	_, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().
		Response(200).JSON(StringSchema()).Done(). // Missing description
		Done().
		Done().
		Build()

	if err == nil {
		t.Fatal("expected error for missing response description")
	}
}

func TestFullPetStoreExample(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Pet Store API").
		Version("1.0.0").
		Description("A sample Pet Store API").
		Contact("API Support", "https://support.example.com", "support@example.com").
		License("Apache 2.0", "https://www.apache.org/licenses/LICENSE-2.0").
		Server("https://api.petstore.io/v1").
		Components().
		Schema("Pet", ObjectSchema().
			Property("id", IntegerSchema().Format("int64")).
			Property("name", StringSchema()).
			Property("tag", StringSchema()).
			Required("id", "name")).
		Schema("Error", ObjectSchema().
			Property("code", IntegerSchema().Format("int32")).
			Property("message", StringSchema()).
			Required("code", "message")).
		SecurityScheme("bearerAuth").BearerAuth().BearerFormat("JWT").Done().
		Done().
		Path("/pets").
		Get().
		Summary("List all pets").
		OperationID("listPets").
		Tags("pets").
		QueryParam("limit").Type("integer").Format("int32").Description("Max items").DoneOp().
		Response(200).Description("A list of pets").JSON(ArraySchema(RefSchema("Pet"))).Done().
		ResponseDefault().Description("Unexpected error").JSON(RefSchema("Error")).Done().
		Done().
		Post().
		Summary("Create a pet").
		OperationID("createPets").
		Tags("pets").
		Security("bearerAuth").
		JSONBody(RefSchema("Pet")).
		Response(201).Description("Pet created").Done().
		ResponseDefault().Description("Unexpected error").JSON(RefSchema("Error")).Done().
		Done().
		Done().
		Path("/pets/{petId}").
		Get().
		Summary("Info for a specific pet").
		OperationID("showPetById").
		Tags("pets").
		PathParam("petId").Type("string").Description("The id of the pet").DoneOp().
		Response(200).Description("Expected response").JSON(RefSchema("Pet")).Done().
		ResponseDefault().Description("Unexpected error").JSON(RefSchema("Error")).Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify basic structure
	if spec.Info.Title != "Pet Store API" {
		t.Error("title mismatch")
	}
	if len(spec.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(spec.Paths))
	}
	if spec.Components == nil || len(spec.Components.Schemas) != 2 {
		t.Error("components schemas mismatch")
	}

	// Verify it can be serialized to JSON
	jsonData, err := ToJSON(spec)
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}
	if len(jsonData) == 0 {
		t.Error("JSON output is empty")
	}

	// Verify JSON structure
	var parsed map[string]any
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("JSON parsing failed: %v", err)
	}
	if parsed["openapi"] != "3.1.0" {
		t.Errorf("openapi = %v, want %q", parsed["openapi"], "3.1.0")
	}
}

func TestStandaloneBuilders(t *testing.T) {
	t.Run("StandaloneParameter", func(t *testing.T) {
		param := PathParam("id").Type("string").Format("uuid").Build()
		if param.Name != "id" {
			t.Errorf("Name = %q, want %q", param.Name, "id")
		}
		if param.In != "path" {
			t.Errorf("In = %q, want %q", param.In, "path")
		}
		if !param.Required {
			t.Error("path parameter should be required")
		}
	})

	t.Run("StandaloneResponse", func(t *testing.T) {
		resp := NewResponse().Description("Success").JSON(RefSchema("Item")).Build()
		if resp.Description != "Success" {
			t.Errorf("Description = %q, want %q", resp.Description, "Success")
		}
		if resp.Content == nil {
			t.Error("Content should not be nil")
		}
	})

	t.Run("StandaloneRequestBody", func(t *testing.T) {
		body := NewRequestBody().Required().JSON(RefSchema("Pet")).Build()
		if !body.Required {
			t.Error("Required should be true")
		}
		if body.Content == nil {
			t.Error("Content should not be nil")
		}
	})

	t.Run("StandaloneServer", func(t *testing.T) {
		server := NewServer("https://api.example.com").Description("Production").Build()
		if server.URL != "https://api.example.com" {
			t.Errorf("URL = %q, want %q", server.URL, "https://api.example.com")
		}
		if server.Description != "Production" {
			t.Errorf("Description = %q, want %q", server.Description, "Production")
		}
	})

	t.Run("StandaloneSecurityScheme", func(t *testing.T) {
		scheme := NewSecurityScheme().BearerAuth().BearerFormat("JWT").Build()
		if scheme.Type != "http" {
			t.Errorf("Type = %q, want %q", scheme.Type, "http")
		}
		if scheme.Scheme != "bearer" {
			t.Errorf("Scheme = %q, want %q", scheme.Scheme, "bearer")
		}
	})
}
