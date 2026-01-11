package main

import (
	"fmt"

	"github.com/grokify/traffic2openapi/pkg/sitegen"
	"github.com/spf13/cobra"
)

var siteCmd = &cobra.Command{
	Use:   "site",
	Short: "Generate static HTML documentation site from IR files",
	Long: `Generate a static HTML website from Intermediate Representation (IR) files.

The site displays API traffic grouped by endpoint (method + path template),
with both deduplicated and distinct views of captured requests/responses.

Features:
  - Index page with all endpoints
  - Per-endpoint pages with request/response details
  - Deduped view showing all captured parameter values
  - Distinct view showing individual requests
  - Light/dark mode toggle
  - Syntax highlighting for JSON bodies
  - Copy buttons for code blocks

Examples:
  # Generate site from IR file
  traffic2openapi site -i traffic.ndjson -o ./site/

  # Generate site from directory of IR files
  traffic2openapi site -i ./logs/ -o ./docs/api/

  # Custom title
  traffic2openapi site -i traffic.ndjson -o ./site/ --title "My API Docs"`,
	RunE: runSite,
}

var (
	siteInputPath  string
	siteOutputPath string
	siteTitle      string
	siteBaseURL    string
)

func init() {
	rootCmd.AddCommand(siteCmd)

	siteCmd.Flags().StringVarP(&siteInputPath, "input", "i", "", "Input IR file or directory (required)")
	siteCmd.Flags().StringVarP(&siteOutputPath, "output", "o", "./site/", "Output directory for generated site")
	siteCmd.Flags().StringVar(&siteTitle, "title", "API Traffic Documentation", "Site title")
	siteCmd.Flags().StringVar(&siteBaseURL, "base-url", "", "Base URL for links (e.g., /docs/api/)")

	if err := siteCmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark input flag required: %v", err))
	}
}

func runSite(cmd *cobra.Command, args []string) error {
	opts := &sitegen.Options{
		Title:   siteTitle,
		BaseURL: siteBaseURL,
	}

	cmd.Printf("Reading IR files from %s...\n", siteInputPath)

	if err := sitegen.GenerateFromFile(siteInputPath, siteOutputPath, opts); err != nil {
		return fmt.Errorf("generating site: %w", err)
	}

	cmd.Printf("Site generated successfully at %s\n", siteOutputPath)
	return nil
}
