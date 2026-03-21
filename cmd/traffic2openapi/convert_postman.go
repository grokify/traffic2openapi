package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/grokify/traffic2openapi/pkg/ir"
	"github.com/grokify/traffic2openapi/pkg/postman"
	"github.com/spf13/cobra"
)

var postmanCmd = &cobra.Command{
	Use:   "postman",
	Short: "Convert Postman Collection to IR format",
	Long: `Convert Postman Collection v2.1 files to Intermediate Representation (IR) format.

Postman Collections contain API requests with saved example responses, organized
into folders. This converter preserves:
  - Request details (method, URL, headers, body, query params)
  - Saved example responses
  - Folder structure as tags
  - Collection metadata (name, description, version)
  - Request descriptions and documentation
  - Variable resolution

Examples:
  # Convert a Postman collection
  traffic2openapi convert postman -i collection.json -o api.ndjson

  # Convert with base URL variable
  traffic2openapi convert postman -i collection.json -o api.ndjson --base-url https://api.example.com

  # Convert with custom variables
  traffic2openapi convert postman -i collection.json -o api.ndjson --var apiVersion=v2 --var env=prod

  # Output as JSON batch instead of NDJSON
  traffic2openapi convert postman -i collection.json -o api.json --format batch

  # Generate OpenAPI directly
  traffic2openapi convert postman -i collection.json -o api.ndjson
  traffic2openapi generate -i api.ndjson -o openapi.yaml`,
	RunE: runPostmanConvert,
}

var (
	// Postman flags
	postmanInputPath      string
	postmanOutputPath     string
	postmanBaseURL        string
	postmanVariables      []string
	postmanIncludeHeaders bool
	postmanFilterHeaders  string
	postmanIncludeAuth    bool
	postmanOutputFormat   string
	postmanFilterHost     string
	postmanFilterMethod   string
)

func init() {
	convertCmd.AddCommand(postmanCmd)

	// Input/output flags
	postmanCmd.Flags().StringVarP(&postmanInputPath, "input", "i", "", "Input Postman collection file (required)")
	postmanCmd.Flags().StringVarP(&postmanOutputPath, "output", "o", "", "Output file path (default: stdout)")
	postmanCmd.Flags().StringVar(&postmanOutputFormat, "format", "ndjson", "Output format: ndjson or batch")

	// Variable flags
	postmanCmd.Flags().StringVar(&postmanBaseURL, "base-url", "", "Base URL for {{url}}/{{baseUrl}} variable resolution")
	postmanCmd.Flags().StringArrayVar(&postmanVariables, "var", []string{}, "Variable in key=value format (can be repeated)")

	// Filter flags
	postmanCmd.Flags().BoolVar(&postmanIncludeHeaders, "headers", true, "Include HTTP headers in output")
	postmanCmd.Flags().StringVar(&postmanFilterHeaders, "filter-headers", "", "Headers to filter out (comma-separated)")
	postmanCmd.Flags().BoolVar(&postmanIncludeAuth, "auth", true, "Convert Postman auth to headers")
	postmanCmd.Flags().StringVar(&postmanFilterHost, "host", "", "Only include requests to this host")
	postmanCmd.Flags().StringVar(&postmanFilterMethod, "method", "", "Only include requests with this method (GET, POST, etc.)")

	_ = postmanCmd.MarkFlagRequired("input")
}

func runPostmanConvert(cmd *cobra.Command, args []string) error {
	if postmanInputPath == "" {
		return fmt.Errorf("--input is required")
	}

	// Read the collection
	cmd.Printf("Reading Postman collection: %s\n", postmanInputPath)
	collection, err := postman.ReadFile(postmanInputPath)
	if err != nil {
		return fmt.Errorf("reading collection: %w", err)
	}

	// Configure converter
	converter := postman.NewConverter()
	configurePostmanConverter(converter)

	// Convert
	result, err := converter.Convert(collection)
	if err != nil {
		return fmt.Errorf("converting collection: %w", err)
	}

	records := result.Records

	// Apply post-conversion filters
	records = filterPostmanRecords(records)

	if len(records) == 0 {
		cmd.Printf("No records found\n")
		return nil
	}

	cmd.Printf("Converted %d records\n", len(records))

	// Print metadata summary
	if result.Metadata != nil {
		if result.Metadata.Title != nil {
			cmd.Printf("API: %s\n", *result.Metadata.Title)
		}
		if result.Metadata.APIVersion != nil {
			cmd.Printf("Version: %s\n", *result.Metadata.APIVersion)
		}
	}

	if len(result.TagDefinitions) > 0 {
		tags := make([]string, 0, len(result.TagDefinitions))
		for _, t := range result.TagDefinitions {
			tags = append(tags, t.Name)
		}
		cmd.Printf("Tags: %s\n", strings.Join(tags, ", "))
	}

	// Write output
	if postmanOutputPath == "" {
		// Write to stdout
		if postmanOutputFormat == "batch" {
			batch := ir.NewBatchWithMetadata(records, result.Metadata)
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(batch)
		}
		return ir.WriteNDJSON(os.Stdout, records)
	}

	// Write to file
	if postmanOutputFormat == "batch" {
		batch := ir.NewBatchWithMetadata(records, result.Metadata)
		data, err := json.MarshalIndent(batch, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling batch: %w", err)
		}
		if err := os.WriteFile(postmanOutputPath, data, 0600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		if err := ir.WriteFile(postmanOutputPath, records); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	}

	cmd.Printf("Wrote IR records to %s\n", postmanOutputPath)
	return nil
}

func configurePostmanConverter(converter *postman.Converter) {
	converter.IncludeHeaders = postmanIncludeHeaders
	converter.PreserveAuth = postmanIncludeAuth

	if postmanBaseURL != "" {
		converter.BaseURL = postmanBaseURL
	}

	// Parse variables
	for _, v := range postmanVariables {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			converter.Variables[parts[0]] = parts[1]
		}
	}

	// Parse filter headers
	if postmanFilterHeaders != "" {
		headers := strings.Split(postmanFilterHeaders, ",")
		for _, h := range headers {
			h = strings.TrimSpace(h)
			if h != "" {
				converter.FilterHeaders = append(converter.FilterHeaders, h)
			}
		}
	}
}

func filterPostmanRecords(records []ir.IRRecord) []ir.IRRecord {
	if postmanFilterHost == "" && postmanFilterMethod == "" {
		return records
	}

	filtered := make([]ir.IRRecord, 0, len(records))
	hostFilter := strings.ToLower(postmanFilterHost)
	methodFilter := strings.ToUpper(postmanFilterMethod)

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
