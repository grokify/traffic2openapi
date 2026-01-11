package ir

import (
	"io"
)

// ChannelReader reads IR records from a channel.
// Useful for consuming records from a ChannelWriter or other channel-based sources.
type ChannelReader struct {
	ch     <-chan *IRRecord
	closed bool
}

// NewChannelReader creates a reader that consumes from the given channel.
func NewChannelReader(ch <-chan *IRRecord) *ChannelReader {
	return &ChannelReader{
		ch: ch,
	}
}

// NewChannelReaderFromWriter creates a reader that consumes from a ChannelWriter.
// This enables pipelines like: LoggingTransport → ChannelWriter → ChannelReader → Engine
func NewChannelReaderFromWriter(w *ChannelWriter) *ChannelReader {
	return &ChannelReader{
		ch: w.Channel(),
	}
}

// Read reads the next IR record from the channel.
// Returns io.EOF when the channel is closed.
// Blocks if no record is available.
func (r *ChannelReader) Read() (*IRRecord, error) {
	if r.closed {
		return nil, io.EOF
	}

	record, ok := <-r.ch
	if !ok {
		r.closed = true
		return nil, io.EOF
	}

	return record, nil
}

// Close marks the reader as closed.
// Note: This does not close the underlying channel.
// The channel should be closed by the writer.
func (r *ChannelReader) Close() error {
	r.closed = true
	return nil
}
