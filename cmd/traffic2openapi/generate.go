package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/grokify/traffic2openapi/pkg/inference"
	"github.com/grokify/traffic2openapi/pkg/ir"
	"github.com/grokify/traffic2openapi/pkg/openapi"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate OpenAPI spec from IR files",
	Long: `Generate an OpenAPI specification from Intermediate Representation (IR) files.

The command reads IR files (JSON batch or NDJSON streaming format) and
generates an OpenAPI 3.0, 3.1, or 3.2 specification through intelligent
inference of API structure.

Examples:
  # Generate from a directory of IR files
  traffic2openapi generate -i ./logs/ -o openapi.yaml

  # Generate from a single file
  traffic2openapi generate -i traffic.ndjson -o api.json

  # Generate OpenAPI 3.0 (for compatibility)
  traffic2openapi generate -i ./logs/ -o api.yaml --version 3.0

  # Specify API title and servers
  traffic2openapi generate -i ./logs/ -o api.yaml \
    --title "My API" \
    --api-version "2.0.0" \
    --server https://api.example.com`,
	RunE: runGenerate,
}

var (
	inputPath      string
	outputPath     string
	openAPIVersion string
	outputFormat   string
	apiTitle       string
	apiDescription string
	apiVersion     string
	servers        []string
	includeErrors  bool
	watchMode      bool
	watchDebounce  time.Duration
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input file or directory containing IR files (required)")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: stdout)")
	generateCmd.Flags().StringVarP(&openAPIVersion, "version", "v", "3.1", "OpenAPI version: 3.0, 3.1, or 3.2")
	generateCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "Output format: json or yaml (default: auto-detect from extension)")
	generateCmd.Flags().StringVar(&apiTitle, "title", "Generated API", "API title")
	generateCmd.Flags().StringVar(&apiDescription, "description", "", "API description")
	generateCmd.Flags().StringVar(&apiVersion, "api-version", "1.0.0", "API version")
	generateCmd.Flags().StringSliceVar(&servers, "server", nil, "Server URL (can be repeated)")
	generateCmd.Flags().BoolVar(&includeErrors, "include-errors", true, "Include 4xx/5xx error responses")
	generateCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "Watch for file changes and regenerate")
	generateCmd.Flags().DurationVar(&watchDebounce, "debounce", 500*time.Millisecond, "Debounce interval for watch mode")

	if err := generateCmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark input flag required: %v", err))
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// If watch mode, run with file watcher
	if watchMode {
		return runGenerateWatch(cmd)
	}

	return doGenerate(cmd)
}

func doGenerate(cmd *cobra.Command) error {
	// Validate input exists
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("input path error: %w", err)
	}

	// Read IR records
	var records []ir.IRRecord
	if info.IsDir() {
		records, err = ir.ReadDir(inputPath)
	} else {
		records, err = ir.ReadFile(inputPath)
	}
	if err != nil {
		return fmt.Errorf("reading IR files: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no records found in input")
	}

	cmd.Printf("Read %d IR records\n", len(records))

	// Configure inference engine
	engineOpts := inference.DefaultEngineOptions()
	engineOpts.IncludeErrorResponses = includeErrors

	// Run inference
	engine := inference.NewEngine(engineOpts)
	engine.ProcessRecords(records)
	result := engine.Finalize()

	cmd.Printf("Inferred %d endpoints\n", len(result.Endpoints))

	// Configure OpenAPI generator
	genOpts := openapi.GeneratorOptions{
		Title:       apiTitle,
		Description: apiDescription,
		APIVersion:  apiVersion,
		Servers:     servers,
	}

	// Set OpenAPI version
	switch openAPIVersion {
	case "3.0", "3.0.3":
		genOpts.Version = openapi.Version30
	case "3.1", "3.1.0":
		genOpts.Version = openapi.Version31
	case "3.2", "3.2.0":
		genOpts.Version = "3.2.0"
	default:
		return fmt.Errorf("unsupported OpenAPI version: %s (use 3.0, 3.1, or 3.2)", openAPIVersion)
	}

	// Generate spec
	spec := openapi.GenerateFromInference(result, genOpts)

	// Determine output format
	format := outputFormat
	if format == "" && outputPath != "" {
		ext := strings.ToLower(filepath.Ext(outputPath))
		switch ext {
		case ".json":
			format = "json"
		case ".yaml", ".yml":
			format = "yaml"
		default:
			format = "yaml"
		}
	}
	if format == "" {
		format = "yaml"
	}

	// Write output
	if outputPath == "" {
		// Write to stdout
		var output string
		if format == "json" {
			output, err = openapi.ToString(spec, openapi.FormatJSON)
		} else {
			output, err = openapi.ToString(spec, openapi.FormatYAML)
		}
		if err != nil {
			return fmt.Errorf("generating output: %w", err)
		}
		fmt.Print(output)
	} else {
		// Write to file
		if err := openapi.WriteFile(outputPath, spec); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		cmd.Printf("Wrote OpenAPI %s spec to %s\n", genOpts.Version, outputPath)
	}

	return nil
}

func runGenerateWatch(cmd *cobra.Command) error {
	// Require output path for watch mode
	if outputPath == "" {
		return fmt.Errorf("--output is required for watch mode")
	}

	// Initial generation
	cmd.Println("Starting watch mode...")
	if err := doGenerate(cmd); err != nil {
		cmd.Printf("Initial generation failed: %v\n", err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	// Add input path(s) to watcher
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("input path error: %w", err)
	}

	if info.IsDir() {
		// Watch directory and all files
		err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("walking directory: %w", err)
		}
		cmd.Printf("Watching directory: %s\n", inputPath)
	} else {
		if err := watcher.Add(inputPath); err != nil {
			return fmt.Errorf("adding file to watcher: %w", err)
		}
		cmd.Printf("Watching file: %s\n", inputPath)
	}

	cmd.Println("Press Ctrl+C to stop")

	// Debounce timer
	var debounceTimer *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only react to write/create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Skip non-IR files
			ext := strings.ToLower(filepath.Ext(event.Name))
			if ext != ".json" && ext != ".ndjson" {
				continue
			}

			// Debounce regeneration
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(watchDebounce, func() {
				cmd.Printf("\nFile changed: %s\n", event.Name)
				if err := doGenerate(cmd); err != nil {
					cmd.Printf("Generation failed: %v\n", err)
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			cmd.Printf("Watcher error: %v\n", err)
		}
	}
}
