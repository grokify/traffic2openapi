// Package inference provides algorithms for inferring OpenAPI schemas from HTTP traffic.
package inference

import (
	"sync"
)

// SchemaStore tracks JSON field paths and their observed values.
// Paths use dot notation (e.g., "user.address.city") with array markers (e.g., "items[].name").
type SchemaStore struct {
	mu          sync.RWMutex
	Examples    map[string][]any  // path -> unique example values
	Types       map[string]string // path -> inferred type (string, number, integer, boolean, array, object)
	Optional    map[string]bool   // path -> true if not present in all observations
	Nullable    map[string]bool   // path -> true if null was observed
	Formats     map[string]string // path -> detected format (email, uuid, date-time, uri, etc.)
	seenCount   map[string]int    // path -> number of times seen
	totalCount  int               // total observations
	maxExamples int
}

// NewSchemaStore creates a new SchemaStore with default settings.
func NewSchemaStore() *SchemaStore {
	return &SchemaStore{
		Examples:    make(map[string][]any),
		Types:       make(map[string]string),
		Optional:    make(map[string]bool),
		Nullable:    make(map[string]bool),
		Formats:     make(map[string]string),
		seenCount:   make(map[string]int),
		maxExamples: 5,
	}
}

// AddObservation records a new observation of the schema.
// This increments the total count for optionality tracking.
func (s *SchemaStore) AddObservation() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalCount++
}

// AddValue adds a value at a given path.
func (s *SchemaStore) AddValue(path string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Track that this path was seen
	s.seenCount[path]++

	// Handle null
	if value == nil {
		s.Nullable[path] = true
		return
	}

	// Infer type
	inferredType := inferType(value)
	if existing, ok := s.Types[path]; ok {
		s.Types[path] = mergeTypes(existing, inferredType)
	} else {
		s.Types[path] = inferredType
	}

	// Detect format for strings
	if str, ok := value.(string); ok {
		if format := detectFormat(str); format != "" {
			s.Formats[path] = format
		}
	}

	// Add example if unique and under limit
	if len(s.Examples[path]) < s.maxExamples {
		if !s.hasExample(path, value) {
			s.Examples[path] = append(s.Examples[path], value)
		}
	}
}

// hasExample checks if a value already exists in examples (no lock, internal use).
func (s *SchemaStore) hasExample(path string, value any) bool {
	for _, ex := range s.Examples[path] {
		if valuesEqual(ex, value) {
			return true
		}
	}
	return false
}

// FinalizeOptional marks paths as optional if they weren't seen in all observations.
func (s *SchemaStore) FinalizeOptional() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for path, count := range s.seenCount {
		if count < s.totalCount {
			s.Optional[path] = true
		}
	}
}

// GetPaths returns all tracked paths.
func (s *SchemaStore) GetPaths() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	paths := make([]string, 0, len(s.Examples))
	for path := range s.Examples {
		paths = append(paths, path)
	}
	// Also include paths that only had null values
	for path := range s.Nullable {
		if _, ok := s.Examples[path]; !ok {
			paths = append(paths, path)
		}
	}
	return paths
}

// EndpointData represents aggregated data for a single API endpoint.
type EndpointData struct {
	Method       string
	PathTemplate string
	PathParams   map[string]*ParamData // parameter name -> data
	QueryParams  map[string]*ParamData // parameter name -> data
	HeaderParams map[string]*ParamData // header name -> data
	RequestBody  *BodyData             // request body schema
	Responses    map[int]*ResponseData // status code -> response data
	RequestCount int                   // number of requests observed
}

// NewEndpointData creates a new EndpointData.
func NewEndpointData(method, pathTemplate string) *EndpointData {
	return &EndpointData{
		Method:       method,
		PathTemplate: pathTemplate,
		PathParams:   make(map[string]*ParamData),
		QueryParams:  make(map[string]*ParamData),
		HeaderParams: make(map[string]*ParamData),
		Responses:    make(map[int]*ResponseData),
	}
}

// ParamData tracks parameter values and infers type/format.
type ParamData struct {
	Name      string
	Examples  []any
	Type      string // string, integer, number, boolean
	Format    string // uuid, email, date-time, etc.
	Required  bool
	seenCount int
}

// NewParamData creates a new ParamData.
func NewParamData(name string) *ParamData {
	return &ParamData{
		Name:     name,
		Examples: make([]any, 0, 5),
		Type:     "string",
		Required: true,
	}
}

// AddValue adds a value to the parameter.
func (p *ParamData) AddValue(value any) {
	p.seenCount++

	// Infer type
	inferredType := inferType(value)
	if p.Type == "" || p.Type == "string" {
		p.Type = inferredType
	} else {
		p.Type = mergeTypes(p.Type, inferredType)
	}

	// Detect format for strings
	if str, ok := value.(string); ok {
		if format := detectFormat(str); format != "" {
			p.Format = format
		}
	}

	// Add example
	if len(p.Examples) < 5 {
		for _, ex := range p.Examples {
			if valuesEqual(ex, value) {
				return
			}
		}
		p.Examples = append(p.Examples, value)
	}
}

// BodyData tracks request/response body schema.
type BodyData struct {
	ContentType string
	Schema      *SchemaStore
}

// NewBodyData creates a new BodyData.
func NewBodyData(contentType string) *BodyData {
	return &BodyData{
		ContentType: contentType,
		Schema:      NewSchemaStore(),
	}
}

// ResponseData tracks response information for a status code.
type ResponseData struct {
	StatusCode  int
	ContentType string
	Headers     map[string]*ParamData
	Body        *SchemaStore
}

// NewResponseData creates a new ResponseData.
func NewResponseData(statusCode int) *ResponseData {
	return &ResponseData{
		StatusCode: statusCode,
		Headers:    make(map[string]*ParamData),
		Body:       NewSchemaStore(),
	}
}

// InferenceResult holds the complete inference results.
type InferenceResult struct {
	Endpoints        map[string]*EndpointData           // key: "METHOD /path/template"
	Hosts            []string                           // observed hosts
	Schemes          []string                           // observed schemes (http, https)
	SecuritySchemes  map[string]*DetectedSecurityScheme // detected authentication schemes
	PaginationParams map[string]*PaginationParam        // detected pagination parameters
	RateLimitHeaders map[string]*RateLimitHeader        // detected rate limit headers
}

// NewInferenceResult creates a new InferenceResult.
func NewInferenceResult() *InferenceResult {
	return &InferenceResult{
		Endpoints:        make(map[string]*EndpointData),
		Hosts:            make([]string, 0),
		Schemes:          make([]string, 0),
		SecuritySchemes:  make(map[string]*DetectedSecurityScheme),
		PaginationParams: make(map[string]*PaginationParam),
		RateLimitHeaders: make(map[string]*RateLimitHeader),
	}
}
