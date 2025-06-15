package palette_test

import (
	"fmt"
	"log"
	"strings"

	_ "github.com/kennyp/palette" // Initialize format importers/exporters
	"github.com/kennyp/palette/color"
	paletteio "github.com/kennyp/palette/io"
	"github.com/kennyp/palette/io/csv"
	"github.com/kennyp/palette/io/json"
	"github.com/kennyp/palette/palette"
)

// Example: Creating a basic color palette and working with colors
func ExamplePalette_basic() {
	// Create a new palette
	p := palette.New("Web Safe Colors")
	p.Description = "A collection of web-safe colors"
	
	// Add colors in different formats
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	p.Add(color.NewRGB(0, 0, 255), "Blue")
	p.Add(color.NewCMYK(100, 0, 100, 0), "Process Green")
	
	// Display palette information
	fmt.Printf("Palette: %s (%d colors)\n", p.Name, p.Len())
	fmt.Printf("Description: %s\n", p.Description)
	
	// Access individual colors
	if color, err := p.Get(0); err == nil {
		fmt.Printf("First color: %s = %s\n", color.Name, color.Color.String())
	}
	
	// Find color by name
	if color, found := p.GetByName("Green"); found {
		fmt.Printf("Green color: %s\n", color.Color.String())
	}
	
	// Output:
	// Palette: Web Safe Colors (4 colors)
	// Description: A collection of web-safe colors
	// First color: Red = RGB(255, 0, 0)
	// Green color: RGB(0, 255, 0)
}

// Example: Color space conversions
func ExampleColor_conversions() {
	// Start with an RGB color
	rgb := color.NewRGB(255, 128, 64)
	fmt.Printf("Original RGB: %s\n", rgb.String())
	
	// Convert to other color spaces
	cmyk := rgb.ToCMYK()
	fmt.Printf("As CMYK: %s\n", cmyk.String())
	
	hsb := rgb.ToHSB()
	fmt.Printf("As HSB: %s\n", hsb.String())
	
	lab := rgb.ToLAB()
	fmt.Printf("As LAB: %s\n", lab.String())
	
	// Convert back to RGB to show round-trip
	backToRGB := cmyk.ToRGB()
	fmt.Printf("CMYK back to RGB: %s\n", backToRGB.String())
	
	// Output:
	// Original RGB: RGB(255, 128, 64)
	// As CMYK: CMYK(0%, 50%, 75%, 0%)
	// As HSB: HSB(20°, 75%, 100%)
	// As LAB: LAB(67, 44, 55)
	// CMYK back to RGB: RGB(255, 128, 64)
}

// Example: Working with different color constructors
func ExampleColor_constructors() {
	// RGB from integers (0-255)
	red := color.NewRGB(255, 0, 0)
	fmt.Printf("RGB from ints: %s\n", red.String())
	
	// RGB from floats (0.0-1.0)
	redFloat := color.NewRGBFromFloat(1.0, 0.0, 0.0)
	fmt.Printf("RGB from floats: %s\n", redFloat.String())
	
	// CMYK (percentages 0-100)
	cyan := color.NewCMYK(100, 0, 0, 0)
	fmt.Printf("CMYK: %s\n", cyan.String())
	
	// HSB (H: 0-359°, S,B: 0-100%)
	blue := color.NewHSB(240, 100, 100)
	fmt.Printf("HSB: %s\n", blue.String())
	
	// LAB (L: 0-100, A,B: -128 to 127)
	lab := color.NewLAB(50, 20, -30)
	fmt.Printf("LAB: %s\n", lab.String())
	
	// Output:
	// RGB from ints: RGB(255, 0, 0)
	// RGB from floats: RGB(255, 0, 0)
	// CMYK: CMYK(100%, 0%, 0%, 0%)
	// HSB: HSB(240°, 100%, 100%)
	// LAB: LAB(50, 20, -30)
}

// Example: Filtering and transforming palettes
func ExamplePalette_filtering() {
	p := palette.New("Mixed Colors")
	p.Add(color.NewRGB(255, 0, 0), "Red RGB")
	p.Add(color.NewCMYK(100, 0, 100, 0), "Green CMYK")
	p.Add(color.NewRGB(0, 0, 255), "Blue RGB")
	p.Add(color.NewHSB(60, 100, 100), "Yellow HSB")
	
	fmt.Printf("Original palette: %d colors\n", p.Len())
	
	// Filter by color space
	rgbOnly := p.FilterByColorSpace("RGB")
	fmt.Printf("RGB colors only: %d colors\n", rgbOnly.Len())
	
	// Custom filter
	primary := p.Filter(func(c palette.NamedColor) bool {
		return strings.Contains(strings.ToLower(c.Name), "red") ||
			   strings.Contains(strings.ToLower(c.Name), "green") ||
			   strings.Contains(strings.ToLower(c.Name), "blue")
	})
	fmt.Printf("Primary colors: %d colors\n", primary.Len())
	
	// Transform colors (make them all CMYK)
	cmykPalette, _ := p.ConvertToColorSpace("CMYK")
	fmt.Printf("Converted to CMYK: %d colors\n", cmykPalette.Len())
	
	// Check first color's color space
	if first, err := cmykPalette.Get(0); err == nil {
		fmt.Printf("First color space: %s\n", first.Color.ColorSpace())
	}
	
	// Output:
	// Original palette: 4 colors
	// RGB colors only: 2 colors
	// Primary colors: 3 colors
	// Converted to CMYK: 4 colors
	// First color space: CMYK
}

// Example: Exporting palette to JSON
func ExampleExport_json() {
	// Create a palette
	p := palette.New("Brand Colors")
	p.Description = "Primary brand color palette"
	p.Add(color.NewRGB(255, 69, 0), "Brand Orange")
	p.Add(color.NewRGB(0, 123, 255), "Brand Blue")
	p.Add(color.NewRGB(40, 167, 69), "Brand Green")
	
	// Export to JSON with all color formats
	var output strings.Builder
	err := paletteio.Export(p, &output, ".json")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("JSON Export:")
	fmt.Println(output.String())
	
	// Output will show JSON with palette name, description, and colors
}

// Example: Exporting palette to CSV
func ExampleExport_csv() {
	// Create a palette
	p := palette.New("Traffic Light Colors")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(255, 255, 0), "Yellow")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	// Export to CSV with hex colors
	exporter := csv.NewExporter()
	exporter.ColorFormat = csv.FormatHex
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("CSV Export (Hex):")
	fmt.Print(output.String())
	
	// Output:
	// CSV Export (Hex):
	// Name,Hex
	// Red,#FF0000
	// Yellow,#FFFF00
	// Green,#00FF00
}

// Example: Importing palette from CSV
func ExampleImport_csv() {
	// Sample CSV data
	csvData := `Name,R,G,B
Crimson,220,20,60
Gold,255,215,0
Forest Green,34,139,34`
	
	// Import the CSV
	reader := strings.NewReader(csvData)
	p, err := paletteio.Import(reader, ".csv")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Imported palette: %s\n", p.Name)
	fmt.Printf("Number of colors: %d\n", p.Len())
	
	// Show each color
	for i := range p.Len() {
		if c, err := p.Get(i); err == nil {
			fmt.Printf("  %s: %s\n", c.Name, c.Color.String())
		}
	}
	
	// Output:
	// Imported palette: CSV Import
	// Number of colors: 3
	//   Crimson: RGB(220, 20, 60)
	//   Gold: RGB(255, 215, 0)
	//   Forest Green: RGB(34, 139, 34)
}

// Example: Working with palette metadata
func ExamplePalette_metadata() {
	p := palette.New("Corporate Palette")
	p.Description = "Official corporate color scheme"
	
	// Set metadata
	p.SetMetadata("version", "2.1")
	p.SetMetadata("author", "Design Team")
	p.SetMetadata("created", "2024-01-15")
	p.SetMetadata("usage", "Web and print materials")
	
	// Retrieve metadata
	if version, ok := p.GetMetadata("version"); ok {
		fmt.Printf("Palette version: %s\n", version)
	}
	
	// List all metadata keys
	keys := p.ListMetadataKeys()
	fmt.Printf("Metadata keys: %v\n", keys)
	
	// Check metadata count
	fmt.Printf("Total metadata items: %d\n", len(keys))
	
	// Output:
	// Palette version: 2.1
	// Metadata keys: [author created usage version]
	// Total metadata items: 4
}

// Example: Advanced JSON export with custom options
func Example_advancedJSONExport() {
	// Create a palette with mixed color types
	p := palette.New("Design System")
	p.Add(color.NewRGB(255, 59, 48), "System Red")
	p.Add(color.NewCMYK(50, 0, 100, 0), "System Green")
	p.Add(color.NewHSB(211, 86, 100), "System Blue")
	
	// Configure JSON exporter for maximum detail
	exporter := json.NewExporter()
	exporter.PrettyPrint = true
	exporter.ColorFormat = json.FormatAll  // Include all color representations
	exporter.IncludeMetadata = true
	
	var output strings.Builder
	err := exporter.Export(p, &output)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Advanced JSON Export:")
	fmt.Println(output.String())
	
	// Output will show formatted JSON with RGB, hex, CMYK, HSB, and LAB for each color
}

// Example: Using the registry system directly
func ExampleRegistry_usage() {
	// Create a custom registry
	registry := paletteio.NewRegistry()
	
	// Register specific importers/exporters
	registry.RegisterImporter(csv.NewImporter())
	registry.RegisterExporter(json.NewExporter())
	
	// List supported formats
	importFormats := registry.ListSupportedImportFormats()
	exportFormats := registry.ListSupportedExportFormats()
	
	fmt.Printf("Supported import formats: %v\n", importFormats)
	fmt.Printf("Supported export formats: %v\n", exportFormats)
	
	// Use the custom registry
	p := palette.New("Test")
	p.Add(color.NewRGB(128, 128, 128), "Gray")
	
	var output strings.Builder
	err := registry.Export(p, &output, ".json")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Custom registry export successful")
	
	// Output:
	// Supported import formats: [.csv]
	// Supported export formats: [.json]
	// Custom registry export successful
}

// Example: Color validation and error handling
func ExampleColor_validation() {
	// Colors automatically clamp values to valid ranges
	
	// RGB values are clamped to 0-255
	rgb := color.NewRGBFromFloat(1.5, -0.5, 0.5) // Over/under range
	fmt.Printf("Clamped RGB: %s\n", rgb.String())
	
	// CMYK percentages are clamped to 0-100
	cmyk := color.NewCMYK(150, 200, 50, 75) // Over 100%
	fmt.Printf("Clamped CMYK: %s\n", cmyk.String())
	
	// HSB hue wraps around 360°
	hsb := color.NewHSB(400, 150, 150) // Over ranges
	fmt.Printf("Wrapped/clamped HSB: %s\n", hsb.String())
	
	// LAB values are clamped appropriately
	lab := color.NewLAB(120, 127, -128) // Out of typical ranges
	fmt.Printf("Clamped LAB: %s\n", lab.String())
	
	// Output:
	// Clamped RGB: RGB(255, 0, 128)
	// Clamped CMYK: CMYK(100%, 100%, 50%, 75%)
	// Wrapped/clamped HSB: HSB(40°, 100%, 100%)
	// Clamped LAB: LAB(100, 127, -128)
}

// Example: Palette validation and error handling
func ExamplePalette_validation() {
	p := palette.New("Test Palette")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	// Valid palette
	if err := p.Validate(); err == nil {
		fmt.Println("Palette is valid")
	}
	
	// Create invalid palette (empty name)
	invalid := palette.New("")
	if err := invalid.Validate(); err != nil {
		fmt.Printf("Invalid palette: %v\n", err)
	}
	
	// Create palette with duplicate names
	duplicate := palette.New("Duplicate Test")
	duplicate.Add(color.NewRGB(255, 0, 0), "Red")
	duplicate.Add(color.NewRGB(0, 255, 0), "Red") // Same name
	
	if err := duplicate.Validate(); err != nil {
		fmt.Printf("Duplicate names: %v\n", err)
	}
	
	// Safe operations
	if c, err := p.Get(0); err == nil {
		fmt.Printf("First color: %s\n", c.Name)
	}
	
	if err := p.Remove(10); err != nil {
		fmt.Printf("Remove error: %v\n", err)
	}
	
	// Output:
	// Palette is valid
	// Invalid palette: palette name cannot be empty
	// Duplicate names: duplicate color name: Red
	// First color: Red
	// Remove error: index 10 out of range [0, 2)
}