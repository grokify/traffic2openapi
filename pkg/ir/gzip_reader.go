package ir

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// GzipNDJSONReader reads IR records from gzip-compressed NDJSON format.
type GzipNDJSONReader struct {
	gr     *gzip.Reader
	reader *NDJSONReader
	closer io.Closer
}

// NewGzipNDJSONReader creates a reader for streaming gzip-compressed NDJSON input.
func NewGzipNDJSONReader(r io.Reader) (*GzipNDJSONReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("creating gzip reader: %w", err)
	}

	return &GzipNDJSONReader{
		gr:     gr,
		reader: NewNDJSONReader(gr),
	}, nil
}

// NewGzipNDJSONFileReader creates a reader for streaming from a gzip-compressed file.
func NewGzipNDJSONFileReader(path string) (*GzipNDJSONReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	r, err := NewGzipNDJSONReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}

	r.closer = f
	return r, nil
}

// Read reads the next IR record.
// Returns io.EOF when no more records are available.
func (r *GzipNDJSONReader) Read() (*IRRecord, error) {
	return r.reader.Read()
}

// Close closes the gzip reader and underlying file if applicable.
func (r *GzipNDJSONReader) Close() error {
	if err := r.gr.Close(); err != nil {
		return fmt.Errorf("closing gzip reader: %w", err)
	}
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

// LineNumber returns the current line number (useful for error reporting).
func (r *GzipNDJSONReader) LineNumber() int {
	return r.reader.LineNumber()
}
