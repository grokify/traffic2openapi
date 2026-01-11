// Package sitegen generates static HTML documentation sites from IR traffic logs.
package sitegen

import (
	"time"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

// SiteData is the top-level data for template rendering.
type SiteData struct {
	Title       string
	GeneratedAt time.Time
	Endpoints   []*EndpointPage
	Stats       *SiteStats
}

// SiteStats contains aggregate statistics for the site.
type SiteStats struct {
	TotalRequests  int
	TotalEndpoints int
	UniqueHosts    []string
}

// EndpointPage represents a single endpoint's page.
type EndpointPage struct {
	Method       string
	PathTemplate string
	Slug         string // URL-safe filename (e.g., "get-users-userid")
	RequestCount int
	StatusGroups []*StatusGroup
}

// StatusGroup groups requests by HTTP status code.
type StatusGroup struct {
	StatusCode int
	Distinct   []*RequestView // All unique requests
	Deduped    *DedupedView   // Collapsed view with all seen values
}

// RequestView represents a single request for display.
type RequestView struct {
	ID              string
	Method          string
	Path            string // Original path (not template)
	PathTemplate    string
	PathParams      map[string]string
	QueryParams     map[string]any
	RequestHeaders  map[string]string
	RequestBody     any
	ResponseHeaders map[string]string
	ResponseBody    any
	StatusCode      int
	ContentType     string
	DurationMs      float64
	DedupKey        string
}

// DedupedView shows all variations in a compact format.
type DedupedView struct {
	Method              string
	PathTemplate        string
	PathParamValues     map[string][]string // param -> all seen values
	QueryParamValues    map[string][]string // param -> all seen values
	RequestBodyExample  any
	ResponseBodyExample any
	Count               int
}

// StoredRecord holds an IR record with its computed metadata.
type StoredRecord struct {
	Record       *ir.IRRecord
	PathTemplate string
	PathParams   map[string]string
	EndpointKey  string
	DedupKey     string
}

// Options configures the site generator.
type Options struct {
	Title   string
	BaseURL string
}

// DefaultOptions returns the default site generation options.
func DefaultOptions() *Options {
	return &Options{
		Title:   "API Traffic Documentation",
		BaseURL: "",
	}
}
