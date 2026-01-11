package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/har"
	"github.com/grokify/traffic2openapi/pkg/ir"
	"github.com/spf13/cobra"
)

var harCmd = &cobra.Command{
	Use:   "har",
	Short: "Convert HAR files to IR format",
	Long: `Convert HAR (HTTP Archive) files to Intermediate Representation (IR) format.

HAR is a standard format for recording HTTP transactions, supported by:
  - Browser DevTools (Chrome, Firefox, Safari)
  - Playwright and Puppeteer
  - Charles Proxy, Fiddler, mitmproxy
  - Postman

Examples:
  # Convert a single HAR file
  traffic2openapi convert har -i recording.har -o traffic.ndjson

  # Convert multiple HAR files from a directory
  traffic2openapi convert har -i ./har-files/ -o traffic.ndjson

  # Convert and filter specific hosts
  traffic2openapi convert har -i recording.har -o traffic.ndjson --host api.example.com

  # Convert without headers
  traffic2openapi convert har -i recording.har -o traffic.ndjson --no-headers

  # Generate OpenAPI directly
  traffic2openapi convert har -i recording.har -o traffic.ndjson
  traffic2openapi generate -i traffic.ndjson -o openapi.yaml`,
	RunE: runHARConvert,
}

var (
	// HAR flags
	harInputPath      string
	harOutputPath     string
	harIncludeHeaders bool
	harFilterHeaders  string
	harFilterHost     string
	harFilterMethod   string
	harIncludeCookies bool
)

func init() {
	convertCmd.AddCommand(harCmd)

	// Input/output flags
	harCmd.Flags().StringVarP(&harInputPath, "input", "i", "", "Input HAR file or directory (required)")
	harCmd.Flags().StringVarP(&harOutputPath, "output", "o", "", "Output file path (default: stdout)")

	// Filter flags
	harCmd.Flags().BoolVar(&harIncludeHeaders, "headers", true, "Include HTTP headers in output")
	harCmd.Flags().StringVar(&harFilterHeaders, "filter-headers", "", "Additional headers to filter (comma-separated)")
	harCmd.Flags().StringVar(&harFilterHost, "host", "", "Only include requests to this host")
	harCmd.Flags().StringVar(&harFilterMethod, "method", "", "Only include requests with this method (GET, POST, etc.)")
	harCmd.Flags().BoolVar(&harIncludeCookies, "cookies", false, "Include cookie headers in output")

	_ = harCmd.MarkFlagRequired("input")
}

func runHARConvert(cmd *cobra.Command, args []string) error {
	if harInputPath == "" {
		return fmt.Errorf("--input is required")
	}

	// Create reader with configured converter
	reader := har.NewReader()
	configureHARConverter(reader.Converter)

	// Check if input is file or directory
	info, err := os.Stat(harInputPath)
	if err != nil {
		return fmt.Errorf("input path error: %w", err)
	}

	var records []ir.IRRecord

	if info.IsDir() {
		cmd.Printf("Reading HAR files from directory: %s\n", harInputPath)
		records, err = reader.ReadDir(harInputPath)
	} else {
		cmd.Printf("Reading HAR file: %s\n", harInputPath)
		records, err = reader.ReadFile(harInputPath)
	}

	if err != nil {
		return err
	}

	// Apply filters
	records = filterRecords(records)

	if len(records) == 0 {
		cmd.Printf("No records found\n")
		return nil
	}

	cmd.Printf("Converted %d records\n", len(records))

	// Write output
	if harOutputPath == "" {
		return ir.WriteNDJSON(os.Stdout, records)
	}

	if err := ir.WriteFile(harOutputPath, records); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	cmd.Printf("Wrote IR records to %s\n", harOutputPath)
	return nil
}

func configureHARConverter(converter *har.Converter) {
	converter.IncludeHeaders = harIncludeHeaders
	converter.IncludeCookies = harIncludeCookies

	if harFilterHeaders != "" {
		additional := strings.Split(harFilterHeaders, ",")
		for _, h := range additional {
			h = strings.TrimSpace(h)
			if h != "" {
				converter.FilterHeaders = append(converter.FilterHeaders, h)
			}
		}
	}
}

func filterRecords(records []ir.IRRecord) []ir.IRRecord {
	if harFilterHost == "" && harFilterMethod == "" {
		return records
	}

	filtered := make([]ir.IRRecord, 0, len(records))
	hostFilter := strings.ToLower(harFilterHost)
	methodFilter := strings.ToUpper(harFilterMethod)

	for _, r := range records {
		// Filter by host
		if hostFilter != "" {
			if r.Request.Host == nil || !strings.Contains(strings.ToLower(*r.Request.Host), hostFilter) {
				continue
			}
		}

		// Filter by method
		if methodFilter != "" {
			if strings.ToUpper(string(r.Request.Method)) != methodFilter {
				continue
			}
		}

		filtered = append(filtered, r)
	}

	return filtered
}
