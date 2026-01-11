package inference

import (
	"strings"
	"sync"
)

// EndpointClusterer groups IR records by endpoint (method + path template).
type EndpointClusterer struct {
	mu                 sync.RWMutex
	pathInferrer       *PathInferrer
	endpoints          map[string]*EndpointData
	hosts              map[string]bool
	schemes            map[string]bool
	securityDetector   *SecurityDetector
	paginationDetector *PaginationDetector
	rateLimitDetector  *RateLimitDetector
}

// NewEndpointClusterer creates a new EndpointClusterer.
func NewEndpointClusterer() *EndpointClusterer {
	return &EndpointClusterer{
		pathInferrer:       NewPathInferrer(),
		endpoints:          make(map[string]*EndpointData),
		hosts:              make(map[string]bool),
		schemes:            make(map[string]bool),
		securityDetector:   NewSecurityDetector(),
		paginationDetector: NewPaginationDetector(),
		rateLimitDetector:  NewRateLimitDetector(),
	}
}

// AddRecord processes an IR record and adds it to the appropriate endpoint.
func (c *EndpointClusterer) AddRecord(method, path string, pathTemplate string, pathParams map[string]string,
	query map[string]any, headers map[string]string, requestBody any, requestContentType string,
	status int, responseBody any, responseContentType string, responseHeaders map[string]string,
	host string, scheme string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Track host and scheme
	if host != "" {
		c.hosts[host] = true
	}
	if scheme != "" {
		c.schemes[scheme] = true
	}

	// Infer path template if not provided
	var inferredParams map[string]string
	if pathTemplate == "" {
		pathTemplate, inferredParams = c.pathInferrer.InferTemplate(path)
	} else {
		inferredParams = pathParams
	}

	// Get or create endpoint
	key := EndpointKey(method, pathTemplate)
	endpoint, exists := c.endpoints[key]
	if !exists {
		endpoint = NewEndpointData(method, pathTemplate)
		c.endpoints[key] = endpoint
	}

	endpoint.RequestCount++

	// Process path parameters
	for name, value := range inferredParams {
		param, exists := endpoint.PathParams[name]
		if !exists {
			param = NewParamData(name)
			param.Required = true // Path params are always required
			endpoint.PathParams[name] = param
		}
		param.AddValue(value)
	}

	// Process query parameters
	for name, value := range query {
		param, exists := endpoint.QueryParams[name]
		if !exists {
			param = NewParamData(name)
			param.Required = false // Query params start as optional
			endpoint.QueryParams[name] = param
		}
		param.AddValue(value)
	}

	// Update query param optionality
	for name, param := range endpoint.QueryParams {
		if _, inThisRequest := query[name]; !inThisRequest {
			param.Required = false
		}
	}

	// Process header parameters (exclude common headers)
	for name, value := range headers {
		if isExcludedHeader(name) {
			continue
		}
		param, exists := endpoint.HeaderParams[name]
		if !exists {
			param = NewParamData(name)
			param.Required = false
			endpoint.HeaderParams[name] = param
		}
		param.AddValue(value)
	}

	// Detect security schemes from request headers
	c.securityDetector.DetectFromHeaders(headers)

	// Detect pagination patterns from query parameters
	c.paginationDetector.DetectFromQuery(query)

	// Process request body
	if requestBody != nil {
		if endpoint.RequestBody == nil {
			ct := requestContentType
			if ct == "" {
				ct = "application/json"
			}
			endpoint.RequestBody = NewBodyData(ct)
		}
		ProcessBody(endpoint.RequestBody.Schema, requestBody)
	}

	// Process response
	if status > 0 {
		resp, exists := endpoint.Responses[status]
		if !exists {
			resp = NewResponseData(status)
			if responseContentType != "" {
				resp.ContentType = responseContentType
			} else {
				resp.ContentType = "application/json"
			}
			endpoint.Responses[status] = resp
		}

		// Process response body
		if responseBody != nil {
			ProcessBody(resp.Body, responseBody)
		}

		// Process response headers
		for name, value := range responseHeaders {
			if isExcludedHeader(name) {
				continue
			}
			param, exists := resp.Headers[name]
			if !exists {
				param = NewParamData(name)
				resp.Headers[name] = param
			}
			param.AddValue(value)
		}

		// Detect rate limit headers from response
		c.rateLimitDetector.DetectFromHeaders(responseHeaders)
	}
}

// Finalize completes the inference process (e.g., marking optional fields).
func (c *EndpointClusterer) Finalize() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, endpoint := range c.endpoints {
		// Finalize request body schema
		if endpoint.RequestBody != nil {
			endpoint.RequestBody.Schema.FinalizeOptional()
		}

		// Finalize response schemas
		for _, resp := range endpoint.Responses {
			resp.Body.FinalizeOptional()
		}
	}
}

// GetResult returns the inference result.
func (c *EndpointClusterer) GetResult() *InferenceResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := NewInferenceResult()

	// Copy endpoints
	for key, endpoint := range c.endpoints {
		result.Endpoints[key] = endpoint
	}

	// Collect hosts
	for host := range c.hosts {
		result.Hosts = append(result.Hosts, host)
	}

	// Collect schemes
	for scheme := range c.schemes {
		result.Schemes = append(result.Schemes, scheme)
	}

	// Copy detected security schemes
	for key, scheme := range c.securityDetector.GetSchemes() {
		result.SecuritySchemes[key] = scheme
	}

	// Copy detected pagination parameters
	for key, param := range c.paginationDetector.GetParams() {
		result.PaginationParams[key] = param
	}

	// Copy detected rate limit headers
	for key, header := range c.rateLimitDetector.GetHeaders() {
		result.RateLimitHeaders[key] = header
	}

	return result
}

// Headers to exclude from documentation
var excludedHeaders = map[string]bool{
	"content-length":                   true,
	"content-type":                     true,
	"date":                             true,
	"server":                           true,
	"connection":                       true,
	"keep-alive":                       true,
	"transfer-encoding":                true,
	"accept":                           true,
	"accept-encoding":                  true,
	"accept-language":                  true,
	"user-agent":                       true,
	"host":                             true,
	"cache-control":                    true,
	"pragma":                           true,
	"expires":                          true,
	"x-request-id":                     true,
	"x-correlation-id":                 true,
	"x-trace-id":                       true,
	"x-forwarded-for":                  true,
	"x-forwarded-proto":                true,
	"x-forwarded-host":                 true,
	"x-real-ip":                        true,
	"cf-ray":                           true,
	"cf-connecting-ip":                 true,
	"cf-ipcountry":                     true,
	"cf-visitor":                       true,
	"cf-request-id":                    true,
	"x-amzn-requestid":                 true,
	"x-amzn-trace-id":                  true,
	"x-cache":                          true,
	"x-cache-hits":                     true,
	"x-served-by":                      true,
	"x-timer":                          true,
	"vary":                             true,
	"etag":                             true,
	"last-modified":                    true,
	"if-none-match":                    true,
	"if-modified-since":                true,
	"access-control-allow-origin":      true,
	"access-control-allow-methods":     true,
	"access-control-allow-headers":     true,
	"access-control-allow-credentials": true,
	"access-control-max-age":           true,
	"access-control-expose-headers":    true,
}

// isExcludedHeader checks if a header should be excluded.
func isExcludedHeader(name string) bool {
	return excludedHeaders[strings.ToLower(name)]
}
