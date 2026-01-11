package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format represents the output format.
type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// WriteFile writes the spec to a file.
// Format is determined by file extension (.json or .yaml/.yml).
func WriteFile(path string, spec *Spec) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return WriteYAML(f, spec)
	case ".json":
		return WriteJSON(f, spec)
	default:
		// Default to YAML
		return WriteYAML(f, spec)
	}
}

// WriteJSON writes the spec as JSON.
func WriteJSON(w io.Writer, spec *Spec) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(spec); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}

// WriteYAML writes the spec as YAML.
func WriteYAML(w io.Writer, spec *Spec) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	if err := encoder.Encode(spec); err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	return nil
}

// ToJSON converts the spec to JSON bytes.
func ToJSON(spec *Spec) ([]byte, error) {
	return json.MarshalIndent(spec, "", "  ")
}

// ToYAML converts the spec to YAML bytes.
func ToYAML(spec *Spec) ([]byte, error) {
	return yaml.Marshal(spec)
}

// ToString converts the spec to a string in the specified format.
func ToString(spec *Spec, format Format) (string, error) {
	var data []byte
	var err error

	switch format {
	case FormatJSON:
		data, err = ToJSON(spec)
	case FormatYAML:
		data, err = ToYAML(spec)
	default:
		data, err = ToYAML(spec)
	}

	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadFile reads an OpenAPI spec from a file.
// Format is determined by file extension (.json or .yaml/.yml).
func ReadFile(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return FromYAML(data)
	case ".json":
		return FromJSON(data)
	default:
		// Try YAML first, then JSON
		spec, err := FromYAML(data)
		if err == nil {
			return spec, nil
		}
		return FromJSON(data)
	}
}

// FromJSON parses a spec from JSON bytes.
func FromJSON(data []byte) (*Spec, error) {
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return &spec, nil
}

// FromYAML parses a spec from YAML bytes.
func FromYAML(data []byte) (*Spec, error) {
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	return &spec, nil
}
