package io

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/kennyp/palette/color"
	"github.com/kennyp/palette/palette"
)

// Mock importer for testing
type mockImporter struct {
	formats []string
	palette *palette.Palette
	err     error
}

func (m *mockImporter) Import(r io.Reader) (*palette.Palette, error) {
	return m.palette, m.err
}

func (m *mockImporter) CanImport(format string) bool {
	for _, f := range m.formats {
		if f == format {
			return true
		}
	}
	return false
}

func (m *mockImporter) SupportedFormats() []string {
	return m.formats
}

// Mock exporter for testing
type mockExporter struct {
	formats []string
	output  string
	err     error
}

func (m *mockExporter) Export(p *palette.Palette, w io.Writer) error {
	if m.err != nil {
		return m.err
	}
	_, err := w.Write([]byte(m.output))
	return err
}

func (m *mockExporter) CanExport(format string) bool {
	for _, f := range m.formats {
		if f == format {
			return true
		}
	}
	return false
}

func (m *mockExporter) SupportedFormats() []string {
	return m.formats
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	
	if registry == nil {
		t.Errorf("NewRegistry() should return non-nil registry")
	}
	
	formats := registry.ListSupportedImportFormats()
	if len(formats) != 0 {
		t.Errorf("NewRegistry() should have no importers, got %d", len(formats))
	}
	
	formats = registry.ListSupportedExportFormats()
	if len(formats) != 0 {
		t.Errorf("NewRegistry() should have no exporters, got %d", len(formats))
	}
}

func TestRegisterImporter(t *testing.T) {
	registry := NewRegistry()
	importer := &mockImporter{
		formats: []string{".test", ".mock"},
		palette: palette.New("Test"),
	}
	
	registry.RegisterImporter(importer)
	
	formats := registry.ListSupportedImportFormats()
	if len(formats) != 2 {
		t.Errorf("RegisterImporter() should add formats, got %d formats", len(formats))
	}
	
	if !contains(formats, ".test") || !contains(formats, ".mock") {
		t.Errorf("RegisterImporter() formats = %v, want [.test .mock]", formats)
	}
}

func TestRegisterExporter(t *testing.T) {
	registry := NewRegistry()
	exporter := &mockExporter{
		formats: []string{".test", ".mock"},
		output:  "test output",
	}
	
	registry.RegisterExporter(exporter)
	
	formats := registry.ListSupportedExportFormats()
	if len(formats) != 2 {
		t.Errorf("RegisterExporter() should add formats, got %d formats", len(formats))
	}
	
	if !contains(formats, ".test") || !contains(formats, ".mock") {
		t.Errorf("RegisterExporter() formats = %v, want [.test .mock]", formats)
	}
}

func TestFindImporter(t *testing.T) {
	registry := NewRegistry()
	importer := &mockImporter{
		formats: []string{".test"},
		palette: palette.New("Test"),
	}
	
	registry.RegisterImporter(importer)
	
	// Test finding existing importer
	found, err := registry.FindImporter(".test")
	if err != nil {
		t.Errorf("FindImporter() error = %v", err)
	}
	
	if found != importer {
		t.Errorf("FindImporter() should return registered importer")
	}
	
	// Test format normalization - "test" should normalize to ".test" and be found
	_, err = registry.FindImporter("test")
	if err == nil {
		t.Errorf("FindImporter() should fail for 'test' since it normalizes to '.test' but we only registered '.test'")
	}
	
	// Test non-existent format
	_, err = registry.FindImporter(".nonexistent")
	if err == nil {
		t.Errorf("FindImporter() should error for non-existent format")
	}
}

func TestFindExporter(t *testing.T) {
	registry := NewRegistry()
	exporter := &mockExporter{
		formats: []string{".test"},
		output:  "test output",
	}
	
	registry.RegisterExporter(exporter)
	
	// Test finding existing exporter
	found, err := registry.FindExporter(".test")
	if err != nil {
		t.Errorf("FindExporter() error = %v", err)
	}
	
	if found != exporter {
		t.Errorf("FindExporter() should return registered exporter")
	}
	
	// Test format normalization - should fail for unknown format
	_, err = registry.FindExporter("test")
	if err == nil {
		t.Errorf("FindExporter() should fail for unknown format 'test'")
	}
	
	// Test non-existent format
	_, err = registry.FindExporter(".nonexistent")
	if err == nil {
		t.Errorf("FindExporter() should error for non-existent format")
	}
}

func TestImport(t *testing.T) {
	registry := NewRegistry()
	testPalette := palette.New("Test")
	testPalette.Add(color.NewRGB(255, 0, 0), "Red")
	
	importer := &mockImporter{
		formats: []string{".test"},
		palette: testPalette,
	}
	
	registry.RegisterImporter(importer)
	
	reader := strings.NewReader("test data")
	p, err := registry.Import(reader, ".test")
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Name != "Test" {
		t.Errorf("Import() palette name = %v, want Test", p.Name)
	}
	
	if p.Len() != 1 {
		t.Errorf("Import() palette length = %d, want 1", p.Len())
	}
}

func TestExport(t *testing.T) {
	registry := NewRegistry()
	exporter := &mockExporter{
		formats: []string{".test"},
		output:  "exported data",
	}
	
	registry.RegisterExporter(exporter)
	
	p := palette.New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	
	var buf bytes.Buffer
	err := registry.Export(p, &buf, ".test")
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	if buf.String() != "exported data" {
		t.Errorf("Export() output = %v, want 'exported data'", buf.String())
	}
}

func TestImportFromFile(t *testing.T) {
	registry := NewRegistry()
	testPalette := palette.New("Test")
	
	importer := &mockImporter{
		formats: []string{".test"},
		palette: testPalette,
	}
	
	registry.RegisterImporter(importer)
	
	reader := strings.NewReader("test data")
	p, err := registry.ImportFromFile("colors.test", reader)
	
	if err != nil {
		t.Errorf("ImportFromFile() error = %v", err)
	}
	
	if p.Name != "Test" {
		t.Errorf("ImportFromFile() palette name = %v, want Test", p.Name)
	}
	
	// Test file without extension
	_, err = registry.ImportFromFile("colors", reader)
	if err == nil {
		t.Errorf("ImportFromFile() should error for file without extension")
	}
}

func TestExportToFile(t *testing.T) {
	registry := NewRegistry()
	exporter := &mockExporter{
		formats: []string{".test"},
		output:  "exported data",
	}
	
	registry.RegisterExporter(exporter)
	
	p := palette.New("Test")
	
	var buf bytes.Buffer
	err := registry.ExportToFile(p, "output.test", &buf)
	
	if err != nil {
		t.Errorf("ExportToFile() error = %v", err)
	}
	
	if buf.String() != "exported data" {
		t.Errorf("ExportToFile() output = %v, want 'exported data'", buf.String())
	}
	
	// Test file without extension
	err = registry.ExportToFile(p, "output", &buf)
	if err == nil {
		t.Errorf("ExportToFile() should error for file without extension")
	}
}

func TestAutoDetectFormat(t *testing.T) {
	registry := NewRegistry()
	
	tests := map[string]struct {
		content  string
		expected string
	}{
		"Adobe Color Book": {"8BCBtest", ".acb"},
		"JSON object":      {`{"name": "test"}`, ".json"},
		"JSON array":       {`[{"color": "red"}]`, ".json"},
		"CSV":              {"name,r,g,b\nred,255,0,0", ".csv"},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			format, err := registry.AutoDetectFormat(reader)
			
			if err != nil {
				t.Errorf("AutoDetectFormat() error = %v", err)
			}
			
			if format != tt.expected {
				t.Errorf("AutoDetectFormat() = %v, want %v", format, tt.expected)
			}
		})
	}
	
	// Test insufficient data
	reader := strings.NewReader("ab")
	_, err := registry.AutoDetectFormat(reader)
	if err == nil {
		t.Errorf("AutoDetectFormat() should error for insufficient data")
	}
	
	// Test unknown format
	reader = strings.NewReader("unknown format data")
	_, err = registry.AutoDetectFormat(reader)
	if err == nil {
		t.Errorf("AutoDetectFormat() should error for unknown format")
	}
}

func TestNormalizeFormat(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"Extension with dot":    {".json", ".json"},
		"Extension without dot": {"json", ".json"},
		"ACB format":            {"acb", ".acb"},
		"Colorbook alias":       {"colorbook", ".acb"},
		"ACO format":            {"aco", ".aco"},
		"Colorswatch alias":     {"colorswatch", ".aco"},
		"Swatch alias":          {"swatch", ".aco"},
		"CSV format":            {"csv", ".csv"},
		"MIME type JSON":        {"application/json", ".json"},
		"MIME type CSV":         {"text/csv", ".csv"},
		"Unknown format":        {"unknown", "unknown"},
		"Case insensitive":      {"JSON", ".json"},
		"With spaces":           {"  json  ", ".json"},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := normalizeFormat(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeFormat(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Save original registry
	originalRegistry := DefaultRegistry
	defer func() { DefaultRegistry = originalRegistry }()
	
	// Create test registry
	DefaultRegistry = NewRegistry()
	testPalette := palette.New("Test")
	
	importer := &mockImporter{
		formats: []string{".test"},
		palette: testPalette,
	}
	exporter := &mockExporter{
		formats: []string{".test"},
		output:  "exported",
	}
	
	DefaultRegistry.RegisterImporter(importer)
	DefaultRegistry.RegisterExporter(exporter)
	
	// Test convenience Import function
	reader := strings.NewReader("test")
	p, err := Import(reader, ".test")
	if err != nil {
		t.Errorf("Import() convenience function error = %v", err)
	}
	if p.Name != "Test" {
		t.Errorf("Import() convenience function palette name = %v, want Test", p.Name)
	}
	
	// Test convenience Export function
	var buf bytes.Buffer
	err = Export(testPalette, &buf, ".test")
	if err != nil {
		t.Errorf("Export() convenience function error = %v", err)
	}
	if buf.String() != "exported" {
		t.Errorf("Export() convenience function output = %v, want exported", buf.String())
	}
	
	// Test convenience ImportFromFile function
	reader = strings.NewReader("test")
	p, err = ImportFromFile("test.test", reader)
	if err != nil {
		t.Errorf("ImportFromFile() convenience function error = %v", err)
	}
	
	// Test convenience ExportToFile function
	buf.Reset()
	err = ExportToFile(testPalette, "output.test", &buf)
	if err != nil {
		t.Errorf("ExportToFile() convenience function error = %v", err)
	}
}

func TestDeduplication(t *testing.T) {
	registry := NewRegistry()
	
	// Register multiple importers with overlapping formats
	importer1 := &mockImporter{formats: []string{".test", ".json"}}
	importer2 := &mockImporter{formats: []string{".json", ".csv"}}
	
	registry.RegisterImporter(importer1)
	registry.RegisterImporter(importer2)
	
	formats := registry.ListSupportedImportFormats()
	
	// Should not have duplicates
	seen := make(map[string]bool)
	for _, format := range formats {
		if seen[format] {
			t.Errorf("ListSupportedImportFormats() contains duplicate: %s", format)
		}
		seen[format] = true
	}
	
	// Should have all unique formats
	expected := []string{".test", ".json", ".csv"}
	if len(formats) != len(expected) {
		t.Errorf("ListSupportedImportFormats() length = %d, want %d", len(formats), len(expected))
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkFindImporter(b *testing.B) {
	registry := NewRegistry()
	
	// Register many importers
	for i := 0; i < 100; i++ {
		importer := &mockImporter{
			formats: []string{fmt.Sprintf(".test%d", i)},
		}
		registry.RegisterImporter(importer)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.FindImporter(".test50")
	}
}

func BenchmarkNormalizeFormat(b *testing.B) {
	formats := []string{"json", ".json", "application/json", "CSV", "  aco  "}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizeFormat(formats[i%len(formats)])
	}
}