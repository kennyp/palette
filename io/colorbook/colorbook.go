package colorbook

import (
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"strings"

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
	return format == ".acb" || format == ".ACB" || format == "colorbook"
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

	// Set BookID from metadata or generate one
	if bookID, ok := p.GetMetadata("book_id"); ok {
		if id, ok := bookID.(colorbook.BookID); ok {
			acb.ID = id
		}
	} else {
		// Generate a unique BookID for user-created palettes
		acb.ID = generateBookID(p)
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
	for i := range p.Len() {
		namedColor, err := p.Get(i)
		if err != nil {
			return fmt.Errorf("failed to get color at index %d: %w", i, err)
		}

		adobeColor, err := convertToAdobeColor(namedColor.Color, namedColor.Name, i, acb.ColorType)
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
	return format == ".acb" || format == ".ACB" || format == "colorbook"
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
		// Adobe CMYK: 0=100%, 255=0% (inverted from percentage)
		cy := uint8(math.Round((255 - float64(c.Components[0])) / 255.0 * 100))
		mg := uint8(math.Round((255 - float64(c.Components[1])) / 255.0 * 100))
		ye := uint8(math.Round((255 - float64(c.Components[2])) / 255.0 * 100))
		k := uint8(math.Round((255 - float64(c.Components[3])) / 255.0 * 100))
		return color.NewCMYK(cy, mg, ye, k), nil

	case colorbook.ColorTypeLab:
		// Adobe LAB: L=0-255 (maps to 0-100%), a/b=0-255 (maps to -128 to 127)
		l := int8(math.Round(float64(c.Components[0]) / 2.55))
		a := int8(int(c.Components[1]) - 128)
		b := int8(int(c.Components[2]) - 128)
		return color.NewLAB(l, a, b), nil

	default:
		return nil, fmt.Errorf("unsupported color type: %v", colorType)
	}
}

// convertToAdobeColor converts a palette color to an Adobe Color Book color.
func convertToAdobeColor(c color.Color, name string, index int, targetType colorbook.ColorType) (*colorbook.Color, error) {
	adobeColor := &colorbook.Color{
		Name: name,
		Key:  generateColorKey(name, index),
	}

	switch targetType {
	case colorbook.ColorTypeRGB:
		rgb := c.ToRGB()
		adobeColor.Components = [4]byte{rgb.R, rgb.G, rgb.B, 0}

	case colorbook.ColorTypeCMYK:
		cmyk := c.ToCMYK()
		// Adobe CMYK: 0=100%, 255=0% (inverted from percentage)
		cy := uint8(255 - math.Round(float64(cmyk.C)/100.0*255))
		mg := uint8(255 - math.Round(float64(cmyk.M)/100.0*255))
		ye := uint8(255 - math.Round(float64(cmyk.Y)/100.0*255))
		k := uint8(255 - math.Round(float64(cmyk.K)/100.0*255))
		adobeColor.Components = [4]byte{cy, mg, ye, k}

	case colorbook.ColorTypeLab:
		lab := c.ToLAB()
		// Adobe LAB: L=0-255 (from 0-100%), a/b=0-255 (from -128 to 127)
		adobeColor.Components = [4]byte{
			byte(math.Round(float64(lab.L) * 2.55)),
			byte(int(lab.A) + 128),
			byte(int(lab.B) + 128),
			0,
		}

	default:
		return nil, fmt.Errorf("unsupported target color type: %v", targetType)
	}

	return adobeColor, nil
}

// generateColorKey creates a 6-character catalog code from the color name and index.
// Format: First 3 chars (uppercased) + index padded to 3 digits (e.g., "RED001", "BLU042").
func generateColorKey(name string, index int) [6]byte {
	var key [6]byte

	// Take first 3 chars of name, uppercase
	prefix := strings.ToUpper(name)
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}

	// Format: PREFIX + NNN (e.g., "RED001", "BLU042")
	formatted := fmt.Sprintf("%-3s%03d", prefix, index%1000)
	if len(formatted) > 6 {
		formatted = formatted[:6]
	}

	copy(key[:], formatted)
	return key
}

// generateBookID creates a unique BookID for user-generated palettes.
// Uses FNV-1a hash of palette name, color count, and first color name.
// Returns a value in range 4000-65535 to avoid Adobe's reserved range (3000-3022).
func generateBookID(p *palette.Palette) colorbook.BookID {
	h := fnv.New32a()
	h.Write([]byte(p.Name))
	h.Write([]byte(fmt.Sprintf(":%d:", p.Len())))

	if p.Len() > 0 {
		if nc, err := p.Get(0); err == nil {
			h.Write([]byte(nc.Name))
		}
	}

	hash := h.Sum32()
	// Map to range 4000-65535
	id := uint16(4000 + (hash % (65535 - 4000)))
	return colorbook.BookID(id)
}
