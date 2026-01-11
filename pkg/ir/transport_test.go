package ir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// MemoryWriter collects IR records in memory for testing.
type MemoryWriter struct {
	Records []*IRRecord
	mu      sync.Mutex
}

func (w *MemoryWriter) Write(record *IRRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Records = append(w.Records, record)
	return nil
}

func (w *MemoryWriter) Flush() error {
	return nil
}

func (w *MemoryWriter) Close() error {
	return nil
}

func TestLoggingTransportBasic(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create logging transport
	writer := &MemoryWriter{}
	transport := NewLoggingTransport(writer)

	client := &http.Client{Transport: transport}

	// Make request
	resp, err := client.Get(server.URL + "/test?foo=bar")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response to ensure it wasn't consumed
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ok") {
		t.Errorf("response body not preserved: %s", body)
	}

	// Check IR record
	if len(writer.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(writer.Records))
	}

	record := writer.Records[0]

	if record.Request.Method != RequestMethodGET {
		t.Errorf("expected GET, got %s", record.Request.Method)
	}

	if record.Request.Path != "/test" {
		t.Errorf("expected /test, got %s", record.Request.Path)
	}

	if record.Response.Status != 200 {
		t.Errorf("expected status 200, got %d", record.Response.Status)
	}

	// Check query params captured
	if record.Request.Query == nil {
		t.Error("query params not captured")
	} else if record.Request.Query["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", record.Request.Query)
	}

	// Check response body captured
	if record.Response.Body == nil {
		t.Error("response body not captured")
	}

	// Check duration is set (may be 0 for very fast requests)
	if record.DurationMs == nil {
		t.Error("duration not captured")
	}
}

func TestLoggingTransportPOST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(body); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	writer := &MemoryWriter{}
	transport := NewLoggingTransport(writer)
	client := &http.Client{Transport: transport}

	reqBody := `{"name":"test"}`
	resp, err := client.Post(server.URL+"/users", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Ensure response body was preserved
	respBody, _ := io.ReadAll(resp.Body)
	if string(respBody) != reqBody {
		t.Errorf("response body not preserved: %s", respBody)
	}

	record := writer.Records[0]

	if record.Request.Method != RequestMethodPOST {
		t.Errorf("expected POST, got %s", record.Request.Method)
	}

	// Check request body captured
	if record.Request.Body == nil {
		t.Error("request body not captured")
	}
}

func TestLoggingTransportHeaderFiltering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	writer := &MemoryWriter{}
	transport := NewLoggingTransport(writer)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("X-Custom-Header", "visible")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	record := writer.Records[0]

	// Authorization should be filtered
	if record.Request.Headers != nil {
		if _, ok := record.Request.Headers["authorization"]; ok {
			t.Error("authorization header should be filtered")
		}
	}

	// Custom header should be present
	if record.Request.Headers == nil || record.Request.Headers["x-custom-header"] != "visible" {
		t.Error("custom header should be captured")
	}
}

func TestLoggingTransportErrorHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a writer that always fails
	failingWriter := &failingWriter{}

	var capturedErr error
	transport := NewLoggingTransport(failingWriter,
		WithTransportErrorHandler(func(err error) {
			capturedErr = err
		}),
	)

	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// Error handler should have been called
	if capturedErr == nil {
		t.Error("error handler was not called")
	}
	if capturedErr.Error() != "simulated write failure" {
		t.Errorf("unexpected error: %v", capturedErr)
	}
}

type failingWriter struct{}

func (w *failingWriter) Write(record *IRRecord) error {
	return fmt.Errorf("simulated write failure")
}

func (w *failingWriter) Flush() error {
	return nil
}

func (w *failingWriter) Close() error {
	return nil
}

func TestAsyncNDJSONWriter(t *testing.T) {
	var buf bytes.Buffer
	syncWriter := NewNDJSONWriter(&buf)

	var capturedErrors []error
	var mu sync.Mutex

	asyncWriter := NewAsyncNDJSONWriter(syncWriter,
		WithBufferSize(10),
		WithErrorHandler(func(err error) {
			mu.Lock()
			capturedErrors = append(capturedErrors, err)
			mu.Unlock()
		}),
	)

	// Write some records (async Write returns nil, errors go to callback)
	for i := 0; i < 5; i++ {
		record := NewRecord(RequestMethodGET, "/test", 200)
		if err := asyncWriter.Write(record); err != nil {
			t.Errorf("async write returned error: %v", err)
		}
	}

	// Close and wait for writes
	if err := asyncWriter.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// Check records were written
	if asyncWriter.Count() != 5 {
		t.Errorf("expected 5 records, got %d", asyncWriter.Count())
	}

	// Check no errors
	if len(capturedErrors) > 0 {
		t.Errorf("unexpected errors: %v", capturedErrors)
	}

	// Verify NDJSON format
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}
}

func TestLoggingTransportSkipPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	writer := &MemoryWriter{}
	opts := DefaultLoggingOptions()
	opts.SkipPaths = []string{"/health", "/metrics"}

	transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
	client := &http.Client{Transport: transport}

	// Request to /health should be skipped
	resp, err := client.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// Request to /metrics/cpu should be skipped (prefix match)
	resp, err = client.Get(server.URL + "/metrics/cpu")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// Request to /api should be logged
	resp, err = client.Get(server.URL + "/api")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if len(writer.Records) != 1 {
		t.Errorf("expected 1 record (skipping health and metrics), got %d", len(writer.Records))
	}

	if writer.Records[0].Request.Path != "/api" {
		t.Errorf("expected /api, got %s", writer.Records[0].Request.Path)
	}
}

func TestLoggingTransportAllowMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	writer := &MemoryWriter{}
	opts := DefaultLoggingOptions()
	opts.AllowMethods = []string{"POST", "PUT"}

	transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
	client := &http.Client{Transport: transport}

	// GET should be skipped
	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// POST should be logged
	resp, err = client.Post(server.URL+"/test", "text/plain", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if len(writer.Records) != 1 {
		t.Errorf("expected 1 record (POST only), got %d", len(writer.Records))
	}

	if writer.Records[0].Request.Method != RequestMethodPOST {
		t.Errorf("expected POST, got %s", writer.Records[0].Request.Method)
	}
}

func TestLoggingTransportSkipStatusCodes(t *testing.T) {
	statusCode := 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
	defer server.Close()

	writer := &MemoryWriter{}
	opts := DefaultLoggingOptions()
	opts.SkipStatusCodes = []int{404, 500}

	transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
	client := &http.Client{Transport: transport}

	// 200 should be logged
	statusCode = 200
	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// 404 should be skipped
	statusCode = 404
	resp, err = client.Get(server.URL + "/notfound")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// 500 should be skipped
	statusCode = 500
	resp, err = client.Get(server.URL + "/error")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if len(writer.Records) != 1 {
		t.Errorf("expected 1 record (200 only), got %d", len(writer.Records))
	}
}

func TestLoggingTransportRequestID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Run("extracts request ID from header", func(t *testing.T) {
		writer := &MemoryWriter{}
		opts := DefaultLoggingOptions()
		opts.RequestIDHeaders = []string{"X-Request-ID", "X-Correlation-ID"}

		transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
		client := &http.Client{Transport: transport}

		req, _ := http.NewRequest("GET", server.URL+"/test", nil)
		req.Header.Set("X-Request-ID", "my-custom-id-123")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		if len(writer.Records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(writer.Records))
		}

		if *writer.Records[0].Id != "my-custom-id-123" {
			t.Errorf("expected ID 'my-custom-id-123', got '%s'", *writer.Records[0].Id)
		}
	})

	t.Run("falls back to second header", func(t *testing.T) {
		writer := &MemoryWriter{}
		opts := DefaultLoggingOptions()
		opts.RequestIDHeaders = []string{"X-Request-ID", "X-Correlation-ID"}

		transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
		client := &http.Client{Transport: transport}

		req, _ := http.NewRequest("GET", server.URL+"/test", nil)
		req.Header.Set("X-Correlation-ID", "correlation-456")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		if *writer.Records[0].Id != "correlation-456" {
			t.Errorf("expected ID 'correlation-456', got '%s'", *writer.Records[0].Id)
		}
	})

	t.Run("generates UUID when no header found", func(t *testing.T) {
		writer := &MemoryWriter{}
		opts := DefaultLoggingOptions()
		opts.RequestIDHeaders = []string{"X-Request-ID"}

		transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
		client := &http.Client{Transport: transport}

		resp, err := client.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		// Should be a UUID (36 chars with dashes)
		if len(*writer.Records[0].Id) != 36 {
			t.Errorf("expected UUID (36 chars), got '%s' (%d chars)", *writer.Records[0].Id, len(*writer.Records[0].Id))
		}
	})
}

func TestLoggingTransportSampling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test with 0.0 sampling (0.0 means "not configured", logs all requests)
	t.Run("zero means log all", func(t *testing.T) {
		writer := &MemoryWriter{}
		opts := DefaultLoggingOptions()
		opts.SampleRate = 0.0 // Per docs: 0.0 means "not configured", logs all

		transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
		client := &http.Client{Transport: transport}

		for i := 0; i < 10; i++ {
			resp, err := client.Get(server.URL + "/test")
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()
		}

		if len(writer.Records) != 10 {
			t.Errorf("expected 10 records with 0.0 sampling (=log all), got %d", len(writer.Records))
		}
	})

	// Test with 100% sampling (log everything)
	t.Run("full sampling", func(t *testing.T) {
		writer := &MemoryWriter{}
		opts := DefaultLoggingOptions()
		opts.SampleRate = 1.0

		transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
		client := &http.Client{Transport: transport}

		for i := 0; i < 10; i++ {
			resp, err := client.Get(server.URL + "/test")
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()
		}

		if len(writer.Records) != 10 {
			t.Errorf("expected 10 records with 100%% sampling, got %d", len(writer.Records))
		}
	})

	// Test with 50% sampling (should log roughly half)
	t.Run("partial sampling", func(t *testing.T) {
		writer := &MemoryWriter{}
		opts := DefaultLoggingOptions()
		opts.SampleRate = 0.5

		transport := NewLoggingTransport(writer, WithLoggingOptions(opts))
		client := &http.Client{Transport: transport}

		numRequests := 1000
		for i := 0; i < numRequests; i++ {
			resp, err := client.Get(server.URL + "/test")
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()
		}

		// With 50% sampling over 1000 requests, expect roughly 400-600 records
		// Using wide margin to avoid flaky tests
		if len(writer.Records) < 300 || len(writer.Records) > 700 {
			t.Errorf("expected ~500 records with 50%% sampling, got %d", len(writer.Records))
		}
	})
}
