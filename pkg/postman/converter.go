// Package postman provides an adapter for converting Postman Collection v2.1 to IR format.
//
// This converter provides lossless conversion, preserving:
//   - Request/response details (method, URL, headers, body)
//   - Folder structure as tags
//   - Collection metadata (name, description, version)
//   - Saved example responses
//   - Request descriptions
//   - Variable resolution
//   - Authentication configuration
package postman

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	postman "github.com/rbretecher/go-postman-collection"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

// Converter converts Postman collections to IR records.
type Converter struct {
	// Variables is a map of variable names to values for resolution.
	// Collection-level variables are added automatically.
	Variables map[string]string

	// BaseURL is the value to use for {{url}} or {{baseUrl}} variables.
	BaseURL string

	// IncludeHeaders controls whether to include HTTP headers in output.
	IncludeHeaders bool

	// IncludeDisabled includes disabled headers and query params.
	IncludeDisabled bool

	// FilterHeaders is a list of header names to exclude (case-insensitive).
	// By default, sensitive auth headers are not filtered since Postman collections
	// typically contain example/placeholder values.
	FilterHeaders []string

	// GenerateIDs controls whether to generate UUIDs for records without IDs.
	GenerateIDs bool

	// PreserveAuth converts Postman auth configurations to request headers.
	PreserveAuth bool
}

// NewConverter creates a new Postman to IR converter with default settings.
func NewConverter() *Converter {
	return &Converter{
		Variables:       make(map[string]string),
		IncludeHeaders:  true,
		IncludeDisabled: false,
		FilterHeaders:   []string{},
		GenerateIDs:     true,
		PreserveAuth:    true,
	}
}

// ConvertResult contains the conversion output.
type ConvertResult struct {
	// Records is the list of converted IR records.
	Records []ir.IRRecord

	// Metadata contains API-level metadata from the collection.
	Metadata *ir.APIMetadata

	// TagDefinitions contains tag definitions from folder structure.
	TagDefinitions []ir.TagDefinition

	// Warnings contains non-fatal issues encountered during conversion.
	Warnings []string
}

// Convert converts a Postman collection to IR records with full metadata.
func (c *Converter) Convert(collection *postman.Collection) (*ConvertResult, error) {
	if collection == nil {
		return nil, fmt.Errorf("collection is nil")
	}

	result := &ConvertResult{
		Records:        make([]ir.IRRecord, 0),
		TagDefinitions: make([]ir.TagDefinition, 0),
		Warnings:       make([]string, 0),
	}

	// Build variables map from collection variables
	variables := c.buildVariables(collection)

	// Extract API metadata from collection info
	result.Metadata = c.extractMetadata(collection, variables)

	// Track tags from folder structure
	tagSet := make(map[string]*ir.TagDefinition)

	// Walk items recursively
	ctx := &walkContext{
		converter: c,
		variables: variables,
		tags:      []string{},
		tagSet:    tagSet,
		counter:   0,
		auth:      collection.Auth,
		warnings:  &result.Warnings,
	}

	c.walkItems(collection.Items, ctx, &result.Records)

	// Convert tag set to slice
	for _, tag := range tagSet {
		result.TagDefinitions = append(result.TagDefinitions, *tag)
	}

	// Update metadata with tag definitions
	if result.Metadata != nil {
		result.Metadata.TagDefinitions = result.TagDefinitions
	}

	return result, nil
}

// ConvertToRecords is a convenience method that returns only the IR records.
func (c *Converter) ConvertToRecords(collection *postman.Collection) ([]ir.IRRecord, error) {
	result, err := c.Convert(collection)
	if err != nil {
		return nil, err
	}
	return result.Records, nil
}

// ConvertToBatch converts a collection to a complete IR batch.
func (c *Converter) ConvertToBatch(collection *postman.Collection) (*ir.Batch, error) {
	result, err := c.Convert(collection)
	if err != nil {
		return nil, err
	}

	batch := ir.NewBatchWithMetadata(result.Records, result.Metadata)
	return batch, nil
}

// walkContext holds state during recursive item traversal.
type walkContext struct {
	converter *Converter
	variables map[string]string
	tags      []string
	tagSet    map[string]*ir.TagDefinition
	counter   int
	auth      *postman.Auth
	warnings  *[]string
}

// walkItems recursively processes items and item groups.
func (c *Converter) walkItems(items []*postman.Items, ctx *walkContext, records *[]ir.IRRecord) {
	for _, item := range items {
		if item == nil {
			continue
		}

		if item.IsGroup() {
			// Item group (folder)
			c.walkItemGroup(item, ctx, records)
		} else if item.Request != nil {
			// Single request item
			c.convertItem(item, ctx, records)
		}
	}
}

// walkItemGroup processes a folder and its children.
func (c *Converter) walkItemGroup(group *postman.Items, ctx *walkContext, records *[]ir.IRRecord) {
	// Add tag for this folder
	tagName := group.Name
	if tagName != "" {
		// Track tag definition
		if _, exists := ctx.tagSet[tagName]; !exists {
			tagDef := &ir.TagDefinition{Name: tagName}
			if group.Description != "" {
				tagDef.Description = ptrString(group.Description)
			}
			ctx.tagSet[tagName] = tagDef
		}
	}

	// Create new context with this tag added
	newCtx := &walkContext{
		converter: ctx.converter,
		variables: ctx.variables,
		tags:      append(append([]string{}, ctx.tags...), tagName),
		tagSet:    ctx.tagSet,
		counter:   ctx.counter,
		auth:      ctx.auth,
		warnings:  ctx.warnings,
	}

	// Override auth if folder has its own
	if group.Auth != nil {
		newCtx.auth = group.Auth
	}

	// Merge folder-level variables
	if len(group.Variables) > 0 {
		newVars := make(map[string]string)
		for k, v := range ctx.variables {
			newVars[k] = v
		}
		for _, v := range group.Variables {
			if v != nil && v.Key != "" && !v.Disabled {
				newVars[v.Key] = v.Value
			}
		}
		newCtx.variables = newVars
	}

	// Process child items
	c.walkItems(group.Items, newCtx, records)

	// Update counter
	ctx.counter = newCtx.counter
}

// convertItem converts a single Postman item (with responses) to IR records.
func (c *Converter) convertItem(item *postman.Items, ctx *walkContext, records *[]ir.IRRecord) {
	if item.Request == nil {
		return
	}

	// Get effective auth (item > folder > collection)
	auth := ctx.auth

	// Parse base request details
	baseReq := c.convertRequest(item.Request, ctx.variables, auth)

	// Generate operation ID from item name
	operationId := sanitizeOperationId(item.Name)

	// Get tags (filter empty)
	tags := make([]string, 0, len(ctx.tags))
	for _, t := range ctx.tags {
		if t != "" {
			tags = append(tags, t)
		}
	}

	// If no saved responses, create a single record with a stub response
	if len(item.Responses) == 0 {
		ctx.counter++
		record := ir.IRRecord{
			Id:          c.generateID(ctx.counter, 0),
			Source:      ptrSource(ir.IRRecordSourcePostman),
			Request:     baseReq,
			Response:    ir.Response{Status: 200},
			OperationId: ptrString(operationId),
			Summary:     ptrString(item.Name),
		}

		if item.Description != "" {
			record.Description = ptrString(item.Description)
		}

		if len(tags) > 0 {
			record.Tags = tags
		}

		*records = append(*records, record)
		return
	}

	// Convert each saved response as a separate record
	for i, resp := range item.Responses {
		if resp == nil {
			continue
		}

		ctx.counter++
		recordId := c.generateID(ctx.counter, i)

		// Use original request from response if available, otherwise use item request
		var reqDetails ir.Request
		if resp.OriginalRequest != nil {
			reqDetails = c.convertRequest(resp.OriginalRequest, ctx.variables, auth)
		} else {
			reqDetails = baseReq
		}

		// Convert response
		respDetails := c.convertResponse(resp, ctx.variables)

		record := ir.IRRecord{
			Id:          recordId,
			Source:      ptrSource(ir.IRRecordSourcePostman),
			Request:     reqDetails,
			Response:    respDetails,
			OperationId: ptrString(operationId),
			Summary:     ptrString(item.Name),
		}

		// Add response name to description if different from item name
		if resp.Name != "" && resp.Name != item.Name {
			desc := item.Description
			if desc != "" {
				desc = desc + "\n\nExample: " + resp.Name
			} else {
				desc = "Example: " + resp.Name
			}
			record.Description = ptrString(desc)
		} else if item.Description != "" {
			record.Description = ptrString(item.Description)
		}

		if len(tags) > 0 {
			record.Tags = tags
		}

		// Extract duration if available
		if resp.ResponseTime != nil {
			switch v := resp.ResponseTime.(type) {
			case float64:
				record.DurationMs = ptrFloat64(v)
			case int:
				record.DurationMs = ptrFloat64(float64(v))
			}
		}

		*records = append(*records, record)
	}
}

// convertRequest converts a Postman request to IR request.
func (c *Converter) convertRequest(req *postman.Request, variables map[string]string, auth *postman.Auth) ir.Request {
	irReq := ir.Request{
		Method: ir.RequestMethod(req.Method),
		Path:   "/",
		Scheme: ir.RequestSchemeHTTPS,
	}

	// Parse URL
	if req.URL != nil {
		c.parseURL(req.URL, variables, &irReq)
	}

	// Convert headers
	if c.IncludeHeaders && len(req.Header) > 0 {
		headers := c.convertHeaders(req.Header, variables)
		if len(headers) > 0 {
			irReq.Headers = headers
		}
	}

	// Add auth headers
	if c.PreserveAuth && auth != nil {
		authHeaders := c.authToHeaders(auth, variables)
		if len(authHeaders) > 0 {
			if irReq.Headers == nil {
				irReq.Headers = make(map[string]string)
			}
			for k, v := range authHeaders {
				irReq.Headers[k] = v
			}
		}
	}

	// Convert body
	if req.Body != nil {
		body, contentType := c.convertBody(req.Body, variables)
		if body != nil {
			irReq.Body = body
		}
		if contentType != "" {
			irReq.ContentType = ptrString(contentType)
			// Also set in headers
			if irReq.Headers == nil {
				irReq.Headers = make(map[string]string)
			}
			if _, exists := irReq.Headers["content-type"]; !exists {
				irReq.Headers["content-type"] = contentType
			}
		}
	}

	return irReq
}

// parseURL extracts URL components into the IR request.
func (c *Converter) parseURL(url *postman.URL, variables map[string]string, irReq *ir.Request) {
	// Protocol/scheme
	protocol := resolveVars(url.Protocol, variables)
	if protocol == "" {
		protocol = "https"
	}
	if strings.ToLower(protocol) == "http" {
		irReq.Scheme = ir.RequestSchemeHTTP
	} else {
		irReq.Scheme = ir.RequestSchemeHTTPS
	}

	// Host
	if len(url.Host) > 0 {
		host := resolveVars(strings.Join(url.Host, "."), variables)
		if host != "" {
			// Include port if present
			if url.Port != "" {
				port := resolveVars(url.Port, variables)
				if port != "" && port != "443" && port != "80" {
					host = host + ":" + port
				}
			}
			irReq.Host = ptrString(host)
		}
	}

	// Path
	if len(url.Path) > 0 {
		pathParts := make([]string, 0, len(url.Path))
		for _, p := range url.Path {
			resolved := resolveVars(p, variables)
			if resolved != "" {
				pathParts = append(pathParts, resolved)
			}
		}
		path := "/" + strings.Join(pathParts, "/")
		// Clean double slashes
		path = regexp.MustCompile(`//+`).ReplaceAllString(path, "/")
		irReq.Path = path
	}

	// Query parameters
	if len(url.Query) > 0 {
		query := make(map[string]interface{})
		for _, q := range url.Query {
			if q == nil {
				continue
			}
			if !c.IncludeDisabled && q.Key == "" {
				continue
			}
			key := resolveVars(q.Key, variables)
			value := resolveVars(q.Value, variables)
			if key != "" {
				query[key] = value
			}
		}
		if len(query) > 0 {
			irReq.Query = query
		}
	}

	// URL variables (path parameters) - these map to pathParams
	if len(url.Variables) > 0 {
		pathParams := make(map[string]string)
		for _, v := range url.Variables {
			if v == nil || v.Disabled {
				continue
			}
			key := v.Key
			if key == "" {
				key = v.Name
			}
			if key != "" {
				value := resolveVars(v.Value, variables)
				pathParams[key] = value
			}
		}
		if len(pathParams) > 0 {
			irReq.PathParams = pathParams
		}
	}
}

// convertHeaders converts Postman headers to IR headers map.
func (c *Converter) convertHeaders(headers []*postman.Header, variables map[string]string) map[string]string {
	result := make(map[string]string)

	for _, h := range headers {
		if h == nil {
			continue
		}
		if !c.IncludeDisabled && h.Disabled {
			continue
		}

		key := strings.ToLower(resolveVars(h.Key, variables))
		if key == "" {
			continue
		}

		// Check filter
		if c.shouldFilterHeader(key) {
			continue
		}

		value := resolveVars(h.Value, variables)
		result[key] = value
	}

	return result
}

// authToHeaders converts Postman auth to HTTP headers.
func (c *Converter) authToHeaders(auth *postman.Auth, variables map[string]string) map[string]string {
	if auth == nil {
		return nil
	}

	headers := make(map[string]string)
	params := auth.GetParams()

	// Build params map for easier lookup
	paramMap := make(map[string]string)
	for _, p := range params {
		if p != nil && p.Key != "" {
			if v, ok := p.Value.(string); ok {
				paramMap[p.Key] = resolveVars(v, variables)
			}
		}
	}

	// Convert based on auth type - use string comparison since authType is private
	authJSON, _ := json.Marshal(auth)
	authStr := string(authJSON)

	switch {
	case strings.Contains(authStr, `"bearer"`):
		if token := paramMap["token"]; token != "" {
			headers["authorization"] = "Bearer " + token
		}
	case strings.Contains(authStr, `"basic"`):
		// Basic auth - username:password encoded in Base64
		// We store as-is since actual encoding happens at request time
		if username := paramMap["username"]; username != "" {
			if password := paramMap["password"]; password != "" {
				headers["authorization"] = "Basic " + username + ":" + password
			}
		}
	case strings.Contains(authStr, `"apikey"`):
		key := paramMap["key"]
		value := paramMap["value"]
		in := paramMap["in"]
		if key != "" && value != "" {
			if in == "header" || in == "" {
				headers[strings.ToLower(key)] = value
			}
			// query params handled elsewhere
		}
	case strings.Contains(authStr, `"oauth2"`):
		if token := paramMap["accessToken"]; token != "" {
			headers["authorization"] = "Bearer " + token
		}
	}

	return headers
}

// convertBody converts Postman body to IR body.
func (c *Converter) convertBody(body *postman.Body, variables map[string]string) (interface{}, string) {
	if body == nil || body.Disabled {
		return nil, ""
	}

	switch body.Mode {
	case "raw":
		raw := resolveVars(body.Raw, variables)
		if raw == "" {
			return nil, ""
		}

		// Determine content type from options
		contentType := "text/plain"
		if body.Options != nil && body.Options.Raw.Language != "" {
			switch body.Options.Raw.Language {
			case "json":
				contentType = "application/json"
			case "xml":
				contentType = "application/xml"
			case "html":
				contentType = "text/html"
			case "javascript":
				contentType = "application/javascript"
			default:
				contentType = "text/plain"
			}
		}

		// Try to parse as JSON
		if strings.Contains(contentType, "json") {
			var parsed interface{}
			if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
				return parsed, contentType
			}
		}

		return raw, contentType

	case "urlencoded":
		if body.URLEncoded == nil {
			return nil, ""
		}
		pairs := c.parseFormData(body.URLEncoded, variables)
		if len(pairs) == 0 {
			return nil, ""
		}
		return pairs, "application/x-www-form-urlencoded"

	case "formdata":
		if body.FormData == nil {
			return nil, ""
		}
		pairs := c.parseFormData(body.FormData, variables)
		if len(pairs) == 0 {
			return nil, ""
		}
		return pairs, "multipart/form-data"

	case "graphql":
		if body.GraphQL == nil {
			return nil, ""
		}
		// GraphQL is typically JSON
		return body.GraphQL, "application/json"
	}

	return nil, ""
}

// parseFormData extracts key-value pairs from URL-encoded or form data.
func (c *Converter) parseFormData(data interface{}, variables map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	// Handle various formats
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				key, _ := m["key"].(string)
				value, _ := m["value"].(string)
				disabled, _ := m["disabled"].(bool)

				if !c.IncludeDisabled && disabled {
					continue
				}
				if key != "" {
					result[resolveVars(key, variables)] = resolveVars(value, variables)
				}
			}
		}
	case []map[string]interface{}:
		for _, m := range v {
			key, _ := m["key"].(string)
			value, _ := m["value"].(string)
			disabled, _ := m["disabled"].(bool)

			if !c.IncludeDisabled && disabled {
				continue
			}
			if key != "" {
				result[resolveVars(key, variables)] = resolveVars(value, variables)
			}
		}
	}

	return result
}

// convertResponse converts a Postman response to IR response.
func (c *Converter) convertResponse(resp *postman.Response, variables map[string]string) ir.Response {
	irResp := ir.Response{
		Status: resp.Code,
	}

	// Default to 200 if not specified
	if irResp.Status == 0 {
		irResp.Status = 200
	}

	// Convert headers
	if c.IncludeHeaders && resp.Headers != nil && len(resp.Headers.Headers) > 0 {
		headers := make(map[string]string)
		for _, h := range resp.Headers.Headers {
			if h == nil {
				continue
			}
			if !c.IncludeDisabled && h.Disabled {
				continue
			}
			key := strings.ToLower(h.Key)
			if key != "" {
				headers[key] = h.Value
				if key == "content-type" {
					irResp.ContentType = ptrString(h.Value)
				}
			}
		}
		if len(headers) > 0 {
			irResp.Headers = headers
		}
	}

	// Parse body
	if resp.Body != "" {
		body := resolveVars(resp.Body, variables)

		// Try to parse as JSON based on preview language or content type
		isJSON := resp.PreviewLanguage == "json" ||
			(irResp.ContentType != nil && strings.Contains(*irResp.ContentType, "json"))

		if isJSON {
			var parsed interface{}
			if err := json.Unmarshal([]byte(body), &parsed); err == nil {
				irResp.Body = parsed
			} else {
				irResp.Body = body
			}
		} else {
			irResp.Body = body
		}

		// Set content type if not already set
		if irResp.ContentType == nil {
			if isJSON {
				irResp.ContentType = ptrString("application/json")
			} else if resp.PreviewLanguage != "" {
				switch resp.PreviewLanguage {
				case "xml":
					irResp.ContentType = ptrString("application/xml")
				case "html":
					irResp.ContentType = ptrString("text/html")
				default:
					irResp.ContentType = ptrString("text/plain")
				}
			}
		}
	}

	return irResp
}

// buildVariables creates a combined variables map.
func (c *Converter) buildVariables(collection *postman.Collection) map[string]string {
	vars := make(map[string]string)

	// Start with converter-provided variables
	for k, v := range c.Variables {
		vars[k] = v
	}

	// Add base URL mappings
	if c.BaseURL != "" {
		vars["url"] = c.BaseURL
		vars["baseUrl"] = c.BaseURL
		vars["base_url"] = c.BaseURL
	}

	// Add collection-level variables (lower priority)
	for _, v := range collection.Variables {
		if v != nil && v.Key != "" && !v.Disabled {
			if _, exists := vars[v.Key]; !exists {
				vars[v.Key] = v.Value
			}
		}
	}

	return vars
}

// extractMetadata extracts API metadata from collection info.
func (c *Converter) extractMetadata(collection *postman.Collection, variables map[string]string) *ir.APIMetadata {
	now := time.Now().UTC()
	source := "postman"

	metadata := &ir.APIMetadata{
		GeneratedAt: &now,
		Source:      &source,
	}

	// Title from collection name
	if collection.Info.Name != "" {
		title := resolveVars(collection.Info.Name, variables)
		metadata.Title = &title
	}

	// Description from collection description
	if collection.Info.Description.Content != "" {
		desc := resolveVars(collection.Info.Description.Content, variables)
		metadata.Description = &desc
	}

	// Version from collection version
	if collection.Info.Version != "" {
		version := resolveVars(collection.Info.Version, variables)
		metadata.APIVersion = &version
	}

	// Try to extract server from base URL variable
	if c.BaseURL != "" {
		metadata.Servers = []ir.Server{
			{URL: c.BaseURL},
		}
	}

	return metadata
}

// shouldFilterHeader checks if a header should be filtered.
func (c *Converter) shouldFilterHeader(name string) bool {
	name = strings.ToLower(name)
	for _, filter := range c.FilterHeaders {
		if strings.ToLower(filter) == name {
			return true
		}
	}
	return false
}

// generateID generates a record ID.
func (c *Converter) generateID(counter, responseIndex int) *string {
	if !c.GenerateIDs {
		return nil
	}

	var id string
	if responseIndex > 0 {
		id = fmt.Sprintf("req-%04d-%d", counter, responseIndex)
	} else {
		id = fmt.Sprintf("req-%04d", counter)
	}
	return &id
}

// resolveVars replaces {{variable}} placeholders with values.
func resolveVars(text string, variables map[string]string) string {
	if text == "" {
		return text
	}

	// Replace {{variable}} patterns
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		// Extract variable name
		name := strings.TrimPrefix(strings.TrimSuffix(match, "}}"), "{{")
		if value, ok := variables[name]; ok {
			return value
		}
		// Leave unresolved variables as-is for transparency
		return match
	})

	return result
}

// sanitizeOperationId converts a name to a valid operation ID.
func sanitizeOperationId(name string) string {
	if name == "" {
		return ""
	}

	// Convert to camelCase
	words := regexp.MustCompile(`[\s\-_/]+`).Split(name, -1)
	result := strings.Builder{}

	for i, word := range words {
		if word == "" {
			continue
		}
		word = strings.ToLower(word)
		// Remove non-alphanumeric characters
		word = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(word, "")
		if word == "" {
			continue
		}
		if i == 0 {
			result.WriteString(word)
		} else {
			// Capitalize first letter
			result.WriteString(strings.ToUpper(word[:1]) + word[1:])
		}
	}

	id := result.String()

	// Ensure it starts with a letter or underscore
	if len(id) > 0 && (id[0] >= '0' && id[0] <= '9') {
		id = "_" + id
	}

	return id
}

// Helper functions
func ptrString(s string) *string {
	return &s
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrSource(s ir.IRRecordSource) *ir.IRRecordSource {
	return &s
}

// GenerateUUID generates a new UUID string.
func GenerateUUID() string {
	return uuid.New().String()
}
