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
	"github.com/grokify/traffic2openapi/pkg/openapi/convert"
	"github.com/grokify/traffic2openapi/pkg/openapi/validate"
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

  # Generate multiple versions at once
  traffic2openapi generate -i ./logs/ -o api.yaml --versions 3.0,3.1,3.2

  # Generate all supported versions
  traffic2openapi generate -i ./logs/ -o api.yaml --all-versions

  # Specify API title and servers
  traffic2openapi generate -i ./logs/ -o api.yaml \
    --title "My API" \
    --api-version "2.0.0" \
    --server https://api.example.com

  # Skip validation for faster generation
  traffic2openapi generate -i ./logs/ -o api.yaml --skip-validation`,
	RunE: runGenerate,
}

var (
	inputPath       string
	outputPath      string
	openAPIVersion  string
	openAPIVersions []string
	allVersions     bool
	outputFormat    string
	apiTitle        string
	apiDescription  string
	apiVersion      string
	servers         []string
	includeErrors   bool
	watchMode       bool
	watchDebounce   time.Duration
	skipValidation  bool
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input file or directory containing IR files (required)")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: stdout)")
	generateCmd.Flags().StringVarP(&openAPIVersion, "version", "v", "3.1", "OpenAPI version: 3.0, 3.1, or 3.2")
	generateCmd.Flags().StringSliceVar(&openAPIVersions, "versions", nil, "Multiple OpenAPI versions (comma-separated: 3.0,3.1,3.2)")
	generateCmd.Flags().BoolVar(&allVersions, "all-versions", false, "Generate all supported versions (3.0.3, 3.1.0, 3.2.0)")
	generateCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "Output format: json or yaml (default: auto-detect from extension)")
	generateCmd.Flags().StringVar(&apiTitle, "title", "Generated API", "API title")
	generateCmd.Flags().StringVar(&apiDescription, "description", "", "API description")
	generateCmd.Flags().StringVar(&apiVersion, "api-version", "1.0.0", "API version")
	generateCmd.Flags().StringSliceVar(&servers, "server", nil, "Server URL (can be repeated)")
	generateCmd.Flags().BoolVar(&includeErrors, "include-errors", true, "Include 4xx/5xx error responses")
	generateCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "Watch for file changes and regenerate")
	generateCmd.Flags().DurationVar(&watchDebounce, "debounce", 500*time.Millisecond, "Debounce interval for watch mode")
	generateCmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "Skip validation of generated spec")

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

// parseTargetVersions parses version strings into TargetVersion values.
func parseTargetVersions(versions []string) ([]convert.TargetVersion, error) {
	var targets []convert.TargetVersion
	for _, v := range versions {
		v = strings.TrimSpace(v)
		switch v {
		case "3.0", "3.0.3":
			targets = append(targets, convert.Version303)
		case "3.1", "3.1.0":
			targets = append(targets, convert.Version310)
		case "3.2", "3.2.0":
			targets = append(targets, convert.Version320)
		default:
			return nil, fmt.Errorf("unsupported version: %s", v)
		}
	}
	return targets, nil
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

	// Check if multi-version output is requested
	if allVersions || len(openAPIVersions) > 0 {
		return doGenerateMultiVersion(cmd, result)
	}

	// Single version output
	return doGenerateSingleVersion(cmd, result)
}

func doGenerateSingleVersion(cmd *cobra.Command, result *inference.InferenceResult) error {
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
		genOpts.Version = openapi.Version32
	default:
		return fmt.Errorf("unsupported OpenAPI version: %s (use 3.0, 3.1, or 3.2)", openAPIVersion)
	}

	// Generate spec
	spec := openapi.GenerateFromInference(result, genOpts)

	// Validate spec unless skipped
	if !skipValidation {
		if err := validateSpec(cmd, spec); err != nil {
			return err
		}
	}

	// Determine output format
	format := getOutputFormat()

	// Write output
	if outputPath == "" {
		// Write to stdout
		var output string
		var err error
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

func doGenerateMultiVersion(cmd *cobra.Command, result *inference.InferenceResult) error {
	// Require output path for multi-version
	if outputPath == "" {
		return fmt.Errorf("--output is required for multi-version output")
	}

	// Determine target versions
	var targets []convert.TargetVersion
	if allVersions {
		targets = []convert.TargetVersion{
			convert.Version303,
			convert.Version310,
			convert.Version320,
		}
	} else {
		var err error
		targets, err = parseTargetVersions(openAPIVersions)
		if err != nil {
			return err
		}
	}

	if len(targets) == 0 {
		return fmt.Errorf("no target versions specified")
	}

	// Generate base spec (use 3.1 as canonical format)
	genOpts := openapi.GeneratorOptions{
		Title:       apiTitle,
		Description: apiDescription,
		APIVersion:  apiVersion,
		Servers:     servers,
		Version:     openapi.Version31,
	}
	spec := openapi.GenerateFromInference(result, genOpts)

	// Convert to multiple versions
	output, err := convert.NewMultiVersionOutput(spec, targets...)
	if err != nil {
		return fmt.Errorf("converting versions: %w", err)
	}

	// Validate each version unless skipped
	if !skipValidation {
		for _, version := range output.Versions() {
			versionSpec := output.Get(version)
			if err := validateSpec(cmd, versionSpec); err != nil {
				return fmt.Errorf("validation failed for %s: %w", version, err)
			}
		}
	}

	// Determine output format and write files
	format := getOutputFormat()
	var oaFormat openapi.Format
	if format == "json" {
		oaFormat = openapi.FormatJSON
	} else {
		oaFormat = openapi.FormatYAML
	}

	// Get output directory and base filename
	dir := filepath.Dir(outputPath)
	ext := filepath.Ext(outputPath)
	base := strings.TrimSuffix(filepath.Base(outputPath), ext)

	if err := output.WriteFilesToDir(dir, base, oaFormat); err != nil {
		return fmt.Errorf("writing output files: %w", err)
	}

	// Report what was written
	for _, version := range output.Versions() {
		filename := convert.VersionedFilename(filepath.Base(outputPath), version)
		cmd.Printf("Wrote OpenAPI %s spec to %s\n", version, filepath.Join(dir, filename))
	}

	return nil
}

func getOutputFormat() string {
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
	return format
}

// validateSpec validates the generated OpenAPI spec using libopenapi.
func validateSpec(cmd *cobra.Command, spec *openapi.Spec) error {
	// Render to YAML for validation
	yamlBytes, err := openapi.ToYAML(spec)
	if err != nil {
		return fmt.Errorf("rendering spec for validation: %w", err)
	}

	result, err := validate.Validate(yamlBytes)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid {
		cmd.PrintErrln("Validation failed:")
		for _, e := range result.Errors {
			cmd.PrintErrf("  ERROR: %s\n", e.Message)
		}
		return fmt.Errorf("generated spec failed validation with %d error(s)", len(result.Errors))
	}

	// Show warnings if any
	if len(result.Warnings) > 0 {
		cmd.Printf("Validation passed with %d warning(s)\n", len(result.Warnings))
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
