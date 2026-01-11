package ir

import (
	"context"

	"github.com/grokify/omnistorage"
)

// StorageProvider provides symmetric read/write access to IR records
// using an omnistorage backend. It automatically handles compression
// based on file path extensions.
type StorageProvider struct {
	backend omnistorage.Backend
	options *ProviderOptions
}

// StorageProviderOption configures a StorageProvider.
type StorageProviderOption func(*StorageProvider)

// Storage creates a new storage provider with the given omnistorage backend.
func Storage(backend omnistorage.Backend, opts ...StorageProviderOption) *StorageProvider {
	p := &StorageProvider{
		backend: backend,
		options: &ProviderOptions{},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// NewWriter creates a writer for the given path.
// If the path ends with .gz, gzip compression is automatically applied.
// Supported path patterns:
//   - *.ndjson - plain NDJSON
//   - *.ndjson.gz - gzip-compressed NDJSON
func (p *StorageProvider) NewWriter(ctx context.Context, path string) (IRWriter, error) {
	return NewStorageWriter(ctx, p.backend, path)
}

// NewReader creates a reader for the given path.
// If the path ends with .gz, gzip decompression is automatically applied.
// Supported path patterns:
//   - *.ndjson - plain NDJSON
//   - *.ndjson.gz - gzip-compressed NDJSON
func (p *StorageProvider) NewReader(ctx context.Context, path string) (IRReader, error) {
	return NewStorageReader(ctx, p.backend, path)
}

// Backend returns the underlying omnistorage backend.
func (p *StorageProvider) Backend() omnistorage.Backend {
	return p.backend
}

// Ensure StorageProvider implements Provider
var _ Provider = (*StorageProvider)(nil)
