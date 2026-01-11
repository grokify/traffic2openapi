package sitegen

import (
	"sort"
	"strings"
	"sync"

	"github.com/grokify/traffic2openapi/pkg/inference"
	"github.com/grokify/traffic2openapi/pkg/ir"
)

// Engine processes IR records and stores them for site generation.
type Engine struct {
	mu      sync.RWMutex
	records map[string][]*StoredRecord // endpointKey -> records
	hosts   map[string]bool
	options *Options
}

// NewEngine creates a new site generation engine.
func NewEngine(opts *Options) *Engine {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &Engine{
		records: make(map[string][]*StoredRecord),
		hosts:   make(map[string]bool),
		options: opts,
	}
}

// ProcessRecord processes a single IR record.
func (e *Engine) ProcessRecord(record *ir.IRRecord) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Infer path template
	pathTemplate, pathParams := inference.InferPathTemplate(record.Request.Path)

	// Create endpoint key
	endpointKey := inference.EndpointKey(string(record.Request.Method), pathTemplate)

	// Compute dedup key
	dedupKey := ComputeDedupKey(record, pathTemplate)

	// Store the record
	stored := &StoredRecord{
		Record:       record,
		PathTemplate: pathTemplate,
		PathParams:   pathParams,
		EndpointKey:  endpointKey,
		DedupKey:     dedupKey,
	}

	e.records[endpointKey] = append(e.records[endpointKey], stored)

	// Track hosts
	if record.Request.Host != nil {
		e.hosts[*record.Request.Host] = true
	}
}

// ProcessRecords processes multiple IR records.
func (e *Engine) ProcessRecords(records []ir.IRRecord) {
	for i := range records {
		e.ProcessRecord(&records[i])
	}
}

// BuildSiteData builds the SiteData from processed records.
func (e *Engine) BuildSiteData() *SiteData {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Collect unique hosts
	hosts := make([]string, 0, len(e.hosts))
	for h := range e.hosts {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)

	// Count total requests
	totalRequests := 0
	for _, recs := range e.records {
		totalRequests += len(recs)
	}

	// Build endpoint pages
	endpoints := e.buildEndpointPages()

	return &SiteData{
		Title:     e.options.Title,
		Endpoints: endpoints,
		Stats: &SiteStats{
			TotalRequests:  totalRequests,
			TotalEndpoints: len(endpoints),
			UniqueHosts:    hosts,
		},
	}
}

// buildEndpointPages builds EndpointPage structures from stored records.
func (e *Engine) buildEndpointPages() []*EndpointPage {
	// Group by endpoint key
	endpointKeys := make([]string, 0, len(e.records))
	for key := range e.records {
		endpointKeys = append(endpointKeys, key)
	}
	sort.Strings(endpointKeys)

	pages := make([]*EndpointPage, 0, len(endpointKeys))

	for _, key := range endpointKeys {
		records := e.records[key]
		if len(records) == 0 {
			continue
		}

		// Parse endpoint key (e.g., "GET /users/{userId}")
		parts := strings.SplitN(key, " ", 2)
		method := parts[0]
		pathTemplate := "/"
		if len(parts) > 1 {
			pathTemplate = parts[1]
		}

		// Group by status code
		statusGroups := e.buildStatusGroups(records)

		pages = append(pages, &EndpointPage{
			Method:       method,
			PathTemplate: pathTemplate,
			Slug:         makeSlug(method, pathTemplate),
			RequestCount: len(records),
			StatusGroups: statusGroups,
		})
	}

	return pages
}

// buildStatusGroups groups records by status code.
func (e *Engine) buildStatusGroups(records []*StoredRecord) []*StatusGroup {
	// Group by status code
	byStatus := make(map[int][]*StoredRecord)
	var statusCodes []int

	for _, rec := range records {
		status := rec.Record.Response.Status
		if _, exists := byStatus[status]; !exists {
			statusCodes = append(statusCodes, status)
		}
		byStatus[status] = append(byStatus[status], rec)
	}

	sort.Ints(statusCodes)

	groups := make([]*StatusGroup, 0, len(statusCodes))
	for _, status := range statusCodes {
		recs := byStatus[status]
		distinct, deduped := DeduplicateRequests(recs)
		groups = append(groups, &StatusGroup{
			StatusCode: status,
			Distinct:   distinct,
			Deduped:    deduped,
		})
	}

	return groups
}

// makeSlug creates a URL-safe slug from method and path template.
func makeSlug(method, pathTemplate string) string {
	// Convert to lowercase
	slug := strings.ToLower(method + "-" + pathTemplate)

	// Replace slashes and braces
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "{", "")
	slug = strings.ReplaceAll(slug, "}", "")

	// Remove leading/trailing dashes
	slug = strings.Trim(slug, "-")

	// Collapse multiple dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	return slug
}
