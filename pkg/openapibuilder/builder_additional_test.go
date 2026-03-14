package openapibuilder

import (
	"testing"
)

func TestVersion320(t *testing.T) {
	spec, err := NewSpec(Version320).
		Title("Test API").
		Version("1.0.0").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.OpenAPI != "3.2.0" {
		t.Errorf("OpenAPI = %q, want %q", spec.OpenAPI, "3.2.0")
	}
}

func TestVersionIs32x(t *testing.T) {
	if !Version320.Is32x() {
		t.Error("Version320.Is32x() = false, want true")
	}
	if Version310.Is32x() {
		t.Error("Version310.Is32x() = true, want false")
	}
}

func TestSchemaDescription(t *testing.T) {
	s := StringSchema().Description("A test string").Build()
	if s.Description != "A test string" {
		t.Errorf("Description = %q, want %q", s.Description, "A test string")
	}
}

func TestSchemaTitle(t *testing.T) {
	s := ObjectSchema().Title("User Object").Build()
	if s.Title != "User Object" {
		t.Errorf("Title = %q, want %q", s.Title, "User Object")
	}
}

func TestSchemaDefault(t *testing.T) {
	s := StringSchema().Default("default_value").Build()
	if s.Default != "default_value" {
		t.Errorf("Default = %v, want %q", s.Default, "default_value")
	}
}

func TestSchemaPattern(t *testing.T) {
	s := StringSchema().Pattern("^[a-z]+$").Build()
	if s.Pattern != "^[a-z]+$" {
		t.Errorf("Pattern = %q, want %q", s.Pattern, "^[a-z]+$")
	}
}

func TestSchemaDeprecated(t *testing.T) {
	s := StringSchema().Deprecated().Build()
	if !s.Deprecated {
		t.Error("Deprecated = false, want true")
	}
}

func TestSchemaReadWriteOnly(t *testing.T) {
	s1 := StringSchema().ReadOnly().Build()
	if !s1.ReadOnly {
		t.Error("ReadOnly = false, want true")
	}

	s2 := StringSchema().WriteOnly().Build()
	if !s2.WriteOnly {
		t.Error("WriteOnly = false, want true")
	}
}

func TestNumberSchema(t *testing.T) {
	s := NumberSchema().Format("double").Minimum(0).Maximum(100).Build()
	if s.Type != "number" {
		t.Errorf("Type = %v, want %q", s.Type, "number")
	}
	if s.Format != "double" {
		t.Errorf("Format = %q, want %q", s.Format, "double")
	}
	if *s.Minimum != 0 {
		t.Errorf("Minimum = %f, want %f", *s.Minimum, 0.0)
	}
	if *s.Maximum != 100 {
		t.Errorf("Maximum = %f, want %f", *s.Maximum, 100.0)
	}
}

func TestBooleanSchema(t *testing.T) {
	s := BooleanSchema().Default(true).Build()
	if s.Type != "boolean" {
		t.Errorf("Type = %v, want %q", s.Type, "boolean")
	}
	if s.Default != true {
		t.Errorf("Default = %v, want true", s.Default)
	}
}

func TestExplicitRef(t *testing.T) {
	s := Ref("#/components/responses/NotFound").Build()
	if s.Ref != "#/components/responses/NotFound" {
		t.Errorf("Ref = %q, want %q", s.Ref, "#/components/responses/NotFound")
	}
}

func TestArraySchemaMaxMinItems(t *testing.T) {
	s := ArraySchema(StringSchema()).MinItems(1).MaxItems(10).Build()
	if *s.MinItems != 1 {
		t.Errorf("MinItems = %d, want %d", *s.MinItems, 1)
	}
	if *s.MaxItems != 10 {
		t.Errorf("MaxItems = %d, want %d", *s.MaxItems, 10)
	}
}

func TestObjectSchemaAdditionalProperties(t *testing.T) {
	s := ObjectSchema().
		Property("name", StringSchema()).
		AdditionalProperties(StringSchema()).
		Build()

	if s.AdditionalProperties == nil {
		t.Error("AdditionalProperties should not be nil")
	}
}

func TestObjectSchemaMinMaxProperties(t *testing.T) {
	s := ObjectSchema().MinProperties(1).MaxProperties(10).Build()
	if *s.MinProperties != 1 {
		t.Errorf("MinProperties = %d, want %d", *s.MinProperties, 1)
	}
	if *s.MaxProperties != 10 {
		t.Errorf("MaxProperties = %d, want %d", *s.MaxProperties, 10)
	}
}

func TestOneOfAnyOf(t *testing.T) {
	s := NewSchema().
		OneOf(StringSchema(), IntegerSchema()).
		Build()

	if len(s.OneOf) != 2 {
		t.Errorf("OneOf count = %d, want %d", len(s.OneOf), 2)
	}

	s2 := NewSchema().
		AnyOf(StringSchema(), IntegerSchema()).
		Build()

	if len(s2.AnyOf) != 2 {
		t.Errorf("AnyOf count = %d, want %d", len(s2.AnyOf), 2)
	}
}

func TestNotSchema(t *testing.T) {
	s := NewSchema().Not(StringSchema()).Build()
	if s.Not == nil {
		t.Error("Not should not be nil")
	}
}

func TestHeaderParameter(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().
		HeaderParam("X-Custom-Header").Type("string").Description("Custom header").DoneOp().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params := spec.Paths["/test"].Get.Parameters
	if len(params) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(params))
	}
	if params[0].In != "header" {
		t.Errorf("In = %q, want %q", params[0].In, "header")
	}
}

func TestCookieParameter(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().
		CookieParam("session_id").Type("string").DoneOp().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params := spec.Paths["/test"].Get.Parameters
	if params[0].In != "cookie" {
		t.Errorf("In = %q, want %q", params[0].In, "cookie")
	}
}

func TestParameterRequired(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().
		QueryParam("required_param").Type("string").Required().DoneOp().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params := spec.Paths["/test"].Get.Parameters
	if !params[0].Required {
		t.Error("Required = false, want true")
	}
}

func TestParameterExample(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().
		QueryParam("limit").Type("integer").Example(10).DoneOp().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params := spec.Paths["/test"].Get.Parameters
	if params[0].Example != 10 {
		t.Errorf("Example = %v, want %d", params[0].Example, 10)
	}
}

func TestParameterDeprecated(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().
		QueryParam("old_param").Type("string").Deprecated().DoneOp().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params := spec.Paths["/test"].Get.Parameters
	if !params[0].Deprecated {
		t.Error("Deprecated = false, want true")
	}
}

func TestAllHTTPMethods(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().Summary("GET").Response(200).Description("OK").Done().Done().
		Put().Summary("PUT").Response(200).Description("OK").Done().Done().
		Post().Summary("POST").Response(200).Description("OK").Done().Done().
		Delete().Summary("DELETE").Response(200).Description("OK").Done().Done().
		Patch().Summary("PATCH").Response(200).Description("OK").Done().Done().
		Options().Summary("OPTIONS").Response(200).Description("OK").Done().Done().
		Head().Summary("HEAD").Response(200).Description("OK").Done().Done().
		Trace().Summary("TRACE").Response(200).Description("OK").Done().Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathItem := spec.Paths["/test"]
	if pathItem.Get == nil || pathItem.Get.Summary != "GET" {
		t.Error("GET operation not set correctly")
	}
	if pathItem.Put == nil || pathItem.Put.Summary != "PUT" {
		t.Error("PUT operation not set correctly")
	}
	if pathItem.Post == nil || pathItem.Post.Summary != "POST" {
		t.Error("POST operation not set correctly")
	}
	if pathItem.Delete == nil || pathItem.Delete.Summary != "DELETE" {
		t.Error("DELETE operation not set correctly")
	}
	if pathItem.Patch == nil || pathItem.Patch.Summary != "PATCH" {
		t.Error("PATCH operation not set correctly")
	}
	if pathItem.Options == nil || pathItem.Options.Summary != "OPTIONS" {
		t.Error("OPTIONS operation not set correctly")
	}
	if pathItem.Head == nil || pathItem.Head.Summary != "HEAD" {
		t.Error("HEAD operation not set correctly")
	}
	if pathItem.Trace == nil || pathItem.Trace.Summary != "TRACE" {
		t.Error("TRACE operation not set correctly")
	}
}

func TestOperationDescription(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().
		Summary("Test endpoint").
		Description("A detailed description").
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	op := spec.Paths["/test"].Get
	if op.Description != "A detailed description" {
		t.Errorf("Description = %q, want %q", op.Description, "A detailed description")
	}
}

func TestOperationDeprecated(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/old").
		Get().
		Deprecated().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	op := spec.Paths["/old"].Get
	if !op.Deprecated {
		t.Error("Deprecated = false, want true")
	}
}

func TestMultipleTags(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Get().
		Tags("tag1", "tag2", "tag3").
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	op := spec.Paths["/test"].Get
	if len(op.Tags) != 3 {
		t.Errorf("Tags count = %d, want %d", len(op.Tags), 3)
	}
}

func TestRequestBodyDescription(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Post().
		RequestBody().Description("Request body description").JSON(ObjectSchema()).Done().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	op := spec.Paths["/test"].Post
	if op.RequestBody.Description != "Request body description" {
		t.Errorf("Description = %q, want %q", op.RequestBody.Description, "Request body description")
	}
}

func TestMultipleContentTypes(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Path("/test").
		Post().
		RequestBody().
		Content("application/json", ObjectSchema()).
		Content("application/xml", ObjectSchema()).
		Done().
		Response(200).Description("Success").Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	op := spec.Paths["/test"].Post
	if len(op.RequestBody.Content) != 2 {
		t.Errorf("Content count = %d, want %d", len(op.RequestBody.Content), 2)
	}
	if _, ok := op.RequestBody.Content["application/json"]; !ok {
		t.Error("application/json content not found")
	}
	if _, ok := op.RequestBody.Content["application/xml"]; !ok {
		t.Error("application/xml content not found")
	}
}

func TestBasicAuth(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		SecurityScheme("basicAuth").BasicAuth().Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scheme := spec.Components.SecuritySchemes["basicAuth"]
	if scheme.Type != "http" {
		t.Errorf("Type = %q, want %q", scheme.Type, "http")
	}
	if scheme.Scheme != "basic" {
		t.Errorf("Scheme = %q, want %q", scheme.Scheme, "basic")
	}
}

func TestAPIKeyQuery(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		SecurityScheme("apiKey").APIKeyQuery("api_key").Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scheme := spec.Components.SecuritySchemes["apiKey"]
	if scheme.Type != "apiKey" {
		t.Errorf("Type = %q, want %q", scheme.Type, "apiKey")
	}
	if scheme.In != "query" {
		t.Errorf("In = %q, want %q", scheme.In, "query")
	}
	if scheme.Name != "api_key" {
		t.Errorf("Name = %q, want %q", scheme.Name, "api_key")
	}
}

func TestAPIKeyCookie(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		SecurityScheme("sessionAuth").APIKeyCookie("session_id").Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scheme := spec.Components.SecuritySchemes["sessionAuth"]
	if scheme.In != "cookie" {
		t.Errorf("In = %q, want %q", scheme.In, "cookie")
	}
}

func TestOpenIDConnect(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		SecurityScheme("openid").OpenIDConnect("https://auth.example.com/.well-known/openid-configuration").Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scheme := spec.Components.SecuritySchemes["openid"]
	if scheme.Type != "openIdConnect" {
		t.Errorf("Type = %q, want %q", scheme.Type, "openIdConnect")
	}
	if scheme.OpenIdConnectUrl != "https://auth.example.com/.well-known/openid-configuration" {
		t.Errorf("OpenIdConnectUrl = %q", scheme.OpenIdConnectUrl)
	}
}

func TestOAuth2Flows(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		SecurityScheme("oauth2").
		OAuth2().
		Implicit("https://auth.example.com/authorize").
		Scope("read", "Read access").
		Done().
		Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scheme := spec.Components.SecuritySchemes["oauth2"]
	if scheme.Flows.Implicit == nil {
		t.Fatal("Implicit flow not set")
	}
	if scheme.Flows.Implicit.AuthorizationUrl != "https://auth.example.com/authorize" {
		t.Errorf("AuthorizationUrl = %q", scheme.Flows.Implicit.AuthorizationUrl)
	}
}

func TestOAuth2ClientCredentials(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		SecurityScheme("oauth2").
		OAuth2().
		ClientCredentials("https://auth.example.com/token").
		Scope("admin", "Admin access").
		Done().
		Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scheme := spec.Components.SecuritySchemes["oauth2"]
	if scheme.Flows.ClientCredentials == nil {
		t.Fatal("ClientCredentials flow not set")
	}
	if scheme.Flows.ClientCredentials.TokenUrl != "https://auth.example.com/token" {
		t.Errorf("TokenUrl = %q", scheme.Flows.ClientCredentials.TokenUrl)
	}
}

func TestOAuth2Password(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Components().
		SecurityScheme("oauth2").
		OAuth2().
		Password("https://auth.example.com/token").
		Scope("user", "User access").
		Done().
		Done().
		Done().
		Done().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scheme := spec.Components.SecuritySchemes["oauth2"]
	if scheme.Flows.Password == nil {
		t.Fatal("Password flow not set")
	}
}

func TestRenderYAML(t *testing.T) {
	spec, err := NewSpec(Version310).
		Title("Test API").
		Version("1.0.0").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	yaml, err := ToYAML(spec)
	if err != nil {
		t.Fatalf("ToYAML error: %v", err)
	}
	if len(yaml) == 0 {
		t.Error("YAML output is empty")
	}
}

func TestNullableSchema30(t *testing.T) {
	s := StringSchema().WithVersion(Version303).Nullable().Build()
	if s.Type != "string" {
		t.Errorf("Type = %v, want %q", s.Type, "string")
	}
	if !s.Nullable {
		t.Error("Nullable = false, want true for 3.0")
	}
}
