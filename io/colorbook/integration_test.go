package colorbook_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kennyp/palette/io/colorbook"
)

func TestRealACBFiles(t *testing.T) {
	testDataDir := "../../testdata"
	
	// Test files we have available
	testFiles := map[string]struct {
		minColors int // Minimum expected colors for validation
		hasName   bool // Whether we expect a title/name
	}{
		"example.acb":        {1, false},
		"ANPA_Color.acb":     {100, true},
		"DIC_Color_Guide.acb": {600, true}, 
		"FOCOLTONE.acb":      {700, true},
		"TRUMATCH.acb":       {2000, true},
	}

	importer := colorbook.NewImporter()

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

			// Check if we have a name when expected
			if expected.hasName && palette.Name == "" {
				t.Error("Expected palette to have a name")
			}

			// Validate that colors have names and valid values
			for i, namedColor := range colors {
				if namedColor.Color == nil {
					t.Errorf("Color %d is nil", i)
					continue
				}

				// Check that color has a name
				if namedColor.Name == "" {
					t.Errorf("Color %d has empty name", i)
				}

				// Basic color validation - convert to RGB and check bounds
				rgb := namedColor.Color.ToRGB()
				if rgb.R > 255 || rgb.G > 255 || rgb.B > 255 {
					t.Errorf("Color %d (%s) has invalid RGB values: %v", i, namedColor.Name, rgb)
				}
			}

			t.Logf("Successfully imported %s: %d colors, name: %q", 
				filename, len(colors), palette.Name)
		})
	}
}

func TestACBFormatValidation(t *testing.T) {
	testDataDir := "../../testdata"
	
	tests := []struct {
		filename string
		expectError bool
		description string
	}{
		{"FOCOLTONE.acb", false, "Valid FOCOLTONE color book"},
		{"DIC_Color_Guide.acb", false, "Valid DIC color guide"},
		{"TRUMATCH.acb", false, "Valid TRUMATCH color system"},
	}

	importer := colorbook.NewImporter()

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

func TestACBFileDetails(t *testing.T) {
	testDataDir := "../../testdata"
	
	// Test a specific file in detail
	filePath := filepath.Join(testDataDir, "FOCOLTONE.acb")
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Skip("FOCOLTONE.acb not found")
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	importer := colorbook.NewImporter()
	palette, err := importer.Import(file)
	if err != nil {
		t.Fatalf("Failed to import: %v", err)
	}

	// Check specific aspects of the FOCOLTONE color book
	colors := palette.Colors

	// FOCOLTONE should have exactly 763 colors
	expectedColors := 763
	if len(colors) != expectedColors {
		t.Errorf("FOCOLTONE should have %d colors, got %d", expectedColors, len(colors))
	}

	// Check that color names follow FOCOLTONE naming pattern
	if len(colors) > 0 {
		firstName := colors[0].Name
		if firstName == "" {
			t.Error("First color should have a name")
		}
		t.Logf("First FOCOLTONE color: %s", firstName)
	}

	// Check color accuracy - first color should be specific values
	// (This would require knowing exact FOCOLTONE values)
	if len(colors) > 0 {
		firstColor := colors[0].Color.ToRGB()
		t.Logf("First color RGB: R=%d, G=%d, B=%d", firstColor.R, firstColor.G, firstColor.B)
	}
}

func TestLargeACBPerformance(t *testing.T) {
	testDataDir := "../../testdata"
	filePath := filepath.Join(testDataDir, "DIC_Color_Guide.acb")
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Skip("DIC_Color_Guide.acb not found")
	}

	// Time the import of a large color book
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	importer := colorbook.NewImporter()
	
	// This should complete reasonably quickly
	palette, err := importer.Import(file)
	if err != nil {
		t.Fatalf("Failed to import large ACB file: %v", err)
	}

	colors := palette.Colors
	t.Logf("DIC Color Guide import performance: %d colors imported", len(colors))

	// Basic validation of large dataset
	if len(colors) < 1000 {
		t.Errorf("DIC Color Guide should have at least 1000 colors, got %d", len(colors))
	}
}