package ir

import (
	"errors"
	"sync"
)

// ErrChannelClosed is returned when writing to a closed ChannelWriter.
var ErrChannelClosed = errors.New("channel writer is closed")

// ChannelWriter writes IR records to a channel for in-memory processing.
// Useful for testing, pipelines, and real-time processing scenarios.
type ChannelWriter struct {
	ch     chan *IRRecord
	closed bool
	mu     sync.RWMutex
}

// ChannelWriterOption configures a ChannelWriter.
type ChannelWriterOption func(*ChannelWriter)

// WithChannelBufferSize sets the channel buffer size.
func WithChannelBufferSize(size int) ChannelWriterOption {
	return func(w *ChannelWriter) {
		w.ch = make(chan *IRRecord, size)
	}
}

// NewChannelWriter creates a new ChannelWriter with default unbuffered channel.
func NewChannelWriter(opts ...ChannelWriterOption) *ChannelWriter {
	w := &ChannelWriter{
		ch: make(chan *IRRecord), // unbuffered by default
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// NewChannelWriterWithChan creates a ChannelWriter using an existing channel.
// This allows the caller to provide their own channel for more control.
func NewChannelWriterWithChan(ch chan *IRRecord) *ChannelWriter {
	return &ChannelWriter{
		ch: ch,
	}
}

// Write sends a record to the channel.
// Returns ErrChannelClosed if the writer has been closed.
// May block if the channel buffer is full (or unbuffered).
func (w *ChannelWriter) Write(record *IRRecord) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return ErrChannelClosed
	}

	w.ch <- record
	return nil
}

// Flush is a no-op for ChannelWriter since writes go directly to the channel.
func (w *ChannelWriter) Flush() error {
	return nil
}

// Close closes the underlying channel.
// After Close, Write will return ErrChannelClosed.
func (w *ChannelWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	close(w.ch)
	return nil
}

// Channel returns the underlying channel for reading records.
// Consumers should range over this channel to receive records.
func (w *ChannelWriter) Channel() <-chan *IRRecord {
	return w.ch
}
