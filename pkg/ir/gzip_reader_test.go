package ir

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGzipNDJSONReader(t *testing.T) {
	// Create gzipped NDJSON data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)

	ndjson := `{"request":{"method":"GET","path":"/test1"},"response":{"status":200}}
{"request":{"method":"POST","path":"/test2"},"response":{"status":201}}`

	if _, err := gw.Write([]byte(ndjson)); err != nil {
		t.Fatalf("failed to write gzip data: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	// Read it back
	reader, err := NewGzipNDJSONReader(&buf)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}
	defer reader.Close()

	// Read first record
	record, err := reader.Read()
	if err != nil {
		t.Fatalf("read 1 failed: %v", err)
	}
	if record.Request.Path != "/test1" {
		t.Errorf("expected /test1, got %s", record.Request.Path)
	}

	// Read second record
	record, err = reader.Read()
	if err != nil {
		t.Fatalf("read 2 failed: %v", err)
	}
	if record.Request.Path != "/test2" {
		t.Errorf("expected /test2, got %s", record.Request.Path)
	}

	// Read should return EOF
	_, err = reader.Read()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestGzipNDJSONFileReader(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.ndjson.gz")

	// Create gzipped file
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	gw := gzip.NewWriter(f)
	ndjson := `{"request":{"method":"GET","path":"/api"},"response":{"status":200}}
{"request":{"method":"POST","path":"/api"},"response":{"status":201}}
{"request":{"method":"DELETE","path":"/api"},"response":{"status":204}}`

	if _, err := gw.Write([]byte(ndjson)); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	// Read it back
	reader, err := NewGzipNDJSONFileReader(path)
	if err != nil {
		t.Fatalf("failed to create file reader: %v", err)
	}
	defer reader.Close()

	count := 0
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		count++
	}

	if count != 3 {
		t.Errorf("expected 3 records, got %d", count)
	}
}

func TestGzipNDJSONReaderLineNumber(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)

	ndjson := `{"request":{"method":"GET","path":"/test1"},"response":{"status":200}}
{"request":{"method":"GET","path":"/test2"},"response":{"status":200}}`

	if _, err := gw.Write([]byte(ndjson)); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	reader, err := NewGzipNDJSONReader(&buf)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	_, _ = reader.Read()
	if reader.LineNumber() != 1 {
		t.Errorf("expected line 1, got %d", reader.LineNumber())
	}

	_, _ = reader.Read()
	if reader.LineNumber() != 2 {
		t.Errorf("expected line 2, got %d", reader.LineNumber())
	}
}

func TestGzipNDJSONReaderImplementsInterface(t *testing.T) {
	var _ IRReader = (*GzipNDJSONReader)(nil)
}
