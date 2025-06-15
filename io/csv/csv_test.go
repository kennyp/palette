package csv

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
	if !importer.CanImport(".csv") {
		t.Errorf("CanImport() should accept .csv")
	}
	
	if importer.CanImport(".json") {
		t.Errorf("CanImport() should not accept .json")
	}
	
	// Test SupportedFormats
	formats := importer.SupportedFormats()
	if len(formats) != 1 || formats[0] != ".csv" {
		t.Errorf("SupportedFormats() = %v, want [.csv]", formats)
	}
}

func TestImportRGB(t *testing.T) {
	importer := NewImporter()
	
	csvData := `Name,R,G,B
Red,255,0,0
Green,0,255,0
Blue,0,0,255`
	
	reader := strings.NewReader(csvData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Name != "CSV Import" {
		t.Errorf("Import() name = %v, want CSV Import", p.Name)
	}
	
	if p.Len() != 3 {
		t.Errorf("Import() length = %d, want 3", p.Len())
	}
	
	// Check first color
	red, _ := p.Get(0)
	if red.Name != "Red" {
		t.Errorf("Import() color 0 name = %v, want Red", red.Name)
	}
	
	expectedRed := color.NewRGB(255, 0, 0)
	if red.Color.ToRGB() != expectedRed {
		t.Errorf("Import() color 0 = %v, want %v", red.Color, expectedRed)
	}
}

func TestImportHex(t *testing.T) {
	importer := NewImporter()
	
	csvData := `Name,Hex
Red,#FF0000
Green,#00FF00
Blue,#0000FF`
	
	reader := strings.NewReader(csvData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Len() != 3 {
		t.Errorf("Import() length = %d, want 3", p.Len())
	}
	
	// Check hex color parsing
	red, _ := p.Get(0)
	expectedRed := color.NewRGB(255, 0, 0)
	if red.Color.ToRGB() != expectedRed {
		t.Errorf("Import() hex color = %v, want %v", red.Color, expectedRed)
	}
}

func TestImportCMYK(t *testing.T) {
	importer := NewImporter()
	importer.ColorFormat = FormatCMYK
	
	csvData := `Name,C,M,Y,K
Cyan,100,0,0,0
Magenta,0,100,0,0
Yellow,0,0,100,0
Black,0,0,0,100`
	
	reader := strings.NewReader(csvData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Len() != 4 {
		t.Errorf("Import() length = %d, want 4", p.Len())
	}
	
	// Check CMYK color
	cyan, _ := p.Get(0)
	expectedCyan := color.NewCMYK(100, 0, 0, 0)
	if cyan.Color.ToCMYK() != expectedCyan {
		t.Errorf("Import() CMYK color = %v, want %v", cyan.Color, expectedCyan)
	}
}

func TestImportHSB(t *testing.T) {
	importer := NewImporter()
	importer.ColorFormat = FormatHSB
	
	csvData := `Name,H,S,B
Red,0,100,100
Green,120,100,100
Blue,240,100,100`
	
	reader := strings.NewReader(csvData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Len() != 3 {
		t.Errorf("Import() length = %d, want 3", p.Len())
	}
	
	// Check HSB color
	red, _ := p.Get(0)
	expectedRed := color.NewHSB(0, 100, 100)
	if red.Color.ToHSB() != expectedRed {
		t.Errorf("Import() HSB color = %v, want %v", red.Color, expectedRed)
	}
}

func TestImportLAB(t *testing.T) {
	importer := NewImporter()
	importer.ColorFormat = FormatLAB
	
	csvData := `Name,L,A,B
Gray,50,0,0
Red,50,50,25
Blue,50,-25,-50`
	
	reader := strings.NewReader(csvData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Len() != 3 {
		t.Errorf("Import() length = %d, want 3", p.Len())
	}
	
	// Check LAB color
	gray, _ := p.Get(0)
	expectedGray := color.NewLAB(50, 0, 0)
	if gray.Color.ToLAB() != expectedGray {
		t.Errorf("Import() LAB color = %v, want %v", gray.Color, expectedGray)
	}
}

func TestImportAutoDetect(t *testing.T) {
	importer := NewImporter()
	importer.ColorFormat = FormatAuto
	
	tests := map[string]struct {
		csvData  string
		expected ColorFormat
	}{
		"RGB detection": {
			"Name,R,G,B\nRed,255,0,0",
			FormatRGB,
		},
		"CMYK detection": {
			"Name,C,M,Y,K\nCyan,100,0,0,0",
			FormatCMYK,
		},
		"Hex detection": {
			"Name,Color\nRed,#FF0000",
			FormatHex,
		},
	}
	
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvData)
			p, err := importer.Import(reader)
			
			if err != nil {
				t.Errorf("Import() error = %v", err)
			}
			
			// Check that format was detected correctly
			if format, ok := p.GetMetadata("color_format"); !ok || format != tt.expected {
				t.Errorf("Import() detected format = %v, want %v", format, tt.expected)
			}
		})
	}
}

func TestImportWithoutHeader(t *testing.T) {
	importer := NewImporter()
	importer.HasHeader = false
	importer.ColorFormat = FormatRGB
	
	csvData := `Red,255,0,0
Green,0,255,0
Blue,0,0,255`
	
	reader := strings.NewReader(csvData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Len() != 3 {
		t.Errorf("Import() length = %d, want 3", p.Len())
	}
}

func TestImportCustomDelimiter(t *testing.T) {
	importer := NewImporter()
	importer.Delimiter = ';'
	
	csvData := `Name;R;G;B
Red;255;0;0
Green;0;255;0`
	
	reader := strings.NewReader(csvData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Len() != 2 {
		t.Errorf("Import() length = %d, want 2", p.Len())
	}
}

func TestImportFloatRGB(t *testing.T) {
	importer := NewImporter()
	importer.ColorFormat = FormatRGBFloat
	
	csvData := `Name,R,G,B
Red,1.0,0.0,0.0
Gray,0.5,0.5,0.5`
	
	reader := strings.NewReader(csvData)
	p, err := importer.Import(reader)
	
	if err != nil {
		t.Errorf("Import() error = %v", err)
	}
	
	if p.Len() != 2 {
		t.Errorf("Import() length = %d, want 2", p.Len())
	}
	
	// Check float conversion
	red, _ := p.Get(0)
	expectedRed := color.NewRGB(255, 0, 0)
	if red.Color.ToRGB() != expectedRed {
		t.Errorf("Import() float RGB = %v, want %v", red.Color, expectedRed)
	}
	
	gray, _ := p.Get(1)
	expectedGray := color.NewRGB(128, 128, 128)
	if gray.Color.ToRGB() != expectedGray {
		t.Errorf("Import() float RGB gray = %v, want %v", gray.Color, expectedGray)
	}
}

func TestExporter(t *testing.T) {
	exporter := NewExporter()
	
	// Test CanExport
	if !exporter.CanExport(".csv") {
		t.Errorf("CanExport() should accept .csv")
	}
	
	if exporter.CanExport(".json") {
		t.Errorf("CanExport() should not accept .json")
	}
	
	// Test SupportedFormats
	formats := exporter.SupportedFormats()
	if len(formats) != 1 || formats[0] != ".csv" {
		t.Errorf("SupportedFormats() = %v, want [.csv]", formats)
	}
}

func TestExportRGB(t *testing.T) {
	exporter := NewExporter()
	exporter.ColorFormat = FormatRGB
	
	p := palette.New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result := output.String()
	
	// Check header
	if !strings.Contains(result, "Name,R,G,B") {
		t.Errorf("Export() should contain RGB header")
	}
	
	// Check data
	if !strings.Contains(result, "Red,255,0,0") {
		t.Errorf("Export() should contain red color data")
	}
	
	if !strings.Contains(result, "Green,0,255,0") {
		t.Errorf("Export() should contain green color data")
	}
}

func TestExportHex(t *testing.T) {
	exporter := NewExporter()
	exporter.ColorFormat = FormatHex
	
	p := palette.New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result := output.String()
	
	// Check header
	if !strings.Contains(result, "Name,Hex") {
		t.Errorf("Export() should contain Hex header")
	}
	
	// Check data
	if !strings.Contains(result, "Red,#FF0000") {
		t.Errorf("Export() should contain red hex color")
	}
	
	if !strings.Contains(result, "Green,#00FF00") {
		t.Errorf("Export() should contain green hex color")
	}
}

func TestExportCMYK(t *testing.T) {
	exporter := NewExporter()
	exporter.ColorFormat = FormatCMYK
	
	p := palette.New("Test")
	p.Add(color.NewCMYK(100, 0, 0, 0), "Cyan")
	p.Add(color.NewCMYK(0, 100, 0, 0), "Magenta")
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result := output.String()
	
	// Check header
	if !strings.Contains(result, "Name,C,M,Y,K") {
		t.Errorf("Export() should contain CMYK header")
	}
	
	// Check data
	if !strings.Contains(result, "Cyan,100,0,0,0") {
		t.Errorf("Export() should contain cyan CMYK data")
	}
}

func TestExportFloatRGB(t *testing.T) {
	exporter := NewExporter()
	exporter.ColorFormat = FormatRGBFloat
	
	p := palette.New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(128, 128, 128), "Gray")
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result := output.String()
	
	// Check data contains float values
	if !strings.Contains(result, "Red,1.000,0.000,0.000") {
		t.Errorf("Export() should contain red float RGB")
	}
	
	if !strings.Contains(result, "Gray,0.502,0.502,0.502") {
		t.Errorf("Export() should contain gray float RGB")
	}
}

func TestExportWithoutHeader(t *testing.T) {
	exporter := NewExporter()
	exporter.IncludeHeader = false
	
	p := palette.New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result := output.String()
	
	// Should not contain header
	if strings.Contains(result, "Name,R,G,B") {
		t.Errorf("Export() should not contain header when disabled")
	}
	
	// Should contain data
	if !strings.Contains(result, "Red,255,0,0") {
		t.Errorf("Export() should contain color data")
	}
}

func TestExportCustomDelimiter(t *testing.T) {
	exporter := NewExporter()
	exporter.Delimiter = ';'
	
	p := palette.New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}
	
	result := output.String()
	
	// Should use semicolon delimiter
	if !strings.Contains(result, "Name;R;G;B") {
		t.Errorf("Export() should use custom delimiter in header")
	}
	
	if !strings.Contains(result, "Red;255;0;0") {
		t.Errorf("Export() should use custom delimiter in data")
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can export and then import a palette
	original := palette.New("Round Trip Test")
	original.Add(color.NewRGB(255, 0, 0), "Red")
	original.Add(color.NewRGB(0, 255, 0), "Green")
	original.Add(color.NewRGB(0, 0, 255), "Blue")
	
	// Export
	exporter := NewExporter()
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
	
	// Compare lengths
	if imported.Len() != original.Len() {
		t.Errorf("Round trip length = %d, want %d", imported.Len(), original.Len())
	}
	
	// Compare colors
	for i := 0; i < original.Len(); i++ {
		origColor, _ := original.Get(i)
		impColor, _ := imported.Get(i)
		
		if origColor.Name != impColor.Name {
			t.Errorf("Round trip color %d name = %v, want %v", i, impColor.Name, origColor.Name)
		}
		
		if origColor.Color.ToRGB() != impColor.Color.ToRGB() {
			t.Errorf("Round trip color %d = %v, want %v", i, impColor.Color, origColor.Color)
		}
	}
}

func TestParseErrors(t *testing.T) {
	importer := NewImporter()
	
	tests := map[string]string{
		"Empty file":        "",
		"Only header":       "Name,R,G,B",
		"Invalid RGB":       "Red,999,0,0",
		"Invalid hex":       "Red,#GGGGGG",
		"Insufficient data": "Red,255",
		"Non-numeric":       "Red,abc,def,ghi",
	}
	
	for name, csvData := range tests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(csvData)
			_, err := importer.Import(reader)
			
			if err == nil {
				t.Errorf("Import() should error for: %s", name)
			}
		})
	}
}

// Benchmark tests
func BenchmarkImportRGB(b *testing.B) {
	importer := NewImporter()
	
	csvData := `Name,R,G,B
Red,255,0,0
Green,0,255,0
Blue,0,0,255
Yellow,255,255,0
Cyan,0,255,255
Magenta,255,0,255`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(csvData)
		_, _ = importer.Import(reader)
	}
}

func BenchmarkExportRGB(b *testing.B) {
	exporter := NewExporter()
	
	p := palette.New("Benchmark")
	for i := 0; i < 100; i++ {
		p.Add(color.NewRGB(uint8(i), uint8(i), uint8(i)), fmt.Sprintf("Color%d", i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var output strings.Builder
		_ = exporter.Export(p, &output)
	}
}