package har

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/har"
	"github.com/grokify/traffic2openapi/pkg/ir"
)

// Reader reads HAR files and converts them to IR format.
type Reader struct {
	Converter *Converter
}

// NewReader creates a new HAR reader with default settings.
func NewReader() *Reader {
	return &Reader{
		Converter: NewConverter(),
	}
}

// ReadFile reads a HAR file and returns IR records.
func (r *Reader) ReadFile(path string) ([]ir.IRRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	return r.Read(f)
}

// Read reads HAR data from an io.Reader and returns IR records.
func (r *Reader) Read(reader io.Reader) ([]ir.IRRecord, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading data: %w", err)
	}

	h, err := Parse(data)
	if err != nil {
		return nil, err
	}

	return r.Converter.ConvertHAR(h), nil
}

// ReadDir reads all HAR files from a directory and returns IR records.
func (r *Reader) ReadDir(path string) ([]ir.IRRecord, error) {
	var allRecords []ir.IRRecord

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		if ext != ".har" {
			return nil
		}

		records, err := r.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", filePath, err)
		}

		allRecords = append(allRecords, records...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allRecords, nil
}

// Parse parses HAR JSON data into a HAR struct.
func Parse(data []byte) (*har.HAR, error) {
	// Handle UTF-8 BOM if present
	data = skipBOM(data)

	var h har.HAR
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("parsing HAR: %w", err)
	}

	if h.Log == nil {
		return nil, fmt.Errorf("invalid HAR: missing log field")
	}

	return &h, nil
}

// ParseFile parses a HAR file into a HAR struct.
func ParseFile(path string) (*har.HAR, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return Parse(data)
}

// skipBOM removes UTF-8 BOM if present at the start of data.
func skipBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

// EntryCount returns the number of entries in a HAR file without fully parsing.
func EntryCount(path string) (int, error) {
	h, err := ParseFile(path)
	if err != nil {
		return 0, err
	}

	if h.Log == nil {
		return 0, nil
	}

	return len(h.Log.Entries), nil
}

// FilterEntries filters HAR entries based on a predicate function.
func FilterEntries(h *har.HAR, predicate func(*har.Entry) bool) []*har.Entry {
	if h == nil || h.Log == nil {
		return nil
	}

	var filtered []*har.Entry
	for _, entry := range h.Log.Entries {
		if predicate(entry) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// FilterByMethod returns entries matching the given HTTP method.
func FilterByMethod(h *har.HAR, method string) []*har.Entry {
	method = strings.ToUpper(method)
	return FilterEntries(h, func(e *har.Entry) bool {
		return e.Request != nil && strings.ToUpper(e.Request.Method) == method
	})
}

// FilterByStatus returns entries matching the given status code.
func FilterByStatus(h *har.HAR, status int) []*har.Entry {
	return FilterEntries(h, func(e *har.Entry) bool {
		return e.Response != nil && int(e.Response.Status) == status
	})
}

// FilterByHost returns entries matching the given host.
func FilterByHost(h *har.HAR, host string) []*har.Entry {
	host = strings.ToLower(host)
	return FilterEntries(h, func(e *har.Entry) bool {
		if e.Request == nil || e.Request.URL == "" {
			return false
		}
		return strings.Contains(strings.ToLower(e.Request.URL), host)
	})
}

// FilterByContentType returns entries with responses matching the content type.
func FilterByContentType(h *har.HAR, contentType string) []*har.Entry {
	contentType = strings.ToLower(contentType)
	return FilterEntries(h, func(e *har.Entry) bool {
		if e.Response == nil || e.Response.Content == nil {
			return false
		}
		return strings.Contains(strings.ToLower(e.Response.Content.MimeType), contentType)
	})
}
