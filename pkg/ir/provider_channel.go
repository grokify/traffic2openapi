package ir

import (
	"context"
	"io"
)

// ChannelProvider provides symmetric read/write access to IR records
// using Go channels. This is useful for in-memory pipelines, testing,
// and real-time processing scenarios.
//
// Unlike file-based providers, ChannelProvider connects writers and readers
// through a shared channel rather than via file paths.
type ChannelProvider struct {
	bufferSize int
	channel    chan *IRRecord
}

// ChannelProviderOption configures a ChannelProvider.
type ChannelProviderOption func(*ChannelProvider)

// WithChannelProviderBufferSize sets the channel buffer size.
// Default is 0 (unbuffered).
func WithChannelProviderBufferSize(size int) ChannelProviderOption {
	return func(p *ChannelProvider) {
		p.bufferSize = size
	}
}

// WithExistingChannel uses an existing channel instead of creating a new one.
// This allows connecting to external channel sources.
func WithExistingChannel(ch chan *IRRecord) ChannelProviderOption {
	return func(p *ChannelProvider) {
		p.channel = ch
	}
}

// Channel creates a new channel provider with the given options.
func Channel(opts ...ChannelProviderOption) *ChannelProvider {
	p := &ChannelProvider{
		bufferSize: 0,
	}
	for _, opt := range opts {
		opt(p)
	}

	// Create channel if not provided
	if p.channel == nil {
		p.channel = make(chan *IRRecord, p.bufferSize)
	}

	return p
}

// NewWriter creates a writer that sends records to the channel.
// The path parameter is ignored for channel providers.
func (p *ChannelProvider) NewWriter(ctx context.Context, _ string) (IRWriter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return NewChannelWriterWithChan(p.channel), nil
}

// NewReader creates a reader that receives records from the channel.
// The path parameter is ignored for channel providers.
func (p *ChannelProvider) NewReader(ctx context.Context, _ string) (IRReader, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return NewChannelReader(p.channel), nil
}

// NewStreamWriter creates a writer that sends records to the channel.
// The io.Writer parameter is ignored for channel providers.
func (p *ChannelProvider) NewStreamWriter(_ io.Writer) IRWriter {
	return NewChannelWriterWithChan(p.channel)
}

// NewStreamReader creates a reader that receives records from the channel.
// The io.Reader parameter is ignored for channel providers.
func (p *ChannelProvider) NewStreamReader(_ io.Reader) (IRReader, error) {
	return NewChannelReader(p.channel), nil
}

// Chan returns the underlying channel.
// This allows direct access for advanced use cases.
func (p *ChannelProvider) Chan() chan *IRRecord {
	return p.channel
}

// Ensure ChannelProvider implements Provider and StreamProvider
var (
	_ Provider       = (*ChannelProvider)(nil)
	_ StreamProvider = (*ChannelProvider)(nil)
)
