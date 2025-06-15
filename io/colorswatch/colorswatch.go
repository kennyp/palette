package colorswatch

import (
	"fmt"
	"io"

	"github.com/kennyp/palette/adobe/colorswatch"
	"github.com/kennyp/palette/color"
	"github.com/kennyp/palette/palette"
)

// Importer implements importing Adobe Color Swatch (.aco) files.
type Importer struct{}

// NewImporter creates a new Adobe Color Swatch importer.
func NewImporter() *Importer {
	return &Importer{}
}

// Import reads an Adobe Color Swatch file and converts it to a palette.
func (i *Importer) Import(r io.Reader) (*palette.Palette, error) {
	// Read all data from reader
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read color swatch data: %w", err)
	}

	// Parse Adobe Color Swatch
	var acs colorswatch.ColorSwatch
	if err := acs.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("failed to parse color swatch: %w", err)
	}

	// Create palette
	paletteName := "Color Swatch"
	if acs.Version == colorswatch.Version2 {
		paletteName = "Color Swatch (with names)"
	}
	p := palette.New(paletteName)

	// Store metadata
	p.SetMetadata("version", acs.Version)
	p.SetMetadata("format", "Adobe Color Swatch")

	// Convert colors
	for i, c := range acs.Colors {
		paletteColor, err := convertAdobeSwatchColor(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert color at index %d: %w", i, err)
		}

		colorName := c.Name
		if colorName == "" {
			colorName = fmt.Sprintf("Color %d", i+1)
		}

		p.Add(paletteColor, colorName)
	}

	return p, nil
}

// CanImport returns true if this importer can handle the given format.
func (i *Importer) CanImport(format string) bool {
	return format == ".aco" || format == "colorswatch" || format == "swatch"
}

// SupportedFormats returns the list of supported formats.
func (i *Importer) SupportedFormats() []string {
	return []string{".aco", "colorswatch", "swatch"}
}

// Exporter implements exporting to Adobe Color Swatch (.aco) files.
type Exporter struct {
	// Version specifies the ACO version to export (1 or 2)
	Version uint16
}

// NewExporter creates a new Adobe Color Swatch exporter.
// By default, exports version 2 (with color names).
func NewExporter() *Exporter {
	return &Exporter{
		Version: colorswatch.Version2,
	}
}

// NewExporterV1 creates a new Adobe Color Swatch exporter for version 1 (no names).
func NewExporterV1() *Exporter {
	return &Exporter{
		Version: colorswatch.Version1,
	}
}

// Export converts a palette to Adobe Color Swatch format and writes it.
func (e *Exporter) Export(p *palette.Palette, w io.Writer) error {
	// Create Adobe Color Swatch
	acs := &colorswatch.ColorSwatch{
		Version: e.Version,
	}

	// Override version from metadata if available
	if version, ok := p.GetMetadata("version"); ok {
		if v, ok := version.(uint16); ok && (v == colorswatch.Version1 || v == colorswatch.Version2) {
			acs.Version = v
		}
	}

	// Convert colors
	acs.Colors = make([]*colorswatch.Color, 0, p.Len())
	for i := range p.Len() {
		namedColor, err := p.Get(i)
		if err != nil {
			return fmt.Errorf("failed to get color at index %d: %w", i, err)
		}

		adobeColor, err := convertToAdobeSwatchColor(namedColor.Color, namedColor.Name)
		if err != nil {
			return fmt.Errorf("failed to convert color %s: %w", namedColor.Name, err)
		}

		// Only include names for version 2
		if acs.Version == colorswatch.Version1 {
			adobeColor.Name = ""
		}

		acs.Colors = append(acs.Colors, adobeColor)
	}

	// Marshal and write
	data, err := acs.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal color swatch: %w", err)
	}

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write color swatch data: %w", err)
	}

	return nil
}

// CanExport returns true if this exporter can handle the given format.
func (e *Exporter) CanExport(format string) bool {
	return format == ".aco" || format == "colorswatch" || format == "swatch"
}

// SupportedFormats returns the list of supported formats.
func (e *Exporter) SupportedFormats() []string {
	return []string{".aco", "colorswatch", "swatch"}
}

// Helper functions

// convertAdobeSwatchColor converts an Adobe Color Swatch color to a palette color.
func convertAdobeSwatchColor(c *colorswatch.Color) (color.Color, error) {
	switch c.ColorSpace {
	case colorswatch.ColorSpaceRGB:
		// Adobe ACO RGB values are 0-65535, convert to 0-255
		r := uint8(c.Values[0] >> 8)
		g := uint8(c.Values[1] >> 8)
		b := uint8(c.Values[2] >> 8)
		return color.NewRGB(r, g, b), nil

	case colorswatch.ColorSpaceHSB:
		// Adobe ACO HSB: H=0-360*182, S=0-100*655, B=0-100*655
		h := uint16(c.Values[0] / 182)
		s := uint8(c.Values[1] / 655)
		b := uint8(c.Values[2] / 655)
		return color.NewHSB(h, s, b), nil

	case colorswatch.ColorSpaceCMYK:
		// Adobe ACO CMYK values are 0-10000 (representing 0-100%)
		cy := uint8(c.Values[0] / 100)
		mg := uint8(c.Values[1] / 100)
		ye := uint8(c.Values[2] / 100)
		k := uint8(c.Values[3] / 100)
		return color.NewCMYK(cy, mg, ye, k), nil

	case colorswatch.ColorSpaceLab:
		// Adobe ACO LAB format
		// L: 0-10000 (0-100), a: -12800 to 12700 (-128 to 127), b: -12800 to 12700 (-128 to 127)
		l := int8(c.Values[0] / 100)
		a := int8((int32(c.Values[1]) - 12800) / 100)
		b := int8((int32(c.Values[2]) - 12800) / 100)
		return color.NewLAB(l, a, b), nil

	case colorswatch.ColorSpaceGrayscale:
		// Grayscale: convert to RGB
		gray := uint8(c.Values[0] / 100)
		return color.NewRGB(gray, gray, gray), nil

	case colorswatch.ColorSpacePantone,
		colorswatch.ColorSpaceFocoltone,
		colorswatch.ColorSpaceTruematch,
		colorswatch.ColorSpaceToyo,
		colorswatch.ColorSpaceHKS:
		// For spot colors, convert to RGB as approximation
		// This is a simplification - in reality these would need proper color management
		r := uint8(c.Values[0] >> 8)
		g := uint8(c.Values[1] >> 8)
		b := uint8(c.Values[2] >> 8)
		return color.NewRGB(r, g, b), nil

	default:
		return nil, fmt.Errorf("unsupported color space: %v", c.ColorSpace)
	}
}

// convertToAdobeSwatchColor converts a palette color to an Adobe Color Swatch color.
func convertToAdobeSwatchColor(c color.Color, name string) (*colorswatch.Color, error) {
	adobeColor := &colorswatch.Color{
		Name: name,
	}

	// Determine the best color space based on the input color type
	switch c.ColorSpace() {
	case "RGB":
		rgb := c.ToRGB()
		adobeColor.ColorSpace = colorswatch.ColorSpaceRGB
		// Convert 0-255 to 0-65535
		adobeColor.Values = [4]uint16{
			uint16(rgb.R) << 8,
			uint16(rgb.G) << 8,
			uint16(rgb.B) << 8,
			0, // Alpha/unused
		}

	case "HSB":
		hsb := c.ToHSB()
		adobeColor.ColorSpace = colorswatch.ColorSpaceHSB
		// Convert to Adobe ACO HSB format
		adobeColor.Values = [4]uint16{
			uint16(hsb.H) * 182, // H: 0-360 -> 0-65520
			uint16(hsb.S) * 655, // S: 0-100 -> 0-65500
			uint16(hsb.B) * 655, // B: 0-100 -> 0-65500
			0,                   // Unused
		}

	case "CMYK":
		cmyk := c.ToCMYK()
		adobeColor.ColorSpace = colorswatch.ColorSpaceCMYK
		// Convert 0-100 to 0-10000
		adobeColor.Values = [4]uint16{
			uint16(cmyk.C) * 100,
			uint16(cmyk.M) * 100,
			uint16(cmyk.Y) * 100,
			uint16(cmyk.K) * 100,
		}

	case "LAB":
		lab := c.ToLAB()
		adobeColor.ColorSpace = colorswatch.ColorSpaceLab
		// Convert to Adobe ACO LAB format
		adobeColor.Values = [4]uint16{
			uint16(lab.L) * 100,              // L: 0-100 -> 0-10000
			uint16(int32(lab.A)*100 + 12800), // a: -128 to 127 -> 0 to 25500
			uint16(int32(lab.B)*100 + 12800), // b: -128 to 127 -> 0 to 25500
			0,                                // Unused
		}

	default:
		// Default to RGB conversion
		rgb := c.ToRGB()
		adobeColor.ColorSpace = colorswatch.ColorSpaceRGB
		adobeColor.Values = [4]uint16{
			uint16(rgb.R) << 8,
			uint16(rgb.G) << 8,
			uint16(rgb.B) << 8,
			0,
		}
	}

	return adobeColor, nil
}
