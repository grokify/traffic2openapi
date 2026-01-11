package ir

import (
	"context"
	"io"
)

// Provider defines the interface for IR record storage providers.
// Providers offer symmetric read/write access to IR records.
type Provider interface {
	// NewWriter creates a writer for the given path.
	// The path interpretation depends on the provider implementation.
	NewWriter(ctx context.Context, path string) (IRWriter, error)

	// NewReader creates a reader for the given path.
	// The path interpretation depends on the provider implementation.
	NewReader(ctx context.Context, path string) (IRReader, error)
}

// StreamProvider defines the interface for providers that work with io.Reader/Writer
// rather than paths. This is useful for in-memory or streaming scenarios.
type StreamProvider interface {
	// NewStreamWriter creates a writer that writes to the given io.Writer.
	NewStreamWriter(w io.Writer) IRWriter

	// NewStreamReader creates a reader that reads from the given io.Reader.
	// Returns an error if the reader cannot be initialized (e.g., invalid format).
	NewStreamReader(r io.Reader) (IRReader, error)
}

// ProviderOptions holds common configuration for providers.
type ProviderOptions struct {
	// BufferSize is the buffer size for I/O operations.
	// 0 means use the provider's default.
	BufferSize int
}

// ProviderOption configures a Provider.
type ProviderOption func(*ProviderOptions)

// WithBufferSize sets the buffer size for I/O operations.
func WithProviderBufferSize(size int) ProviderOption {
	return func(o *ProviderOptions) {
		o.BufferSize = size
	}
}

// ApplyProviderOptions applies options to ProviderOptions.
func ApplyProviderOptions(opts ...ProviderOption) *ProviderOptions {
	options := &ProviderOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}
