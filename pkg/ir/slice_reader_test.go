package ir

import (
	"io"
	"testing"
)

func TestSliceReader(t *testing.T) {
	records := []IRRecord{
		*NewRecord(RequestMethodGET, "/test1", 200),
		*NewRecord(RequestMethodPOST, "/test2", 201),
		*NewRecord(RequestMethodDELETE, "/test3", 204),
	}

	reader := NewSliceReader(records)

	if reader.Len() != 3 {
		t.Errorf("expected len 3, got %d", reader.Len())
	}

	if reader.Remaining() != 3 {
		t.Errorf("expected remaining 3, got %d", reader.Remaining())
	}

	// Read first record
	record, err := reader.Read()
	if err != nil {
		t.Fatalf("read 1 failed: %v", err)
	}
	if record.Request.Path != "/test1" {
		t.Errorf("expected /test1, got %s", record.Request.Path)
	}
	if reader.Remaining() != 2 {
		t.Errorf("expected remaining 2, got %d", reader.Remaining())
	}

	// Read second record
	record, err = reader.Read()
	if err != nil {
		t.Fatalf("read 2 failed: %v", err)
	}
	if record.Request.Path != "/test2" {
		t.Errorf("expected /test2, got %s", record.Request.Path)
	}

	// Read third record
	record, err = reader.Read()
	if err != nil {
		t.Fatalf("read 3 failed: %v", err)
	}
	if record.Request.Path != "/test3" {
		t.Errorf("expected /test3, got %s", record.Request.Path)
	}

	// Read should return EOF
	_, err = reader.Read()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}

	if reader.Remaining() != 0 {
		t.Errorf("expected remaining 0, got %d", reader.Remaining())
	}
}

func TestSliceReaderReset(t *testing.T) {
	records := []IRRecord{
		*NewRecord(RequestMethodGET, "/test1", 200),
		*NewRecord(RequestMethodGET, "/test2", 200),
	}

	reader := NewSliceReader(records)

	// Read all records
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
	}

	if reader.Remaining() != 0 {
		t.Errorf("expected remaining 0, got %d", reader.Remaining())
	}

	// Reset and read again
	reader.Reset()

	if reader.Remaining() != 2 {
		t.Errorf("expected remaining 2 after reset, got %d", reader.Remaining())
	}

	record, err := reader.Read()
	if err != nil {
		t.Fatalf("read after reset failed: %v", err)
	}
	if record.Request.Path != "/test1" {
		t.Errorf("expected /test1, got %s", record.Request.Path)
	}
}

func TestSliceReaderEmpty(t *testing.T) {
	reader := NewSliceReader(nil)

	if reader.Len() != 0 {
		t.Errorf("expected len 0, got %d", reader.Len())
	}

	_, err := reader.Read()
	if err != io.EOF {
		t.Errorf("expected io.EOF for empty slice, got %v", err)
	}
}

func TestSliceReaderClose(t *testing.T) {
	records := []IRRecord{
		*NewRecord(RequestMethodGET, "/test", 200),
	}

	reader := NewSliceReader(records)

	// Close should be a no-op
	if err := reader.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}

	// Should still be able to read after close (it's a no-op)
	_, err := reader.Read()
	if err != nil {
		t.Errorf("read after close failed: %v", err)
	}
}

func TestSliceReaderImplementsInterface(t *testing.T) {
	var _ IRReader = (*SliceReader)(nil)
}
