package palette

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/kennyp/palette/color"
)

// Palette represents a collection of colors with metadata.
type Palette struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Colors      []NamedColor `json:"colors"`
	metadata    map[string]any
}

// NamedColor represents a color with an optional name.
type NamedColor struct {
	Name  string      `json:"name,omitempty"`
	Color color.Color `json:"color"`
}

// New creates a new empty palette with the given name.
func New(name string) *Palette {
	return &Palette{
		Name:     name,
		Colors:   make([]NamedColor, 0),
		metadata: make(map[string]any),
	}
}

// NewWithColors creates a new palette with the given name and colors.
func NewWithColors(name string, colors ...NamedColor) *Palette {
	p := New(name)
	p.Colors = append(p.Colors, colors...)
	return p
}

// Add adds a color to the palette.
func (p *Palette) Add(color color.Color, name string) {
	p.Colors = append(p.Colors, NamedColor{
		Name:  name,
		Color: color,
	})
}

// AddColor adds a color to the palette without a name.
func (p *Palette) AddColor(color color.Color) {
	p.Colors = append(p.Colors, NamedColor{
		Color: color,
	})
}

// Remove removes the color at the given index.
func (p *Palette) Remove(index int) error {
	if index < 0 || index >= len(p.Colors) {
		return fmt.Errorf("index %d out of range [0, %d)", index, len(p.Colors))
	}

	p.Colors = slices.Delete(p.Colors, index, index+1)
	return nil
}

// RemoveByName removes the first color with the given name.
func (p *Palette) RemoveByName(name string) bool {
	for i, c := range p.Colors {
		if c.Name == name {
			p.Colors = slices.Delete(p.Colors, i, i+1)
			return true
		}
	}
	return false
}

// Get returns the color at the given index.
func (p *Palette) Get(index int) (NamedColor, error) {
	if index < 0 || index >= len(p.Colors) {
		return NamedColor{}, fmt.Errorf("index %d out of range [0, %d)", index, len(p.Colors))
	}
	return p.Colors[index], nil
}

// GetByName returns the first color with the given name.
func (p *Palette) GetByName(name string) (NamedColor, bool) {
	for _, c := range p.Colors {
		if c.Name == name {
			return c, true
		}
	}
	return NamedColor{}, false
}

// Len returns the number of colors in the palette.
func (p *Palette) Len() int {
	return len(p.Colors)
}

// IsEmpty returns true if the palette has no colors.
func (p *Palette) IsEmpty() bool {
	return len(p.Colors) == 0
}

// Clear removes all colors from the palette.
func (p *Palette) Clear() {
	p.Colors = p.Colors[:0]
}

// Clone creates a deep copy of the palette.
func (p *Palette) Clone() *Palette {
	clone := &Palette{
		Name:        p.Name,
		Description: p.Description,
		Colors:      make([]NamedColor, len(p.Colors)),
		metadata:    make(map[string]any),
	}

	copy(clone.Colors, p.Colors)

	// Copy metadata
	maps.Copy(clone.metadata, p.metadata)

	return clone
}

// Filter returns a new palette containing only colors that match the predicate.
func (p *Palette) Filter(predicate func(NamedColor) bool) *Palette {
	filtered := New(p.Name + " (filtered)")
	filtered.Description = p.Description

	for _, c := range p.Colors {
		if predicate(c) {
			filtered.Colors = append(filtered.Colors, c)
		}
	}

	return filtered
}

// FilterByColorSpace returns a new palette containing only colors in the given color space.
func (p *Palette) FilterByColorSpace(colorSpace string) *Palette {
	return p.Filter(func(c NamedColor) bool {
		return c.Color.ColorSpace() == colorSpace
	})
}

// Map applies a function to each color and returns a new palette.
func (p *Palette) Map(mapper func(NamedColor) NamedColor) *Palette {
	mapped := New(p.Name + " (mapped)")
	mapped.Description = p.Description
	mapped.Colors = make([]NamedColor, len(p.Colors))

	for i, c := range p.Colors {
		mapped.Colors[i] = mapper(c)
	}

	return mapped
}

// ConvertToColorSpace returns a new palette with all colors converted to the specified color space.
func (p *Palette) ConvertToColorSpace(colorSpace string) (*Palette, error) {
	return p.Map(func(c NamedColor) NamedColor {
		var convertedColor color.Color

		switch strings.ToUpper(colorSpace) {
		case "RGB":
			convertedColor = c.Color.ToRGB()
		case "CMYK":
			convertedColor = c.Color.ToCMYK()
		case "LAB":
			convertedColor = c.Color.ToLAB()
		case "HSB":
			convertedColor = c.Color.ToHSB()
		default:
			convertedColor = c.Color // Keep original if unknown color space
		}

		return NamedColor{
			Name:  c.Name,
			Color: convertedColor,
		}
	}), nil
}

// String returns a string representation of the palette.
func (p *Palette) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Palette: %s (%d colors)\n", p.Name, len(p.Colors)))

	if p.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}

	for i, c := range p.Colors {
		if c.Name != "" {
			sb.WriteString(fmt.Sprintf("  [%d] %s: %s\n", i, c.Name, c.Color.String()))
		} else {
			sb.WriteString(fmt.Sprintf("  [%d] %s\n", i, c.Color.String()))
		}
	}

	return sb.String()
}

// SetMetadata sets a metadata value for the palette.
func (p *Palette) SetMetadata(key string, value any) {
	if p.metadata == nil {
		p.metadata = make(map[string]any)
	}
	p.metadata[key] = value
}

// GetMetadata gets a metadata value from the palette.
func (p *Palette) GetMetadata(key string) (any, bool) {
	if p.metadata == nil {
		return nil, false
	}
	value, exists := p.metadata[key]
	return value, exists
}

// RemoveMetadata removes a metadata key from the palette.
func (p *Palette) RemoveMetadata(key string) {
	if p.metadata != nil {
		delete(p.metadata, key)
	}
}

// ListMetadataKeys returns all metadata keys.
func (p *Palette) ListMetadataKeys() []string {
	if p.metadata == nil {
		return nil
	}

	keys := make([]string, 0, len(p.metadata))
	for k := range p.metadata {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Validate checks if the palette is valid.
func (p *Palette) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("palette name cannot be empty")
	}

	// Check for duplicate named colors
	nameCount := make(map[string]int)
	for _, c := range p.Colors {
		if c.Name != "" {
			nameCount[c.Name]++
			if nameCount[c.Name] > 1 {
				return fmt.Errorf("duplicate color name: %s", c.Name)
			}
		}
	}

	return nil
}
