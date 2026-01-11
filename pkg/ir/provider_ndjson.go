package ir

import (
	"context"
	"fmt"
	"io"
	"os"
)

// NDJSONProvider provides symmetric read/write access to NDJSON-formatted IR records.
type NDJSONProvider struct {
	options *ProviderOptions
}

// NDJSON creates a new NDJSON provider with default options.
func NDJSON(opts ...ProviderOption) *NDJSONProvider {
	return &NDJSONProvider{
		options: ApplyProviderOptions(opts...),
	}
}

// NewWriter creates a writer for the given file path.
func (p *NDJSONProvider) NewWriter(ctx context.Context, path string) (IRWriter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	w := NewNDJSONWriter(f)
	w.closer = f
	return w, nil
}

// NewReader creates a reader for the given file path.
func (p *NDJSONProvider) NewReader(ctx context.Context, path string) (IRReader, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	r := NewNDJSONReader(f)
	r.closer = f
	return r, nil
}

// NewStreamWriter creates a writer that writes to the given io.Writer.
func (p *NDJSONProvider) NewStreamWriter(w io.Writer) IRWriter {
	return NewNDJSONWriter(w)
}

// NewStreamReader creates a reader that reads from the given io.Reader.
func (p *NDJSONProvider) NewStreamReader(r io.Reader) (IRReader, error) {
	return NewNDJSONReader(r), nil
}

// Ensure NDJSONProvider implements Provider and StreamProvider
var (
	_ Provider       = (*NDJSONProvider)(nil)
	_ StreamProvider = (*NDJSONProvider)(nil)
)
