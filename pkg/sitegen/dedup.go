package sitegen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

// ComputeDedupKey computes a deduplication key for a request.
// The key is based on: method + pathTemplate + sorted query keys + body structure + status.
func ComputeDedupKey(record *ir.IRRecord, pathTemplate string) string {
	h := sha256.New()

	// Method + path template
	h.Write([]byte(string(record.Request.Method)))
	h.Write([]byte(pathTemplate))

	// Sorted query param keys (structure, not values)
	if record.Request.Query != nil {
		queryKeys := sortedMapKeys(record.Request.Query)
		for _, k := range queryKeys {
			h.Write([]byte(k))
		}
	}

	// Body structure fingerprint
	if record.Request.Body != nil {
		h.Write([]byte(structureFingerprint(record.Request.Body)))
	}

	// Response status
	h.Write([]byte(strconv.Itoa(record.Response.Status)))

	return hex.EncodeToString(h.Sum(nil))[:16]
}

// structureFingerprint computes a fingerprint of a JSON structure.
// It captures the shape (keys, types) but not the actual values.
func structureFingerprint(v any) string {
	switch val := v.(type) {
	case map[string]any:
		keys := sortedMapKeys(val)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = k + ":" + structureFingerprint(val[k])
		}
		result := "{"
		for i, p := range parts {
			if i > 0 {
				result += ","
			}
			result += p
		}
		result += "}"
		return result
	case []any:
		if len(val) > 0 {
			return "[]" + structureFingerprint(val[0])
		}
		return "[]"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "bool"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// sortedMapKeys returns the keys of a map sorted alphabetically.
func sortedMapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// DeduplicateRequests groups requests and creates both distinct and deduped views.
func DeduplicateRequests(records []*StoredRecord) (distinct []*RequestView, deduped *DedupedView) {
	if len(records) == 0 {
		return nil, nil
	}

	// Group by dedup key
	groups := make(map[string][]*StoredRecord)
	var orderedKeys []string

	for _, rec := range records {
		if _, exists := groups[rec.DedupKey]; !exists {
			orderedKeys = append(orderedKeys, rec.DedupKey)
		}
		groups[rec.DedupKey] = append(groups[rec.DedupKey], rec)
	}

	// Create distinct views - one per unique dedup key
	distinct = make([]*RequestView, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		recs := groups[key]
		// Use first record as representative
		rec := recs[0]
		distinct = append(distinct, recordToRequestView(rec))
	}

	// Create deduped view - aggregate all values
	deduped = createDedupedView(records)

	return distinct, deduped
}

// recordToRequestView converts a StoredRecord to a RequestView.
func recordToRequestView(rec *StoredRecord) *RequestView {
	r := rec.Record

	var durationMs float64
	if r.DurationMs != nil {
		durationMs = *r.DurationMs
	}

	var id string
	if r.Id != nil {
		id = *r.Id
	}

	var contentType string
	if r.Response.ContentType != nil {
		contentType = *r.Response.ContentType
	}

	return &RequestView{
		ID:              id,
		Method:          string(r.Request.Method),
		Path:            r.Request.Path,
		PathTemplate:    rec.PathTemplate,
		PathParams:      rec.PathParams,
		QueryParams:     r.Request.Query,
		RequestHeaders:  r.Request.Headers,
		RequestBody:     r.Request.Body,
		ResponseHeaders: r.Response.Headers,
		ResponseBody:    r.Response.Body,
		StatusCode:      r.Response.Status,
		ContentType:     contentType,
		DurationMs:      durationMs,
		DedupKey:        rec.DedupKey,
	}
}

// createDedupedView creates an aggregated view of all requests.
func createDedupedView(records []*StoredRecord) *DedupedView {
	if len(records) == 0 {
		return nil
	}

	first := records[0]
	dv := &DedupedView{
		Method:           string(first.Record.Request.Method),
		PathTemplate:     first.PathTemplate,
		PathParamValues:  make(map[string][]string),
		QueryParamValues: make(map[string][]string),
		Count:            len(records),
	}

	// Collect all unique path param values
	pathParamSeen := make(map[string]map[string]bool)
	// Collect all unique query param values
	queryParamSeen := make(map[string]map[string]bool)

	for _, rec := range records {
		// Path params
		for k, v := range rec.PathParams {
			if pathParamSeen[k] == nil {
				pathParamSeen[k] = make(map[string]bool)
			}
			if !pathParamSeen[k][v] {
				pathParamSeen[k][v] = true
				dv.PathParamValues[k] = append(dv.PathParamValues[k], v)
			}
		}

		// Query params
		for k, v := range rec.Record.Request.Query {
			if queryParamSeen[k] == nil {
				queryParamSeen[k] = make(map[string]bool)
			}
			strVal := fmt.Sprintf("%v", v)
			if !queryParamSeen[k][strVal] {
				queryParamSeen[k][strVal] = true
				dv.QueryParamValues[k] = append(dv.QueryParamValues[k], strVal)
			}
		}

		// Use first non-nil body as example
		if dv.RequestBodyExample == nil && rec.Record.Request.Body != nil {
			dv.RequestBodyExample = rec.Record.Request.Body
		}
		if dv.ResponseBodyExample == nil && rec.Record.Response.Body != nil {
			dv.ResponseBodyExample = rec.Record.Response.Body
		}
	}

	// Sort values for consistent output
	for k := range dv.PathParamValues {
		sort.Strings(dv.PathParamValues[k])
	}
	for k := range dv.QueryParamValues {
		sort.Strings(dv.QueryParamValues[k])
	}

	return dv
}
