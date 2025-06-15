package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kennyp/palette/color"
	"github.com/kennyp/palette/palette"
)

// Importer implements importing JSON files containing palette data.
type Importer struct {
	// StrictMode determines if unknown fields should cause an error
	StrictMode bool
}

// NewImporter creates a new JSON importer.
func NewImporter() *Importer {
	return &Importer{
		StrictMode: false,
	}
}

// Import reads a JSON file and converts it to a palette.
func (i *Importer) Import(r io.Reader) (*palette.Palette, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON data: %w", err)
	}

	// Try to parse as our standard palette format first
	var paletteData PaletteJSON
	if err := json.Unmarshal(data, &paletteData); err == nil && paletteData.Name != "" {
		return i.convertFromPaletteJSON(paletteData)
	}

	// Try to parse as array of colors
	var colorsArray []ColorJSON
	if err := json.Unmarshal(data, &colorsArray); err == nil && len(colorsArray) > 0 {
		return i.convertFromColorArray(colorsArray)
	}

	// Try to parse as generic color object
	var colorObj map[string]interface{}
	if err := json.Unmarshal(data, &colorObj); err == nil {
		return i.convertFromGenericJSON(colorObj)
	}

	return nil, fmt.Errorf("unable to parse JSON as palette data")
}

// CanImport returns true if this importer can handle the given format.
func (i *Importer) CanImport(format string) bool {
	return format == ".json"
}

// SupportedFormats returns the list of supported formats.
func (i *Importer) SupportedFormats() []string {
	return []string{".json"}
}

// PaletteJSON represents the JSON structure for a complete palette.
type PaletteJSON struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Colors      []ColorJSON `json:"colors"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// ColorJSON represents the JSON structure for a color.
type ColorJSON struct {
	Name       string                 `json:"name,omitempty"`
	ColorSpace string                 `json:"color_space,omitempty"`
	RGB        *RGBValues             `json:"rgb,omitempty"`
	CMYK       *CMYKValues            `json:"cmyk,omitempty"`
	HSB        *HSBValues             `json:"hsb,omitempty"`
	LAB        *LABValues             `json:"lab,omitempty"`
	Hex        string                 `json:"hex,omitempty"`
	Values     interface{}            `json:"values,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RGBValues represents RGB color values.
type RGBValues struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// CMYKValues represents CMYK color values.
type CMYKValues struct {
	C uint8 `json:"c"`
	M uint8 `json:"m"`
	Y uint8 `json:"y"`
	K uint8 `json:"k"`
}

// HSBValues represents HSB color values.
type HSBValues struct {
	H uint16 `json:"h"`
	S uint8  `json:"s"`
	B uint8  `json:"b"`
}

// LABValues represents LAB color values.
type LABValues struct {
	L int8 `json:"l"`
	A int8 `json:"a"`
	B int8 `json:"b"`
}

// convertFromPaletteJSON converts a PaletteJSON to a palette.
func (i *Importer) convertFromPaletteJSON(data PaletteJSON) (*palette.Palette, error) {
	p := palette.New(data.Name)
	p.Description = data.Description

	// Set metadata
	p.SetMetadata("format", "JSON")
	if data.Metadata != nil {
		p.SetMetadata("original_metadata", data.Metadata)
	}

	// Convert colors
	for idx, colorData := range data.Colors {
		c, err := i.convertColorJSON(colorData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert color at index %d: %w", idx, err)
		}

		colorName := colorData.Name
		if colorName == "" {
			colorName = fmt.Sprintf("Color %d", idx+1)
		}

		p.Add(c, colorName)
	}

	return p, nil
}

// convertFromColorArray converts an array of ColorJSON to a palette.
func (i *Importer) convertFromColorArray(colors []ColorJSON) (*palette.Palette, error) {
	p := palette.New("JSON Color Array")
	p.SetMetadata("format", "JSON")

	for idx, colorData := range colors {
		c, err := i.convertColorJSON(colorData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert color at index %d: %w", idx, err)
		}

		colorName := colorData.Name
		if colorName == "" {
			colorName = fmt.Sprintf("Color %d", idx+1)
		}

		p.Add(c, colorName)
	}

	return p, nil
}

// convertFromGenericJSON attempts to parse generic JSON color data.
func (i *Importer) convertFromGenericJSON(data map[string]interface{}) (*palette.Palette, error) {
	p := palette.New("JSON Import")
	p.SetMetadata("format", "JSON")

	// Look for known color fields
	colorCount := 0

	for key, value := range data {
		if c := i.tryParseColorValue(key, value); c != nil {
			p.Add(c, key)
			colorCount++
		}
	}

	if colorCount == 0 {
		return nil, fmt.Errorf("no recognizable color data found in JSON")
	}

	return p, nil
}

// convertColorJSON converts a ColorJSON to a color.Color.
func (i *Importer) convertColorJSON(data ColorJSON) (color.Color, error) {
	// Try each color space in order of preference
	if data.RGB != nil {
		return color.NewRGB(data.RGB.R, data.RGB.G, data.RGB.B), nil
	}

	if data.Hex != "" {
		return i.parseHexColor(data.Hex)
	}

	if data.CMYK != nil {
		return color.NewCMYK(data.CMYK.C, data.CMYK.M, data.CMYK.Y, data.CMYK.K), nil
	}

	if data.HSB != nil {
		return color.NewHSB(data.HSB.H, data.HSB.S, data.HSB.B), nil
	}

	if data.LAB != nil {
		return color.NewLAB(data.LAB.L, data.LAB.A, data.LAB.B), nil
	}

	// Try to parse from generic values field
	if data.Values != nil {
		return i.parseGenericValues(data.Values, data.ColorSpace)
	}

	return nil, fmt.Errorf("no valid color data found")
}

// parseHexColor parses a hex color string.
func (i *Importer) parseHexColor(hex string) (color.Color, error) {
	if len(hex) == 0 {
		return nil, fmt.Errorf("empty hex color")
	}

	// Remove # prefix if present
	if hex[0] == '#' {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color length: %d", len(hex))
	}

	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return nil, fmt.Errorf("failed to parse hex color: %w", err)
	}

	return color.NewRGB(r, g, b), nil
}

// parseGenericValues parses generic color values based on color space.
func (i *Importer) parseGenericValues(values interface{}, colorSpace string) (color.Color, error) {
	// Try to convert to float64 slice
	var nums []float64

	switch v := values.(type) {
	case []interface{}:
		nums = make([]float64, len(v))
		for i, val := range v {
			if f, ok := val.(float64); ok {
				nums[i] = f
			} else {
				return nil, fmt.Errorf("non-numeric value in color array")
			}
		}
	case []float64:
		nums = v
	default:
		return nil, fmt.Errorf("unsupported values format")
	}

	// Parse based on color space and number of values
	switch colorSpace {
	case "RGB", "rgb":
		if len(nums) < 3 {
			return nil, fmt.Errorf("insufficient RGB values")
		}
		return color.NewRGB(uint8(nums[0]), uint8(nums[1]), uint8(nums[2])), nil

	case "CMYK", "cmyk":
		if len(nums) < 4 {
			return nil, fmt.Errorf("insufficient CMYK values")
		}
		return color.NewCMYK(uint8(nums[0]), uint8(nums[1]), uint8(nums[2]), uint8(nums[3])), nil

	case "HSB", "hsb", "HSV", "hsv":
		if len(nums) < 3 {
			return nil, fmt.Errorf("insufficient HSB values")
		}
		return color.NewHSB(uint16(nums[0]), uint8(nums[1]), uint8(nums[2])), nil

	case "LAB", "lab":
		if len(nums) < 3 {
			return nil, fmt.Errorf("insufficient LAB values")
		}
		return color.NewLAB(int8(nums[0]), int8(nums[1]), int8(nums[2])), nil

	default:
		// Default to RGB if no color space specified
		if len(nums) >= 3 {
			return color.NewRGB(uint8(nums[0]), uint8(nums[1]), uint8(nums[2])), nil
		}
	}

	return nil, fmt.Errorf("unable to parse color values")
}

// tryParseColorValue attempts to parse a value as a color.
func (i *Importer) tryParseColorValue(key string, value interface{}) color.Color {
	// Try hex string
	if str, ok := value.(string); ok {
		if c, err := i.parseHexColor(str); err == nil {
			return c
		}
	}

	// Try array of numbers
	if arr, ok := value.([]interface{}); ok && len(arr) >= 3 {
		var nums []float64
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				nums = append(nums, f)
			} else {
				return nil
			}
		}

		if len(nums) >= 3 {
			return color.NewRGB(uint8(nums[0]), uint8(nums[1]), uint8(nums[2]))
		}
	}

	return nil
}

// Exporter implements exporting palettes to JSON format.
type Exporter struct {
	// PrettyPrint determines if the JSON should be formatted with indentation
	PrettyPrint bool
	// IncludeMetadata determines if palette metadata should be included
	IncludeMetadata bool
	// ColorFormat specifies which color representations to include
	ColorFormat ColorFormatFlags
}

// ColorFormatFlags represents which color formats to include in the JSON.
type ColorFormatFlags int

const (
	// FormatPrimary includes only the primary color space for each color
	FormatPrimary ColorFormatFlags = 1 << iota
	// FormatRGB includes RGB values
	FormatRGB
	// FormatHex includes hex representation
	FormatHex
	// FormatCMYK includes CMYK values
	FormatCMYK
	// FormatHSB includes HSB values
	FormatHSB
	// FormatLAB includes LAB values
	FormatLAB
	// FormatAll includes all color representations
	FormatAll = FormatRGB | FormatHex | FormatCMYK | FormatHSB | FormatLAB
)

// NewExporter creates a new JSON exporter with default settings.
func NewExporter() *Exporter {
	return &Exporter{
		PrettyPrint:     true,
		IncludeMetadata: true,
		ColorFormat:     FormatRGB | FormatHex,
	}
}

// Export converts a palette to JSON format and writes it.
func (e *Exporter) Export(p *palette.Palette, w io.Writer) error {
	paletteData := PaletteJSON{
		Name:        p.Name,
		Description: p.Description,
		Colors:      make([]ColorJSON, 0, p.Len()),
	}

	// Include metadata if requested
	if e.IncludeMetadata {
		metadata := make(map[string]interface{})
		for _, key := range p.ListMetadataKeys() {
			if value, ok := p.GetMetadata(key); ok {
				metadata[key] = value
			}
		}
		if len(metadata) > 0 {
			paletteData.Metadata = metadata
		}
	}

	// Convert colors
	for i := 0; i < p.Len(); i++ {
		namedColor, err := p.Get(i)
		if err != nil {
			return fmt.Errorf("failed to get color at index %d: %w", i, err)
		}

		colorJSON := e.convertColorToJSON(namedColor)
		paletteData.Colors = append(paletteData.Colors, colorJSON)
	}

	// Marshal JSON
	var data []byte
	var err error

	if e.PrettyPrint {
		data, err = json.MarshalIndent(paletteData, "", "  ")
	} else {
		data, err = json.Marshal(paletteData)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to output
	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON data: %w", err)
	}

	return nil
}

// CanExport returns true if this exporter can handle the given format.
func (e *Exporter) CanExport(format string) bool {
	return format == ".json"
}

// SupportedFormats returns the list of supported formats.
func (e *Exporter) SupportedFormats() []string {
	return []string{".json"}
}

// convertColorToJSON converts a named color to JSON representation.
func (e *Exporter) convertColorToJSON(namedColor palette.NamedColor) ColorJSON {
	colorJSON := ColorJSON{
		Name:       namedColor.Name,
		ColorSpace: namedColor.Color.ColorSpace(),
	}

	// Include requested color formats
	if e.ColorFormat&FormatRGB != 0 || e.ColorFormat&FormatPrimary != 0 {
		rgb := namedColor.Color.ToRGB()
		colorJSON.RGB = &RGBValues{R: rgb.R, G: rgb.G, B: rgb.B}
	}

	if e.ColorFormat&FormatHex != 0 {
		rgb := namedColor.Color.ToRGB()
		colorJSON.Hex = fmt.Sprintf("#%02X%02X%02X", rgb.R, rgb.G, rgb.B)
	}

	if e.ColorFormat&FormatCMYK != 0 {
		cmyk := namedColor.Color.ToCMYK()
		colorJSON.CMYK = &CMYKValues{C: cmyk.C, M: cmyk.M, Y: cmyk.Y, K: cmyk.K}
	}

	if e.ColorFormat&FormatHSB != 0 {
		hsb := namedColor.Color.ToHSB()
		colorJSON.HSB = &HSBValues{H: hsb.H, S: hsb.S, B: hsb.B}
	}

	if e.ColorFormat&FormatLAB != 0 {
		lab := namedColor.Color.ToLAB()
		colorJSON.LAB = &LABValues{L: lab.L, A: lab.A, B: lab.B}
	}

	return colorJSON
}
