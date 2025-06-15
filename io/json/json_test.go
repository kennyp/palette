package json

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kennyp/palette/color"
	"github.com/kennyp/palette/palette"
)

func TestImporter(t *testing.T) {
	importer := NewImporter()
	
	// Test CanImport
	if !importer.CanImport(".json") {
		t.Errorf("CanImport() should accept .json")
	}
	
	if importer.CanImport(".csv") {
		t.Errorf("CanImport() should not accept .csv")
	}
	
	// Test SupportedFormats
	formats := importer.SupportedFormats()
	if len(formats) != 1 || formats[0] != ".json" {
		t.Errorf("SupportedFormats() = %v, want [.json]", formats)
	}
}

func TestImportPaletteJSON(t *testing.T) {
	importer := NewImporter()
	
	jsonData := `{
		"name": "Test Palette",
		"description": "A test palette",
		"colors": [
			{
				"name": "Red",
				"rgb": {"r": 255, "g": 0, "b": 0}
			},
			{
				"name": "Green",
				"hex": "#00FF00"
			},
			{
				"name": "Blue",
				"cmyk": {"c": 100, "m": 100, "y": 0, "k": 0}
			}
		],
		"metadata": {
			"version": 1,
			"author": "test"
		}
	}`
	
	reader := strings.NewReader(jsonData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Name != "Test Palette" {
		t.Errorf("Import() name = %v, want Test Palette", p.Name)
	}
	
	if p.Description != "A test palette" {
		t.Errorf("Import() description = %v, want A test palette", p.Description)
	}
	
	if p.Len() != 3 {
		t.Errorf("Import() length = %d, want 3", p.Len())
	}
	
	// Check colors
	red, _ := p.Get(0)
	if red.Name != "Red" {
		t.Errorf("Import() color 0 name = %v, want Red", red.Name)
	}
	
	expectedRed := color.NewRGB(255, 0, 0)
	if red.Color.ToRGB() != expectedRed {
		t.Errorf("Import() color 0 = %v, want %v", red.Color, expectedRed)
	}
	
	// Check metadata
	if format, ok := p.GetMetadata("format"); !ok || format != "JSON" {
		t.Errorf("Import() should set format metadata")
	}
}

func TestImportColorArray(t *testing.T) {
	importer := NewImporter()
	
	jsonData := `[
		{
			"name": "Red",
			"rgb": {"r": 255, "g": 0, "b": 0}
		},
		{
			"name": "Green",
			"hex": "#00FF00"
		}
	]`
	
	reader := strings.NewReader(jsonData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Name != "JSON Color Array" {
		t.Errorf("Import() name = %v, want JSON Color Array", p.Name)
	}
	
	if p.Len() != 2 {
		t.Errorf("Import() length = %d, want 2", p.Len())
	}
}

func TestImportGenericJSON(t *testing.T) {
	importer := NewImporter()
	
	jsonData := `{
		"red": "#FF0000",
		"green": [0, 255, 0],
		"blue": "#0000FF",
		"other": "not a color"
	}`
	
	reader := strings.NewReader(jsonData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Name != "JSON Import" {
		t.Errorf("Import() name = %v, want JSON Import", p.Name)
	}
	
	// Should import 3 colors (red, green, blue) and ignore "other"
	if p.Len() != 3 {
		t.Errorf("Import() length = %d, want 3", p.Len())
	}
}

func TestImportHexColor(t *testing.T) {
	importer := NewImporter()
	
	tests := map[string]struct {
		hex      string
		expected color.RGB
	}{
		"Red":         {"#FF0000", color.NewRGB(255, 0, 0)},
		"Green":       {"#00FF00", color.NewRGB(0, 255, 0)},
		"Blue":        {"#0000FF", color.NewRGB(0, 0, 255)},
		"Without hash": {"FF0000", color.NewRGB(255, 0, 0)},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := importer.parseHexColor(tt.hex)
			if err != nil {
				t.Errorf("parseHexColor() error = %v", err)
			}
			
			if result.ToRGB() != tt.expected {
				t.Errorf("parseHexColor() = %v, want %v", result, tt.expected)
			}
		})
	}
	
	// Test invalid hex
	_, err := importer.parseHexColor("#GGGGGG")
	if err == nil {
		t.Errorf("parseHexColor() should error for invalid hex")
	}
	
	_, err = importer.parseHexColor("#FF00")
	if err == nil {
		t.Errorf("parseHexColor() should error for wrong length")
	}
}

func TestImportGenericValues(t *testing.T) {
	importer := NewImporter()
	
	tests := map[string]struct {
		values     any
		colorSpace string
		expected   color.Color
		shouldErr  bool
	}{
		"RGB float slice":  {[]any{255.0, 0.0, 0.0}, "RGB", color.NewRGB(255, 0, 0), false},
		"CMYK values":      {[]any{100.0, 0.0, 100.0, 0.0}, "CMYK", color.NewCMYK(100, 0, 100, 0), false},
		"HSB values":       {[]any{240.0, 100.0, 100.0}, "HSB", color.NewHSB(240, 100, 100), false},
		"LAB values":       {[]any{50.0, 20.0, -30.0}, "LAB", color.NewLAB(50, 20, -30), false},
		"Default to RGB":   {[]any{128.0, 64.0, 192.0}, "unknown", color.NewRGB(128, 64, 192), false},
		"Insufficient RGB": {[]any{255.0, 0.0}, "RGB", nil, true},
		"Invalid values":   {"not an array", "RGB", nil, true},
		"Non-numeric":      {[]any{"red", "green", "blue"}, "RGB", nil, true},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := importer.parseGenericValues(tt.values, tt.colorSpace)
			
			if tt.shouldErr {
				if err == nil {
					t.Errorf("parseGenericValues() should error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("parseGenericValues() error = %v", err)
			}
			
			if result.String() != tt.expected.String() {
				t.Errorf("parseGenericValues() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExporter(t *testing.T) {
	exporter := NewExporter()
	
	// Test CanExport
	if !exporter.CanExport(".json") {
		t.Errorf("CanExport() should accept .json")
	}
	
	if exporter.CanExport(".csv") {
		t.Errorf("CanExport() should not accept .csv")
	}
	
	// Test SupportedFormats
	formats := exporter.SupportedFormats()
	if len(formats) != 1 || formats[0] != ".json" {
		t.Errorf("SupportedFormats() = %v, want [.json]", formats)
	}
}

func TestExportBasic(t *testing.T) {
	exporter := NewExporter()
	
	p := palette.New("Test Palette")
	p.Description = "A test palette"
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewCMYK(100, 0, 100, 0), "Green")
	p.SetMetadata("version", 1)
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result := output.String()
	
	// Check that JSON contains expected fields
	if !strings.Contains(result, `"name": "Test Palette"`) {
		t.Errorf("Export() should contain palette name")
	}
	
	if !strings.Contains(result, `"description": "A test palette"`) {
		t.Errorf("Export() should contain description")
	}
	
	if !strings.Contains(result, `"Red"`) {
		t.Errorf("Export() should contain color name")
	}
	
	if !strings.Contains(result, `"rgb"`) {
		t.Errorf("Export() should contain RGB values by default")
	}
	
	if !strings.Contains(result, `"hex"`) {
		t.Errorf("Export() should contain hex values by default")
	}
	
	if !strings.Contains(result, `"metadata"`) {
		t.Errorf("Export() should contain metadata by default")
	}
}

func TestExportColorFormats(t *testing.T) {
	tests := map[string]struct {
		colorFormat   ColorFormatFlags
		shouldContain []string
		shouldNotContain []string
	}{
		"RGB only": {
			FormatRGB,
			[]string{`"rgb"`},
			[]string{`"hex"`, `"cmyk"`, `"hsb"`, `"lab"`},
		},
		"Hex only": {
			FormatHex,
			[]string{`"hex"`},
			[]string{`"rgb"`, `"cmyk"`, `"hsb"`, `"lab"`},
		},
		"All formats": {
			FormatAll,
			[]string{`"rgb"`, `"hex"`, `"cmyk"`, `"hsb"`, `"lab"`},
			[]string{},
		},
		"RGB and CMYK": {
			FormatRGB | FormatCMYK,
			[]string{`"rgb"`, `"cmyk"`},
			[]string{`"hex"`, `"hsb"`, `"lab"`},
		},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exporter := NewExporter()
			exporter.ColorFormat = tt.colorFormat
			
			p := palette.New("Test")
			p.Add(color.NewRGB(255, 0, 0), "Red")
			
			var output strings.Builder
			err := exporter.Export(p, &output)
			
			if err != nil {
				t.Errorf("Export() error = %v", err)
			}
			
			result := output.String()
			
			for _, should := range tt.shouldContain {
				if !strings.Contains(result, should) {
					t.Errorf("Export() should contain %s", should)
				}
			}
			
			for _, shouldNot := range tt.shouldNotContain {
				if strings.Contains(result, shouldNot) {
					t.Errorf("Export() should not contain %s", shouldNot)
				}
			}
		})
	}
}

func TestExportOptions(t *testing.T) {
	// Test without pretty print
	exporter := NewExporter()
	exporter.PrettyPrint = false
	
	p := palette.New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result := output.String()
	
	// Should not contain indentation
	if strings.Contains(result, "  ") {
		t.Errorf("Export() without pretty print should not contain indentation")
	}
	
	// Test without metadata
	exporter.IncludeMetadata = false
	p.SetMetadata("test", "value")
	
	output.Reset()
	err = exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result = output.String()
	
	if strings.Contains(result, `"metadata"`) {
		t.Errorf("Export() without metadata should not contain metadata field")
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can export and then import a palette
	original := palette.New("Round Trip Test")
	original.Description = "Testing round trip conversion"
	original.Add(color.NewRGB(255, 0, 0), "Red")
	original.Add(color.NewCMYK(100, 0, 100, 0), "Green")
	original.Add(color.NewHSB(240, 100, 100), "Blue")
	original.Add(color.NewLAB(50, 20, -30), "Gray")
	original.SetMetadata("version", 2)
	
	// Export
	exporter := NewExporter()
	exporter.ColorFormat = FormatAll // Include all color formats
	
	var exported strings.Builder
	err := exporter.Export(original, &exported)
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	// Import
	importer := NewImporter()
	reader := strings.NewReader(exported.String())
	imported, err := importer.Import(reader)
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	// Compare
	if imported.Name != original.Name {
		t.Errorf("Round trip name = %v, want %v", imported.Name, original.Name)
	}
	
	if imported.Description != original.Description {
		t.Errorf("Round trip description = %v, want %v", imported.Description, original.Description)
	}
	
	if imported.Len() != original.Len() {
		t.Errorf("Round trip length = %d, want %d", imported.Len(), original.Len())
	}
	
	// Check first color (RGB should be exact)
	origColor, _ := original.Get(0)
	impColor, _ := imported.Get(0)
	
	if origColor.Name != impColor.Name {
		t.Errorf("Round trip color name = %v, want %v", impColor.Name, origColor.Name)
	}
	
	if origColor.Color.ToRGB() != impColor.Color.ToRGB() {
		t.Errorf("Round trip color = %v, want %v", impColor.Color, origColor.Color)
	}
}

func TestInvalidJSON(t *testing.T) {
	importer := NewImporter()
	
	tests := map[string]string{
		"Invalid syntax":    `{"invalid": json}`,
		"Empty object":      `{}`,
		"Empty array":       `[]`,
		"No valid color":    `[{"invalid": "color"}]`,
	}
	
	for name, jsonData := range tests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(jsonData)
			_, err := importer.Import(reader)
			
			if err == nil {
				t.Errorf("Import() should error for invalid JSON: %s", jsonData)
			}
		})
	}
}

// Benchmark tests
func BenchmarkImport(b *testing.B) {
	importer := NewImporter()
	
	jsonData := `{
		"name": "Benchmark Palette",
		"colors": [
			{"name": "Red", "rgb": {"r": 255, "g": 0, "b": 0}},
			{"name": "Green", "rgb": {"r": 0, "g": 255, "b": 0}},
			{"name": "Blue", "rgb": {"r": 0, "g": 0, "b": 255}}
		]
	}`
	
	b.ResetTimer()
	for b.Loop() {
		reader := strings.NewReader(jsonData)
		_, _ = importer.Import(reader)
	}
}

func BenchmarkExport(b *testing.B) {
	exporter := NewExporter()
	
	p := palette.New("Benchmark")
	for i := range 100 {
		p.Add(color.NewRGB(uint8(i), 0, 0), fmt.Sprintf("Color%d", i))
	}
	
	b.ResetTimer()
	for b.Loop() {
		var output strings.Builder
		_ = exporter.Export(p, &output)
	}
}

func TestJSONErrorCases(t *testing.T) {
	importer := NewImporter()
	
	tests := map[string]struct {
		json   string
		hasErr bool
	}{
		"invalid_json": {
			json:   "{invalid json",
			hasErr: true,
		},
		"empty_object": {
			json:   "{}",
			hasErr: true, // No recognizable color data
		},
		"null_values": {
			json:   `{"colors": [{"name": "Red", "rgb": null}]}`,
			hasErr: true,
		},
		"invalid_hex": {
			json:   `{"colors": [{"name": "Red", "hex": "#XYZ"}]}`,
			hasErr: true,
		},
		"invalid_rgb_values": {
			json:   `{"colors": [{"name": "Red", "rgb": {"r": "invalid", "g": 0, "b": 0}}]}`,
			hasErr: true,
		},
		"insufficient_values": {
			json:   `{"colors": [{"name": "Red", "values": [255]}]}`,
			hasErr: true,
		},
		"empty_colors_array": {
			json:   `{"colors": []}`,
			hasErr: false, // Should create empty palette
		},
		"mixed_valid_invalid": {
			json: `{
				"colors": [
					{"name": "Red", "hex": "#FF0000"},
					{"name": "Invalid", "rgb": {"r": "bad", "g": 0, "b": 0}}
				]
			}`,
			hasErr: true, // Should fail on any invalid color
		},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(tt.json)
			_, err := importer.Import(reader)
			
			if tt.hasErr && err == nil {
				t.Errorf("Expected error for %s, got none", name)
			}
			if !tt.hasErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", name, err)
			}
		})
	}
}

func TestJSONGenericFormat(t *testing.T) {
	// Test parsing generic JSON objects that aren't structured palettes
	tests := map[string]struct {
		json   string
		hasErr bool
	}{
		"color_object": {
			json: `{
				"red": "#FF0000",
				"green": [0, 255, 0],
				"blue": {"r": 0, "g": 0, "b": 255}
			}`,
			hasErr: false,
		},
		"nested_invalid": {
			json: `{
				"colors": {
					"red": "invalid_color"
				}
			}`,
			hasErr: false, // Should skip invalid colors
		},
		"array_format": {
			json: `[
				{"name": "Red", "hex": "#FF0000"},
				{"name": "Green", "rgb": {"r": 0, "g": 255, "b": 0}}
			]`,
			hasErr: false,
		},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			importer := NewImporter()
			reader := strings.NewReader(tt.json)
			palette, err := importer.Import(reader)
			
			if tt.hasErr && err == nil {
				t.Errorf("Expected error for %s, got none", name)
			}
			if !tt.hasErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", name, err)
			}
			if !tt.hasErr && palette == nil {
				t.Errorf("Expected palette for %s, got nil", name)
			}
		})
	}
}

func TestJSONExporterOptions(t *testing.T) {
	p := palette.New("Test Palette")
	p.Description = "A test palette"
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	p.SetMetadata("source", "test")
	p.SetMetadata("version", "1.0")
	
	tests := map[string]struct {
		setupExporter func() *Exporter
		checkOutput   func(string) error
	}{
		"with_metadata": {
			setupExporter: func() *Exporter {
				e := NewExporter()
				e.IncludeMetadata = true
				return e
			},
			checkOutput: func(output string) error {
				if !strings.Contains(output, "metadata") {
					return fmt.Errorf("output should contain metadata")
				}
				if !strings.Contains(output, "source") {
					return fmt.Errorf("output should contain source metadata")
				}
				return nil
			},
		},
		"without_metadata": {
			setupExporter: func() *Exporter {
				e := NewExporter()
				e.IncludeMetadata = false
				return e
			},
			checkOutput: func(output string) error {
				if strings.Contains(output, "metadata") {
					return fmt.Errorf("output should not contain metadata")
				}
				return nil
			},
		},
		"pretty_format": {
			setupExporter: func() *Exporter {
				e := NewExporter()
				e.PrettyPrint = true
				return e
			},
			checkOutput: func(output string) error {
				if !strings.Contains(output, "\n") {
					return fmt.Errorf("pretty print should contain newlines")
				}
				return nil
			},
		},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exporter := tt.setupExporter()
			
			var output strings.Builder
			err := exporter.Export(p, &output)
			if err != nil {
				t.Errorf("Export failed: %v", err)
			}
			
			if err := tt.checkOutput(output.String()); err != nil {
				t.Errorf("Output check failed: %v", err)
			}
		})
	}
}

func TestJSONRoundTrip(t *testing.T) {
	// Test that we can export and re-import various color formats
	colors := []struct {
		name  string
		color color.Color
	}{
		{"Red", color.NewRGB(255, 0, 0)},
		{"Green", color.NewRGB(0, 255, 0)},
		{"Blue", color.NewRGB(0, 0, 255)},
		{"CMYK", color.NewCMYK(50, 25, 0, 0)},
		{"LAB", color.NewLAB(50, 20, -30)},
		{"HSB", color.NewHSB(240, 100, 100)},
	}
	
	original := palette.New("Round Trip Test")
	for _, c := range colors {
		original.Add(c.color, c.name)
	}
	
	// Export
	exporter := NewExporter()
	exporter.PrettyPrint = true
	var output strings.Builder
	err := exporter.Export(original, &output)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	
	// Re-import
	importer := NewImporter()
	reader := strings.NewReader(output.String())
	imported, err := importer.Import(reader)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	
	// Verify
	if imported.Len() != original.Len() {
		t.Errorf("Round trip changed palette length: got %d, want %d", imported.Len(), original.Len())
	}
	
	for i := range original.Len() {
		origColor, _ := original.Get(i)
		impColor, _ := imported.Get(i)
		
		if origColor.Name != impColor.Name {
			t.Errorf("Round trip changed color name at %d: got %s, want %s", i, impColor.Name, origColor.Name)
		}
		
		// Colors should be close (may not be exact due to conversions)
		if origColor.Color.ColorSpace() == "RGB" && impColor.Color.ColorSpace() == "RGB" {
			origRGB := origColor.Color.ToRGB()
			impRGB := impColor.Color.ToRGB()
			if origRGB != impRGB {
				t.Errorf("Round trip changed RGB color at %d: got %v, want %v", i, impRGB, origRGB)
			}
		}
	}
}