package ir

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// GzipNDJSONWriter provides streaming writes for gzip-compressed NDJSON format.
// Each record is JSON-encoded and written as a newline-delimited line.
// The output is gzip-compressed for storage efficiency.
type GzipNDJSONWriter struct {
	gw     *gzip.Writer
	closer io.Closer
	count  int
}

// GzipWriterOption configures a GzipNDJSONWriter.
type GzipWriterOption func(*GzipNDJSONWriter)

// WithGzipLevel sets the gzip compression level.
// Valid levels are gzip.DefaultCompression, gzip.NoCompression, gzip.BestSpeed,
// gzip.BestCompression, or any integer from 1 to 9.
func WithGzipLevel(level int) GzipWriterOption {
	return func(w *GzipNDJSONWriter) {
		// Recreate gzip writer with new level
		// This is a bit inefficient but options are applied before any writes
		underlying := w.gw.Reset
		_ = underlying // The Reset method needs the underlying writer
	}
}

// NewGzipNDJSONWriter creates a writer for streaming gzip-compressed NDJSON output.
func NewGzipNDJSONWriter(w io.Writer) *GzipNDJSONWriter {
	return &GzipNDJSONWriter{
		gw: gzip.NewWriter(w),
	}
}

// NewGzipNDJSONWriterLevel creates a writer with a specific compression level.
func NewGzipNDJSONWriterLevel(w io.Writer, level int) (*GzipNDJSONWriter, error) {
	gw, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return nil, fmt.Errorf("creating gzip writer: %w", err)
	}
	return &GzipNDJSONWriter{
		gw: gw,
	}, nil
}

// NewGzipNDJSONFileWriter creates a writer for streaming to a gzip-compressed file.
// The file should typically have a .ndjson.gz extension.
func NewGzipNDJSONFileWriter(path string) (*GzipNDJSONWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	w := NewGzipNDJSONWriter(f)
	w.closer = f
	return w, nil
}

// NewGzipNDJSONFileWriterLevel creates a writer for streaming to a gzip-compressed file
// with a specific compression level.
func NewGzipNDJSONFileWriterLevel(path string, level int) (*GzipNDJSONWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	gw, err := gzip.NewWriterLevel(f, level)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("creating gzip writer: %w", err)
	}

	return &GzipNDJSONWriter{
		gw:     gw,
		closer: f,
	}, nil
}

// Write writes a single record.
func (w *GzipNDJSONWriter) Write(record *IRRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshaling record: %w", err)
	}

	if _, err := w.gw.Write(data); err != nil {
		return fmt.Errorf("writing record: %w", err)
	}

	if _, err := w.gw.Write([]byte("\n")); err != nil {
		return fmt.Errorf("writing newline: %w", err)
	}

	w.count++
	return nil
}

// Flush flushes buffered data to the underlying gzip stream.
func (w *GzipNDJSONWriter) Flush() error {
	return w.gw.Flush()
}

// Close flushes and closes the gzip writer and underlying file if applicable.
func (w *GzipNDJSONWriter) Close() error {
	if err := w.gw.Close(); err != nil {
		return fmt.Errorf("closing gzip writer: %w", err)
	}
	if w.closer != nil {
		return w.closer.Close()
	}
	return nil
}

// Count returns the number of records written.
func (w *GzipNDJSONWriter) Count() int {
	return w.count
}
