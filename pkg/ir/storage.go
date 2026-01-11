package ir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/grokify/omnistorage"
	"github.com/grokify/omnistorage/compress/gzip"
	"github.com/grokify/omnistorage/format/ndjson"
)

// StorageWriter writes IR records to an omnistorage backend.
// It automatically handles compression based on the file path extension.
type StorageWriter struct {
	ndjsonWriter *ndjson.Writer
	count        int
}

// NewStorageWriter creates an IR writer using an omnistorage backend.
// If the path ends with .gz, gzip compression is automatically applied.
// Supported path patterns:
//   - *.ndjson - plain NDJSON
//   - *.ndjson.gz - gzip-compressed NDJSON
func NewStorageWriter(ctx context.Context, backend omnistorage.Backend, path string) (*StorageWriter, error) {
	w, err := backend.NewWriter(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("creating writer: %w", err)
	}

	var writer io.WriteCloser = w

	// Apply gzip compression if path ends with .gz
	if strings.HasSuffix(strings.ToLower(path), ".gz") {
		gzWriter, err := gzip.NewWriter(w)
		if err != nil {
			_ = w.Close()
			return nil, fmt.Errorf("creating gzip writer: %w", err)
		}
		writer = gzWriter
	}

	return &StorageWriter{
		ndjsonWriter: ndjson.NewWriter(writer),
	}, nil
}

// Write writes a single IR record.
func (w *StorageWriter) Write(record *IRRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshaling record: %w", err)
	}

	if err := w.ndjsonWriter.Write(data); err != nil {
		return fmt.Errorf("writing record: %w", err)
	}

	w.count++
	return nil
}

// Flush flushes any buffered data.
func (w *StorageWriter) Flush() error {
	return w.ndjsonWriter.Flush()
}

// Close flushes and closes the writer.
func (w *StorageWriter) Close() error {
	return w.ndjsonWriter.Close()
}

// Count returns the number of records written.
func (w *StorageWriter) Count() int {
	return w.count
}

// StorageReader reads IR records from an omnistorage backend.
// It automatically handles decompression based on the file path extension.
type StorageReader struct {
	ndjsonReader *ndjson.Reader
	lineNum      int
}

// NewStorageReader creates an IR reader using an omnistorage backend.
// If the path ends with .gz, gzip decompression is automatically applied.
// Supported path patterns:
//   - *.ndjson - plain NDJSON
//   - *.ndjson.gz - gzip-compressed NDJSON
func NewStorageReader(ctx context.Context, backend omnistorage.Backend, path string) (*StorageReader, error) {
	r, err := backend.NewReader(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("creating reader: %w", err)
	}

	var reader io.ReadCloser = r

	// Apply gzip decompression if path ends with .gz
	if strings.HasSuffix(strings.ToLower(path), ".gz") {
		gzReader, err := gzip.NewReader(r)
		if err != nil {
			_ = r.Close()
			return nil, fmt.Errorf("creating gzip reader: %w", err)
		}
		reader = gzReader
	}

	return &StorageReader{
		ndjsonReader: ndjson.NewReader(reader),
	}, nil
}

// Read reads the next IR record.
// Returns io.EOF when no more records are available.
func (r *StorageReader) Read() (*IRRecord, error) {
	data, err := r.ndjsonReader.Read()
	if err != nil {
		return nil, err
	}

	r.lineNum++

	var record IRRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("line %d: %w", r.lineNum, err)
	}

	return &record, nil
}

// Close closes the reader.
func (r *StorageReader) Close() error {
	return r.ndjsonReader.Close()
}

// LineNumber returns the current line number (useful for error reporting).
func (r *StorageReader) LineNumber() int {
	return r.lineNum
}

// Ensure StorageWriter implements IRWriter
var _ IRWriter = (*StorageWriter)(nil)

// Ensure StorageReader implements IRReader
var _ IRReader = (*StorageReader)(nil)
