package ir

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// NDJSONReader reads IR records from newline-delimited JSON format.
type NDJSONReader struct {
	scanner *bufio.Scanner
	closer  io.Closer
	lineNum int
}

// NewNDJSONReader creates a reader for streaming NDJSON input.
func NewNDJSONReader(r io.Reader) *NDJSONReader {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line size

	return &NDJSONReader{
		scanner: scanner,
	}
}

// NewNDJSONFileReader creates a reader for streaming from a file.
func NewNDJSONFileReader(path string) (*NDJSONReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	r := NewNDJSONReader(f)
	r.closer = f
	return r, nil
}

// Read reads the next IR record.
// Returns io.EOF when no more records are available.
func (r *NDJSONReader) Read() (*IRRecord, error) {
	for r.scanner.Scan() {
		r.lineNum++
		line := strings.TrimSpace(r.scanner.Text())
		if line == "" {
			continue
		}

		var record IRRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("line %d: %w", r.lineNum, err)
		}
		return &record, nil
	}

	if err := r.scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning NDJSON: %w", err)
	}

	return nil, io.EOF
}

// Close closes the underlying reader if it implements io.Closer.
func (r *NDJSONReader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

// LineNumber returns the current line number (useful for error reporting).
func (r *NDJSONReader) LineNumber() int {
	return r.lineNum
}
