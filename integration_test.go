package palette_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	_ "github.com/kennyp/palette" // Initialize format importers/exporters
	"github.com/kennyp/palette/color"
	paletteio "github.com/kennyp/palette/io"
	"github.com/kennyp/palette/palette"
)

// TestFullWorkflow tests the complete workflow of creating, manipulating, and converting palettes
func TestFullWorkflow(t *testing.T) {
	// 1. Create a palette programmatically
	original := palette.New("Integration Test Palette")
	original.Description = "A palette for testing the full workflow"
	
	// Add colors in different color spaces
	original.Add(color.NewRGB(255, 0, 0), "Bright Red")
	original.Add(color.NewCMYK(100, 0, 100, 0), "Process Green")
	original.Add(color.NewHSB(240, 100, 100), "Pure Blue")
	original.Add(color.NewLAB(50, 0, 0), "Mid Gray")
	
	// Add metadata
	original.SetMetadata("version", "1.0")
	original.SetMetadata("author", "Integration Test")
	
	// 2. Export to JSON
	var jsonOutput bytes.Buffer
	err := paletteio.Export(original, &jsonOutput, ".json")
	if err != nil {
		t.Fatalf("Failed to export to JSON: %v", err)
	}
	
	// 3. Import from JSON
	jsonImported, err := paletteio.Import(strings.NewReader(jsonOutput.String()), ".json")
	if err != nil {
		t.Fatalf("Failed to import from JSON: %v", err)
	}
	
	// 4. Verify JSON round-trip
	if jsonImported.Name != original.Name {
		t.Errorf("JSON round-trip name mismatch: got %s, want %s", jsonImported.Name, original.Name)
	}
	
	if jsonImported.Len() != original.Len() {
		t.Errorf("JSON round-trip length mismatch: got %d, want %d", jsonImported.Len(), original.Len())
	}
	
	// 5. Convert all colors to RGB
	rgbPalette, err := jsonImported.ConvertToColorSpace("RGB")
	if err != nil {
		t.Fatalf("Failed to convert to RGB: %v", err)
	}
	
	// Verify all colors are now RGB
	for i := range rgbPalette.Len() {
		c, _ := rgbPalette.Get(i)
		if c.Color.ColorSpace() != "RGB" {
			t.Errorf("Color %d should be RGB, got %s", i, c.Color.ColorSpace())
		}
	}
	
	// 6. Export RGB palette to CSV
	var csvOutput bytes.Buffer
	err = paletteio.Export(rgbPalette, &csvOutput, ".csv")
	if err != nil {
		t.Fatalf("Failed to export to CSV: %v", err)
	}
	
	// 7. Import from CSV
	csvImported, err := paletteio.Import(strings.NewReader(csvOutput.String()), ".csv")
	if err != nil {
		t.Fatalf("Failed to import from CSV: %v", err)
	}
	
	// 8. Verify CSV round-trip preserves color count
	if csvImported.Len() != rgbPalette.Len() {
		t.Errorf("CSV round-trip length mismatch: got %d, want %d", csvImported.Len(), rgbPalette.Len())
	}
	
	// 9. Filter and transform
	filtered := csvImported.Filter(func(c palette.NamedColor) bool {
		return strings.Contains(strings.ToLower(c.Name), "red") ||
			   strings.Contains(strings.ToLower(c.Name), "green") ||
			   strings.Contains(strings.ToLower(c.Name), "blue")
	})
	
	if filtered.Len() != 3 {
		t.Errorf("Filtered palette should have 3 colors, got %d", filtered.Len())
	}
	
	// 10. Final export test
	var finalOutput bytes.Buffer
	err = paletteio.Export(filtered, &finalOutput, ".json")
	if err != nil {
		t.Fatalf("Failed final export: %v", err)
	}
	
	t.Logf("Full workflow completed successfully")
	t.Logf("Original: %d colors", original.Len())
	t.Logf("JSON round-trip: %d colors", jsonImported.Len())
	t.Logf("RGB converted: %d colors", rgbPalette.Len())
	t.Logf("CSV round-trip: %d colors", csvImported.Len())
	t.Logf("Filtered: %d colors", filtered.Len())
}

// TestFormatCompatibility tests importing and exporting between different formats
func TestFormatCompatibility(t *testing.T) {
	// Create test palette
	p := palette.New("Format Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	p.Add(color.NewRGB(0, 0, 255), "Blue")
	
	formats := []string{".json", ".csv"}
	
	for _, fromFormat := range formats {
		for _, toFormat := range formats {
			t.Run(fmt.Sprintf("%s_to_%s", fromFormat, toFormat), func(t *testing.T) {
				// Export to first format
				var intermediate bytes.Buffer
				err := paletteio.Export(p, &intermediate, fromFormat)
				if err != nil {
					t.Fatalf("Failed to export to %s: %v", fromFormat, err)
				}
				
				// Import from first format
				imported, err := paletteio.Import(strings.NewReader(intermediate.String()), fromFormat)
				if err != nil {
					t.Fatalf("Failed to import from %s: %v", fromFormat, err)
				}
				
				// Export to second format
				var final bytes.Buffer
				err = paletteio.Export(imported, &final, toFormat)
				if err != nil {
					t.Fatalf("Failed to export to %s: %v", toFormat, err)
				}
				
				// Import from second format
				final_imported, err := paletteio.Import(strings.NewReader(final.String()), toFormat)
				if err != nil {
					t.Fatalf("Failed to import from %s: %v", toFormat, err)
				}
				
				// Verify color count is preserved
				if final_imported.Len() != p.Len() {
					t.Errorf("Color count not preserved: got %d, want %d", final_imported.Len(), p.Len())
				}
			})
		}
	}
}

// TestColorSpaceConversions tests accuracy of color space conversions
func TestColorSpaceConversions(t *testing.T) {
	// Test with known color values
	testCases := []struct {
		name string
		rgb  color.RGB
	}{
		{"White", color.NewRGB(255, 255, 255)},
		{"Black", color.NewRGB(0, 0, 0)},
		{"Red", color.NewRGB(255, 0, 0)},
		{"Green", color.NewRGB(0, 255, 0)},
		{"Blue", color.NewRGB(0, 0, 255)},
		{"Gray", color.NewRGB(128, 128, 128)},
		{"Orange", color.NewRGB(255, 165, 0)},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			original := tc.rgb
			
			// Test RGB -> CMYK -> RGB
			cmyk := original.ToCMYK()
			backToRGB := cmyk.ToRGB()
			
			tolerance := uint8(2) // Allow small rounding errors
			if abs(int(original.R)-int(backToRGB.R)) > int(tolerance) ||
				abs(int(original.G)-int(backToRGB.G)) > int(tolerance) ||
				abs(int(original.B)-int(backToRGB.B)) > int(tolerance) {
				t.Errorf("RGB->CMYK->RGB conversion failed: %v -> %v -> %v", original, cmyk, backToRGB)
			}
			
			// Test RGB -> HSB -> RGB
			hsb := original.ToHSB()
			backToRGB2 := hsb.ToRGB()
			
			if abs(int(original.R)-int(backToRGB2.R)) > int(tolerance) ||
				abs(int(original.G)-int(backToRGB2.G)) > int(tolerance) ||
				abs(int(original.B)-int(backToRGB2.B)) > int(tolerance) {
				t.Errorf("RGB->HSB->RGB conversion failed: %v -> %v -> %v", original, hsb, backToRGB2)
			}
		})
	}
}

// TestLargeDatasets tests performance and accuracy with large palettes
func TestLargeDatasets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}
	
	// Create a large palette
	large := palette.New("Large Test Palette")
	large.Description = "Testing with many colors"
	
	// Add 1000 colors with different patterns
	for i := range 1000 {
		r := uint8(i % 256)
		g := uint8((i * 7) % 256)
		b := uint8((i * 13) % 256)
		
		large.Add(color.NewRGB(r, g, b), fmt.Sprintf("Color_%04d", i))
	}
	
	t.Logf("Created palette with %d colors", large.Len())
	
	// Test JSON export/import
	var jsonOutput bytes.Buffer
	err := paletteio.Export(large, &jsonOutput, ".json")
	if err != nil {
		t.Fatalf("Failed to export large palette to JSON: %v", err)
	}
	
	jsonImported, err := paletteio.Import(strings.NewReader(jsonOutput.String()), ".json")
	if err != nil {
		t.Fatalf("Failed to import large palette from JSON: %v", err)
	}
	
	if jsonImported.Len() != large.Len() {
		t.Errorf("Large palette JSON round-trip failed: got %d colors, want %d", jsonImported.Len(), large.Len())
	}
	
	// Test CSV export/import
	var csvOutput bytes.Buffer
	err = paletteio.Export(large, &csvOutput, ".csv")
	if err != nil {
		t.Fatalf("Failed to export large palette to CSV: %v", err)
	}
	
	csvImported, err := paletteio.Import(strings.NewReader(csvOutput.String()), ".csv")
	if err != nil {
		t.Fatalf("Failed to import large palette from CSV: %v", err)
	}
	
	if csvImported.Len() != large.Len() {
		t.Errorf("Large palette CSV round-trip failed: got %d colors, want %d", csvImported.Len(), large.Len())
	}
	
	// Test color space conversion on large dataset
	cmykConverted, err := large.ConvertToColorSpace("CMYK")
	if err != nil {
		t.Fatalf("Failed to convert large palette to CMYK: %v", err)
	}
	
	if cmykConverted.Len() != large.Len() {
		t.Errorf("Large palette CMYK conversion failed: got %d colors, want %d", cmykConverted.Len(), large.Len())
	}
	
	// Verify all colors are CMYK
	for i := range min(10, cmykConverted.Len()) { // Check first 10 colors
		c, _ := cmykConverted.Get(i)
		if c.Color.ColorSpace() != "CMYK" {
			t.Errorf("Color %d should be CMYK after conversion, got %s", i, c.Color.ColorSpace())
		}
	}
	
	t.Logf("Large dataset test completed successfully")
}

// TestMetadataPreservation tests that metadata is preserved through export/import cycles
func TestMetadataPreservation(t *testing.T) {
	// Create palette with extensive metadata
	p := palette.New("Metadata Test")
	p.Description = "Testing metadata preservation"
	p.Add(color.NewRGB(255, 0, 0), "Red")
	
	// Add various types of metadata
	p.SetMetadata("version", "2.1")
	p.SetMetadata("author", "Test Suite")
	p.SetMetadata("created", "2024-01-15")
	p.SetMetadata("revision", 42)
	p.SetMetadata("tags", []string{"test", "metadata", "preservation"})
	
	// Export to JSON (which supports metadata)
	var output bytes.Buffer
	err := paletteio.Export(p, &output, ".json")
	if err != nil {
		t.Fatalf("Failed to export palette with metadata: %v", err)
	}
	
	// Import back
	imported, err := paletteio.Import(strings.NewReader(output.String()), ".json")
	if err != nil {
		t.Fatalf("Failed to import palette with metadata: %v", err)
	}
	
	// Check that basic properties are preserved
	if imported.Name != p.Name {
		t.Errorf("Name not preserved: got %s, want %s", imported.Name, p.Name)
	}
	
	if imported.Description != p.Description {
		t.Errorf("Description not preserved: got %s, want %s", imported.Description, p.Description)
	}
	
	// Check that format metadata was added
	if format, ok := imported.GetMetadata("format"); !ok || format != "JSON" {
		t.Errorf("Format metadata not set correctly")
	}
	
	// Check that original metadata is stored
	if _, ok := imported.GetMetadata("original_metadata"); !ok {
		t.Errorf("Original metadata not preserved")
	}
}

// TestErrorHandling tests error conditions and edge cases
func TestErrorHandling(t *testing.T) {
	// Test invalid format
	p := palette.New("Error Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	
	var output bytes.Buffer
	err := paletteio.Export(p, &output, ".invalid")
	if err == nil {
		t.Errorf("Export should fail for invalid format")
	}
	
	// Test empty palette
	empty := palette.New("Empty")
	err = paletteio.Export(empty, &output, ".json")
	if err != nil {
		t.Errorf("Export should succeed for empty palette: %v", err)
	}
	
	// Test invalid JSON import
	invalidJSON := `{"invalid": json}`
	_, err = paletteio.Import(strings.NewReader(invalidJSON), ".json")
	if err == nil {
		t.Errorf("Import should fail for invalid JSON")
	}
	
	// Test invalid CSV import
	invalidCSV := `Name,R,G,B
Red,999,0,0` // Invalid RGB value
	_, err = paletteio.Import(strings.NewReader(invalidCSV), ".csv")
	if err == nil {
		t.Errorf("Import should fail for invalid CSV data")
	}
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

