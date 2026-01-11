package sitegen

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/grokify/traffic2openapi/pkg/ir"
)

// Generator generates static HTML sites from IR records.
type Generator struct {
	engine    *Engine
	outputDir string
	options   *Options
}

// NewGenerator creates a new site generator.
func NewGenerator(outputDir string, opts *Options) *Generator {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &Generator{
		engine:    NewEngine(opts),
		outputDir: outputDir,
		options:   opts,
	}
}

// ProcessRecords processes IR records for site generation.
func (g *Generator) ProcessRecords(records []ir.IRRecord) {
	g.engine.ProcessRecords(records)
}

// Generate generates the static HTML site.
func (g *Generator) Generate() error {
	// Build site data
	siteData := g.engine.BuildSiteData()
	siteData.GeneratedAt = time.Now()

	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Create assets directory and write static files
	assetsDir := filepath.Join(g.outputDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return fmt.Errorf("creating assets directory: %w", err)
	}

	//nolint:gosec // G306: Static web assets need to be readable by web servers
	if err := os.WriteFile(filepath.Join(assetsDir, "style.css"), []byte(styleCSS), 0644); err != nil {
		return fmt.Errorf("writing style.css: %w", err)
	}

	//nolint:gosec // G306: Static web assets need to be readable by web servers
	if err := os.WriteFile(filepath.Join(assetsDir, "script.js"), []byte(scriptJS), 0644); err != nil {
		return fmt.Errorf("writing script.js: %w", err)
	}

	// Parse templates
	funcMap := template.FuncMap{
		"json":          toJSON,
		"jsonPretty":    toJSONPretty,
		"statusClass":   statusClass,
		"methodClass":   methodClass,
		"truncate":      truncate,
		"joinStrings":   joinStrings,
		"hasContent":    hasContent,
		"formatHeaders": formatHeaders,
	}

	indexTmpl, err := template.New("index").Funcs(funcMap).Parse(indexTemplate)
	if err != nil {
		return fmt.Errorf("parsing index template: %w", err)
	}

	endpointTmpl, err := template.New("endpoint").Funcs(funcMap).Parse(endpointTemplate)
	if err != nil {
		return fmt.Errorf("parsing endpoint template: %w", err)
	}

	// Generate index page
	indexPath := filepath.Join(g.outputDir, "index.html")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("creating index.html: %w", err)
	}
	defer indexFile.Close()

	if err := indexTmpl.Execute(indexFile, siteData); err != nil {
		return fmt.Errorf("executing index template: %w", err)
	}

	// Generate endpoint pages
	for _, ep := range siteData.Endpoints {
		epPath := filepath.Join(g.outputDir, ep.Slug+".html")
		epFile, err := os.Create(epPath)
		if err != nil {
			return fmt.Errorf("creating %s.html: %w", ep.Slug, err)
		}

		data := struct {
			*EndpointPage
			SiteTitle string
			BaseURL   string
		}{
			EndpointPage: ep,
			SiteTitle:    siteData.Title,
			BaseURL:      g.options.BaseURL,
		}

		if err := endpointTmpl.Execute(epFile, data); err != nil {
			epFile.Close()
			return fmt.Errorf("executing endpoint template for %s: %w", ep.Slug, err)
		}
		epFile.Close()
	}

	return nil
}

// Template helper functions

func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}

func toJSONPretty(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}

func statusClass(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "status-2xx"
	case status >= 300 && status < 400:
		return "status-3xx"
	case status >= 400 && status < 500:
		return "status-4xx"
	case status >= 500:
		return "status-5xx"
	default:
		return "status-other"
	}
}

func methodClass(method string) string {
	switch method {
	case "GET":
		return "method-get"
	case "POST":
		return "method-post"
	case "PUT":
		return "method-put"
	case "PATCH":
		return "method-patch"
	case "DELETE":
		return "method-delete"
	default:
		return "method-other"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func hasContent(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case string:
		return val != ""
	case map[string]any:
		return len(val) > 0
	case map[string]string:
		return len(val) > 0
	case []any:
		return len(val) > 0
	default:
		return true
	}
}

func formatHeaders(headers map[string]string) string {
	if len(headers) == 0 {
		return ""
	}
	result := ""
	keys := sortedMapKeys(headers)
	for _, k := range keys {
		result += fmt.Sprintf("%s: %s\n", k, headers[k])
	}
	return result
}

// GenerateFromFile generates a site from an IR file or directory.
func GenerateFromFile(inputPath, outputDir string, opts *Options) error {
	records, err := ir.ReadFile(inputPath)
	if err != nil {
		// Try reading as directory
		records, err = ir.ReadDir(inputPath)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
	}

	gen := NewGenerator(outputDir, opts)
	gen.ProcessRecords(records)
	return gen.Generate()
}
