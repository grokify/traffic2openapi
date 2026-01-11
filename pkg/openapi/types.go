// Package openapi provides OpenAPI 3.1 specification generation.
package openapi

// Spec represents an OpenAPI 3.x specification.
type Spec struct {
	OpenAPI    string               `json:"openapi" yaml:"openapi"`
	Info       Info                 `json:"info" yaml:"info"`
	Servers    []Server             `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths      map[string]*PathItem `json:"paths" yaml:"paths"`
	Components *Components          `json:"components,omitempty" yaml:"components,omitempty"`
}

// Info provides metadata about the API.
type Info struct {
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string   `json:"version" yaml:"version"`
	Contact     *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License     *License `json:"license,omitempty" yaml:"license,omitempty"`
}

// Contact information for the API.
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License information for the API.
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server represents an API server.
type Server struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// PathItem describes operations available on a single path.
type PathItem struct {
	Summary     string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation  `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation  `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation  `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation  `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *Operation  `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *Operation  `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *Operation  `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *Operation  `json:"trace,omitempty" yaml:"trace,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	Tags        []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses" yaml:"responses"`
	Deprecated  bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
}

// Parameter describes a single operation parameter.
type Parameter struct {
	Name            string  `json:"name" yaml:"name"`
	In              string  `json:"in" yaml:"in"` // query, header, path, cookie
	Description     string  `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool    `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool    `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool    `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Schema          *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         any     `json:"example,omitempty" yaml:"example,omitempty"`
}

// RequestBody describes a single request body.
type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
}

// Response describes a single response from an API operation.
type Response struct {
	Description string               `json:"description" yaml:"description"`
	Headers     map[string]Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// Header describes a single header.
type Header struct {
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool    `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// MediaType provides schema and examples for the media type.
type MediaType struct {
	Schema   *Schema            `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  any                `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]Example `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// Example describes an example value.
type Example struct {
	Summary       string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string `json:"description,omitempty" yaml:"description,omitempty"`
	Value         any    `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Schema represents a JSON Schema (OpenAPI 3.1 uses JSON Schema 2020-12).
type Schema struct {
	// Core
	Type   any    `json:"type,omitempty" yaml:"type,omitempty"` // string or []string for nullable
	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	// Metadata
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Default     any    `json:"default,omitempty" yaml:"default,omitempty"`

	// Validation
	Enum  []any `json:"enum,omitempty" yaml:"enum,omitempty"`
	Const any   `json:"const,omitempty" yaml:"const,omitempty"`

	// Numeric
	MultipleOf       *float64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`

	// String
	MaxLength *int   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength *int   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern   string `json:"pattern,omitempty" yaml:"pattern,omitempty"`

	// Array
	Items       *Schema `json:"items,omitempty" yaml:"items,omitempty"`
	MaxItems    *int    `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems    *int    `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems bool    `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`

	// Object
	Properties           map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required             []string           `json:"required,omitempty" yaml:"required,omitempty"`
	AdditionalProperties any                `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	MaxProperties        *int               `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        *int               `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`

	// Composition
	AllOf []*Schema `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf []*Schema `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf []*Schema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not   *Schema   `json:"not,omitempty" yaml:"not,omitempty"`

	// Reference
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	// Examples (OpenAPI 3.1)
	Examples []any `json:"examples,omitempty" yaml:"examples,omitempty"`

	// Deprecated
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`

	// Read/Write only
	ReadOnly  bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly bool `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
}

// Components holds reusable objects.
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]*Example        `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]*Header         `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

// SecurityScheme defines a security scheme.
type SecurityScheme struct {
	Type             string      `json:"type" yaml:"type"` // apiKey, http, oauth2, openIdConnect
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`                         // for apiKey
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`                             // for apiKey: query, header, cookie
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`                     // for http
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`         // for http bearer
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`                       // for oauth2
	OpenIdConnectUrl string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"` // for openIdConnect
}

// OAuthFlows defines OAuth 2.0 flows.
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow defines a single OAuth 2.0 flow.
type OAuthFlow struct {
	AuthorizationUrl string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenUrl         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshUrl       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}

// SecurityRequirement defines security requirements for an operation.
type SecurityRequirement map[string][]string
