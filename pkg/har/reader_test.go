package har

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	chromedphar "github.com/chromedp/cdproto/har"
	"github.com/grokify/traffic2openapi/pkg/ir"
)

func TestReadFile(t *testing.T) {
	// Find the examples directory
	examplesPath := findExamplesDir()
	if examplesPath == "" {
		t.Skip("examples directory not found")
	}

	harFile := filepath.Join(examplesPath, "har", "sample.har")
	if _, err := os.Stat(harFile); os.IsNotExist(err) {
		t.Skipf("sample HAR file not found: %s", harFile)
	}

	reader := NewReader()
	records, err := reader.ReadFile(harFile)
	if err != nil {
		t.Fatalf("failed to read HAR file: %v", err)
	}

	if len(records) != 4 {
		t.Errorf("expected 4 records, got %d", len(records))
	}

	// Check first record (GET /users)
	if records[0].Request.Method != ir.RequestMethodGET {
		t.Errorf("expected GET, got %s", records[0].Request.Method)
	}
	if records[0].Request.Path != "/users" {
		t.Errorf("expected /users, got %s", records[0].Request.Path)
	}
	if records[0].Response.Status != 200 {
		t.Errorf("expected 200, got %d", records[0].Response.Status)
	}

	// Check third record (POST /users)
	if records[2].Request.Method != ir.RequestMethodPOST {
		t.Errorf("expected POST, got %s", records[2].Request.Method)
	}
	if records[2].Response.Status != 201 {
		t.Errorf("expected 201, got %d", records[2].Response.Status)
	}

	// Verify request body was parsed
	if records[2].Request.Body == nil {
		t.Error("expected request body for POST")
	}
}

func TestReadFileHeaderFiltering(t *testing.T) {
	examplesPath := findExamplesDir()
	if examplesPath == "" {
		t.Skip("examples directory not found")
	}

	harFile := filepath.Join(examplesPath, "har", "sample.har")
	if _, err := os.Stat(harFile); os.IsNotExist(err) {
		t.Skipf("sample HAR file not found: %s", harFile)
	}

	reader := NewReader()
	records, err := reader.ReadFile(harFile)
	if err != nil {
		t.Fatalf("failed to read HAR file: %v", err)
	}

	// Authorization header should be filtered
	for _, r := range records {
		if r.Request.Headers != nil {
			if _, ok := r.Request.Headers["authorization"]; ok {
				t.Error("authorization header should be filtered")
			}
		}
	}
}

func TestParse(t *testing.T) {
	harJSON := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "Test", "version": "1.0"},
			"entries": [
				{
					"request": {
						"method": "GET",
						"url": "https://example.com/test",
						"httpVersion": "HTTP/1.1",
						"headers": [],
						"queryString": [],
						"cookies": [],
						"headersSize": 0,
						"bodySize": 0
					},
					"response": {
						"status": 200,
						"statusText": "OK",
						"httpVersion": "HTTP/1.1",
						"headers": [],
						"cookies": [],
						"content": {"size": 0, "mimeType": ""},
						"redirectURL": "",
						"headersSize": 0,
						"bodySize": 0
					},
					"cache": {},
					"timings": {"send": 0, "wait": 0, "receive": 0}
				}
			]
		}
	}`

	h, err := Parse([]byte(harJSON))
	if err != nil {
		t.Fatalf("failed to parse HAR: %v", err)
	}

	if h.Log == nil {
		t.Fatal("expected log")
	}

	if len(h.Log.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(h.Log.Entries))
	}

	if h.Log.Entries[0].Request.Method != "GET" {
		t.Errorf("expected GET, got %s", h.Log.Entries[0].Request.Method)
	}
}

func TestParseInvalid(t *testing.T) {
	// Invalid JSON
	_, err := Parse([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	// Missing log field
	_, err = Parse([]byte(`{}`))
	if err == nil {
		t.Error("expected error for missing log")
	}
}

func TestSkipBOM(t *testing.T) {
	// UTF-8 BOM + valid JSON
	bomJSON := []byte{0xEF, 0xBB, 0xBF}
	bomJSON = append(bomJSON, []byte(`{"log":{"version":"1.2","entries":[]}}`)...)

	h, err := Parse(bomJSON)
	if err != nil {
		t.Fatalf("failed to parse HAR with BOM: %v", err)
	}

	if h.Log == nil {
		t.Error("expected log")
	}
}

func TestRead(t *testing.T) {
	harJSON := `{
		"log": {
			"version": "1.2",
			"entries": [
				{
					"request": {"method": "GET", "url": "https://example.com/a", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0},
					"response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0},
					"cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}
				}
			]
		}
	}`

	reader := NewReader()
	records, err := reader.Read(strings.NewReader(harJSON))
	if err != nil {
		t.Fatalf("failed to read HAR: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
}

func TestFilterByMethod(t *testing.T) {
	h, _ := Parse([]byte(`{
		"log": {
			"version": "1.2",
			"entries": [
				{"request": {"method": "GET", "url": "https://example.com/a", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}},
				{"request": {"method": "POST", "url": "https://example.com/b", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 201, "statusText": "Created", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}},
				{"request": {"method": "GET", "url": "https://example.com/c", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}}
			]
		}
	}`))

	getEntries := FilterByMethod(h, "GET")
	if len(getEntries) != 2 {
		t.Errorf("expected 2 GET entries, got %d", len(getEntries))
	}

	postEntries := FilterByMethod(h, "POST")
	if len(postEntries) != 1 {
		t.Errorf("expected 1 POST entry, got %d", len(postEntries))
	}
}

func TestFilterByHost(t *testing.T) {
	h, _ := Parse([]byte(`{
		"log": {
			"version": "1.2",
			"entries": [
				{"request": {"method": "GET", "url": "https://api.example.com/a", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}},
				{"request": {"method": "GET", "url": "https://cdn.example.com/b", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}},
				{"request": {"method": "GET", "url": "https://api.example.com/c", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}}
			]
		}
	}`))

	apiEntries := FilterByHost(h, "api.example.com")
	if len(apiEntries) != 2 {
		t.Errorf("expected 2 api.example.com entries, got %d", len(apiEntries))
	}

	cdnEntries := FilterByHost(h, "cdn")
	if len(cdnEntries) != 1 {
		t.Errorf("expected 1 cdn entry, got %d", len(cdnEntries))
	}
}

// findExamplesDir locates the examples directory relative to the test file.
func findExamplesDir() string {
	// Try relative paths from test location
	paths := []string{
		"../../../examples",
		"../../../../examples",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

func TestReadFileNotFound(t *testing.T) {
	reader := NewReader()
	_, err := reader.ReadFile("/nonexistent/path/to/file.har")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two HAR files
	harContent := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "test", "version": "1.0"},
			"entries": [
				{
					"request": {"method": "GET", "url": "https://example.com/test", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0},
					"response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0},
					"cache": {},
					"timings": {"send": 0, "wait": 0, "receive": 0}
				}
			]
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "test1.har"), []byte(harContent), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test2.har"), []byte(harContent), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	// Write a non-HAR file that should be skipped
	if err := os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte("not a har"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	reader := NewReader()
	records, err := reader.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	harContent := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "test", "version": "1.0"},
			"entries": []
		}
	}`

	harPath := filepath.Join(tmpDir, "test.har")
	if err := os.WriteFile(harPath, []byte(harContent), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	h, err := ParseFile(harPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if h.Log == nil {
		t.Error("Log should not be nil")
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/to/file.har")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestEntryCount(t *testing.T) {
	tmpDir := t.TempDir()
	harContent := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "test", "version": "1.0"},
			"entries": [
				{
					"request": {"method": "GET", "url": "https://example.com/test", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0},
					"response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0},
					"cache": {},
					"timings": {"send": 0, "wait": 0, "receive": 0}
				},
				{
					"request": {"method": "POST", "url": "https://example.com/test", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0},
					"response": {"status": 201, "statusText": "Created", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0},
					"cache": {},
					"timings": {"send": 0, "wait": 0, "receive": 0}
				}
			]
		}
	}`

	harPath := filepath.Join(tmpDir, "test.har")
	if err := os.WriteFile(harPath, []byte(harContent), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	count, err := EntryCount(harPath)
	if err != nil {
		t.Fatalf("EntryCount failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want %d", count, 2)
	}
}

func TestFilterByStatus(t *testing.T) {
	h, _ := Parse([]byte(`{
		"log": {
			"version": "1.2",
			"entries": [
				{"request": {"method": "GET", "url": "https://example.com/a", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}},
				{"request": {"method": "GET", "url": "https://example.com/b", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 404, "statusText": "Not Found", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}},
				{"request": {"method": "GET", "url": "https://example.com/c", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": ""}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}}
			]
		}
	}`))

	entries200 := FilterByStatus(h, 200)
	if len(entries200) != 2 {
		t.Errorf("expected 2 200 entries, got %d", len(entries200))
	}

	entries404 := FilterByStatus(h, 404)
	if len(entries404) != 1 {
		t.Errorf("expected 1 404 entry, got %d", len(entries404))
	}
}

func TestFilterByContentType(t *testing.T) {
	h, _ := Parse([]byte(`{
		"log": {
			"version": "1.2",
			"entries": [
				{"request": {"method": "GET", "url": "https://example.com/a", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": "application/json"}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}},
				{"request": {"method": "GET", "url": "https://example.com/b", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": "text/html"}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}},
				{"request": {"method": "GET", "url": "https://example.com/c", "httpVersion": "HTTP/1.1", "headers": [], "queryString": [], "cookies": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "headers": [], "cookies": [], "content": {"size": 0, "mimeType": "application/json; charset=utf-8"}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"send": 0, "wait": 0, "receive": 0}}
			]
		}
	}`))

	jsonEntries := FilterByContentType(h, "application/json")
	if len(jsonEntries) != 2 {
		t.Errorf("expected 2 json entries, got %d", len(jsonEntries))
	}

	htmlEntries := FilterByContentType(h, "text/html")
	if len(htmlEntries) != 1 {
		t.Errorf("expected 1 html entry, got %d", len(htmlEntries))
	}
}

func TestFilterEntriesNil(t *testing.T) {
	var nilHAR *chromedphar.HAR = nil

	// Test nil HAR
	filtered := FilterEntries(nilHAR, func(e *chromedphar.Entry) bool { return true })
	if filtered != nil {
		t.Error("expected nil for nil HAR")
	}

	// Test nil Log
	h, _ := Parse([]byte(`{"log":{"version":"1.2","entries":[]}}`))
	h.Log = nil
	filtered = FilterEntries(h, func(e *chromedphar.Entry) bool { return true })
	if filtered != nil {
		t.Error("expected nil for nil Log")
	}
}
