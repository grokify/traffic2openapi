package openapi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/inference"
)

// Version represents the OpenAPI version to generate.
type Version string

const (
	Version30 Version = "3.0.3"
	Version31 Version = "3.1.0"
)

// GeneratorOptions configures the OpenAPI generator.
type GeneratorOptions struct {
	Version     Version
	Title       string
	Description string
	APIVersion  string
	Servers     []string
}

// DefaultGeneratorOptions returns default options.
func DefaultGeneratorOptions() GeneratorOptions {
	return GeneratorOptions{
		Version:    Version31,
		Title:      "Generated API",
		APIVersion: "1.0.0",
	}
}

// Generator converts inference results to OpenAPI specs.
type Generator struct {
	options GeneratorOptions
}

// NewGenerator creates a new OpenAPI generator.
func NewGenerator(options GeneratorOptions) *Generator {
	return &Generator{options: options}
}

// Generate creates an OpenAPI spec from inference results.
func (g *Generator) Generate(result *inference.InferenceResult) *Spec {
	spec := &Spec{
		OpenAPI: string(g.options.Version),
		Info: Info{
			Title:       g.options.Title,
			Description: g.options.Description,
			Version:     g.options.APIVersion,
		},
		Paths: make(map[string]*PathItem),
	}

	// Add servers
	if len(g.options.Servers) > 0 {
		for _, serverURL := range g.options.Servers {
			spec.Servers = append(spec.Servers, Server{URL: serverURL})
		}
	} else if len(result.Hosts) > 0 && len(result.Schemes) > 0 {
		// Generate servers from observed hosts/schemes
		for _, host := range result.Hosts {
			for _, scheme := range result.Schemes {
				spec.Servers = append(spec.Servers, Server{
					URL: fmt.Sprintf("%s://%s", scheme, host),
				})
			}
		}
	}

	// Add detected security schemes
	if len(result.SecuritySchemes) > 0 {
		if spec.Components == nil {
			spec.Components = &Components{}
		}
		spec.Components.SecuritySchemes = make(map[string]*SecurityScheme)

		for key, detected := range result.SecuritySchemes {
			scheme := &SecurityScheme{
				Type: detected.Type,
			}

			switch detected.Type {
			case "http":
				scheme.Scheme = detected.Scheme
				if detected.BearerFormat != "" {
					scheme.BearerFormat = detected.BearerFormat
				}
			case "apiKey":
				scheme.Name = detected.Name
				scheme.In = detected.In
			}

			spec.Components.SecuritySchemes[key] = scheme
		}
	}

	// Store security scheme keys for operation-level security
	securityKeys := make([]string, 0, len(result.SecuritySchemes))
	for key := range result.SecuritySchemes {
		securityKeys = append(securityKeys, key)
	}
	sort.Strings(securityKeys)

	// Generate paths
	for _, endpoint := range result.Endpoints {
		g.addEndpoint(spec, endpoint, securityKeys)
	}

	return spec
}

// addEndpoint adds an endpoint to the spec.
func (g *Generator) addEndpoint(spec *Spec, endpoint *inference.EndpointData, securityKeys []string) {
	path := endpoint.PathTemplate

	// Get or create path item
	pathItem, exists := spec.Paths[path]
	if !exists {
		pathItem = &PathItem{}
		spec.Paths[path] = pathItem
	}

	// Create operation
	operation := g.createOperation(endpoint, securityKeys)

	// Assign to correct method
	switch strings.ToUpper(endpoint.Method) {
	case "GET":
		pathItem.Get = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "DELETE":
		pathItem.Delete = operation
	case "PATCH":
		pathItem.Patch = operation
	case "HEAD":
		pathItem.Head = operation
	case "OPTIONS":
		pathItem.Options = operation
	case "TRACE":
		pathItem.Trace = operation
	}
}

// createOperation creates an Operation from endpoint data.
func (g *Generator) createOperation(endpoint *inference.EndpointData, securityKeys []string) *Operation {
	op := &Operation{
		Summary:     fmt.Sprintf("%s %s", endpoint.Method, endpoint.PathTemplate),
		OperationID: generateOperationID(endpoint.Method, endpoint.PathTemplate),
		Parameters:  make([]Parameter, 0),
		Responses:   make(map[string]Response),
	}

	// Add security requirements if any were detected
	if len(securityKeys) > 0 {
		op.Security = make([]SecurityRequirement, 0, len(securityKeys))
		for _, key := range securityKeys {
			op.Security = append(op.Security, SecurityRequirement{key: []string{}})
		}
	}

	// Add path parameters
	for _, param := range endpoint.PathParams {
		op.Parameters = append(op.Parameters, g.createParameter(param, "path", true))
	}

	// Add query parameters
	for _, param := range endpoint.QueryParams {
		op.Parameters = append(op.Parameters, g.createParameter(param, "query", param.Required))
	}

	// Add header parameters
	for _, param := range endpoint.HeaderParams {
		op.Parameters = append(op.Parameters, g.createParameter(param, "header", param.Required))
	}

	// Sort parameters for consistent output
	sort.Slice(op.Parameters, func(i, j int) bool {
		if op.Parameters[i].In != op.Parameters[j].In {
			return paramInOrder(op.Parameters[i].In) < paramInOrder(op.Parameters[j].In)
		}
		return op.Parameters[i].Name < op.Parameters[j].Name
	})

	// Add request body
	if endpoint.RequestBody != nil && len(endpoint.RequestBody.Schema.Examples) > 0 {
		op.RequestBody = g.createRequestBody(endpoint.RequestBody)
	}

	// Add responses
	for statusCode, respData := range endpoint.Responses {
		op.Responses[fmt.Sprintf("%d", statusCode)] = g.createResponse(respData)
	}

	// Ensure at least one response
	if len(op.Responses) == 0 {
		op.Responses["200"] = Response{Description: "Successful response"}
	}

	return op
}

// createParameter creates a Parameter from param data.
func (g *Generator) createParameter(param *inference.ParamData, in string, required bool) Parameter {
	p := Parameter{
		Name:     param.Name,
		In:       in,
		Required: required,
		Schema:   &Schema{Type: param.Type},
	}

	if param.Format != "" {
		p.Schema.Format = param.Format
	}

	// Add example
	if len(param.Examples) > 0 {
		p.Example = param.Examples[0]
	}

	return p
}

// createRequestBody creates a RequestBody from body data.
func (g *Generator) createRequestBody(body *inference.BodyData) *RequestBody {
	contentType := body.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	schema := g.convertSchemaNode(inference.BuildSchemaTree(body.Schema))

	return &RequestBody{
		Required: true,
		Content: map[string]MediaType{
			contentType: {Schema: schema},
		},
	}
}

// createResponse creates a Response from response data.
func (g *Generator) createResponse(respData *inference.ResponseData) Response {
	resp := Response{
		Description: fmt.Sprintf("Status %d response", respData.StatusCode),
	}

	// Add headers
	if len(respData.Headers) > 0 {
		resp.Headers = make(map[string]Header)
		for name, param := range respData.Headers {
			resp.Headers[name] = Header{
				Schema: &Schema{Type: param.Type},
			}
		}
	}

	// Add content
	if len(respData.Body.Examples) > 0 || len(respData.Body.Types) > 0 {
		contentType := respData.ContentType
		if contentType == "" {
			contentType = "application/json"
		}

		schema := g.convertSchemaNode(inference.BuildSchemaTree(respData.Body))

		resp.Content = map[string]MediaType{
			contentType: {Schema: schema},
		}
	}

	return resp
}

// convertSchemaNode converts an inference SchemaNode to an OpenAPI Schema.
func (g *Generator) convertSchemaNode(node *inference.SchemaNode) *Schema {
	if node == nil {
		return &Schema{Type: "object"}
	}

	schema := &Schema{}

	// Set type (handle nullable for OpenAPI 3.1)
	if node.Nullable && g.options.Version == Version31 {
		schema.Type = []string{node.Type, "null"}
	} else {
		schema.Type = node.Type
	}

	// Set format
	if node.Format != "" {
		schema.Format = node.Format
	}

	// Set enum
	if len(node.Enum) > 0 {
		schema.Enum = make([]any, len(node.Enum))
		for i, v := range node.Enum {
			schema.Enum[i] = v
		}
	}

	// Set examples (OpenAPI 3.1) or example (OpenAPI 3.0)
	if len(node.Examples) > 0 {
		if g.options.Version == Version31 {
			schema.Examples = node.Examples
		} else {
			// OpenAPI 3.0 uses singular example at the schema level
			// We don't add it here as it's not standard
		}
	}

	// Set array items
	if node.Type == "array" && node.Items != nil {
		schema.Items = g.convertSchemaNode(node.Items)
	}

	// Set object properties
	if node.Type == "object" && len(node.Properties) > 0 {
		schema.Properties = make(map[string]*Schema)
		for name, prop := range node.Properties {
			schema.Properties[name] = g.convertSchemaNode(prop)
		}

		if len(node.Required) > 0 {
			schema.Required = node.Required
		}
	}

	return schema
}

// generateOperationID creates an operation ID from method and path.
func generateOperationID(method, path string) string {
	// Convert path to camelCase
	// e.g., GET /users/{userId}/posts -> getUsersByUserIdPosts
	method = strings.ToLower(method)

	// Remove leading slash and split
	path = strings.TrimPrefix(path, "/")
	segments := strings.Split(path, "/")

	var parts []string
	parts = append(parts, method)

	for _, seg := range segments {
		if seg == "" {
			continue
		}

		// Handle path parameters
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			paramName := seg[1 : len(seg)-1]
			parts = append(parts, "By"+capitalize(paramName))
		} else {
			parts = append(parts, capitalize(seg))
		}
	}

	return strings.Join(parts, "")
}

// capitalize capitalizes the first letter.
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// paramInOrder returns the sort order for parameter locations.
func paramInOrder(in string) int {
	switch in {
	case "path":
		return 0
	case "query":
		return 1
	case "header":
		return 2
	case "cookie":
		return 3
	default:
		return 4
	}
}

// GenerateFromInference is a convenience function.
func GenerateFromInference(result *inference.InferenceResult, options GeneratorOptions) *Spec {
	return NewGenerator(options).Generate(result)
}
