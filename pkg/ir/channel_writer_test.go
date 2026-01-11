package ir

import (
	"sync"
	"testing"
	"time"
)

func TestChannelWriterBasic(t *testing.T) {
	w := NewChannelWriter(WithChannelBufferSize(10))

	// Write records
	for i := 0; i < 5; i++ {
		record := NewRecord(RequestMethodGET, "/test", 200)
		if err := w.Write(record); err != nil {
			t.Errorf("write failed: %v", err)
		}
	}

	// Read records
	ch := w.Channel()
	for i := 0; i < 5; i++ {
		select {
		case record := <-ch:
			if record.Request.Path != "/test" {
				t.Errorf("expected /test, got %s", record.Request.Path)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for record")
		}
	}

	if err := w.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestChannelWriterWithExistingChannel(t *testing.T) {
	ch := make(chan *IRRecord, 5)
	w := NewChannelWriterWithChan(ch)

	record := NewRecord(RequestMethodPOST, "/users", 201)
	if err := w.Write(record); err != nil {
		t.Errorf("write failed: %v", err)
	}

	// Read from the original channel
	select {
	case r := <-ch:
		if r.Request.Method != RequestMethodPOST {
			t.Errorf("expected POST, got %s", r.Request.Method)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for record")
	}

	if err := w.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestChannelWriterClose(t *testing.T) {
	w := NewChannelWriter(WithChannelBufferSize(1))

	if err := w.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}

	// Write after close should return error
	record := NewRecord(RequestMethodGET, "/test", 200)
	if err := w.Write(record); err != ErrChannelClosed {
		t.Errorf("expected ErrChannelClosed, got %v", err)
	}

	// Double close should be safe
	if err := w.Close(); err != nil {
		t.Errorf("double close failed: %v", err)
	}
}

func TestChannelWriterFlush(t *testing.T) {
	w := NewChannelWriter(WithChannelBufferSize(1))

	// Flush should be a no-op
	if err := w.Flush(); err != nil {
		t.Errorf("flush failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestChannelWriterConcurrent(t *testing.T) {
	w := NewChannelWriter(WithChannelBufferSize(100))

	var wg sync.WaitGroup
	numWriters := 10
	recordsPerWriter := 10

	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < recordsPerWriter; j++ {
				record := NewRecord(RequestMethodGET, "/concurrent", 200)
				if err := w.Write(record); err != nil {
					t.Errorf("concurrent write failed: %v", err)
				}
			}
		}()
	}

	// Start reader
	received := 0
	done := make(chan struct{})
	go func() {
		for range w.Channel() {
			received++
			if received == numWriters*recordsPerWriter {
				close(done)
				return
			}
		}
	}()

	wg.Wait()

	select {
	case <-done:
		// All records received
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout: only received %d of %d records", received, numWriters*recordsPerWriter)
	}

	if err := w.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestChannelWriterImplementsInterface(t *testing.T) {
	var _ IRWriter = (*ChannelWriter)(nil)
}
