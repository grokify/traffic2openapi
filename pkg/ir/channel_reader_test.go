package ir

import (
	"io"
	"testing"
)

func TestChannelReader(t *testing.T) {
	ch := make(chan *IRRecord, 3)

	// Send records
	ch <- NewRecord(RequestMethodGET, "/test1", 200)
	ch <- NewRecord(RequestMethodPOST, "/test2", 201)
	ch <- NewRecord(RequestMethodDELETE, "/test3", 204)
	close(ch)

	reader := NewChannelReader(ch)

	// Read first record
	record, err := reader.Read()
	if err != nil {
		t.Fatalf("read 1 failed: %v", err)
	}
	if record.Request.Path != "/test1" {
		t.Errorf("expected /test1, got %s", record.Request.Path)
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
}

func TestChannelReaderFromWriter(t *testing.T) {
	writer := NewChannelWriter(WithChannelBufferSize(10))

	// Write records
	for i := 0; i < 5; i++ {
		if err := writer.Write(NewRecord(RequestMethodGET, "/test", 200)); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	// Create reader from writer
	reader := NewChannelReaderFromWriter(writer)

	// Close writer (signals no more records)
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	// Read all records
	count := 0
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		count++
	}

	if count != 5 {
		t.Errorf("expected 5 records, got %d", count)
	}
}

func TestChannelReaderClose(t *testing.T) {
	ch := make(chan *IRRecord, 1)
	ch <- NewRecord(RequestMethodGET, "/test", 200)

	reader := NewChannelReader(ch)

	// Close the reader
	if err := reader.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}

	// Read should return EOF after close
	_, err := reader.Read()
	if err != io.EOF {
		t.Errorf("expected io.EOF after close, got %v", err)
	}
}

func TestChannelReaderImplementsInterface(t *testing.T) {
	var _ IRReader = (*ChannelReader)(nil)
}
