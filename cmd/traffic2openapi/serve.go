package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/grokify/traffic2openapi/pkg/openapi"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve <spec-file>",
	Short: "Serve OpenAPI spec with interactive documentation",
	Long: `Serve an OpenAPI specification with interactive documentation UI.

Supports Swagger UI and Redoc for browsing and testing the API.

Examples:
  # Serve with Swagger UI (default)
  traffic2openapi serve openapi.yaml

  # Serve on a specific port
  traffic2openapi serve openapi.yaml --port 8080

  # Serve with Redoc
  traffic2openapi serve openapi.yaml --ui redoc

  # Auto-reload when spec changes
  traffic2openapi serve openapi.yaml --watch`,
	Args: cobra.ExactArgs(1),
	RunE: runServe,
}

var (
	servePort  int
	serveUI    string
	serveWatch bool
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to serve on")
	serveCmd.Flags().StringVar(&serveUI, "ui", "swagger", "Documentation UI: swagger or redoc")
	serveCmd.Flags().BoolVarP(&serveWatch, "watch", "w", false, "Watch for file changes and auto-reload")
}

// HTML templates for documentation UIs
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.Title}} - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "/spec.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`

const redocHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.Title}} - Redoc</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
  <style>
    body { margin: 0; padding: 0; }
  </style>
</head>
<body>
  <redoc spec-url='/spec.json'></redoc>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`

type templateData struct {
	Title string
}

func runServe(cmd *cobra.Command, args []string) error {
	specPath := args[0]

	// Validate spec file exists
	if _, err := os.Stat(specPath); err != nil {
		return fmt.Errorf("spec file error: %w", err)
	}

	// Read spec to get title
	spec, err := openapi.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("reading spec: %w", err)
	}

	title := spec.Info.Title
	if title == "" {
		title = "API Documentation"
	}

	// Create HTTP handlers
	mux := http.NewServeMux()

	// Serve the spec as JSON
	mux.HandleFunc("/spec.json", func(w http.ResponseWriter, r *http.Request) {
		// Re-read spec each time if watching
		currentSpec := spec
		if serveWatch {
			var readErr error
			currentSpec, readErr = openapi.ReadFile(specPath)
			if readErr != nil {
				http.Error(w, readErr.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		data, err := openapi.ToJSON(currentSpec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	})

	// Serve the spec as YAML
	mux.HandleFunc("/spec.yaml", func(w http.ResponseWriter, r *http.Request) {
		currentSpec := spec
		if serveWatch {
			var readErr error
			currentSpec, readErr = openapi.ReadFile(specPath)
			if readErr != nil {
				http.Error(w, readErr.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		data, err := openapi.ToYAML(currentSpec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	})

	// Serve the documentation UI
	var htmlTemplate string
	switch serveUI {
	case "redoc":
		htmlTemplate = redocHTML
	default:
		htmlTemplate = swaggerUIHTML
	}

	tmpl, err := template.New("ui").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		data := templateData{Title: title}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Start server
	addr := fmt.Sprintf(":%d", servePort)
	cmd.Printf("Serving %s at http://localhost%s\n", filepath.Base(specPath), addr)
	cmd.Printf("UI: %s\n", serveUI)
	if serveWatch {
		cmd.Println("Watching for file changes...")
	}
	cmd.Println("\nPress Ctrl+C to stop")

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return server.ListenAndServe()
}
