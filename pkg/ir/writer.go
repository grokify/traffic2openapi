package ir

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IRWriter is the interface for writing IR records to any destination.
// Implementations may be synchronous or asynchronous internally.
// For async implementations, errors are delivered via an error handler
// callback rather than returned from Write().
type IRWriter interface {
	// Write writes a single IR record.
	// For sync implementations, returns any write error.
	// For async implementations, returns nil and delivers errors via callback.
	Write(record *IRRecord) error

	// Flush flushes any buffered data to the underlying destination.
	// For async implementations, blocks until all pending writes complete.
	Flush() error

	// Close flushes any buffered data and releases resources.
	// For async implementations, blocks until all pending writes complete.
	Close() error
}

// WriteFile writes IR records to a file.
// Format is determined by file extension:
// - .ndjson: newline-delimited JSON
// - .json: batch format
func WriteFile(path string, records []IRRecord) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".ndjson":
		return WriteNDJSON(f, records)
	case ".json":
		return WriteBatch(f, records)
	default:
		return WriteBatch(f, records) // Default to batch
	}
}

// WriteBatch writes records in batch format.
func WriteBatch(w io.Writer, records []IRRecord) error {
	batch := NewBatch(records)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(batch); err != nil {
		return fmt.Errorf("encoding batch: %w", err)
	}

	return nil
}

// WriteNDJSON writes records in newline-delimited JSON format.
func WriteNDJSON(w io.Writer, records []IRRecord) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for i, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("record %d: %w", i, err)
		}

		if _, err := bw.Write(data); err != nil {
			return fmt.Errorf("writing record %d: %w", i, err)
		}

		if _, err := bw.WriteString("\n"); err != nil {
			return fmt.Errorf("writing newline: %w", err)
		}
	}

	return nil
}

// NDJSONWriter provides streaming writes for NDJSON format.
type NDJSONWriter struct {
	w      *bufio.Writer
	closer io.Closer
	count  int
}

// NewNDJSONWriter creates a writer for streaming NDJSON output.
func NewNDJSONWriter(w io.Writer) *NDJSONWriter {
	bw := bufio.NewWriter(w)
	return &NDJSONWriter{
		w: bw,
	}
}

// NewNDJSONFileWriter creates a writer for streaming to a file.
func NewNDJSONFileWriter(path string) (*NDJSONWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	w := NewNDJSONWriter(f)
	w.closer = f
	return w, nil
}

// Write writes a single record.
func (w *NDJSONWriter) Write(record *IRRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshaling record: %w", err)
	}

	if _, err := w.w.Write(data); err != nil {
		return fmt.Errorf("writing record: %w", err)
	}

	if _, err := w.w.WriteString("\n"); err != nil {
		return fmt.Errorf("writing newline: %w", err)
	}

	w.count++
	return nil
}

// Flush flushes buffered data.
func (w *NDJSONWriter) Flush() error {
	return w.w.Flush()
}

// Close flushes and closes the underlying writer if it implements io.Closer.
func (w *NDJSONWriter) Close() error {
	if err := w.w.Flush(); err != nil {
		return err
	}
	if w.closer != nil {
		return w.closer.Close()
	}
	return nil
}

// Count returns the number of records written.
func (w *NDJSONWriter) Count() int {
	return w.count
}
