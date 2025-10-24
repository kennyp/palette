# Palette

A Go library for working with collections of colors. It provides a unified interface for importing and exporting color palettes in various formats including Adobe Color Book (.acb), Adobe Color Swatch (.aco), CSV, and JSON.

## Features

- **Multiple Color Spaces**: Support for RGB, CMYK, LAB, and HSB color spaces with automatic conversion between them
- **Palette Management**: Create, manipulate, and organize collections of colors
- **Multiple Format Support**: Import and export palettes in various formats:
  - Adobe Color Book (.acb)
  - Adobe Color Swatch (.aco) 
  - CSV with flexible color representations
  - JSON with extensible schema
- **Extensible Architecture**: Pluggable import/export system for easy format additions
- **Color Space Conversion**: High-quality color space conversions with proper gamma correction and illuminant handling
- **CLI & Web Interface**: Command-line tool and web server for easy palette conversion without writing code ([see CLI docs](cmd/palette/README.md))

## Installation

### As a Library

```bash
go get github.com/kennyp/palette
```

### As a CLI Tool

```bash
go install github.com/kennyp/palette/cmd/palette@latest
```

See the [CLI documentation](cmd/palette/README.md) for command-line usage and web server features.

## Quick Start

```go
package main

import (
	"fmt"
	"strings"

	_ "github.com/kennyp/palette" // Initialize format importers/exporters
	"github.com/kennyp/palette/color"
	paletteio "github.com/kennyp/palette/io"
	"github.com/kennyp/palette/palette"
)

func main() {
	// Create a new palette
	p := palette.New("My Palette")
	
	// Add colors in different formats
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewCMYK(100, 0, 100, 0), "Green")
	p.Add(color.NewHSB(240, 100, 100), "Blue")
	
	// Convert all colors to a specific color space
	cmykPalette, _ := p.ConvertToColorSpace("CMYK")
	
	// Export to JSON
	var output strings.Builder
	paletteio.Export(p, &output, ".json")
	fmt.Println(output.String())
}
```

## Color Types

The library supports four main color spaces:

### RGB
```go
red := color.NewRGB(255, 0, 0)
redFloat := color.NewRGBFromFloat(1.0, 0.0, 0.0)
```

### CMYK
```go
cyan := color.NewCMYK(100, 0, 0, 0) // C, M, Y, K as percentages (0-100)
```

### LAB
```go
lab := color.NewLAB(50, 20, -30) // L: 0-100, A,B: -128 to 127
```

### HSB (HSV)
```go
blue := color.NewHSB(240, 100, 100) // H: 0-359°, S,B: 0-100%
```

All color types implement the `Color` interface and can be converted between formats:

```go
rgb := color.NewRGB(255, 0, 0)
cmyk := rgb.ToCMYK()  // Automatic conversion
lab := rgb.ToLAB()    // With proper color space handling
hsb := rgb.ToHSB()
```

## Palette Operations

```go
// Create palette
p := palette.New("My Colors")
p.Description = "A collection of brand colors"

// Add colors
p.Add(color.NewRGB(255, 0, 0), "Brand Red")
p.AddColor(color.NewRGB(0, 255, 0)) // Without name

// Access colors
if c, err := p.Get(0); err == nil {
	fmt.Println(c.Name, c.Color.String())
}

// Filter colors
rgbOnly := p.FilterByColorSpace("RGB")

// Transform colors
brightened := p.Map(func(nc palette.NamedColor) palette.NamedColor {
	hsb := nc.Color.ToHSB()
	hsb.B = 100 // Set brightness to 100%
	return palette.NamedColor{Name: nc.Name, Color: hsb}
})

// Clone palette
backup := p.Clone()
```

## Import/Export

The library uses a registry-based system for format support:

```go
// List supported formats
importFormats := paletteio.DefaultRegistry.ListSupportedImportFormats()
exportFormats := paletteio.DefaultRegistry.ListSupportedExportFormats()

// Import from reader
palette, err := paletteio.Import(reader, ".acb")

// Export to writer  
err = paletteio.Export(palette, writer, ".json")

// Auto-detect format from filename
palette, err = paletteio.ImportFromFile("colors.aco", reader)
err = paletteio.ExportToFile(palette, "output.csv", writer)
```

### Format-Specific Features

#### CSV Export Options
```go
exporter := csv.NewExporter()
exporter.ColorFormat = csv.FormatHex        // Export as hex colors
exporter.IncludeHeader = true               // Include column headers
exporter.Delimiter = ';'                    // Use semicolon delimiter
```

#### JSON Export Options  
```go
exporter := json.NewExporter()
exporter.PrettyPrint = true                 // Format with indentation
exporter.ColorFormat = json.FormatAll       // Include all color representations
exporter.IncludeMetadata = true             // Include palette metadata
```

#### Adobe Color Swatch Versions
```go
// Export ACO version 1 (no names)
exporter := colorswatch.NewExporterV1()

// Export ACO version 2 (with names) - default
exporter := colorswatch.NewExporter()
```

## CLI Tool

The Palette library includes a command-line tool for converting palette files and a web server with a user-friendly interface:

```bash
# Convert files via CLI
palette convert -i colors.aco -o colors.json
palette convert -i palette.acb -o palette.csv --colorspace RGB

# Start web server
palette serve --port 8080
```

The web interface provides drag-and-drop file upload, format selection, and instant conversion. See the [CLI documentation](cmd/palette/README.md) for complete usage details.

## Architecture

The library is organized around three main concepts:

1. **Color Types** (`color/`): Core color representations with conversion methods
2. **Palette Collections** (`palette/`): Manage groups of named colors with metadata
3. **Import/Export System** (`io/`): Pluggable format support through registries

This design enables:
- Easy extension with new color formats
- Consistent API across all supported formats  
- Composition for building higher-level applications (CLI tools, web services, etc.)

## Supported Formats

| Format | Extension | Import | Export | Notes |
|--------|-----------|--------|--------|-------|
| Adobe Color Book | .acb | ✅ | ✅ | Binary format with metadata |
| Adobe Color Swatch | .aco | ✅ | ✅ | Version 1 & 2 support |
| CSV | .csv | ✅ | ✅ | Multiple color representations |
| JSON | .json | ✅ | ✅ | Flexible schema support |

## Contributing

The library is designed to be easily extensible. To add support for a new format:

1. Create a new package under `io/yourformat/`
2. Implement the `Importer` and/or `Exporter` interfaces
3. Register your implementations with the default registry

See existing format implementations for examples.

## License

[Add your license here]
