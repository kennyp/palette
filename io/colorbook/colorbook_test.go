package colorbook

import (
	"strings"
	"testing"

	"github.com/kennyp/palette/color"
	"github.com/kennyp/palette/palette"
)

func TestNewImporter(t *testing.T) {
	importer := NewImporter()
	if importer == nil {
		t.Error("NewImporter() returned nil")
	}
}

func TestNewExporter(t *testing.T) {
	exporter := NewExporter()
	if exporter == nil {
		t.Error("NewExporter() returned nil")
	}
}

func TestImporterCanImport(t *testing.T) {
	importer := NewImporter()
	
	tests := map[string]struct {
		format   string
		expected bool
	}{
		"acb_extension":     {".acb", true},
		"ACB_extension":     {".ACB", true},
		"colorbook_format":  {"colorbook", true},
		"json_format":       {".json", false},
		"csv_format":        {".csv", false},
		"unknown_format":    {".xyz", false},
		"empty_format":      {"", false},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := importer.CanImport(tt.format)
			if result != tt.expected {
				t.Errorf("CanImport(%s) = %v, want %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestExporterCanExport(t *testing.T) {
	exporter := NewExporter()
	
	tests := map[string]struct {
		format   string
		expected bool
	}{
		"acb_extension":     {".acb", true},
		"ACB_extension":     {".ACB", true},
		"colorbook_format":  {"colorbook", true},
		"json_format":       {".json", false},
		"csv_format":        {".csv", false},
		"unknown_format":    {".xyz", false},
		"empty_format":      {"", false},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := exporter.CanExport(tt.format)
			if result != tt.expected {
				t.Errorf("CanExport(%s) = %v, want %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestSupportedFormats(t *testing.T) {
	importer := NewImporter()
	exporter := NewExporter()
	
	importFormats := importer.SupportedFormats()
	exportFormats := exporter.SupportedFormats()
	
	expectedFormats := []string{".acb", "colorbook"}
	
	if len(importFormats) != len(expectedFormats) {
		t.Errorf("Importer supported formats length = %d, want %d", len(importFormats), len(expectedFormats))
	}
	
	if len(exportFormats) != len(expectedFormats) {
		t.Errorf("Exporter supported formats length = %d, want %d", len(exportFormats), len(expectedFormats))
	}
	
	for _, expected := range expectedFormats {
		found := false
		for _, format := range importFormats {
			if format == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %s not found in importer formats", expected)
		}
	}
}

func TestExportBasic(t *testing.T) {
	// Test basic export functionality
	p := palette.New("Test Palette")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	p.Add(color.NewRGB(0, 0, 255), "Blue")
	
	exporter := NewExporter()
	var output strings.Builder
	
	err := exporter.Export(p, &output)
	if err != nil {
		t.Errorf("Export failed: %v", err)
	}
	
	result := output.String()
	if len(result) == 0 {
		t.Error("Export produced empty output")
	}
	
	// ACB files are binary, so we just check that something was written
	if len(result) < 50 { // ACB files should be at least this big with header
		t.Errorf("Export output seems too small: %d bytes", len(result))
	}
}

func TestExportEmptyPalette(t *testing.T) {
	// Test exporting an empty palette
	p := palette.New("Empty Palette")
	
	exporter := NewExporter()
	var output strings.Builder
	
	err := exporter.Export(p, &output)
	if err != nil {
		t.Errorf("Export of empty palette failed: %v", err)
	}
	
	result := output.String()
	if len(result) == 0 {
		t.Error("Export of empty palette produced no output")
	}
}

func TestImportInvalidData(t *testing.T) {
	importer := NewImporter()
	
	tests := map[string]struct {
		data string
	}{
		"empty_data":    {""},
		"invalid_data":  {"not a valid ACB file"},
		"short_data":    {"8BCB"}, // Valid header but incomplete
		"wrong_header":  {"INVALID_HEADER_DATA"},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			_, err := importer.Import(reader)
			
			// We expect all of these to fail
			if err == nil {
				t.Errorf("Expected error for %s, got none", name)
			}
		})
	}
}

func TestColorConversion(t *testing.T) {
	// Test that we handle different color spaces appropriately
	testColors := []struct {
		name  string
		color color.Color
	}{
		{"RGB Red", color.NewRGB(255, 0, 0)},
		{"CMYK Cyan", color.NewCMYK(100, 0, 0, 0)},
		{"LAB White", color.NewLAB(100, 0, 0)},
		{"HSB Blue", color.NewHSB(240, 100, 100)},
	}
	
	for _, tc := range testColors {
		t.Run(tc.name, func(t *testing.T) {
			p := palette.New("Test")
			p.Add(tc.color, tc.name)
			
			exporter := NewExporter()
			var output strings.Builder
			
			err := exporter.Export(p, &output)
			if err != nil {
				t.Errorf("Failed to export %s: %v", tc.name, err)
			}
			
			if output.Len() == 0 {
				t.Errorf("Export of %s produced no output", tc.name)
			}
		})
	}
}