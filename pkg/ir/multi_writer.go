package ir

import (
	"errors"
)

// MultiWriter fans out writes to multiple IRWriter destinations.
// Writes are performed sequentially to each writer in order.
type MultiWriter struct {
	writers []IRWriter
}

// NewMultiWriter creates a MultiWriter that writes to all provided writers.
// At least one writer must be provided.
func NewMultiWriter(writers ...IRWriter) (*MultiWriter, error) {
	if len(writers) == 0 {
		return nil, errors.New("at least one writer is required")
	}

	return &MultiWriter{
		writers: writers,
	}, nil
}

// Write writes a record to all underlying writers.
// If any writer returns an error, it continues to the next writer.
// Returns a combined error if any writes failed.
func (w *MultiWriter) Write(record *IRRecord) error {
	var errs []error
	for _, writer := range w.writers {
		if err := writer.Write(record); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Flush flushes all underlying writers.
// If any writer returns an error, it continues to the next writer.
// Returns a combined error if any flushes failed.
func (w *MultiWriter) Flush() error {
	var errs []error
	for _, writer := range w.writers {
		if err := writer.Flush(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Close closes all underlying writers.
// If any writer returns an error, it continues to the next writer.
// Returns a combined error if any closes failed.
func (w *MultiWriter) Close() error {
	var errs []error
	for _, writer := range w.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
