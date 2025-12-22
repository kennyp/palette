package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kennyp/palette/adobe/colorbook"
	paletteio "github.com/kennyp/palette/io"
	"github.com/kennyp/palette/palette"
	_ "github.com/kennyp/palette/palette/all" // Initialize format importers/exporters
)

// ConvertFile converts a palette file from one format to another.
// If fromFormat is empty, it will be detected from the input file extension.
// If toFormat is empty, it will be detected from the output file extension.
// If colorSpace is non-empty, all colors will be converted to that color space.
// If bookID is non-empty and toFormat is .acb, it will be used as the BookID (must be 4000-65535).
func ConvertFile(inputPath, outputPath, fromFormat, toFormat, colorSpace, bookID string) error {
	// Detect formats from file extensions if not specified
	if fromFormat == "" {
		fromFormat = filepath.Ext(inputPath)
	}
	if toFormat == "" {
		toFormat = filepath.Ext(outputPath)
	}

	// Ensure formats have leading dots
	if fromFormat != "" && !strings.HasPrefix(fromFormat, ".") {
		fromFormat = "." + fromFormat
	}
	if toFormat != "" && !strings.HasPrefix(toFormat, ".") {
		toFormat = "." + toFormat
	}

	// Validate formats
	if fromFormat == "" {
		return fmt.Errorf("cannot detect input format from file: %s", inputPath)
	}
	if toFormat == "" {
		return fmt.Errorf("cannot detect output format from file: %s", outputPath)
	}

	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Import palette
	p, err := paletteio.Import(inputFile, fromFormat)
	if err != nil {
		return fmt.Errorf("failed to import palette from %s: %w", fromFormat, err)
	}

	// Convert color space if requested
	if colorSpace != "" {
		p, err = p.ConvertToColorSpace(colorSpace)
		if err != nil {
			return fmt.Errorf("failed to convert to color space %s: %w", colorSpace, err)
		}
	}

	// Set custom BookID for ACB export if provided
	if bookID != "" && (toFormat == ".acb" || toFormat == ".ACB") {
		id, err := strconv.ParseUint(bookID, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid book_id: must be a number between 4000-65535")
		}
		if id < 4000 || id > 65535 {
			return fmt.Errorf("invalid book_id: must be between 4000-65535 (got %d)", id)
		}
		p.SetMetadata("book_id", colorbook.BookID(id))
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Export palette
	if err := paletteio.Export(p, outputFile, toFormat); err != nil {
		return fmt.Errorf("failed to export palette to %s: %w", toFormat, err)
	}

	return nil
}

// GetSupportedFormats returns a list of file extensions for supported formats.
func GetSupportedFormats() []string {
	return []string{".acb", ".aco", ".csv", ".json"}
}

// DetectFormat attempts to detect the format from a file extension.
func DetectFormat(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		return ext
	}
	return ""
}

// ValidateColorSpace checks if the provided color space is valid.
func ValidateColorSpace(cs string) error {
	if cs == "" {
		return nil
	}
	validSpaces := []string{"RGB", "CMYK", "LAB", "HSB"}
	for _, valid := range validSpaces {
		if strings.EqualFold(cs, valid) {
			return nil
		}
	}
	return fmt.Errorf("invalid color space: %s (must be one of: %s)", cs, strings.Join(validSpaces, ", "))
}

// ConvertFromReader converts a palette from a reader to a writer.
// This is useful for web handlers that work with io.Reader/io.Writer.
func ConvertFromReader(input interface{ Read([]byte) (int, error) }, output interface{ Write([]byte) (int, error) }, fromFormat, toFormat, colorSpace string) (*palette.Palette, error) {
	// Ensure formats have leading dots
	if fromFormat != "" && !strings.HasPrefix(fromFormat, ".") {
		fromFormat = "." + fromFormat
	}
	if toFormat != "" && !strings.HasPrefix(toFormat, ".") {
		toFormat = "." + toFormat
	}

	// Import palette
	p, err := paletteio.Import(input, fromFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to import palette from %s: %w", fromFormat, err)
	}

	// Convert color space if requested
	if colorSpace != "" {
		if err := ValidateColorSpace(colorSpace); err != nil {
			return nil, err
		}
		p, err = p.ConvertToColorSpace(colorSpace)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to color space %s: %w", colorSpace, err)
		}
	}

	// Export palette
	if err := paletteio.Export(p, output, toFormat); err != nil {
		return nil, fmt.Errorf("failed to export palette to %s: %w", toFormat, err)
	}

	return p, nil
}
