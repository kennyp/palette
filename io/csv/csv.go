package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/kennyp/palette/color"
	"github.com/kennyp/palette/palette"
)

// Importer implements importing CSV files containing color data.
type Importer struct {
	// Delimiter specifies the field delimiter (default: comma)
	Delimiter rune
	// HasHeader indicates if the first row contains column headers
	HasHeader bool
	// ColorFormat specifies the expected color format in the CSV
	ColorFormat ColorFormat
}

// ColorFormat represents different ways colors can be represented in CSV.
type ColorFormat int

const (
	// FormatAuto attempts to auto-detect the color format
	FormatAuto ColorFormat = iota
	// FormatRGB expects R,G,B columns (0-255)
	FormatRGB
	// FormatRGBFloat expects R,G,B columns (0.0-1.0)
	FormatRGBFloat
	// FormatHex expects a single hex color column (#RRGGBB)
	FormatHex
	// FormatCMYK expects C,M,Y,K columns (0-100)
	FormatCMYK
	// FormatHSB expects H,S,B columns (H: 0-360, S,B: 0-100)
	FormatHSB
	// FormatLAB expects L,A,B columns
	FormatLAB
)

// NewImporter creates a new CSV importer with default settings.
func NewImporter() *Importer {
	return &Importer{
		Delimiter:   ',',
		HasHeader:   true,
		ColorFormat: FormatAuto,
	}
}

// Import reads a CSV file and converts it to a palette.
func (i *Importer) Import(r io.Reader) (*palette.Palette, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = i.Delimiter

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Skip header if present
	startRow := 0
	if i.HasHeader {
		startRow = 1
	}

	if len(records) <= startRow {
		return nil, fmt.Errorf("CSV file contains no data rows")
	}

	// Auto-detect format if needed
	format := i.ColorFormat
	if format == FormatAuto {
		format = i.detectFormat(records[startRow])
	}

	// Create palette
	p := palette.New("CSV Import")
	p.SetMetadata("format", "CSV")
	p.SetMetadata("color_format", format)

	// Parse colors
	for rowIndex := startRow; rowIndex < len(records); rowIndex++ {
		record := records[rowIndex]
		if len(record) == 0 {
			continue // Skip empty rows
		}

		colorName, c, err := i.parseRow(record, format)
		if err != nil {
			return nil, fmt.Errorf("failed to parse row %d: %w", rowIndex+1, err)
		}

		p.Add(c, colorName)
	}

	return p, nil
}

// CanImport returns true if this importer can handle the given format.
func (i *Importer) CanImport(format string) bool {
	return format == ".csv"
}

// SupportedFormats returns the list of supported formats.
func (i *Importer) SupportedFormats() []string {
	return []string{".csv"}
}

// detectFormat attempts to auto-detect the color format from a sample row.
func (i *Importer) detectFormat(record []string) ColorFormat {
	if len(record) == 0 {
		return FormatRGB // Default fallback
	}

	// Check for hex color (single column with # prefix)
	if len(record) >= 1 {
		if strings.HasPrefix(strings.TrimSpace(record[0]), "#") ||
			strings.HasPrefix(strings.TrimSpace(record[len(record)-1]), "#") {
			return FormatHex
		}
	}

	// Count numeric columns
	numericCols := 0
	for _, field := range record {
		if _, err := strconv.ParseFloat(strings.TrimSpace(field), 64); err == nil {
			numericCols++
		}
	}

	// Determine format based on number of numeric columns
	switch numericCols {
	case 3:
		return FormatRGB // Could also be HSB or LAB, but RGB is most common
	case 4:
		return FormatCMYK
	default:
		return FormatRGB // Default fallback
	}
}

// parseRow parses a single CSV row into a color name and color.
func (i *Importer) parseRow(record []string, format ColorFormat) (string, color.Color, error) {
	if len(record) == 0 {
		return "", nil, fmt.Errorf("empty row")
	}

	// Try to find color name (usually first or last column that's not numeric)
	colorName := ""
	colorData := record

	// Check if first column is non-numeric (likely a name)
	if _, err := strconv.ParseFloat(strings.TrimSpace(record[0]), 64); err != nil {
		colorName = strings.TrimSpace(record[0])
		colorData = record[1:]
	} else if len(record) > 3 {
		// Check if last column is non-numeric (name at end)
		if _, err := strconv.ParseFloat(strings.TrimSpace(record[len(record)-1]), 64); err != nil {
			colorName = strings.TrimSpace(record[len(record)-1])
			colorData = record[:len(record)-1]
		}
	}

	// Parse color based on format
	c, err := i.parseColor(colorData, format)
	if err != nil {
		return "", nil, err
	}

	if colorName == "" {
		colorName = c.String()
	}

	return colorName, c, nil
}

// parseColor parses color data from CSV fields.
func (i *Importer) parseColor(fields []string, format ColorFormat) (color.Color, error) {
	switch format {
	case FormatHex:
		return i.parseHexColor(fields)
	case FormatRGB:
		return i.parseRGBColor(fields, false)
	case FormatRGBFloat:
		return i.parseRGBColor(fields, true)
	case FormatCMYK:
		return i.parseCMYKColor(fields)
	case FormatHSB:
		return i.parseHSBColor(fields)
	case FormatLAB:
		return i.parseLABColor(fields)
	default:
		return nil, fmt.Errorf("unsupported color format: %v", format)
	}
}

// parseHexColor parses a hex color string.
func (i *Importer) parseHexColor(fields []string) (color.Color, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("no hex color data")
	}

	// Find the hex color in the fields
	var hexStr string
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if strings.HasPrefix(field, "#") {
			hexStr = field
			break
		}
	}

	if hexStr == "" {
		return nil, fmt.Errorf("no hex color found")
	}

	// Remove # prefix
	hexStr = strings.TrimPrefix(hexStr, "#")

	if len(hexStr) != 6 {
		return nil, fmt.Errorf("invalid hex color format: %s", hexStr)
	}

	// Parse RGB components
	r, err := strconv.ParseUint(hexStr[0:2], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid red component: %w", err)
	}

	g, err := strconv.ParseUint(hexStr[2:4], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid green component: %w", err)
	}

	b, err := strconv.ParseUint(hexStr[4:6], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid blue component: %w", err)
	}

	return color.NewRGB(uint8(r), uint8(g), uint8(b)), nil
}

// parseRGBColor parses RGB color components.
func (i *Importer) parseRGBColor(fields []string, isFloat bool) (color.Color, error) {
	if len(fields) < 3 {
		return nil, fmt.Errorf("insufficient RGB data: need 3 values, got %d", len(fields))
	}

	if isFloat {
		// Parse float values (0.0-1.0)
		r, err := strconv.ParseFloat(strings.TrimSpace(fields[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid red component: %w", err)
		}

		g, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid green component: %w", err)
		}

		b, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid blue component: %w", err)
		}

		return color.NewRGBFromFloat(r, g, b), nil
	} else {
		// Parse integer values (0-255)
		r, err := strconv.ParseUint(strings.TrimSpace(fields[0]), 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid red component: %w", err)
		}

		g, err := strconv.ParseUint(strings.TrimSpace(fields[1]), 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid green component: %w", err)
		}

		b, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid blue component: %w", err)
		}

		return color.NewRGB(uint8(r), uint8(g), uint8(b)), nil
	}
}

// parseCMYKColor parses CMYK color components.
func (i *Importer) parseCMYKColor(fields []string) (color.Color, error) {
	if len(fields) < 4 {
		return nil, fmt.Errorf("insufficient CMYK data: need 4 values, got %d", len(fields))
	}

	c, err := strconv.ParseUint(strings.TrimSpace(fields[0]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid cyan component: %w", err)
	}

	m, err := strconv.ParseUint(strings.TrimSpace(fields[1]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid magenta component: %w", err)
	}

	y, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid yellow component: %w", err)
	}

	k, err := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid key component: %w", err)
	}

	return color.NewCMYK(uint8(c), uint8(m), uint8(y), uint8(k)), nil
}

// parseHSBColor parses HSB color components.
func (i *Importer) parseHSBColor(fields []string) (color.Color, error) {
	if len(fields) < 3 {
		return nil, fmt.Errorf("insufficient HSB data: need 3 values, got %d", len(fields))
	}

	h, err := strconv.ParseUint(strings.TrimSpace(fields[0]), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid hue component: %w", err)
	}

	s, err := strconv.ParseUint(strings.TrimSpace(fields[1]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid saturation component: %w", err)
	}

	b, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid brightness component: %w", err)
	}

	return color.NewHSB(uint16(h), uint8(s), uint8(b)), nil
}

// parseLABColor parses LAB color components.
func (i *Importer) parseLABColor(fields []string) (color.Color, error) {
	if len(fields) < 3 {
		return nil, fmt.Errorf("insufficient LAB data: need 3 values, got %d", len(fields))
	}

	l, err := strconv.ParseInt(strings.TrimSpace(fields[0]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid L component: %w", err)
	}

	a, err := strconv.ParseInt(strings.TrimSpace(fields[1]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid A component: %w", err)
	}

	b, err := strconv.ParseInt(strings.TrimSpace(fields[2]), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid B component: %w", err)
	}

	return color.NewLAB(int8(l), int8(a), int8(b)), nil
}

// Exporter implements exporting palettes to CSV format.
type Exporter struct {
	// Delimiter specifies the field delimiter (default: comma)
	Delimiter rune
	// IncludeHeader indicates if column headers should be written
	IncludeHeader bool
	// ColorFormat specifies how colors should be formatted in the CSV
	ColorFormat ColorFormat
}

// NewExporter creates a new CSV exporter with default settings.
func NewExporter() *Exporter {
	return &Exporter{
		Delimiter:     ',',
		IncludeHeader: true,
		ColorFormat:   FormatRGB,
	}
}

// Export converts a palette to CSV format and writes it.
func (e *Exporter) Export(p *palette.Palette, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = e.Delimiter
	defer csvWriter.Flush()

	// Write header if requested
	if e.IncludeHeader {
		header := e.getHeader()
		if err := csvWriter.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write colors
	for i := range p.Len() {
		namedColor, err := p.Get(i)
		if err != nil {
			return fmt.Errorf("failed to get color at index %d: %w", i, err)
		}

		record := e.formatColor(namedColor)
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write color %s: %w", namedColor.Name, err)
		}
	}

	return nil
}

// CanExport returns true if this exporter can handle the given format.
func (e *Exporter) CanExport(format string) bool {
	return format == ".csv"
}

// SupportedFormats returns the list of supported formats.
func (e *Exporter) SupportedFormats() []string {
	return []string{".csv"}
}

// getHeader returns the appropriate CSV header for the color format.
func (e *Exporter) getHeader() []string {
	switch e.ColorFormat {
	case FormatHex:
		return []string{"Name", "Hex"}
	case FormatRGB, FormatRGBFloat:
		return []string{"Name", "R", "G", "B"}
	case FormatCMYK:
		return []string{"Name", "C", "M", "Y", "K"}
	case FormatHSB:
		return []string{"Name", "H", "S", "B"}
	case FormatLAB:
		return []string{"Name", "L", "A", "B"}
	default:
		return []string{"Name", "R", "G", "B"}
	}
}

// formatColor formats a named color according to the export format.
func (e *Exporter) formatColor(namedColor palette.NamedColor) []string {
	name := namedColor.Name
	if name == "" {
		name = namedColor.Color.String()
	}

	switch e.ColorFormat {
	case FormatHex:
		rgb := namedColor.Color.ToRGB()
		hex := fmt.Sprintf("#%02X%02X%02X", rgb.R, rgb.G, rgb.B)
		return []string{name, hex}

	case FormatRGB:
		rgb := namedColor.Color.ToRGB()
		return []string{name, fmt.Sprintf("%d", rgb.R), fmt.Sprintf("%d", rgb.G), fmt.Sprintf("%d", rgb.B)}

	case FormatRGBFloat:
		rgb := namedColor.Color.ToRGB()
		return []string{
			name,
			fmt.Sprintf("%.3f", float64(rgb.R)/255.0),
			fmt.Sprintf("%.3f", float64(rgb.G)/255.0),
			fmt.Sprintf("%.3f", float64(rgb.B)/255.0),
		}

	case FormatCMYK:
		cmyk := namedColor.Color.ToCMYK()
		return []string{name, fmt.Sprintf("%d", cmyk.C), fmt.Sprintf("%d", cmyk.M), fmt.Sprintf("%d", cmyk.Y), fmt.Sprintf("%d", cmyk.K)}

	case FormatHSB:
		hsb := namedColor.Color.ToHSB()
		return []string{name, fmt.Sprintf("%d", hsb.H), fmt.Sprintf("%d", hsb.S), fmt.Sprintf("%d", hsb.B)}

	case FormatLAB:
		lab := namedColor.Color.ToLAB()
		return []string{name, fmt.Sprintf("%d", lab.L), fmt.Sprintf("%d", lab.A), fmt.Sprintf("%d", lab.B)}

	default:
		rgb := namedColor.Color.ToRGB()
		return []string{name, fmt.Sprintf("%d", rgb.R), fmt.Sprintf("%d", rgb.G), fmt.Sprintf("%d", rgb.B)}
	}
}
