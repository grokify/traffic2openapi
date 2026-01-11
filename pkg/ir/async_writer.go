package ir

import (
	"sync"
)

// ErrorHandler is a callback function for handling async write errors.
type ErrorHandler func(err error)

// AsyncWriterOption configures an AsyncNDJSONWriter.
type AsyncWriterOption func(*AsyncNDJSONWriter)

// WithBufferSize sets the channel buffer size for async writes.
func WithBufferSize(size int) AsyncWriterOption {
	return func(w *AsyncNDJSONWriter) {
		w.bufferSize = size
	}
}

// WithErrorHandler sets the error handler callback.
func WithErrorHandler(handler ErrorHandler) AsyncWriterOption {
	return func(w *AsyncNDJSONWriter) {
		w.errorHandler = handler
	}
}

// AsyncNDJSONWriter provides async streaming writes for NDJSON format.
// Records are buffered in a channel and written by a background goroutine.
// Errors are delivered via an error handler callback.
type AsyncNDJSONWriter struct {
	writer       *NDJSONWriter
	ch           chan *IRRecord
	flushCh      chan chan error // Channel for flush requests
	done         chan struct{}
	wg           sync.WaitGroup
	errorHandler ErrorHandler
	bufferSize   int
	closed       bool
	mu           sync.Mutex
}

// NewAsyncNDJSONWriter creates an async writer wrapping an existing NDJSONWriter.
func NewAsyncNDJSONWriter(writer *NDJSONWriter, opts ...AsyncWriterOption) *AsyncNDJSONWriter {
	w := &AsyncNDJSONWriter{
		writer:     writer,
		bufferSize: 100, // default buffer size
		errorHandler: func(err error) {
			// Default: silently ignore errors
			// Users should provide their own handler
		},
	}

	for _, opt := range opts {
		opt(w)
	}

	w.ch = make(chan *IRRecord, w.bufferSize)
	w.flushCh = make(chan chan error)
	w.done = make(chan struct{})

	w.wg.Add(1)
	go w.writeLoop()

	return w
}

// NewAsyncNDJSONFileWriter creates an async writer for streaming to a file.
func NewAsyncNDJSONFileWriter(path string, opts ...AsyncWriterOption) (*AsyncNDJSONWriter, error) {
	writer, err := NewNDJSONFileWriter(path)
	if err != nil {
		return nil, err
	}
	return NewAsyncNDJSONWriter(writer, opts...), nil
}

// writeLoop runs in a goroutine and processes records from the channel.
func (w *AsyncNDJSONWriter) writeLoop() {
	defer w.wg.Done()

	for {
		select {
		case record, ok := <-w.ch:
			if !ok {
				return
			}
			if err := w.writer.Write(record); err != nil {
				w.errorHandler(err)
			}
		case respCh := <-w.flushCh:
			// Drain pending records before flushing
			w.drainPending()
			respCh <- w.writer.Flush()
		case <-w.done:
			// Drain remaining records
			w.drainPending()
			return
		}
	}
}

// drainPending drains all pending records from the channel.
func (w *AsyncNDJSONWriter) drainPending() {
	for {
		select {
		case record, ok := <-w.ch:
			if !ok {
				return // Channel closed
			}
			if err := w.writer.Write(record); err != nil {
				w.errorHandler(err)
			}
		default:
			return
		}
	}
}

// Write queues a record for async writing.
// Returns nil immediately; errors are delivered via the error handler.
func (w *AsyncNDJSONWriter) Write(record *IRRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil // Silently ignore writes after close
	}

	w.ch <- record
	return nil
}

// Flush waits for all pending writes to complete and flushes the underlying writer.
func (w *AsyncNDJSONWriter) Flush() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	respCh := make(chan error, 1)
	w.flushCh <- respCh
	return <-respCh
}

// Close signals the writer to stop, waits for pending writes to complete,
// and closes the underlying writer.
func (w *AsyncNDJSONWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.mu.Unlock()

	close(w.ch)   // Signal no more records
	close(w.done) // Signal shutdown
	w.wg.Wait()   // Wait for writeLoop to finish

	return w.writer.Close()
}

// Count returns the number of records written.
func (w *AsyncNDJSONWriter) Count() int {
	return w.writer.Count()
}
