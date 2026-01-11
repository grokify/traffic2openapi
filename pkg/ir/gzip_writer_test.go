package ir

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGzipNDJSONWriterBasic(t *testing.T) {
	var buf bytes.Buffer
	w := NewGzipNDJSONWriter(&buf)

	// Write records
	for i := 0; i < 5; i++ {
		record := NewRecord(RequestMethodGET, "/test", 200)
		if err := w.Write(record); err != nil {
			t.Errorf("write failed: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	if w.Count() != 5 {
		t.Errorf("expected count 5, got %d", w.Count())
	}

	// Decompress and verify
	gr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read gzip data: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var record IRRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
		if record.Request.Path != "/test" {
			t.Errorf("line %d: expected /test, got %s", i, record.Request.Path)
		}
	}
}

func TestGzipNDJSONWriterLevel(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewGzipNDJSONWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	record := NewRecord(RequestMethodGET, "/test", 200)
	if err := w.Write(record); err != nil {
		t.Errorf("write failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// Verify it's valid gzip
	gr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	_, err = io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read gzip data: %v", err)
	}
}

func TestGzipNDJSONFileWriter(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.ndjson.gz")

	w, err := NewGzipNDJSONFileWriter(path)
	if err != nil {
		t.Fatalf("failed to create file writer: %v", err)
	}

	for i := 0; i < 3; i++ {
		record := NewRecord(RequestMethodPOST, "/users", 201)
		if err := w.Write(record); err != nil {
			t.Errorf("write failed: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// Read and verify
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read gzip data: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestGzipNDJSONWriterFlush(t *testing.T) {
	var buf bytes.Buffer
	w := NewGzipNDJSONWriter(&buf)

	record := NewRecord(RequestMethodGET, "/test", 200)
	if err := w.Write(record); err != nil {
		t.Errorf("write failed: %v", err)
	}

	// Flush should not error
	if err := w.Flush(); err != nil {
		t.Errorf("flush failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

func TestGzipNDJSONWriterImplementsInterface(t *testing.T) {
	var _ IRWriter = (*GzipNDJSONWriter)(nil)
}

func TestGzipNDJSONWriterInvalidLevel(t *testing.T) {
	var buf bytes.Buffer
	_, err := NewGzipNDJSONWriterLevel(&buf, 100) // Invalid level
	if err == nil {
		t.Error("expected error for invalid compression level")
	}
}
