package colorswatch_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kennyp/palette/io/colorswatch"
)

func TestRealACOFiles(t *testing.T) {
	testDataDir := "../../testdata"
	
	// Test files we have available
	testFiles := map[string]struct {
		minColors int // Minimum expected colors for validation
		maxColors int // Maximum expected colors (0 = no limit)
	}{
		"example.aco":        {1, 10},
		"primary_colors.aco": {6, 6},    // Exactly 6 primary colors
		"grayscale.aco":      {8, 10},   // 9 grayscale levels (0-255 step 32)
	}

	importer := colorswatch.NewImporter()

	for filename, expected := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(testDataDir, filename)
			
			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", filename)
			}

			// Open and import the file
			file, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", filename, err)
			}
			defer file.Close()

			palette, err := importer.Import(file)
			if err != nil {
				t.Fatalf("Failed to import %s: %v", filename, err)
			}

			// Basic validation
			if palette == nil {
				t.Fatal("Imported palette is nil")
			}

			// Check color count
			colors := palette.Colors
			if len(colors) < expected.minColors {
				t.Errorf("Expected at least %d colors, got %d", expected.minColors, len(colors))
			}
			if expected.maxColors > 0 && len(colors) > expected.maxColors {
				t.Errorf("Expected at most %d colors, got %d", expected.maxColors, len(colors))
			}

			// Validate that colors have valid values
			for i, namedColor := range colors {
				if namedColor.Color == nil {
					t.Errorf("Color %d is nil", i)
					continue
				}

				// Basic color validation - convert to RGB and check bounds
				rgb := namedColor.Color.ToRGB()
				if rgb.R > 255 || rgb.G > 255 || rgb.B > 255 {
					t.Errorf("Color %d has invalid RGB values: %v", i, rgb)
				}
			}

			t.Logf("Successfully imported %s: %d colors, name: %q", 
				filename, len(colors), palette.Name)
		})
	}
}

func TestACOFormatValidation(t *testing.T) {
	testDataDir := "../../testdata"
	
	tests := []struct {
		filename string
		expectError bool
		description string
	}{
		{"primary_colors.aco", false, "Valid primary colors swatch"},
		{"grayscale.aco", false, "Valid grayscale swatch"},
		{"example.aco", false, "Valid example swatch"},
	}

	importer := colorswatch.NewImporter()

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			filePath := filepath.Join(testDataDir, tt.filename)
			
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", tt.filename)
			}

			file, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			_, err = importer.Import(file)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestACOColorAccuracy(t *testing.T) {
	testDataDir := "../../testdata"
	filePath := filepath.Join(testDataDir, "primary_colors.aco")
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Skip("primary_colors.aco not found")
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	importer := colorswatch.NewImporter()
	palette, err := importer.Import(file)
	if err != nil {
		t.Fatalf("Failed to import: %v", err)
	}

	colors := palette.Colors
	
	// We know the primary_colors.aco should have these exact colors
	expectedColors := []struct {
		r, g, b uint8
		name    string
	}{
		{255, 0, 0, "Red"},
		{0, 255, 0, "Green"},
		{0, 0, 255, "Blue"},
		{255, 255, 0, "Yellow"},
		{255, 0, 255, "Magenta"},
		{0, 255, 255, "Cyan"},
	}

	if len(colors) != len(expectedColors) {
		t.Fatalf("Expected %d colors, got %d", len(expectedColors), len(colors))
	}

	for i, expected := range expectedColors {
		rgb := colors[i].Color.ToRGB()
		
		// Allow for minor variations due to color space conversions
		tolerance := uint8(2)
		
		if abs(rgb.R, expected.r) > tolerance ||
		   abs(rgb.G, expected.g) > tolerance ||
		   abs(rgb.B, expected.b) > tolerance {
			t.Errorf("Color %d: expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
				i, expected.r, expected.g, expected.b, rgb.R, rgb.G, rgb.B)
		}
	}
}

func TestGrayscaleAccuracy(t *testing.T) {
	testDataDir := "../../testdata"
	filePath := filepath.Join(testDataDir, "grayscale.aco")
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Skip("grayscale.aco not found")
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	importer := colorswatch.NewImporter()
	palette, err := importer.Import(file)
	if err != nil {
		t.Fatalf("Failed to import: %v", err)
	}

	colors := palette.Colors
	
	// Verify grayscale values are reasonable
	for i, namedColor := range colors {
		rgb := namedColor.Color.ToRGB()
		
		// In a true grayscale, R=G=B
		tolerance := uint8(5) // Allow some tolerance for conversion artifacts
		
		if abs(rgb.R, rgb.G) > tolerance || abs(rgb.G, rgb.B) > tolerance {
			t.Errorf("Color %d is not grayscale: RGB(%d,%d,%d)", 
				i, rgb.R, rgb.G, rgb.B)
		}
		
		t.Logf("Grayscale color %d: RGB(%d,%d,%d)", i, rgb.R, rgb.G, rgb.B)
	}
}

func TestRoundTripACO(t *testing.T) {
	testDataDir := "../../testdata"
	testFiles := []string{"primary_colors.aco", "grayscale.aco"}
	
	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(testDataDir, filename)
			
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", filename)
			}

			// Import original
			file, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			importer := colorswatch.NewImporter()
			originalPalette, err := importer.Import(file)
			if err != nil {
				t.Fatalf("Failed to import: %v", err)
			}

			// Export to new file
			tempFile := filepath.Join(testDataDir, "temp_"+filename)
			defer os.Remove(tempFile) // Clean up

			outFile, err := os.Create(tempFile)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			exporter := colorswatch.NewExporter()
			err = exporter.Export(originalPalette, outFile)
			outFile.Close()
			if err != nil {
				t.Fatalf("Failed to export: %v", err)
			}

			// Import the exported file
			roundTripFile, err := os.Open(tempFile)
			if err != nil {
				t.Fatalf("Failed to open round-trip file: %v", err)
			}
			defer roundTripFile.Close()

			roundTripPalette, err := importer.Import(roundTripFile)
			if err != nil {
				t.Fatalf("Failed to import round-trip file: %v", err)
			}

			// Compare palettes
			originalColors := originalPalette.Colors
			roundTripColors := roundTripPalette.Colors

			if len(originalColors) != len(roundTripColors) {
				t.Fatalf("Color count mismatch: original %d, round-trip %d",
					len(originalColors), len(roundTripColors))
			}

			// Compare colors with tolerance for conversion artifacts
			tolerance := uint8(3)
			for i := range originalColors {
				origRGB := originalColors[i].Color.ToRGB()
				rtRGB := roundTripColors[i].Color.ToRGB()

				if abs(origRGB.R, rtRGB.R) > tolerance ||
				   abs(origRGB.G, rtRGB.G) > tolerance ||
				   abs(origRGB.B, rtRGB.B) > tolerance {
					t.Errorf("Color %d mismatch: original RGB(%d,%d,%d), round-trip RGB(%d,%d,%d)",
						i, origRGB.R, origRGB.G, origRGB.B, rtRGB.R, rtRGB.G, rtRGB.B)
				}
			}
		})
	}
}

// Helper function to calculate absolute difference between two uint8 values
func abs(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}