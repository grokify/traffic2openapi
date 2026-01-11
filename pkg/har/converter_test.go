package har

import (
	"testing"

	"github.com/chromedp/cdproto/har"
	"github.com/grokify/traffic2openapi/pkg/ir"
)

func TestConverterBasic(t *testing.T) {
	converter := NewConverter()

	entry := &har.Entry{
		StartedDateTime: "2024-12-30T10:00:00Z",
		Time:            45.5,
		Request: &har.Request{
			Method:      "GET",
			URL:         "https://api.example.com/users?limit=10",
			HTTPVersion: "HTTP/1.1",
			Headers: []*har.NameValuePair{
				{Name: "Accept", Value: "application/json"},
			},
			QueryString: []*har.NameValuePair{
				{Name: "limit", Value: "10"},
			},
		},
		Response: &har.Response{
			Status:     200,
			StatusText: "OK",
			Headers: []*har.NameValuePair{
				{Name: "Content-Type", Value: "application/json"},
			},
			Content: &har.Content{
				MimeType: "application/json",
				Text:     `{"users":[],"total":0}`,
			},
		},
	}

	record := converter.Convert(entry)

	if record == nil {
		t.Fatal("expected record, got nil")
	}

	if record.Request.Method != ir.RequestMethodGET {
		t.Errorf("expected GET, got %s", record.Request.Method)
	}

	if record.Request.Path != "/users" {
		t.Errorf("expected /users, got %s", record.Request.Path)
	}

	if record.Request.Host == nil || *record.Request.Host != "api.example.com" {
		t.Errorf("expected api.example.com, got %v", record.Request.Host)
	}

	if record.Request.Scheme != ir.RequestSchemeHttps {
		t.Errorf("expected https, got %s", record.Request.Scheme)
	}

	if record.Response.Status != 200 {
		t.Errorf("expected 200, got %d", record.Response.Status)
	}

	if record.DurationMs == nil || *record.DurationMs != 45.5 {
		t.Errorf("expected 45.5, got %v", record.DurationMs)
	}

	// Check query params
	if record.Request.Query == nil {
		t.Fatal("expected query params")
	}
	if record.Request.Query["limit"] != "10" {
		t.Errorf("expected limit=10, got %v", record.Request.Query["limit"])
	}

	// Check response body was parsed as JSON
	if record.Response.Body == nil {
		t.Fatal("expected response body")
	}
	bodyMap, ok := record.Response.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", record.Response.Body)
	}
	if bodyMap["total"] != float64(0) {
		t.Errorf("expected total=0, got %v", bodyMap["total"])
	}
}

func TestConverterHeaderFiltering(t *testing.T) {
	converter := NewConverter()

	entry := &har.Entry{
		Request: &har.Request{
			Method: "GET",
			URL:    "https://api.example.com/test",
			Headers: []*har.NameValuePair{
				{Name: "Accept", Value: "application/json"},
				{Name: "Authorization", Value: "Bearer secret"},
				{Name: "Cookie", Value: "session=abc123"},
				{Name: "X-Custom", Value: "value"},
			},
		},
		Response: &har.Response{
			Status: 200,
			Headers: []*har.NameValuePair{
				{Name: "Content-Type", Value: "application/json"},
				{Name: "Set-Cookie", Value: "session=xyz"},
			},
		},
	}

	record := converter.Convert(entry)

	// Authorization should be filtered
	if _, ok := record.Request.Headers["authorization"]; ok {
		t.Error("authorization header should be filtered")
	}

	// Cookie should be filtered (IncludeCookies is false by default)
	if _, ok := record.Request.Headers["cookie"]; ok {
		t.Error("cookie header should be filtered")
	}

	// Accept should be present
	if record.Request.Headers["accept"] != "application/json" {
		t.Error("accept header should be present")
	}

	// X-Custom should be present
	if record.Request.Headers["x-custom"] != "value" {
		t.Error("x-custom header should be present")
	}

	// Set-Cookie should be filtered
	if _, ok := record.Response.Headers["set-cookie"]; ok {
		t.Error("set-cookie header should be filtered")
	}
}

func TestConverterWithPostData(t *testing.T) {
	converter := NewConverter()

	entry := &har.Entry{
		Request: &har.Request{
			Method: "POST",
			URL:    "https://api.example.com/users",
			PostData: &har.PostData{
				MimeType: "application/json",
				Text:     `{"name":"Alice","email":"alice@example.com"}`,
			},
		},
		Response: &har.Response{
			Status: 201,
			Content: &har.Content{
				MimeType: "application/json",
				Text:     `{"id":"123","name":"Alice"}`,
			},
		},
	}

	record := converter.Convert(entry)

	if record.Request.Method != ir.RequestMethodPOST {
		t.Errorf("expected POST, got %s", record.Request.Method)
	}

	// Check request body was parsed as JSON
	if record.Request.Body == nil {
		t.Fatal("expected request body")
	}
	bodyMap, ok := record.Request.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", record.Request.Body)
	}
	if bodyMap["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %v", bodyMap["name"])
	}

	// Check response body
	if record.Response.Body == nil {
		t.Fatal("expected response body")
	}
	respMap, ok := record.Response.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", record.Response.Body)
	}
	if respMap["id"] != "123" {
		t.Errorf("expected id=123, got %v", respMap["id"])
	}
}

func TestConverterBase64Body(t *testing.T) {
	converter := NewConverter()

	// Base64 encoded: {"message":"hello"}
	entry := &har.Entry{
		Request: &har.Request{
			Method: "GET",
			URL:    "https://api.example.com/test",
		},
		Response: &har.Response{
			Status: 200,
			Content: &har.Content{
				MimeType: "application/json",
				Text:     "eyJtZXNzYWdlIjoiaGVsbG8ifQ==",
				Encoding: "base64",
			},
		},
	}

	record := converter.Convert(entry)

	if record.Response.Body == nil {
		t.Fatal("expected response body")
	}

	bodyMap, ok := record.Response.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", record.Response.Body)
	}
	if bodyMap["message"] != "hello" {
		t.Errorf("expected message=hello, got %v", bodyMap["message"])
	}
}

func TestConverterNilEntry(t *testing.T) {
	converter := NewConverter()

	record := converter.Convert(nil)
	if record != nil {
		t.Error("expected nil for nil entry")
	}

	record = converter.Convert(&har.Entry{})
	if record != nil {
		t.Error("expected nil for entry without request/response")
	}
}

func TestConvertBatch(t *testing.T) {
	converter := NewConverter()

	entries := []*har.Entry{
		{
			Request:  &har.Request{Method: "GET", URL: "https://api.example.com/a"},
			Response: &har.Response{Status: 200},
		},
		{
			Request:  &har.Request{Method: "POST", URL: "https://api.example.com/b"},
			Response: &har.Response{Status: 201},
		},
		nil, // Should be skipped
		{
			Request:  &har.Request{Method: "DELETE", URL: "https://api.example.com/c"},
			Response: &har.Response{Status: 204},
		},
	}

	records := converter.ConvertBatch(entries)

	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}

	if records[0].Request.Method != ir.RequestMethodGET {
		t.Errorf("expected GET, got %s", records[0].Request.Method)
	}
	if records[1].Request.Method != ir.RequestMethodPOST {
		t.Errorf("expected POST, got %s", records[1].Request.Method)
	}
	if records[2].Request.Method != ir.RequestMethodDELETE {
		t.Errorf("expected DELETE, got %s", records[2].Request.Method)
	}
}
