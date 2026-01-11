package ir

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/grokify/omnistorage/backend/file"
)

// testRecords creates a set of test records.
func testRecords() []*IRRecord {
	return []*IRRecord{
		NewRecord(RequestMethodGET, "/users", 200).SetID("test-1"),
		NewRecord(RequestMethodPOST, "/users", 201).SetID("test-2"),
		NewRecord(RequestMethodDELETE, "/users/1", 204).SetID("test-3"),
	}
}

// testProviderRoundTrip tests that records can be written and read back.
func testProviderRoundTrip(t *testing.T, name string, provider Provider, path string) {
	t.Helper()

	ctx := context.Background()
	records := testRecords()

	// Write records
	w, err := provider.NewWriter(ctx, path)
	if err != nil {
		t.Fatalf("%s: NewWriter failed: %v", name, err)
	}

	for _, record := range records {
		if err := w.Write(record); err != nil {
			t.Fatalf("%s: Write failed: %v", name, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("%s: Close writer failed: %v", name, err)
	}

	// Read records back
	r, err := provider.NewReader(ctx, path)
	if err != nil {
		t.Fatalf("%s: NewReader failed: %v", name, err)
	}

	var readRecords []*IRRecord
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("%s: Read failed: %v", name, err)
		}
		readRecords = append(readRecords, record)
	}

	if err := r.Close(); err != nil {
		t.Fatalf("%s: Close reader failed: %v", name, err)
	}

	// Verify
	if len(readRecords) != len(records) {
		t.Fatalf("%s: Read %d records, want %d", name, len(readRecords), len(records))
	}

	for i, record := range records {
		if *readRecords[i].Id != *record.Id {
			t.Errorf("%s: Record %d ID = %q, want %q", name, i, *readRecords[i].Id, *record.Id)
		}
		if readRecords[i].Request.Method != record.Request.Method {
			t.Errorf("%s: Record %d Method = %q, want %q", name, i, readRecords[i].Request.Method, record.Request.Method)
		}
		if readRecords[i].Request.Path != record.Request.Path {
			t.Errorf("%s: Record %d Path = %q, want %q", name, i, readRecords[i].Request.Path, record.Request.Path)
		}
		if readRecords[i].Response.Status != record.Response.Status {
			t.Errorf("%s: Record %d Status = %d, want %d", name, i, readRecords[i].Response.Status, record.Response.Status)
		}
	}
}

func TestNDJSONProvider(t *testing.T) {
	tmpDir := t.TempDir()
	provider := NDJSON()
	path := filepath.Join(tmpDir, "test.ndjson")

	testProviderRoundTrip(t, "NDJSONProvider", provider, path)
}

func TestGzipNDJSONProvider(t *testing.T) {
	tmpDir := t.TempDir()
	provider := GzipNDJSON()
	path := filepath.Join(tmpDir, "test.ndjson.gz")

	testProviderRoundTrip(t, "GzipNDJSONProvider", provider, path)
}

func TestGzipNDJSONProviderWithLevel(t *testing.T) {
	tmpDir := t.TempDir()
	provider := GzipNDJSON(WithGzipCompressionLevel(9))
	path := filepath.Join(tmpDir, "test-best.ndjson.gz")

	testProviderRoundTrip(t, "GzipNDJSONProvider(BestCompression)", provider, path)
}

func TestStorageProvider(t *testing.T) {
	tmpDir := t.TempDir()

	backend := file.New(file.Config{Root: tmpDir})
	defer func() { _ = backend.Close() }()

	provider := Storage(backend)

	t.Run("NDJSON", func(t *testing.T) {
		testProviderRoundTrip(t, "StorageProvider/NDJSON", provider, "test.ndjson")
	})

	t.Run("GzipNDJSON", func(t *testing.T) {
		testProviderRoundTrip(t, "StorageProvider/GzipNDJSON", provider, "test.ndjson.gz")
	})
}

func TestChannelProvider(t *testing.T) {
	ctx := context.Background()
	records := testRecords()

	t.Run("Unbuffered", func(t *testing.T) {
		provider := Channel()

		// Channel operations need goroutines since they're synchronous
		errCh := make(chan error, 1)
		go func() {
			w, err := provider.NewWriter(ctx, "")
			if err != nil {
				errCh <- err
				return
			}
			for _, record := range records {
				if err := w.Write(record); err != nil {
					errCh <- err
					return
				}
			}
			errCh <- w.Close()
		}()

		r, err := provider.NewReader(ctx, "")
		if err != nil {
			t.Fatalf("NewReader failed: %v", err)
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

		if err := <-errCh; err != nil {
			t.Fatalf("Writer error: %v", err)
		}

		if len(readRecords) != len(records) {
			t.Errorf("Read %d records, want %d", len(readRecords), len(records))
		}
	})

	t.Run("Buffered", func(t *testing.T) {
		provider := Channel(WithChannelProviderBufferSize(10))

		w, err := provider.NewWriter(ctx, "")
		if err != nil {
			t.Fatalf("NewWriter failed: %v", err)
		}

		// With buffer, we can write without blocking
		for _, record := range records {
			if err := w.Write(record); err != nil {
				t.Fatalf("Write failed: %v", err)
			}
		}
		_ = w.Close()

		r, err := provider.NewReader(ctx, "")
		if err != nil {
			t.Fatalf("NewReader failed: %v", err)
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

		if len(readRecords) != len(records) {
			t.Errorf("Read %d records, want %d", len(readRecords), len(records))
		}
	})
}

func TestChannelProviderWithExistingChannel(t *testing.T) {
	ctx := context.Background()
	records := testRecords()

	// Create external channel
	ch := make(chan *IRRecord, 10)

	provider := Channel(WithExistingChannel(ch))

	// Verify it's using our channel
	if provider.Chan() != ch {
		t.Error("Provider not using provided channel")
	}

	w, err := provider.NewWriter(ctx, "")
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}

	for _, record := range records {
		if err := w.Write(record); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}
	_ = w.Close()

	// Read from external channel directly
	count := 0
	for range ch {
		count++
	}

	if count != len(records) {
		t.Errorf("Read %d records, want %d", count, len(records))
	}
}

func TestProviderContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	t.Run("NDJSONProvider", func(t *testing.T) {
		provider := NDJSON()
		_, err := provider.NewWriter(ctx, "/tmp/test.ndjson")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
		_, err = provider.NewReader(ctx, "/tmp/test.ndjson")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})

	t.Run("GzipNDJSONProvider", func(t *testing.T) {
		provider := GzipNDJSON()
		_, err := provider.NewWriter(ctx, "/tmp/test.ndjson.gz")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
		_, err = provider.NewReader(ctx, "/tmp/test.ndjson.gz")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})

	t.Run("ChannelProvider", func(t *testing.T) {
		provider := Channel()
		_, err := provider.NewWriter(ctx, "")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
		_, err = provider.NewReader(ctx, "")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})
}

func TestStreamProviders(t *testing.T) {
	records := testRecords()

	testCases := []struct {
		name     string
		suffix   string
		provider StreamProvider
	}{
		{"NDJSON", ".ndjson", NDJSON()},
		{"GzipNDJSON", ".ndjson.gz", GzipNDJSON()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-*"+tc.suffix)
			if err != nil {
				t.Fatalf("CreateTemp failed: %v", err)
			}
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			// Write using stream writer
			w := tc.provider.NewStreamWriter(tmpFile)
			for _, record := range records {
				if err := w.Write(record); err != nil {
					t.Fatalf("Write failed: %v", err)
				}
			}
			_ = w.Close()
			_ = tmpFile.Close()

			// Read using stream reader
			readFile, err := os.Open(tmpFile.Name())
			if err != nil {
				t.Fatalf("Open failed: %v", err)
			}
			defer func() { _ = readFile.Close() }()

			r, err := tc.provider.NewStreamReader(readFile)
			if err != nil {
				t.Fatalf("NewStreamReader failed: %v", err)
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

			if len(readRecords) != len(records) {
				t.Errorf("Read %d records, want %d", len(readRecords), len(records))
			}
		})
	}
}
