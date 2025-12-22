package colorbook_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	adobeColorbook "github.com/kennyp/palette/adobe/colorbook"
	"github.com/kennyp/palette/color"
	"github.com/kennyp/palette/io/colorbook"
	"github.com/kennyp/palette/palette"
)

func TestNewImporter(t *testing.T) {
	importer := colorbook.NewImporter()
	if importer == nil {
		t.Error("colorbook.NewImporter() returned nil")
	}
}

func TestNewExporter(t *testing.T) {
	exporter := colorbook.NewExporter()
	if exporter == nil {
		t.Error("colorbook.NewExporter() returned nil")
	}
}

func TestImporterCanImport(t *testing.T) {
	importer := colorbook.NewImporter()

	tests := map[string]struct {
		format   string
		expected bool
	}{
		"acb_extension":    {".acb", true},
		"ACB_extension":    {".ACB", true},
		"colorbook_format": {"colorbook", true},
		"json_format":      {".json", false},
		"csv_format":       {".csv", false},
		"unknown_format":   {".xyz", false},
		"empty_format":     {"", false},
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
	exporter := colorbook.NewExporter()

	tests := map[string]struct {
		format   string
		expected bool
	}{
		"acb_extension":    {".acb", true},
		"ACB_extension":    {".ACB", true},
		"colorbook_format": {"colorbook", true},
		"json_format":      {".json", false},
		"csv_format":       {".csv", false},
		"unknown_format":   {".xyz", false},
		"empty_format":     {"", false},
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
	importer := colorbook.NewImporter()
	exporter := colorbook.NewExporter()

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

	exporter := colorbook.NewExporter()
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

	exporter := colorbook.NewExporter()
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
	importer := colorbook.NewImporter()

	tests := map[string]struct {
		data string
	}{
		"empty_data":   {""},
		"invalid_data": {"not a valid ACB file"},
		"short_data":   {"8BCB"}, // Valid header but incomplete
		"wrong_header": {"INVALID_HEADER_DATA"},
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

			exporter := colorbook.NewExporter()
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

func TestCMYKRoundTrip(t *testing.T) {
	tests := map[string]struct {
		c, m, y, k uint8
	}{
		"pure_cyan":    {100, 0, 0, 0},
		"pure_magenta": {0, 100, 0, 0},
		"pure_yellow":  {0, 0, 100, 0},
		"pure_black":   {0, 0, 0, 100},
		"white":        {0, 0, 0, 0},
		"mixed":        {50, 25, 75, 10},
		"all_50":       {50, 50, 50, 50},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create palette with CMYK color
			p := palette.New("CMYK Test")
			p.SetMetadata("color_type", adobeColorbook.ColorTypeCMYK)
			p.Add(color.NewCMYK(tt.c, tt.m, tt.y, tt.k), "Test Color")

			// Export to ACB
			exporter := colorbook.NewExporter()
			var buf bytes.Buffer
			if err := exporter.Export(p, &buf); err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			// Import back
			importer := colorbook.NewImporter()
			imported, err := importer.Import(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("Import failed: %v", err)
			}

			// Verify color values
			if imported.Len() != 1 {
				t.Fatalf("Expected 1 color, got %d", imported.Len())
			}

			nc, _ := imported.Get(0)
			cmyk := nc.Color.ToCMYK()

			if cmyk.C != tt.c || cmyk.M != tt.m || cmyk.Y != tt.y || cmyk.K != tt.k {
				t.Errorf("CMYK mismatch: got (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					cmyk.C, cmyk.M, cmyk.Y, cmyk.K, tt.c, tt.m, tt.y, tt.k)
			}
		})
	}
}

func TestLABRoundTrip(t *testing.T) {
	tests := map[string]struct {
		l    int8
		a, b int8
	}{
		"white":       {100, 0, 0},
		"black":       {0, 0, 0},
		"mid_gray":    {50, 0, 0},
		"positive_ab": {50, 100, 100},
		"negative_ab": {50, -100, -100},
		"mixed":       {75, -50, 25},
		"extreme_a":   {50, 127, 0},
		"extreme_b":   {50, 0, -128},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create palette with LAB color
			p := palette.New("LAB Test")
			p.SetMetadata("color_type", adobeColorbook.ColorTypeLab)
			p.Add(color.NewLAB(tt.l, tt.a, tt.b), "Test Color")

			// Export to ACB
			exporter := colorbook.NewExporter()
			var buf bytes.Buffer
			if err := exporter.Export(p, &buf); err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			// Import back
			importer := colorbook.NewImporter()
			imported, err := importer.Import(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("Import failed: %v", err)
			}

			// Verify color values
			if imported.Len() != 1 {
				t.Fatalf("Expected 1 color, got %d", imported.Len())
			}

			nc, _ := imported.Get(0)
			lab := nc.Color.ToLAB()

			if lab.L != tt.l || lab.A != tt.a || lab.B != tt.b {
				t.Errorf("LAB mismatch: got (%d,%d,%d), want (%d,%d,%d)",
					lab.L, lab.A, lab.B, tt.l, tt.a, tt.b)
			}
		})
	}
}

func TestRGBRoundTrip(t *testing.T) {
	tests := map[string]struct {
		r, g, b uint8
	}{
		"red":   {255, 0, 0},
		"green": {0, 255, 0},
		"blue":  {0, 0, 255},
		"white": {255, 255, 255},
		"black": {0, 0, 0},
		"gray":  {128, 128, 128},
		"mixed": {100, 150, 200},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create palette with RGB color
			p := palette.New("RGB Test")
			p.Add(color.NewRGB(tt.r, tt.g, tt.b), "Test Color")

			// Export to ACB
			exporter := colorbook.NewExporter()
			var buf bytes.Buffer
			if err := exporter.Export(p, &buf); err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			// Import back
			importer := colorbook.NewImporter()
			imported, err := importer.Import(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("Import failed: %v", err)
			}

			// Verify color values
			if imported.Len() != 1 {
				t.Fatalf("Expected 1 color, got %d", imported.Len())
			}

			nc, _ := imported.Get(0)
			rgb := nc.Color.ToRGB()

			if rgb.R != tt.r || rgb.G != tt.g || rgb.B != tt.b {
				t.Errorf("RGB mismatch: got (%d,%d,%d), want (%d,%d,%d)",
					rgb.R, rgb.G, rgb.B, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestSpotFunctionSuffix(t *testing.T) {
	tests := map[string]struct {
		colorType      adobeColorbook.ColorType
		expectedSuffix string
	}{
		"rgb":  {adobeColorbook.ColorTypeRGB, adobeColorbook.SpotFunctionProcess},
		"cmyk": {adobeColorbook.ColorTypeCMYK, adobeColorbook.SpotFunctionProcess},
		"lab":  {adobeColorbook.ColorTypeLab, adobeColorbook.SpotFunctionSpot},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create palette
			p := palette.New("Test")
			p.SetMetadata("color_type", tt.colorType)

			// Add a color appropriate for the type
			switch tt.colorType {
			case adobeColorbook.ColorTypeRGB:
				p.Add(color.NewRGB(255, 0, 0), "Red")
			case adobeColorbook.ColorTypeCMYK:
				p.Add(color.NewCMYK(100, 0, 0, 0), "Cyan")
			case adobeColorbook.ColorTypeLab:
				p.Add(color.NewLAB(50, 0, 0), "Gray")
			}

			// Export
			exporter := colorbook.NewExporter()
			var buf bytes.Buffer
			if err := exporter.Export(p, &buf); err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			// Check last 8 bytes
			data := buf.Bytes()
			if len(data) < 8 {
				t.Fatalf("Output too small: %d bytes", len(data))
			}

			suffix := string(data[len(data)-8:])
			if suffix != tt.expectedSuffix {
				t.Errorf("Suffix mismatch: got %q, want %q", suffix, tt.expectedSuffix)
			}
		})
	}
}

func TestBookIDGeneration(t *testing.T) {
	tests := map[string]struct {
		name       string
		colorCount int
	}{
		"simple":        {"My Palette", 3},
		"empty":         {"Empty", 0},
		"large":         {"Large Palette", 100},
		"special_chars": {"Palette with spaces & symbols!", 5},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create palette
			p := palette.New(tt.name)
			for i := 0; i < tt.colorCount; i++ {
				p.Add(color.NewRGB(uint8(i), uint8(i), uint8(i)), "Color")
			}

			// Export
			exporter := colorbook.NewExporter()
			var buf bytes.Buffer
			if err := exporter.Export(p, &buf); err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			// Parse the raw ACB to check BookID
			var acb adobeColorbook.ColorBook
			if err := acb.UnmarshalBinary(buf.Bytes()); err != nil {
				t.Fatalf("Failed to parse ACB: %v", err)
			}

			// Verify BookID is in valid range (4000-65535, avoiding Adobe's 3000-3022)
			id := uint16(acb.ID)
			if id < 4000 || id > 65535 {
				t.Errorf("BookID %d is outside valid range (4000-65535)", id)
			}

			// Verify it doesn't collide with Adobe reserved range
			if id >= 3000 && id <= 3022 {
				t.Errorf("BookID %d collides with Adobe reserved range (3000-3022)", id)
			}
		})
	}
}

func TestBookIDDeterministic(t *testing.T) {
	// Same palette should produce same BookID
	p1 := palette.New("Test Palette")
	p1.Add(color.NewRGB(255, 0, 0), "Red")

	p2 := palette.New("Test Palette")
	p2.Add(color.NewRGB(255, 0, 0), "Red")

	exporter := colorbook.NewExporter()

	var buf1, buf2 bytes.Buffer
	exporter.Export(p1, &buf1)
	exporter.Export(p2, &buf2)

	var acb1, acb2 adobeColorbook.ColorBook
	acb1.UnmarshalBinary(buf1.Bytes())
	acb2.UnmarshalBinary(buf2.Bytes())

	if acb1.ID != acb2.ID {
		t.Errorf("Same palette produced different BookIDs: %d vs %d", acb1.ID, acb2.ID)
	}
}

func TestColorKeyGeneration(t *testing.T) {
	// Create palette with multiple colors
	p := palette.New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 0, 255), "Blue")
	p.Add(color.NewRGB(0, 255, 0), "Green")

	// Export
	exporter := colorbook.NewExporter()
	var buf bytes.Buffer
	if err := exporter.Export(p, &buf); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Parse the raw ACB to check color keys
	var acb adobeColorbook.ColorBook
	if err := acb.UnmarshalBinary(buf.Bytes()); err != nil {
		t.Fatalf("Failed to parse ACB: %v", err)
	}

	// Verify keys are non-empty and properly formatted
	for i, c := range acb.Colors {
		key := string(c.Key[:])

		// Key should be exactly 6 characters
		if len(key) != 6 {
			t.Errorf("Color %d key length is %d, want 6", i, len(key))
		}

		// Key should not be all zeros/empty
		allZero := true
		for _, b := range c.Key {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Errorf("Color %d has empty key", i)
		}
	}
}

func TestImportKnownAdobeFiles(t *testing.T) {
	tests := map[string]struct {
		filename   string
		colorType  adobeColorbook.ColorType
		checkFirst func(t *testing.T, c color.Color)
	}{
		"focoltone_cmyk": {
			filename:  "../../testdata/FOCOLTONE.acb",
			colorType: adobeColorbook.ColorTypeCMYK,
			checkFirst: func(t *testing.T, c color.Color) {
				// First color "1070" should be pure Cyan (C=100, M=0, Y=0, K=0)
				cmyk := c.ToCMYK()
				if cmyk.C != 100 || cmyk.M != 0 || cmyk.Y != 0 || cmyk.K != 0 {
					t.Errorf("FOCOLTONE first color: got CMYK(%d,%d,%d,%d), want (100,0,0,0)",
						cmyk.C, cmyk.M, cmyk.Y, cmyk.K)
				}
			},
		},
		"dic_lab": {
			filename:  "../../testdata/DIC_Color_Guide.acb",
			colorType: adobeColorbook.ColorTypeLab,
			checkFirst: func(t *testing.T, c color.Color) {
				// First color "1s" - verify L is reasonable (should be around 87)
				lab := c.ToLAB()
				if lab.L < 80 || lab.L > 95 {
					t.Errorf("DIC first color L=%d, expected around 87", lab.L)
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Skipf("Test file not found: %v", err)
			}

			importer := colorbook.NewImporter()
			p, err := importer.Import(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Import failed: %v", err)
			}

			if p.Len() == 0 {
				t.Fatal("Imported palette has no colors")
			}

			// Check first color
			nc, _ := p.Get(0)
			tt.checkFirst(t, nc.Color)
		})
	}
}
