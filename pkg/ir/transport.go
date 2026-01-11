package ir

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LoggingTransport is an http.RoundTripper that logs HTTP traffic as IR records.
type LoggingTransport struct {
	// Base is the underlying transport. If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	// Writer receives IR records for each request/response.
	Writer IRWriter

	// Options configures logging behavior.
	Options LoggingOptions

	// ErrorHandler is called when writing an IR record fails.
	// If nil, write errors are silently ignored (HTTP request still succeeds).
	ErrorHandler ErrorHandler
}

// LoggingOptions configures the LoggingTransport behavior.
type LoggingOptions struct {
	// FilterHeaders are headers to exclude from logging (case-insensitive).
	// Defaults to common sensitive headers if nil.
	FilterHeaders []string

	// IncludeRequestBody controls whether request bodies are captured.
	IncludeRequestBody bool

	// IncludeResponseBody controls whether response bodies are captured.
	IncludeResponseBody bool

	// MaxBodySize limits body capture size. 0 means no limit.
	MaxBodySize int64

	// Source is the source identifier for IR records.
	Source IRRecordSource

	// --- Request Filtering ---

	// SkipPaths are path prefixes to skip logging (e.g., "/health", "/metrics").
	// If a request path starts with any of these prefixes, it won't be logged.
	SkipPaths []string

	// AllowMethods limits logging to specific HTTP methods (e.g., "GET", "POST").
	// If empty, all methods are logged.
	AllowMethods []string

	// AllowHosts limits logging to specific hosts.
	// If empty, all hosts are logged.
	AllowHosts []string

	// SkipStatusCodes are status codes to skip logging (e.g., 404, 500).
	SkipStatusCodes []int

	// SampleRate is the percentage of requests to log (0.0 to 1.0).
	// Values > 0.0 and < 1.0 enable probabilistic sampling (e.g., 0.5 = 50%).
	// Values <= 0.0 or >= 1.0 log all requests.
	// The zero value (0.0) means "not configured" and logs all requests,
	// making it safe to use partial LoggingOptions without setting SampleRate.
	SampleRate float64

	// --- Context Support ---

	// RequestIDHeaders are headers to check for request ID (in order of priority).
	// The first non-empty value found will be used as the record ID.
	// If empty or no header found, a UUID is generated.
	// Common headers: "X-Request-ID", "X-Correlation-ID", "X-Trace-ID"
	RequestIDHeaders []string
}

// DefaultLoggingOptions returns sensible defaults for logging.
func DefaultLoggingOptions() LoggingOptions {
	return LoggingOptions{
		FilterHeaders: []string{
			"authorization",
			"cookie",
			"set-cookie",
			"x-api-key",
			"x-auth-token",
		},
		IncludeRequestBody:  true,
		IncludeResponseBody: true,
		MaxBodySize:         1 << 20, // 1MB
		Source:              IRRecordSourceProxy,
		SampleRate:          1.0, // Log all requests by default
	}
}

// LoggingTransportOption configures a LoggingTransport.
type LoggingTransportOption func(*LoggingTransport)

// WithBase sets the base transport.
func WithBase(base http.RoundTripper) LoggingTransportOption {
	return func(t *LoggingTransport) {
		t.Base = base
	}
}

// WithLoggingOptions sets the logging options.
func WithLoggingOptions(opts LoggingOptions) LoggingTransportOption {
	return func(t *LoggingTransport) {
		t.Options = opts
	}
}

// WithTransportErrorHandler sets the error handler for write failures.
func WithTransportErrorHandler(handler ErrorHandler) LoggingTransportOption {
	return func(t *LoggingTransport) {
		t.ErrorHandler = handler
	}
}

// NewLoggingTransport creates a new logging transport.
func NewLoggingTransport(writer IRWriter, opts ...LoggingTransportOption) *LoggingTransport {
	t := &LoggingTransport{
		Base:    http.DefaultTransport,
		Writer:  writer,
		Options: DefaultLoggingOptions(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RoundTrip implements http.RoundTripper.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check pre-request filters (path, method, host)
	if !t.shouldLogRequest(req) {
		return t.Base.RoundTrip(req)
	}

	startTime := time.Now()

	// Capture request
	irReq, reqBody := t.captureRequest(req)

	// Restore request body if we consumed it
	if reqBody != nil {
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	// Execute actual request
	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Check post-request filters (status code)
	if !t.shouldLogResponse(resp) {
		return resp, nil
	}

	duration := time.Since(startTime)

	// Capture response
	irResp, respBody := t.captureResponse(resp)

	// Restore response body
	if respBody != nil {
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
	}

	// Extract request ID from headers if configured
	requestID := t.extractRequestID(req)

	// Build and write IR record
	record := t.buildRecord(irReq, irResp, startTime, duration, requestID)
	if err := t.Writer.Write(record); err != nil && t.ErrorHandler != nil {
		t.ErrorHandler(err)
	}

	return resp, nil
}

// shouldLogRequest checks if a request should be logged based on filters.
func (t *LoggingTransport) shouldLogRequest(req *http.Request) bool {
	// Check sampling rate.
	// SampleRate <= 0.0 means "not configured", treat as 1.0 (log all requests).
	// SampleRate between 0.0 and 1.0 enables probabilistic sampling.
	// SampleRate >= 1.0 logs all requests.
	if t.Options.SampleRate > 0.0 && t.Options.SampleRate < 1.0 {
		if rand.Float64() > t.Options.SampleRate { //nolint:gosec // G404: sampling doesn't need crypto rand
			return false
		}
	}

	// Check path filters
	if len(t.Options.SkipPaths) > 0 {
		for _, prefix := range t.Options.SkipPaths {
			if strings.HasPrefix(req.URL.Path, prefix) {
				return false
			}
		}
	}

	// Check method filters
	if len(t.Options.AllowMethods) > 0 {
		allowed := false
		for _, m := range t.Options.AllowMethods {
			if strings.EqualFold(req.Method, m) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// Check host filters
	if len(t.Options.AllowHosts) > 0 {
		host := req.URL.Host
		if host == "" {
			host = req.Host
		}
		allowed := false
		for _, h := range t.Options.AllowHosts {
			if strings.EqualFold(host, h) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	return true
}

// shouldLogResponse checks if a response should be logged based on filters.
func (t *LoggingTransport) shouldLogResponse(resp *http.Response) bool {
	// Check status code filters
	if len(t.Options.SkipStatusCodes) > 0 {
		for _, code := range t.Options.SkipStatusCodes {
			if resp.StatusCode == code {
				return false
			}
		}
	}

	return true
}

func (t *LoggingTransport) captureRequest(req *http.Request) (Request, []byte) {
	irReq := Request{
		Method: RequestMethod(req.Method),
		Path:   req.URL.Path,
	}

	// Scheme
	scheme := req.URL.Scheme
	if scheme == "" {
		scheme = "https"
	}
	if scheme == "https" {
		irReq.Scheme = RequestSchemeHttps
	} else {
		irReq.Scheme = RequestSchemeHttp
	}

	// Host
	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	irReq.Host = &host

	// Query parameters
	if len(req.URL.Query()) > 0 {
		query := make(map[string]interface{})
		for k, v := range req.URL.Query() {
			if len(v) == 1 {
				query[k] = v[0]
			} else {
				query[k] = v
			}
		}
		irReq.Query = query
	}

	// Headers
	headers := t.filterHeaders(req.Header)
	if len(headers) > 0 {
		irReq.Headers = headers
	}

	// Content-Type
	if ct := req.Header.Get("Content-Type"); ct != "" {
		irReq.ContentType = &ct
	}

	// Request body
	var bodyBytes []byte
	if t.Options.IncludeRequestBody && req.Body != nil {
		bodyBytes = t.readBody(req.Body, t.Options.MaxBodySize)
		if len(bodyBytes) > 0 {
			irReq.Body = t.parseBody(bodyBytes, req.Header.Get("Content-Type"))
		}
	}

	return irReq, bodyBytes
}

func (t *LoggingTransport) captureResponse(resp *http.Response) (Response, []byte) {
	irResp := Response{
		Status: resp.StatusCode,
	}

	// Headers
	headers := t.filterHeaders(resp.Header)
	if len(headers) > 0 {
		irResp.Headers = headers
	}

	// Content-Type
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		irResp.ContentType = &ct
	}

	// Response body
	var bodyBytes []byte
	if t.Options.IncludeResponseBody && resp.Body != nil {
		bodyBytes = t.readBody(resp.Body, t.Options.MaxBodySize)
		if len(bodyBytes) > 0 {
			irResp.Body = t.parseBody(bodyBytes, resp.Header.Get("Content-Type"))
		}
	}

	return irResp, bodyBytes
}

// extractRequestID extracts request ID from configured headers.
// Returns empty string if no headers configured or no value found.
func (t *LoggingTransport) extractRequestID(req *http.Request) string {
	for _, header := range t.Options.RequestIDHeaders {
		if val := req.Header.Get(header); val != "" {
			return val
		}
	}
	return ""
}

func (t *LoggingTransport) buildRecord(req Request, resp Response, startTime time.Time, duration time.Duration, requestID string) *IRRecord {
	var id string
	if requestID != "" {
		id = requestID
	} else {
		id = uuid.New().String()
	}
	ts := startTime.UTC()
	durationMs := float64(duration.Milliseconds())
	source := t.Options.Source

	return &IRRecord{
		Id:         &id,
		Timestamp:  &ts,
		Source:     &source,
		Request:    req,
		Response:   resp,
		DurationMs: &durationMs,
	}
}

func (t *LoggingTransport) filterHeaders(h http.Header) map[string]string {
	if h == nil {
		return nil
	}

	filterSet := make(map[string]bool)
	for _, f := range t.Options.FilterHeaders {
		filterSet[strings.ToLower(f)] = true
	}

	result := make(map[string]string)
	for k, v := range h {
		key := strings.ToLower(k)
		if !filterSet[key] && len(v) > 0 {
			result[key] = v[0]
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func (t *LoggingTransport) readBody(body io.ReadCloser, maxSize int64) []byte {
	if body == nil {
		return nil
	}

	var reader io.Reader = body
	if maxSize > 0 {
		reader = io.LimitReader(body, maxSize)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil
	}

	return data
}

func (t *LoggingTransport) parseBody(data []byte, contentType string) interface{} {
	if len(data) == 0 {
		return nil
	}

	// Try to parse as JSON
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "+json") {
		var v interface{}
		if err := json.Unmarshal(data, &v); err == nil {
			return v
		}
	}

	// Return as string
	return string(data)
}
