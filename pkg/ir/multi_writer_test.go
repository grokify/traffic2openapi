package ir

import (
	"errors"
	"sync"
	"testing"
)

// testWriter is a simple writer for testing that tracks operations.
type testWriter struct {
	records  []*IRRecord
	flushed  int
	closed   bool
	writeErr error
	flushErr error
	closeErr error
	mu       sync.Mutex
}

func (w *testWriter) Write(record *IRRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.writeErr != nil {
		return w.writeErr
	}
	w.records = append(w.records, record)
	return nil
}

func (w *testWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.flushErr != nil {
		return w.flushErr
	}
	w.flushed++
	return nil
}

func (w *testWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closeErr != nil {
		return w.closeErr
	}
	w.closed = true
	return nil
}

func TestMultiWriterBasic(t *testing.T) {
	w1 := &testWriter{}
	w2 := &testWriter{}

	multi, err := NewMultiWriter(w1, w2)
	if err != nil {
		t.Fatalf("failed to create MultiWriter: %v", err)
	}

	record := NewRecord(RequestMethodGET, "/test", 200)
	if err := multi.Write(record); err != nil {
		t.Errorf("write failed: %v", err)
	}

	// Both writers should have the record
	if len(w1.records) != 1 {
		t.Errorf("w1: expected 1 record, got %d", len(w1.records))
	}
	if len(w2.records) != 1 {
		t.Errorf("w2: expected 1 record, got %d", len(w2.records))
	}

	// Flush both
	if err := multi.Flush(); err != nil {
		t.Errorf("flush failed: %v", err)
	}
	if w1.flushed != 1 {
		t.Errorf("w1: expected 1 flush, got %d", w1.flushed)
	}
	if w2.flushed != 1 {
		t.Errorf("w2: expected 1 flush, got %d", w2.flushed)
	}

	// Close both
	if err := multi.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}
	if !w1.closed {
		t.Error("w1 should be closed")
	}
	if !w2.closed {
		t.Error("w2 should be closed")
	}
}

func TestMultiWriterNoWriters(t *testing.T) {
	_, err := NewMultiWriter()
	if err == nil {
		t.Error("expected error when creating MultiWriter with no writers")
	}
}

func TestMultiWriterPartialError(t *testing.T) {
	w1 := &testWriter{}
	w2 := &testWriter{writeErr: errors.New("write error")}
	w3 := &testWriter{}

	multi, err := NewMultiWriter(w1, w2, w3)
	if err != nil {
		t.Fatalf("failed to create MultiWriter: %v", err)
	}

	record := NewRecord(RequestMethodGET, "/test", 200)
	err = multi.Write(record)

	// Should get an error
	if err == nil {
		t.Error("expected error from failing writer")
	}

	// But w1 and w3 should still have the record
	if len(w1.records) != 1 {
		t.Errorf("w1: expected 1 record, got %d", len(w1.records))
	}
	if len(w3.records) != 1 {
		t.Errorf("w3: expected 1 record, got %d", len(w3.records))
	}
}

func TestMultiWriterFlushError(t *testing.T) {
	w1 := &testWriter{}
	w2 := &testWriter{flushErr: errors.New("flush error")}

	multi, err := NewMultiWriter(w1, w2)
	if err != nil {
		t.Fatalf("failed to create MultiWriter: %v", err)
	}

	err = multi.Flush()
	if err == nil {
		t.Error("expected error from failing flush")
	}

	// w1 should still have been flushed
	if w1.flushed != 1 {
		t.Errorf("w1: expected 1 flush, got %d", w1.flushed)
	}
}

func TestMultiWriterCloseError(t *testing.T) {
	w1 := &testWriter{}
	w2 := &testWriter{closeErr: errors.New("close error")}

	multi, err := NewMultiWriter(w1, w2)
	if err != nil {
		t.Fatalf("failed to create MultiWriter: %v", err)
	}

	err = multi.Close()
	if err == nil {
		t.Error("expected error from failing close")
	}

	// w1 should still have been closed
	if !w1.closed {
		t.Error("w1 should be closed")
	}
}

func TestMultiWriterImplementsInterface(t *testing.T) {
	var _ IRWriter = (*MultiWriter)(nil)
}
