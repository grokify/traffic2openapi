package ir

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IRReader is the interface for reading IR records from any source.
// Implementations should return io.EOF when no more records are available.
type IRReader interface {
	// Read reads the next IR record.
	// Returns io.EOF when no more records are available.
	Read() (*IRRecord, error)

	// Close releases any resources held by the reader.
	Close() error
}

// ReadFile reads IR records from a file.
// Automatically detects format based on file extension:
// - .ndjson: newline-delimited JSON (one record per line)
// - .json: batch format with version and records array
func ReadFile(path string) ([]IRRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".ndjson":
		return ReadNDJSON(f)
	case ".json":
		return ReadBatch(f)
	default:
		// Try to auto-detect by peeking at first byte
		return readAutoDetect(f)
	}
}

// ReadBatch reads a batch-format JSON file.
func ReadBatch(r io.Reader) ([]IRRecord, error) {
	var batch Batch
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&batch); err != nil {
		return nil, fmt.Errorf("decoding batch JSON: %w", err)
	}

	if batch.Version != Version {
		return nil, fmt.Errorf("unsupported IR version: %s (expected %s)", batch.Version, Version)
	}

	return batch.Records, nil
}

// ReadNDJSON reads newline-delimited JSON records.
func ReadNDJSON(r io.Reader) ([]IRRecord, error) {
	var records []IRRecord
	scanner := bufio.NewScanner(r)

	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line size

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record IRRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning NDJSON: %w", err)
	}

	return records, nil
}

// readAutoDetect tries to detect the format by looking at the first character.
func readAutoDetect(r io.Reader) ([]IRRecord, error) {
	// Read into buffer so we can peek and then re-read
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading data: %w", err)
	}

	// Trim leading whitespace to find first significant character
	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	// If starts with '{' followed by "version", it's batch format
	// If starts with '{' and contains newlines with more '{', it's NDJSON
	// If starts with '[', it could be a raw array of records
	switch trimmed[0] {
	case '{':
		// Check if it looks like batch format
		if strings.Contains(trimmed[:min(100, len(trimmed))], `"version"`) {
			return ReadBatch(strings.NewReader(string(data)))
		}
		// Otherwise assume NDJSON
		return ReadNDJSON(strings.NewReader(string(data)))
	case '[':
		// Raw array of records (not wrapped in batch)
		var records []IRRecord
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, fmt.Errorf("decoding JSON array: %w", err)
		}
		return records, nil
	default:
		return nil, fmt.Errorf("unrecognized format: expected JSON object or array")
	}
}

// ReadDir reads all IR files from a directory.
func ReadDir(dir string) ([]IRRecord, error) {
	var allRecords []IRRecord

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".json" && ext != ".ndjson" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		records, err := ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}
		allRecords = append(allRecords, records...)
	}

	return allRecords, nil
}

// StreamNDJSON streams NDJSON records through a channel.
// Useful for processing large files without loading all into memory.
func StreamNDJSON(r io.Reader) (<-chan IRRecord, <-chan error) {
	records := make(chan IRRecord, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(records)
		defer close(errs)

		scanner := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var record IRRecord
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				errs <- err
				return
			}
			records <- record
		}

		if err := scanner.Err(); err != nil {
			errs <- err
		}
	}()

	return records, errs
}
