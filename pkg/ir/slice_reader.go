package ir

import (
	"io"
)

// SliceReader reads IR records from an in-memory slice.
// Useful for converting existing []IRRecord data to the IRReader interface.
type SliceReader struct {
	records []IRRecord
	index   int
}

// NewSliceReader creates a reader from an existing slice of records.
func NewSliceReader(records []IRRecord) *SliceReader {
	return &SliceReader{
		records: records,
		index:   0,
	}
}

// Read reads the next IR record from the slice.
// Returns io.EOF when all records have been read.
func (r *SliceReader) Read() (*IRRecord, error) {
	if r.index >= len(r.records) {
		return nil, io.EOF
	}

	record := &r.records[r.index]
	r.index++
	return record, nil
}

// Close is a no-op for SliceReader.
func (r *SliceReader) Close() error {
	return nil
}

// Reset resets the reader to the beginning of the slice.
func (r *SliceReader) Reset() {
	r.index = 0
}

// Remaining returns the number of unread records.
func (r *SliceReader) Remaining() int {
	return len(r.records) - r.index
}

// Len returns the total number of records.
func (r *SliceReader) Len() int {
	return len(r.records)
}
