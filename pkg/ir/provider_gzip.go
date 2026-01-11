package ir

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
)

// GzipNDJSONProvider provides symmetric read/write access to gzip-compressed
// NDJSON-formatted IR records.
type GzipNDJSONProvider struct {
	options      *ProviderOptions
	gzipLevel    int
	hasGzipLevel bool
}

// GzipNDJSONOption configures a GzipNDJSONProvider.
type GzipNDJSONOption func(*GzipNDJSONProvider)

// WithGzipCompressionLevel sets the gzip compression level.
// Valid levels: gzip.NoCompression, gzip.BestSpeed, gzip.BestCompression,
// gzip.DefaultCompression, or 1-9.
func WithGzipCompressionLevel(level int) GzipNDJSONOption {
	return func(p *GzipNDJSONProvider) {
		p.gzipLevel = level
		p.hasGzipLevel = true
	}
}

// GzipNDJSON creates a new gzip-compressed NDJSON provider.
func GzipNDJSON(opts ...GzipNDJSONOption) *GzipNDJSONProvider {
	p := &GzipNDJSONProvider{
		options:   &ProviderOptions{},
		gzipLevel: gzip.DefaultCompression,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// NewWriter creates a writer for the given file path.
func (p *GzipNDJSONProvider) NewWriter(ctx context.Context, path string) (IRWriter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	var w *GzipNDJSONWriter
	if p.hasGzipLevel {
		w, err = NewGzipNDJSONWriterLevel(f, p.gzipLevel)
		if err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("creating gzip writer: %w", err)
		}
	} else {
		w = NewGzipNDJSONWriter(f)
	}
	w.closer = f
	return w, nil
}

// NewReader creates a reader for the given file path.
func (p *GzipNDJSONProvider) NewReader(ctx context.Context, path string) (IRReader, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	r, err := NewGzipNDJSONReader(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	r.closer = f
	return r, nil
}

// NewStreamWriter creates a writer that writes to the given io.Writer.
func (p *GzipNDJSONProvider) NewStreamWriter(w io.Writer) IRWriter {
	if p.hasGzipLevel {
		gw, err := NewGzipNDJSONWriterLevel(w, p.gzipLevel)
		if err != nil {
			// Fall back to default compression on error
			return NewGzipNDJSONWriter(w)
		}
		return gw
	}
	return NewGzipNDJSONWriter(w)
}

// NewStreamReader creates a reader that reads from the given io.Reader.
func (p *GzipNDJSONProvider) NewStreamReader(r io.Reader) (IRReader, error) {
	return NewGzipNDJSONReader(r)
}

// Ensure GzipNDJSONProvider implements Provider and StreamProvider
var (
	_ Provider       = (*GzipNDJSONProvider)(nil)
	_ StreamProvider = (*GzipNDJSONProvider)(nil)
)
