package colorbook

import (
	"fmt"
	"io"

	"github.com/kennyp/palette/adobe/colorbook"
	"github.com/kennyp/palette/color"
	"github.com/kennyp/palette/palette"
)

// Importer implements importing Adobe Color Book (.acb) files.
type Importer struct{}

// NewImporter creates a new Adobe Color Book importer.
func NewImporter() *Importer {
	return &Importer{}
}

// Import reads an Adobe Color Book file and converts it to a palette.
func (i *Importer) Import(r io.Reader) (*palette.Palette, error) {
	// Read all data from reader
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read color book data: %w", err)
	}

	// Parse Adobe Color Book
	var acb colorbook.ColorBook
	if err := acb.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("failed to parse color book: %w", err)
	}

	// Create palette
	p := palette.New(acb.Title)
	if acb.Description != "" {
		p.Description = acb.Description
	}

	// Store metadata
	p.SetMetadata("book_id", acb.ID)
	p.SetMetadata("version", acb.Version)
	p.SetMetadata("prefix", acb.Prefix)
	p.SetMetadata("postfix", acb.Postfix)
	p.SetMetadata("colors_per_page", acb.ColorsPerPage)
	p.SetMetadata("key_color_page", acb.KeyColorPage)
	p.SetMetadata("color_type", acb.ColorType)
	p.SetMetadata("format", "Adobe Color Book")

	// Convert colors
	for _, c := range acb.Colors {
		paletteColor, err := convertAdobeColor(c, acb.ColorType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert color %s: %w", c.Name, err)
		}

		p.Add(paletteColor, c.Name)
	}

	return p, nil
}

// CanImport returns true if this importer can handle the given format.
func (i *Importer) CanImport(format string) bool {
	return format == ".acb" || format == "colorbook"
}

// SupportedFormats returns the list of supported formats.
func (i *Importer) SupportedFormats() []string {
	return []string{".acb", "colorbook"}
}

// Exporter implements exporting to Adobe Color Book (.acb) files.
type Exporter struct{}

// NewExporter creates a new Adobe Color Book exporter.
func NewExporter() *Exporter {
	return &Exporter{}
}

// Export converts a palette to Adobe Color Book format and writes it.
func (e *Exporter) Export(p *palette.Palette, w io.Writer) error {
	// Create Adobe Color Book
	acb := &colorbook.ColorBook{
		Title:       p.Name,
		Description: p.Description,
		Version:     colorbook.DefaultVersion,
	}

	// Set metadata if available
	if bookID, ok := p.GetMetadata("book_id"); ok {
		if id, ok := bookID.(colorbook.BookID); ok {
			acb.ID = id
		}
	}

	if version, ok := p.GetMetadata("version"); ok {
		if v, ok := version.(uint16); ok {
			acb.Version = v
		}
	}

	if prefix, ok := p.GetMetadata("prefix"); ok {
		if s, ok := prefix.(string); ok {
			acb.Prefix = s
		}
	}

	if postfix, ok := p.GetMetadata("postfix"); ok {
		if s, ok := postfix.(string); ok {
			acb.Postfix = s
		}
	}

	if colorsPerPage, ok := p.GetMetadata("colors_per_page"); ok {
		if cpp, ok := colorsPerPage.(uint16); ok {
			acb.ColorsPerPage = cpp
		}
	}

	if keyColorPage, ok := p.GetMetadata("key_color_page"); ok {
		if kcp, ok := keyColorPage.(uint16); ok {
			acb.KeyColorPage = kcp
		}
	}

	if colorType, ok := p.GetMetadata("color_type"); ok {
		if ct, ok := colorType.(colorbook.ColorType); ok {
			acb.ColorType = ct
		}
	} else {
		// Default to RGB if no color type specified
		acb.ColorType = colorbook.ColorTypeRGB
	}

	// Convert colors
	acb.Colors = make([]*colorbook.Color, 0, p.Len())
	for i := 0; i < p.Len(); i++ {
		namedColor, err := p.Get(i)
		if err != nil {
			return fmt.Errorf("failed to get color at index %d: %w", i, err)
		}

		adobeColor, err := convertToAdobeColor(namedColor.Color, namedColor.Name, acb.ColorType)
		if err != nil {
			return fmt.Errorf("failed to convert color %s: %w", namedColor.Name, err)
		}

		acb.Colors = append(acb.Colors, adobeColor)
	}

	// Marshal and write
	data, err := acb.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal color book: %w", err)
	}

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write color book data: %w", err)
	}

	return nil
}

// CanExport returns true if this exporter can handle the given format.
func (e *Exporter) CanExport(format string) bool {
	return format == ".acb" || format == "colorbook"
}

// SupportedFormats returns the list of supported formats.
func (e *Exporter) SupportedFormats() []string {
	return []string{".acb", "colorbook"}
}

// Helper functions

// convertAdobeColor converts an Adobe Color Book color to a palette color.
func convertAdobeColor(c *colorbook.Color, colorType colorbook.ColorType) (color.Color, error) {
	switch colorType {
	case colorbook.ColorTypeRGB:
		// Adobe RGB uses 0-255 values stored in first 3 bytes
		return color.NewRGB(c.Components[0], c.Components[1], c.Components[2]), nil

	case colorbook.ColorTypeCMYK:
		// Adobe CMYK uses 0-100 values stored in first 4 bytes
		// But Components is only 3 bytes, so we need to handle this differently
		// For now, assume the values are scaled 0-255 and convert to 0-100
		cy := uint8((float64(c.Components[0]) / 255.0) * 100)
		mg := uint8((float64(c.Components[1]) / 255.0) * 100)
		ye := uint8((float64(c.Components[2]) / 255.0) * 100)
		// K value might be stored elsewhere or derived
		k := uint8(0) // Default to 0 for now
		return color.NewCMYK(cy, mg, ye, k), nil

	case colorbook.ColorTypeLab:
		// Adobe LAB format
		l := int8(c.Components[0])
		a := int8(c.Components[1])
		b := int8(c.Components[2])
		return color.NewLAB(l, a, b), nil

	default:
		return nil, fmt.Errorf("unsupported color type: %v", colorType)
	}
}

// convertToAdobeColor converts a palette color to an Adobe Color Book color.
func convertToAdobeColor(c color.Color, name string, targetType colorbook.ColorType) (*colorbook.Color, error) {
	adobeColor := &colorbook.Color{
		Name: name,
	}

	switch targetType {
	case colorbook.ColorTypeRGB:
		rgb := c.ToRGB()
		adobeColor.Components = [3]byte{rgb.R, rgb.G, rgb.B}

	case colorbook.ColorTypeCMYK:
		cmyk := c.ToCMYK()
		// Convert CMYK percentages to 0-255 scale for storage
		cy := uint8((float64(cmyk.C) / 100.0) * 255)
		mg := uint8((float64(cmyk.M) / 100.0) * 255)
		ye := uint8((float64(cmyk.Y) / 100.0) * 255)
		adobeColor.Components = [3]byte{cy, mg, ye}

	case colorbook.ColorTypeLab:
		lab := c.ToLAB()
		adobeColor.Components = [3]byte{byte(lab.L), byte(lab.A), byte(lab.B)}

	default:
		return nil, fmt.Errorf("unsupported target color type: %v", targetType)
	}

	return adobeColor, nil
}
