// Package har provides an adapter for converting HAR (HTTP Archive) files to IR format.
//
// HAR is a standard format for recording HTTP transactions, supported by:
//   - Browser DevTools (Chrome, Firefox, Safari)
//   - Playwright and Puppeteer
//   - Charles Proxy, Fiddler, mitmproxy
//   - Postman
package har

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/har"
	"github.com/grokify/traffic2openapi/pkg/ir"
)

// Converter converts HAR entries to IR records.
type Converter struct {
	// IncludeHeaders controls whether to include HTTP headers in output.
	IncludeHeaders bool

	// FilterHeaders is a list of header names to exclude (case-insensitive).
	FilterHeaders []string

	// IncludeCookies controls whether to include cookies in headers.
	IncludeCookies bool
}

// NewConverter creates a new HAR to IR converter with default settings.
func NewConverter() *Converter {
	return &Converter{
		IncludeHeaders: true,
		IncludeCookies: false,
		FilterHeaders: []string{
			"authorization",
			"cookie",
			"set-cookie",
			"x-api-key",
			"x-auth-token",
			"x-csrf-token",
			"proxy-authorization",
		},
	}
}

// Convert converts a single HAR entry to an IR record.
func (c *Converter) Convert(entry *har.Entry) *ir.IRRecord {
	if entry == nil || entry.Request == nil || entry.Response == nil {
		return nil
	}

	record := &ir.IRRecord{
		Source: ptrSource(ir.IRRecordSourceHar),
		Request: ir.Request{
			Method: ir.RequestMethod(entry.Request.Method),
			Path:   "/", // Will be overwritten below
		},
		Response: ir.Response{
			Status: int(entry.Response.Status),
		},
	}

	// Parse timestamp
	if entry.StartedDateTime != "" {
		if t, err := time.Parse(time.RFC3339, entry.StartedDateTime); err == nil {
			record.Timestamp = ptrTime(t.UTC())
		} else if t, err := time.Parse("2006-01-02T15:04:05.000Z", entry.StartedDateTime); err == nil {
			record.Timestamp = ptrTime(t.UTC())
		}
	}

	// Parse URL
	if parsedURL, err := url.Parse(entry.Request.URL); err == nil {
		record.Request.Scheme = schemeFromString(parsedURL.Scheme)
		record.Request.Host = ptrString(parsedURL.Host)
		record.Request.Path = parsedURL.Path
		if record.Request.Path == "" {
			record.Request.Path = "/"
		}

		// Extract query parameters
		if len(parsedURL.Query()) > 0 {
			record.Request.Query = make(map[string]interface{})
			for k, v := range parsedURL.Query() {
				if len(v) > 0 {
					record.Request.Query[k] = v[0]
				}
			}
		}
	}

	// Also check HAR QueryString field (may have more accurate data)
	if len(entry.Request.QueryString) > 0 {
		if record.Request.Query == nil {
			record.Request.Query = make(map[string]interface{})
		}
		for _, nvp := range entry.Request.QueryString {
			record.Request.Query[nvp.Name] = nvp.Value
		}
	}

	// Convert request headers
	if c.IncludeHeaders && len(entry.Request.Headers) > 0 {
		headers := c.convertHeaders(entry.Request.Headers)
		if len(headers) > 0 {
			record.Request.Headers = headers
			if ct, ok := headers["content-type"]; ok {
				record.Request.ContentType = ptrString(ct)
			}
		}
	}

	// Convert request body
	if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
		record.Request.Body = parseBody(
			entry.Request.PostData.Text,
			entry.Request.PostData.MimeType,
			"",
		)
		if record.Request.ContentType == nil && entry.Request.PostData.MimeType != "" {
			record.Request.ContentType = ptrString(entry.Request.PostData.MimeType)
		}
	}

	// Convert response headers
	if c.IncludeHeaders && len(entry.Response.Headers) > 0 {
		headers := c.convertHeaders(entry.Response.Headers)
		if len(headers) > 0 {
			record.Response.Headers = headers
			if ct, ok := headers["content-type"]; ok {
				record.Response.ContentType = ptrString(ct)
			}
		}
	}

	// Convert response body
	if entry.Response.Content != nil && entry.Response.Content.Text != "" {
		record.Response.Body = parseBody(
			entry.Response.Content.Text,
			entry.Response.Content.MimeType,
			entry.Response.Content.Encoding,
		)
		if record.Response.ContentType == nil && entry.Response.Content.MimeType != "" {
			record.Response.ContentType = ptrString(entry.Response.Content.MimeType)
		}
	}

	// Set duration
	if entry.Time > 0 {
		record.DurationMs = ptrFloat64(entry.Time)
	}

	return record
}

// ConvertBatch converts multiple HAR entries to IR records.
func (c *Converter) ConvertBatch(entries []*har.Entry) []ir.IRRecord {
	records := make([]ir.IRRecord, 0, len(entries))
	for _, entry := range entries {
		if record := c.Convert(entry); record != nil {
			records = append(records, *record)
		}
	}
	return records
}

// ConvertHAR converts a complete HAR file to IR records.
func (c *Converter) ConvertHAR(h *har.HAR) []ir.IRRecord {
	if h == nil || h.Log == nil {
		return nil
	}
	return c.ConvertBatch(h.Log.Entries)
}

// convertHeaders converts HAR NameValuePair headers to a string map.
func (c *Converter) convertHeaders(headers []*har.NameValuePair) map[string]string {
	result := make(map[string]string)

	for _, h := range headers {
		name := strings.ToLower(h.Name)

		// Skip filtered headers
		if c.shouldFilterHeader(name) {
			continue
		}

		// Skip cookie headers if not including cookies
		if !c.IncludeCookies && (name == "cookie" || name == "set-cookie") {
			continue
		}

		result[name] = h.Value
	}

	return result
}

// shouldFilterHeader checks if a header should be filtered out.
func (c *Converter) shouldFilterHeader(name string) bool {
	name = strings.ToLower(name)
	for _, filter := range c.FilterHeaders {
		if strings.ToLower(filter) == name {
			return true
		}
	}
	return false
}

// parseBody parses a body string, handling JSON and base64 encoding.
func parseBody(text, mimeType, encoding string) interface{} {
	if text == "" {
		return nil
	}

	// Handle base64 encoding
	if encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return text // Return original if decode fails
		}
		text = string(decoded)
	}

	// Try to parse as JSON if mime type suggests it
	if strings.Contains(mimeType, "json") || strings.Contains(mimeType, "javascript") {
		var v interface{}
		if err := json.Unmarshal([]byte(text), &v); err == nil {
			return v
		}
	}

	// For text content, return as string
	if strings.Contains(mimeType, "text") ||
		strings.Contains(mimeType, "xml") ||
		strings.Contains(mimeType, "html") {
		return text
	}

	// Try JSON parsing for unknown types
	var v interface{}
	if err := json.Unmarshal([]byte(text), &v); err == nil {
		return v
	}

	return text
}

func ptrString(s string) *string {
	return &s
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func ptrSource(s ir.IRRecordSource) *ir.IRRecordSource {
	return &s
}

func schemeFromString(s string) ir.RequestScheme {
	switch strings.ToLower(s) {
	case "https":
		return ir.RequestSchemeHttps
	case "http":
		return ir.RequestSchemeHttp
	default:
		return ir.RequestScheme(s)
	}
}
