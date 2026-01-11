package ir

import (
	"context"
	"io"
	"testing"

	"github.com/grokify/omnistorage/backend/file"
)

func TestStorageWriterReader(t *testing.T) {
	tmpDir := t.TempDir()

	backend := file.New(file.Config{Root: tmpDir})
	defer func() { _ = backend.Close() }()

	ctx := context.Background()

	// Create test records using helper function
	records := []*IRRecord{
		NewRecord(RequestMethodGET, "/users", 200).SetID("test-1"),
		NewRecord(RequestMethodPOST, "/users", 201).SetID("test-2"),
		NewRecord(RequestMethodDELETE, "/users/1", 204).SetID("test-3"),
	}

	// Test plain NDJSON
	t.Run("NDJSON", func(t *testing.T) {
		path := "records.ndjson"

		// Write records
		w, err := NewStorageWriter(ctx, backend, path)
		if err != nil {
			t.Fatalf("NewStorageWriter failed: %v", err)
		}

		for _, record := range records {
			if err := w.Write(record); err != nil {
				t.Fatalf("Write failed: %v", err)
			}
		}

		if err := w.Close(); err != nil {
			t.Fatalf("Close writer failed: %v", err)
		}

		if w.Count() != len(records) {
			t.Errorf("Count = %d, want %d", w.Count(), len(records))
		}

		// Read records back
		r, err := NewStorageReader(ctx, backend, path)
		if err != nil {
			t.Fatalf("NewStorageReader failed: %v", err)
		}

		var readRecords []*IRRecord
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}
			readRecords = append(readRecords, record)
		}

		if err := r.Close(); err != nil {
			t.Fatalf("Close reader failed: %v", err)
		}

		// Verify
		if len(readRecords) != len(records) {
			t.Fatalf("Read %d records, want %d", len(readRecords), len(records))
		}

		for i, record := range records {
			if *readRecords[i].Id != *record.Id {
				t.Errorf("Record %d ID = %q, want %q", i, *readRecords[i].Id, *record.Id)
			}
			if readRecords[i].Request.Method != record.Request.Method {
				t.Errorf("Record %d Method = %q, want %q", i, readRecords[i].Request.Method, record.Request.Method)
			}
			if readRecords[i].Request.Path != record.Request.Path {
				t.Errorf("Record %d Path = %q, want %q", i, readRecords[i].Request.Path, record.Request.Path)
			}
			if readRecords[i].Response.Status != record.Response.Status {
				t.Errorf("Record %d Status = %d, want %d", i, readRecords[i].Response.Status, record.Response.Status)
			}
		}
	})

	// Test gzip-compressed NDJSON
	t.Run("GzipNDJSON", func(t *testing.T) {
		path := "records.ndjson.gz"

		// Write records
		w, err := NewStorageWriter(ctx, backend, path)
		if err != nil {
			t.Fatalf("NewStorageWriter failed: %v", err)
		}

		for _, record := range records {
			if err := w.Write(record); err != nil {
				t.Fatalf("Write failed: %v", err)
			}
		}

		if err := w.Close(); err != nil {
			t.Fatalf("Close writer failed: %v", err)
		}

		// Read records back
		r, err := NewStorageReader(ctx, backend, path)
		if err != nil {
			t.Fatalf("NewStorageReader failed: %v", err)
		}

		var readRecords []*IRRecord
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}
			readRecords = append(readRecords, record)
		}

		if err := r.Close(); err != nil {
			t.Fatalf("Close reader failed: %v", err)
		}

		// Verify
		if len(readRecords) != len(records) {
			t.Fatalf("Read %d records, want %d", len(readRecords), len(records))
		}

		for i, record := range records {
			if *readRecords[i].Id != *record.Id {
				t.Errorf("Record %d ID = %q, want %q", i, *readRecords[i].Id, *record.Id)
			}
		}
	})
}

func TestStorageWriterFlush(t *testing.T) {
	tmpDir := t.TempDir()

	backend := file.New(file.Config{Root: tmpDir})
	defer func() { _ = backend.Close() }()

	ctx := context.Background()

	w, err := NewStorageWriter(ctx, backend, "flush-test.ndjson")
	if err != nil {
		t.Fatalf("NewStorageWriter failed: %v", err)
	}

	record := NewRecord(RequestMethodGET, "/test", 200).SetID("test-1")

	if err := w.Write(record); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Flush should not error
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestStorageReaderNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	backend := file.New(file.Config{Root: tmpDir})
	defer func() { _ = backend.Close() }()

	ctx := context.Background()

	_, err := NewStorageReader(ctx, backend, "nonexistent.ndjson")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestStorageReaderLineNumber(t *testing.T) {
	tmpDir := t.TempDir()

	backend := file.New(file.Config{Root: tmpDir})
	defer func() { _ = backend.Close() }()

	ctx := context.Background()

	// Write some records
	w, err := NewStorageWriter(ctx, backend, "linenum-test.ndjson")
	if err != nil {
		t.Fatalf("NewStorageWriter failed: %v", err)
	}

	for i := 0; i < 5; i++ {
		record := NewRecord(RequestMethodGET, "/test", 200).SetID("test")
		if err := w.Write(record); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}
	_ = w.Close()

	// Read and check line numbers
	r, err := NewStorageReader(ctx, backend, "linenum-test.ndjson")
	if err != nil {
		t.Fatalf("NewStorageReader failed: %v", err)
	}

	for i := 1; i <= 5; i++ {
		_, err := r.Read()
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if r.LineNumber() != i {
			t.Errorf("LineNumber = %d, want %d", r.LineNumber(), i)
		}
	}

	_ = r.Close()
}
