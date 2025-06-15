package colorswatch_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kennyp/palette/io/colorswatch"
)

func TestAllColorSpaces(t *testing.T) {
	testDataDir := "../../testdata"
	
	tests := []struct {
		filename      string
		description   string
		expectedColors int
		colorSpace    string
	}{
		{"hsb_colors.aco", "HSB color space", 6, "HSB"},
		{"cmyk_colors.aco", "CMYK color space", 6, "CMYK"},
		{"lab_colors.aco", "LAB color space", 6, "LAB"},
		{"grayscale_rgb.aco", "Grayscale RGB", 11, "RGB"},
	}

	importer := colorswatch.NewImporter()
	exporter := colorswatch.NewExporter()

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			filePath := filepath.Join(testDataDir, tt.filename)
			
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", tt.filename)
			}

			// Import
			file, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			palette, err := importer.Import(file)
			if err != nil {
				t.Fatalf("Failed to import: %v", err)
			}

			// Verify color count
			if len(palette.Colors) != tt.expectedColors {
				t.Errorf("Expected %d colors, got %d", tt.expectedColors, len(palette.Colors))
			}

			// Round-trip test
			tempFile := filepath.Join(testDataDir, "temp_"+tt.filename)
			defer os.Remove(tempFile)

			outFile, err := os.Create(tempFile)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			err = exporter.Export(palette, outFile)
			outFile.Close()
			if err != nil {
				t.Fatalf("Failed to export: %v", err)
			}

			// Re-import
			roundTripFile, err := os.Open(tempFile)
			if err != nil {
				t.Fatalf("Failed to open round-trip file: %v", err)
			}
			defer roundTripFile.Close()

			roundTripPalette, err := importer.Import(roundTripFile)
			if err != nil {
				t.Fatalf("Failed to import round-trip file: %v", err)
			}

			// Verify round-trip accuracy
			if len(roundTripPalette.Colors) != len(palette.Colors) {
				t.Fatalf("Round-trip color count mismatch: original %d, round-trip %d",
					len(palette.Colors), len(roundTripPalette.Colors))
			}

			// Test specific color space behaviors
			for i, namedColor := range palette.Colors {
				rgb := namedColor.Color.ToRGB()
				name := namedColor.Name
				
				// Specific tests for HSB
				if tt.filename == "hsb_colors.aco" {
					if name == "Pure Red" && (rgb.R < 200 || rgb.G > 50 || rgb.B > 50) {
						t.Errorf("HSB Pure Red failed: RGB(%d,%d,%d)", rgb.R, rgb.G, rgb.B)
					}
					if name == "Pure Green" && (rgb.R > 50 || rgb.G < 200 || rgb.B > 50) {
						t.Errorf("HSB Pure Green failed: RGB(%d,%d,%d)", rgb.R, rgb.G, rgb.B)
					}
					if name == "Pure Blue" && (rgb.R > 50 || rgb.G > 50 || rgb.B < 200) {
						t.Errorf("HSB Pure Blue failed: RGB(%d,%d,%d)", rgb.R, rgb.G, rgb.B)
					}
				}

				// Specific tests for CMYK
				if tt.filename == "cmyk_colors.aco" {
					if name == "Black" && (rgb.R > 30 || rgb.G > 30 || rgb.B > 30) {
						t.Errorf("CMYK Black failed: RGB(%d,%d,%d)", rgb.R, rgb.G, rgb.B)
					}
					if name == "Cyan" && (rgb.R > 50 || rgb.G < 150 || rgb.B < 150) {
						t.Errorf("CMYK Cyan failed: RGB(%d,%d,%d)", rgb.R, rgb.G, rgb.B)
					}
				}

				// Specific tests for LAB
				if tt.filename == "lab_colors.aco" {
					if name == "White" && (rgb.R < 200 || rgb.G < 200 || rgb.B < 200) {
						t.Errorf("LAB White failed: RGB(%d,%d,%d)", rgb.R, rgb.G, rgb.B)
					}
					if name == "Black" && (rgb.R > 50 || rgb.G > 50 || rgb.B > 50) {
						t.Errorf("LAB Black failed: RGB(%d,%d,%d)", rgb.R, rgb.G, rgb.B)
					}
				}

				// Basic conversion sanity checks
				cmyk := namedColor.Color.ToCMYK()
				hsb := namedColor.Color.ToHSB()
				lab := namedColor.Color.ToLAB()

				if rgb.R > 255 || rgb.G > 255 || rgb.B > 255 {
					t.Errorf("Color %d has invalid RGB: %v", i, rgb)
				}
				if cmyk.C > 100 || cmyk.M > 100 || cmyk.Y > 100 || cmyk.K > 100 {
					t.Errorf("Color %d has invalid CMYK: %v", i, cmyk)
				}
				if hsb.H > 360 || hsb.S > 100 || hsb.B > 100 {
					t.Errorf("Color %d has invalid HSB: %v", i, hsb)
				}
				if lab.L < -128 || lab.L > 127 || lab.A < -128 || lab.A > 127 || lab.B < -128 || lab.B > 127 {
					t.Errorf("Color %d has invalid LAB: %v", i, lab)
				}
			}

			t.Logf("%s: Successfully tested %d colors", tt.colorSpace, len(palette.Colors))
		})
	}
}