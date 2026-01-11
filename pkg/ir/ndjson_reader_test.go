package ir

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNDJSONReader(t *testing.T) {
	ndjson := `{"request":{"method":"GET","path":"/test1"},"response":{"status":200}}
{"request":{"method":"POST","path":"/test2"},"response":{"status":201}}
{"request":{"method":"DELETE","path":"/test3"},"response":{"status":204}}`

	reader := NewNDJSONReader(strings.NewReader(ndjson))

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

	// Read third record
	record, err = reader.Read()
	if err != nil {
		t.Fatalf("read 3 failed: %v", err)
	}
	if record.Request.Path != "/test3" {
		t.Errorf("expected /test3, got %s", record.Request.Path)
	}

	// Read should return EOF
	_, err = reader.Read()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}

	if err := reader.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestNDJSONReaderSkipsEmptyLines(t *testing.T) {
	ndjson := `{"request":{"method":"GET","path":"/test1"},"response":{"status":200}}

{"request":{"method":"GET","path":"/test2"},"response":{"status":200}}

{"request":{"method":"GET","path":"/test3"},"response":{"status":200}}`

	reader := NewNDJSONReader(strings.NewReader(ndjson))

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

func TestNDJSONFileReader(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.ndjson")

	content := `{"request":{"method":"GET","path":"/api"},"response":{"status":200}}
{"request":{"method":"POST","path":"/api"},"response":{"status":201}}`

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	reader, err := NewNDJSONFileReader(path)
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

	if count != 2 {
		t.Errorf("expected 2 records, got %d", count)
	}
}

func TestNDJSONReaderLineNumber(t *testing.T) {
	ndjson := `{"request":{"method":"GET","path":"/test1"},"response":{"status":200}}
{"request":{"method":"GET","path":"/test2"},"response":{"status":200}}`

	reader := NewNDJSONReader(strings.NewReader(ndjson))

	if reader.LineNumber() != 0 {
		t.Errorf("expected line 0 initially, got %d", reader.LineNumber())
	}

	_, _ = reader.Read()
	if reader.LineNumber() != 1 {
		t.Errorf("expected line 1 after first read, got %d", reader.LineNumber())
	}

	_, _ = reader.Read()
	if reader.LineNumber() != 2 {
		t.Errorf("expected line 2 after second read, got %d", reader.LineNumber())
	}
}

func TestNDJSONReaderInvalidJSON(t *testing.T) {
	ndjson := `{"request":{"method":"GET","path":"/test1"},"response":{"status":200}}
{invalid json}
{"request":{"method":"GET","path":"/test3"},"response":{"status":200}}`

	reader := NewNDJSONReader(strings.NewReader(ndjson))

	// First read should succeed
	_, err := reader.Read()
	if err != nil {
		t.Fatalf("read 1 failed: %v", err)
	}

	// Second read should fail with line number
	_, err = reader.Read()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected error to mention line 2, got: %v", err)
	}
}

func TestNDJSONReaderImplementsInterface(t *testing.T) {
	var _ IRReader = (*NDJSONReader)(nil)
}
