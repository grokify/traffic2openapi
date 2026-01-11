package inference

import (
	"regexp"
	"strconv"
	"strings"
)

// SecurityDetector detects authentication schemes from request headers.
type SecurityDetector struct {
	schemes map[string]*DetectedSecurityScheme
}

// DetectedSecurityScheme represents a detected security scheme.
type DetectedSecurityScheme struct {
	Type         string // http, apiKey, oauth2
	Scheme       string // bearer, basic (for http type)
	Name         string // header name (for apiKey type)
	In           string // header, query, cookie
	BearerFormat string // JWT, etc.
	Count        int    // number of times observed
}

// NewSecurityDetector creates a new SecurityDetector.
func NewSecurityDetector() *SecurityDetector {
	return &SecurityDetector{
		schemes: make(map[string]*DetectedSecurityScheme),
	}
}

// DetectFromHeaders analyzes request headers for security schemes.
func (d *SecurityDetector) DetectFromHeaders(headers map[string]string) {
	for name, value := range headers {
		nameLower := strings.ToLower(name)

		switch nameLower {
		case "authorization":
			d.detectAuthorizationHeader(value)
		case "x-api-key", "api-key", "apikey":
			d.addScheme("apiKeyHeader", &DetectedSecurityScheme{
				Type: "apiKey",
				Name: name,
				In:   "header",
			})
		case "x-auth-token", "x-access-token":
			d.addScheme("tokenHeader", &DetectedSecurityScheme{
				Type: "apiKey",
				Name: name,
				In:   "header",
			})
		}
	}
}

func (d *SecurityDetector) detectAuthorizationHeader(value string) {
	valueLower := strings.ToLower(value)

	if strings.HasPrefix(valueLower, "bearer ") {
		token := strings.TrimPrefix(value, "Bearer ")
		token = strings.TrimPrefix(token, "bearer ")

		scheme := &DetectedSecurityScheme{
			Type:   "http",
			Scheme: "bearer",
		}

		// Detect JWT format
		if isJWT(token) {
			scheme.BearerFormat = "JWT"
		}

		d.addScheme("bearerAuth", scheme)
	} else if strings.HasPrefix(valueLower, "basic ") {
		d.addScheme("basicAuth", &DetectedSecurityScheme{
			Type:   "http",
			Scheme: "basic",
		})
	} else if strings.HasPrefix(valueLower, "digest ") {
		d.addScheme("digestAuth", &DetectedSecurityScheme{
			Type:   "http",
			Scheme: "digest",
		})
	}
}

func (d *SecurityDetector) addScheme(key string, scheme *DetectedSecurityScheme) {
	if existing, ok := d.schemes[key]; ok {
		existing.Count++
		// Merge bearer format if detected
		if scheme.BearerFormat != "" && existing.BearerFormat == "" {
			existing.BearerFormat = scheme.BearerFormat
		}
	} else {
		scheme.Count = 1
		d.schemes[key] = scheme
	}
}

// GetSchemes returns all detected security schemes.
func (d *SecurityDetector) GetSchemes() map[string]*DetectedSecurityScheme {
	return d.schemes
}

// isJWT checks if a token appears to be a JWT (has 3 base64 parts separated by dots).
func isJWT(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	// Each part should be base64-like
	base64Pattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	for _, part := range parts {
		if !base64Pattern.MatchString(part) {
			return false
		}
	}
	return true
}

// PaginationDetector detects pagination patterns from query parameters.
type PaginationDetector struct {
	params map[string]*PaginationParam
}

// PaginationParam represents a detected pagination parameter.
type PaginationParam struct {
	Name        string
	Type        string   // offset, cursor, page
	Examples    []string // observed values
	Min         *int     // minimum observed value
	Max         *int     // maximum observed value
	Description string
}

// Common pagination parameter names
var paginationParams = map[string]string{
	"page":      "page",
	"page_num":  "page",
	"pagenum":   "page",
	"p":         "page",
	"limit":     "offset",
	"per_page":  "offset",
	"perpage":   "offset",
	"page_size": "offset",
	"pagesize":  "offset",
	"size":      "offset",
	"count":     "offset",
	"offset":    "offset",
	"skip":      "offset",
	"cursor":    "cursor",
	"after":     "cursor",
	"before":    "cursor",
	"start":     "offset",
	"from":      "offset",
	"next":      "cursor",
	"prev":      "cursor",
}

// NewPaginationDetector creates a new PaginationDetector.
func NewPaginationDetector() *PaginationDetector {
	return &PaginationDetector{
		params: make(map[string]*PaginationParam),
	}
}

// DetectFromQuery analyzes query parameters for pagination patterns.
func (d *PaginationDetector) DetectFromQuery(query map[string]any) {
	for name, value := range query {
		nameLower := strings.ToLower(name)

		if paginationType, ok := paginationParams[nameLower]; ok {
			d.addParam(name, paginationType, value)
		}
	}
}

func (d *PaginationDetector) addParam(name, paginationType string, value any) {
	param, exists := d.params[name]
	if !exists {
		param = &PaginationParam{
			Name:     name,
			Type:     paginationType,
			Examples: make([]string, 0, 5),
		}
		param.Description = d.getDescription(name)
		d.params[name] = param
	}

	// Track value
	strValue := toString(value)
	if len(param.Examples) < 5 {
		found := false
		for _, ex := range param.Examples {
			if ex == strValue {
				found = true
				break
			}
		}
		if !found {
			param.Examples = append(param.Examples, strValue)
		}
	}

	// Track min/max for numeric values
	if intVal, err := strconv.Atoi(strValue); err == nil {
		if param.Min == nil || intVal < *param.Min {
			param.Min = &intVal
		}
		if param.Max == nil || intVal > *param.Max {
			param.Max = &intVal
		}
	}
}

func (d *PaginationDetector) getDescription(name string) string {
	nameLower := strings.ToLower(name)

	switch nameLower {
	case "page", "page_num", "p":
		return "Page number (1-indexed)"
	case "limit", "per_page", "page_size", "size":
		return "Number of items per page"
	case "offset", "skip":
		return "Number of items to skip"
	case "cursor", "after":
		return "Cursor for pagination (from previous response)"
	case "before":
		return "Cursor for reverse pagination"
	default:
		return ""
	}
}

// GetParams returns all detected pagination parameters.
func (d *PaginationDetector) GetParams() map[string]*PaginationParam {
	return d.params
}

// RateLimitDetector detects rate limiting patterns from response headers.
type RateLimitDetector struct {
	headers map[string]*RateLimitHeader
}

// RateLimitHeader represents a detected rate limit header.
type RateLimitHeader struct {
	Name        string
	Description string
	Type        string // integer, string
	Example     string
	Count       int
}

// Common rate limit headers
var rateLimitHeaders = map[string]string{
	"x-ratelimit-limit":      "Maximum number of requests allowed in the time window",
	"x-ratelimit-remaining":  "Number of requests remaining in the current time window",
	"x-ratelimit-reset":      "Unix timestamp when the rate limit resets",
	"x-rate-limit-limit":     "Maximum number of requests allowed in the time window",
	"x-rate-limit-remaining": "Number of requests remaining in the current time window",
	"x-rate-limit-reset":     "Unix timestamp when the rate limit resets",
	"ratelimit-limit":        "Maximum number of requests allowed in the time window",
	"ratelimit-remaining":    "Number of requests remaining in the current time window",
	"ratelimit-reset":        "Unix timestamp when the rate limit resets",
	"retry-after":            "Number of seconds to wait before retrying",
	"x-retry-after":          "Number of seconds to wait before retrying",
}

// NewRateLimitDetector creates a new RateLimitDetector.
func NewRateLimitDetector() *RateLimitDetector {
	return &RateLimitDetector{
		headers: make(map[string]*RateLimitHeader),
	}
}

// DetectFromHeaders analyzes response headers for rate limiting patterns.
func (d *RateLimitDetector) DetectFromHeaders(headers map[string]string) {
	for name, value := range headers {
		nameLower := strings.ToLower(name)

		if desc, ok := rateLimitHeaders[nameLower]; ok {
			d.addHeader(name, desc, value)
		}
	}
}

func (d *RateLimitDetector) addHeader(name, description, value string) {
	header, exists := d.headers[name]
	if !exists {
		header = &RateLimitHeader{
			Name:        name,
			Description: description,
			Example:     value,
		}

		// Detect type
		if _, err := strconv.Atoi(value); err == nil {
			header.Type = "integer"
		} else {
			header.Type = "string"
		}

		d.headers[name] = header
	}
	header.Count++
}

// GetHeaders returns all detected rate limit headers.
func (d *RateLimitDetector) GetHeaders() map[string]*RateLimitHeader {
	return d.headers
}

// toString converts a value to string.
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return ""
	}
}
