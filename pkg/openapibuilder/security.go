package openapibuilder

import (
	"maps"

	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// SecuritySchemeBuilder builds an OpenAPI SecurityScheme object.
type SecuritySchemeBuilder struct {
	scheme  *openapi.SecurityScheme
	name    string
	parent  *ComponentsBuilder
	version Version
}

// newSecuritySchemeBuilder creates a new security scheme builder.
func newSecuritySchemeBuilder(name string, parent *ComponentsBuilder) *SecuritySchemeBuilder {
	return &SecuritySchemeBuilder{
		scheme:  &openapi.SecurityScheme{},
		name:    name,
		parent:  parent,
		version: Version310,
	}
}

// BearerAuth configures HTTP Bearer authentication.
func (b *SecuritySchemeBuilder) BearerAuth() *SecuritySchemeBuilder {
	b.scheme.Type = "http"
	b.scheme.Scheme = "bearer"
	return b
}

// BasicAuth configures HTTP Basic authentication.
func (b *SecuritySchemeBuilder) BasicAuth() *SecuritySchemeBuilder {
	b.scheme.Type = "http"
	b.scheme.Scheme = "basic"
	return b
}

// APIKeyHeader configures API key authentication via header.
func (b *SecuritySchemeBuilder) APIKeyHeader(headerName string) *SecuritySchemeBuilder {
	b.scheme.Type = "apiKey"
	b.scheme.In = "header"
	b.scheme.Name = headerName
	return b
}

// APIKeyQuery configures API key authentication via query parameter.
func (b *SecuritySchemeBuilder) APIKeyQuery(paramName string) *SecuritySchemeBuilder {
	b.scheme.Type = "apiKey"
	b.scheme.In = "query"
	b.scheme.Name = paramName
	return b
}

// APIKeyCookie configures API key authentication via cookie.
func (b *SecuritySchemeBuilder) APIKeyCookie(cookieName string) *SecuritySchemeBuilder {
	b.scheme.Type = "apiKey"
	b.scheme.In = "cookie"
	b.scheme.Name = cookieName
	return b
}

// OpenIDConnect configures OpenID Connect authentication.
func (b *SecuritySchemeBuilder) OpenIDConnect(url string) *SecuritySchemeBuilder {
	b.scheme.Type = "openIdConnect"
	b.scheme.OpenIdConnectUrl = url
	return b
}

// OAuth2 starts configuring OAuth2 authentication.
func (b *SecuritySchemeBuilder) OAuth2() *OAuthFlowsBuilder {
	b.scheme.Type = "oauth2"
	b.scheme.Flows = &openapi.OAuthFlows{}
	return &OAuthFlowsBuilder{
		flows:  b.scheme.Flows,
		parent: b,
	}
}

// BearerFormat sets the bearer format (e.g., "JWT").
func (b *SecuritySchemeBuilder) BearerFormat(format string) *SecuritySchemeBuilder {
	b.scheme.BearerFormat = format
	return b
}

// Description sets the security scheme description.
func (b *SecuritySchemeBuilder) Description(d string) *SecuritySchemeBuilder {
	b.scheme.Description = d
	return b
}

// Done returns to the parent ComponentsBuilder.
func (b *SecuritySchemeBuilder) Done() *ComponentsBuilder {
	if b.parent != nil {
		b.parent.addSecurityScheme(b.name, b.scheme)
	}
	return b.parent
}

// Build returns the constructed security scheme.
func (b *SecuritySchemeBuilder) Build() *openapi.SecurityScheme {
	return b.scheme
}

// OAuthFlowsBuilder builds OAuth2 flows.
type OAuthFlowsBuilder struct {
	flows  *openapi.OAuthFlows
	parent *SecuritySchemeBuilder
}

// Implicit configures the implicit flow.
func (b *OAuthFlowsBuilder) Implicit(authURL string) *OAuthFlowBuilder {
	b.flows.Implicit = &openapi.OAuthFlow{
		AuthorizationUrl: authURL,
		Scopes:           make(map[string]string),
	}
	return &OAuthFlowBuilder{
		flow:   b.flows.Implicit,
		parent: b,
	}
}

// Password configures the password flow.
func (b *OAuthFlowsBuilder) Password(tokenURL string) *OAuthFlowBuilder {
	b.flows.Password = &openapi.OAuthFlow{
		TokenUrl: tokenURL,
		Scopes:   make(map[string]string),
	}
	return &OAuthFlowBuilder{
		flow:   b.flows.Password,
		parent: b,
	}
}

// ClientCredentials configures the client credentials flow.
func (b *OAuthFlowsBuilder) ClientCredentials(tokenURL string) *OAuthFlowBuilder {
	b.flows.ClientCredentials = &openapi.OAuthFlow{
		TokenUrl: tokenURL,
		Scopes:   make(map[string]string),
	}
	return &OAuthFlowBuilder{
		flow:   b.flows.ClientCredentials,
		parent: b,
	}
}

// AuthorizationCode configures the authorization code flow.
func (b *OAuthFlowsBuilder) AuthorizationCode(authURL, tokenURL string) *OAuthFlowBuilder {
	b.flows.AuthorizationCode = &openapi.OAuthFlow{
		AuthorizationUrl: authURL,
		TokenUrl:         tokenURL,
		Scopes:           make(map[string]string),
	}
	return &OAuthFlowBuilder{
		flow:   b.flows.AuthorizationCode,
		parent: b,
	}
}

// Done returns to the parent SecuritySchemeBuilder.
func (b *OAuthFlowsBuilder) Done() *SecuritySchemeBuilder {
	return b.parent
}

// OAuthFlowBuilder builds a single OAuth2 flow.
type OAuthFlowBuilder struct {
	flow   *openapi.OAuthFlow
	parent *OAuthFlowsBuilder
}

// RefreshURL sets the refresh URL.
func (b *OAuthFlowBuilder) RefreshURL(url string) *OAuthFlowBuilder {
	b.flow.RefreshUrl = url
	return b
}

// Scope adds a scope to the flow.
func (b *OAuthFlowBuilder) Scope(name, description string) *OAuthFlowBuilder {
	b.flow.Scopes[name] = description
	return b
}

// Scopes adds multiple scopes to the flow.
func (b *OAuthFlowBuilder) Scopes(scopes map[string]string) *OAuthFlowBuilder {
	maps.Copy(b.flow.Scopes, scopes)
	return b
}

// Done returns to the parent OAuthFlowsBuilder.
func (b *OAuthFlowBuilder) Done() *OAuthFlowsBuilder {
	return b.parent
}

// StandaloneSecuritySchemeBuilder builds security schemes without a parent.
type StandaloneSecuritySchemeBuilder struct {
	scheme  *openapi.SecurityScheme
	version Version
}

// NewSecurityScheme creates a standalone security scheme builder.
func NewSecurityScheme() *StandaloneSecuritySchemeBuilder {
	return &StandaloneSecuritySchemeBuilder{
		scheme:  &openapi.SecurityScheme{},
		version: Version310,
	}
}

// BearerAuth configures HTTP Bearer authentication.
func (b *StandaloneSecuritySchemeBuilder) BearerAuth() *StandaloneSecuritySchemeBuilder {
	b.scheme.Type = "http"
	b.scheme.Scheme = "bearer"
	return b
}

// BasicAuth configures HTTP Basic authentication.
func (b *StandaloneSecuritySchemeBuilder) BasicAuth() *StandaloneSecuritySchemeBuilder {
	b.scheme.Type = "http"
	b.scheme.Scheme = "basic"
	return b
}

// APIKeyHeader configures API key authentication via header.
func (b *StandaloneSecuritySchemeBuilder) APIKeyHeader(headerName string) *StandaloneSecuritySchemeBuilder {
	b.scheme.Type = "apiKey"
	b.scheme.In = "header"
	b.scheme.Name = headerName
	return b
}

// BearerFormat sets the bearer format (e.g., "JWT").
func (b *StandaloneSecuritySchemeBuilder) BearerFormat(format string) *StandaloneSecuritySchemeBuilder {
	b.scheme.BearerFormat = format
	return b
}

// Description sets the security scheme description.
func (b *StandaloneSecuritySchemeBuilder) Description(d string) *StandaloneSecuritySchemeBuilder {
	b.scheme.Description = d
	return b
}

// Build returns the constructed security scheme.
func (b *StandaloneSecuritySchemeBuilder) Build() *openapi.SecurityScheme {
	return b.scheme
}
