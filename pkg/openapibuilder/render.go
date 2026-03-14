package openapibuilder

import (
	"io"

	"github.com/grokify/traffic2openapi/pkg/openapi"
)

// ToJSON converts the spec to JSON bytes.
func ToJSON(spec *openapi.Spec) ([]byte, error) {
	return openapi.ToJSON(spec)
}

// ToYAML converts the spec to YAML bytes.
func ToYAML(spec *openapi.Spec) ([]byte, error) {
	return openapi.ToYAML(spec)
}

// WriteJSON writes the spec as JSON to the given writer.
func WriteJSON(w io.Writer, spec *openapi.Spec) error {
	return openapi.WriteJSON(w, spec)
}

// WriteYAML writes the spec as YAML to the given writer.
func WriteYAML(w io.Writer, spec *openapi.Spec) error {
	return openapi.WriteYAML(w, spec)
}

// WriteFile writes the spec to a file.
// Format is determined by file extension (.json or .yaml/.yml).
func WriteFile(path string, spec *openapi.Spec) error {
	return openapi.WriteFile(path, spec)
}

// ToString converts the spec to a string in the specified format.
func ToString(spec *openapi.Spec, format openapi.Format) (string, error) {
	return openapi.ToString(spec, format)
}
