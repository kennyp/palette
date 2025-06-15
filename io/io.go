package io

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kennyp/palette/palette"
)

// Importer defines the interface for importing palettes from various formats.
type Importer interface {
	// Import reads a palette from the given reader.
	Import(r io.Reader) (*palette.Palette, error)

	// CanImport returns true if this importer can handle the given format.
	// The format parameter is typically a file extension or MIME type.
	CanImport(format string) bool

	// SupportedFormats returns a list of supported formats (extensions/MIME types).
	SupportedFormats() []string
}

// Exporter defines the interface for exporting palettes to various formats.
type Exporter interface {
	// Export writes a palette to the given writer.
	Export(p *palette.Palette, w io.Writer) error

	// CanExport returns true if this exporter can handle the given format.
	// The format parameter is typically a file extension or MIME type.
	CanExport(format string) bool

	// SupportedFormats returns a list of supported formats (extensions/MIME types).
	SupportedFormats() []string
}

// Registry manages importers and exporters for different formats.
type Registry struct {
	importers []Importer
	exporters []Exporter
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		importers: make([]Importer, 0),
		exporters: make([]Exporter, 0),
	}
}

// RegisterImporter adds an importer to the registry.
func (r *Registry) RegisterImporter(importer Importer) {
	r.importers = append(r.importers, importer)
}

// RegisterExporter adds an exporter to the registry.
func (r *Registry) RegisterExporter(exporter Exporter) {
	r.exporters = append(r.exporters, exporter)
}

// FindImporter finds an importer that can handle the given format.
func (r *Registry) FindImporter(format string) (Importer, error) {
	normalizedFormat := normalizeFormat(format)

	for _, importer := range r.importers {
		if importer.CanImport(normalizedFormat) {
			return importer, nil
		}
	}

	return nil, fmt.Errorf("no importer found for format: %s", format)
}

// FindExporter finds an exporter that can handle the given format.
func (r *Registry) FindExporter(format string) (Exporter, error) {
	normalizedFormat := normalizeFormat(format)

	for _, exporter := range r.exporters {
		if exporter.CanExport(normalizedFormat) {
			return exporter, nil
		}
	}

	return nil, fmt.Errorf("no exporter found for format: %s", format)
}

// Import imports a palette using the appropriate importer for the given format.
func (r *Registry) Import(reader io.Reader, format string) (*palette.Palette, error) {
	importer, err := r.FindImporter(format)
	if err != nil {
		return nil, err
	}

	return importer.Import(reader)
}

// Export exports a palette using the appropriate exporter for the given format.
func (r *Registry) Export(p *palette.Palette, writer io.Writer, format string) error {
	exporter, err := r.FindExporter(format)
	if err != nil {
		return err
	}

	return exporter.Export(p, writer)
}

// ImportFromFile imports a palette from a file, detecting the format from the file extension.
func (r *Registry) ImportFromFile(filename string, reader io.Reader) (*palette.Palette, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return nil, fmt.Errorf("cannot determine format from filename: %s", filename)
	}

	return r.Import(reader, ext)
}

// ExportToFile exports a palette to a file, detecting the format from the file extension.
func (r *Registry) ExportToFile(p *palette.Palette, filename string, writer io.Writer) error {
	ext := filepath.Ext(filename)
	if ext == "" {
		return fmt.Errorf("cannot determine format from filename: %s", filename)
	}

	return r.Export(p, writer, ext)
}

// ListSupportedImportFormats returns a list of all supported import formats.
func (r *Registry) ListSupportedImportFormats() []string {
	var formats []string
	seen := make(map[string]bool)

	for _, importer := range r.importers {
		for _, format := range importer.SupportedFormats() {
			normalized := normalizeFormat(format)
			if !seen[normalized] {
				formats = append(formats, normalized)
				seen[normalized] = true
			}
		}
	}

	return formats
}

// ListSupportedExportFormats returns a list of all supported export formats.
func (r *Registry) ListSupportedExportFormats() []string {
	var formats []string
	seen := make(map[string]bool)

	for _, exporter := range r.exporters {
		for _, format := range exporter.SupportedFormats() {
			normalized := normalizeFormat(format)
			if !seen[normalized] {
				formats = append(formats, normalized)
				seen[normalized] = true
			}
		}
	}

	return formats
}

// AutoDetectFormat attempts to detect the format from the content.
// This is a basic implementation that can be extended with magic number detection.
func (r *Registry) AutoDetectFormat(reader io.Reader) (string, error) {
	// Read the first few bytes to detect magic numbers
	buffer := make([]byte, 16)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read header: %w", err)
	}

	if n < 4 {
		return "", fmt.Errorf("insufficient data to detect format")
	}

	// Check for known magic numbers
	header := string(buffer[:4])
	switch header {
	case "8BCB": // Adobe Color Book
		return ".acb", nil
	default:
		// Check for JSON (starts with '{' or '[')
		if buffer[0] == '{' || buffer[0] == '[' {
			return ".json", nil
		}

		// Check for CSV (first line might contain commas)
		firstLine := strings.Split(string(buffer[:n]), "\n")[0]
		if strings.Contains(firstLine, ",") {
			return ".csv", nil
		}
	}

	return "", fmt.Errorf("unable to detect format from content")
}

// DefaultRegistry returns a registry with all built-in importers and exporters registered.
var DefaultRegistry = NewRegistry()

// Convenience functions that use the default registry

// Import imports a palette using the default registry.
func Import(reader io.Reader, format string) (*palette.Palette, error) {
	return DefaultRegistry.Import(reader, format)
}

// Export exports a palette using the default registry.
func Export(p *palette.Palette, writer io.Writer, format string) error {
	return DefaultRegistry.Export(p, writer, format)
}

// ImportFromFile imports a palette from a file using the default registry.
func ImportFromFile(filename string, reader io.Reader) (*palette.Palette, error) {
	return DefaultRegistry.ImportFromFile(filename, reader)
}

// ExportToFile exports a palette to a file using the default registry.
func ExportToFile(p *palette.Palette, filename string, writer io.Writer) error {
	return DefaultRegistry.ExportToFile(p, filename, writer)
}

// Helper functions

// normalizeFormat normalizes a format string (file extension or MIME type).
func normalizeFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))

	// Handle file extensions
	if strings.HasPrefix(format, ".") {
		return format
	}

	// Handle common extensions without dot
	switch format {
	case "acb", "colorbook":
		return ".acb"
	case "aco", "colorswatch", "swatch":
		return ".aco"
	case "csv":
		return ".csv"
	case "json":
		return ".json"
	}

	// Handle MIME types
	switch format {
	case "application/json":
		return ".json"
	case "text/csv":
		return ".csv"
	}

	// Return as-is if we don't recognize it
	return format
}
