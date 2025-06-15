// Package colorswatch provides types for reading and writing Adobe Color Swatch files.
//
// Implements Adobe Color Swatch (.aco) format specification.
package colorswatch

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	FileTypeMac     = "8BCO" // Signature and filetype for a Mac Color Swatch file
	FileTypeWindows = "ACO"  // Extension for a Windows Color Swatch file
	Version1        = 1      // Version 1 format (no color names)
	Version2        = 2      // Version 2 format (with color names)
)

//go:generate go tool stringer -type=ColorSpace -trimprefix=ColorSpace
type ColorSpace uint16

const (
	ColorSpaceRGB       ColorSpace = 0
	ColorSpaceHSB       ColorSpace = 1
	ColorSpaceCMYK      ColorSpace = 2
	ColorSpaceLab       ColorSpace = 7
	ColorSpaceGrayscale ColorSpace = 8
	ColorSpacePantone   ColorSpace = 3
	ColorSpaceFocoltone ColorSpace = 4
	ColorSpaceTruematch ColorSpace = 5
	ColorSpaceToyo      ColorSpace = 6
	ColorSpaceHKS       ColorSpace = 10
)

type ColorSwatch struct {
	Version uint16   `json:"version"`
	Colors  []*Color `json:"colors"`
}

type Color struct {
	ColorSpace ColorSpace `json:"color_space"`
	Values     [4]uint16  `json:"values"`
	Name       string     `json:"name"`
}

func (cs *ColorSwatch) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write version
	if err := binary.Write(buf, binary.BigEndian, cs.Version); err != nil {
		return nil, fmt.Errorf("failed to write version: %w", err)
	}

	// Write number of colors
	numColors := uint16(len(cs.Colors))
	if err := binary.Write(buf, binary.BigEndian, numColors); err != nil {
		return nil, fmt.Errorf("failed to write number of colors: %w", err)
	}

	// Write colors (version 1 data)
	for i, color := range cs.Colors {
		if err := binary.Write(buf, binary.BigEndian, color.ColorSpace); err != nil {
			return nil, fmt.Errorf("failed to write color space for color %d: %w", i, err)
		}

		if err := binary.Write(buf, binary.BigEndian, color.Values); err != nil {
			return nil, fmt.Errorf("failed to write values for color %d: %w", i, err)
		}
	}

	// If version 2, write color names
	if cs.Version == Version2 {
		// Write version again
		if err := binary.Write(buf, binary.BigEndian, cs.Version); err != nil {
			return nil, fmt.Errorf("failed to write version 2 header: %w", err)
		}

		// Write number of colors again
		if err := binary.Write(buf, binary.BigEndian, numColors); err != nil {
			return nil, fmt.Errorf("failed to write version 2 color count: %w", err)
		}

		// Write colors with names
		for i, color := range cs.Colors {
			if err := binary.Write(buf, binary.BigEndian, color.ColorSpace); err != nil {
				return nil, fmt.Errorf("failed to write v2 color space for color %d: %w", i, err)
			}

			if err := binary.Write(buf, binary.BigEndian, color.Values); err != nil {
				return nil, fmt.Errorf("failed to write v2 values for color %d: %w", i, err)
			}

			if err := writeString(buf, color.Name); err != nil {
				return nil, fmt.Errorf("failed to write name for color %d: %w", i, err)
			}
		}
	}

	return buf.Bytes(), nil
}

func writeString(w io.Writer, s string) error {
	runes := []rune(s)
	length := uint32(len(runes) + 1) // +1 for null terminator

	// Write length
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write string length: %w", err)
	}

	// Encode to UTF-16 and write
	encoded := utf16.Encode(runes)
	for _, u16 := range encoded {
		if err := binary.Write(w, binary.BigEndian, u16); err != nil {
			return fmt.Errorf("failed to write UTF-16 character: %w", err)
		}
	}

	// Write null terminator
	if err := binary.Write(w, binary.BigEndian, uint16(0)); err != nil {
		return fmt.Errorf("failed to write null terminator: %w", err)
	}

	return nil
}

func (cs *ColorSwatch) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)

	// Read version
	if err := binary.Read(buf, binary.BigEndian, &cs.Version); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	slog.Debug("parsed version", slog.Int("version", int(cs.Version)))

	if cs.Version != Version1 && cs.Version != Version2 {
		return fmt.Errorf("unsupported version: %d", cs.Version)
	}

	// Read number of colors
	var numColors uint16
	if err := binary.Read(buf, binary.BigEndian, &numColors); err != nil {
		return fmt.Errorf("failed to read number of colors: %w", err)
	}

	slog.Debug("parsed color count", slog.Int("count", int(numColors)))

	cs.Colors = make([]*Color, numColors)

	// Read colors
	for i := range cs.Colors {
		color := &Color{}

		// Read color space
		if err := binary.Read(buf, binary.BigEndian, &color.ColorSpace); err != nil {
			return fmt.Errorf("failed to read color space for color %d: %w", i, err)
		}

		// Read color values
		if err := binary.Read(buf, binary.BigEndian, &color.Values); err != nil {
			return fmt.Errorf("failed to read values for color %d: %w", i, err)
		}

		slog.Debug("parsed color", slog.Int("index", i), slog.String("colorspace", color.ColorSpace.String()), slog.Any("values", color.Values))

		cs.Colors[i] = color
	}

	// If version 2, read color names
	if cs.Version == Version2 {
		// Skip to start of version 2 data by reading version and count again
		if err := binary.Read(buf, binary.BigEndian, &cs.Version); err != nil {
			return fmt.Errorf("failed to read version 2 header: %w", err)
		}

		var numColors2 uint16
		if err := binary.Read(buf, binary.BigEndian, &numColors2); err != nil {
			return fmt.Errorf("failed to read version 2 color count: %w", err)
		}

		if numColors2 != numColors {
			return fmt.Errorf("version 2 color count mismatch: expected %d, got %d", numColors, numColors2)
		}

		// Read color names
		for i := range cs.Colors {
			// Read color space and values again (version 2 repeats this data)
			var colorSpace ColorSpace
			var values [4]uint16

			if err := binary.Read(buf, binary.BigEndian, &colorSpace); err != nil {
				return fmt.Errorf("failed to read v2 color space for color %d: %w", i, err)
			}

			if err := binary.Read(buf, binary.BigEndian, &values); err != nil {
				return fmt.Errorf("failed to read v2 values for color %d: %w", i, err)
			}

			// Read color name
			name, err := readString(slog.With("color", i, "field", "name"), buf)
			if err != nil {
				return fmt.Errorf("failed to read name for color %d: %w", i, err)
			}

			cs.Colors[i].Name = name

			slog.Debug("parsed color name", slog.Int("index", i), slog.String("name", name))
		}
	}

	return nil
}

func readString(log *slog.Logger, r io.Reader) (string, error) {
	var length uint32

	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}

	log.Debug("read string length", slog.Uint64("length", uint64(length)))

	if length == 0 {
		return "", nil
	}

	// Read UTF-16 encoded string
	s := make([]byte, 2*length)
	if _, err := io.ReadFull(r, s); err != nil {
		return "", fmt.Errorf("failed to read string bytes: %w", err)
	}

	// Convert UTF-16 to UTF-8
	ret := &bytes.Buffer{}
	u16s := make([]uint16, 1)
	b8s := make([]byte, 4)

	for i, l := 0, len(s); i < l; i += 2 {
		u16s[0] = uint16(s[i+1]) + (uint16(s[i]) << 8)
		if u16s[0] == 0 { // Null terminator
			break
		}
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8s, r[0])
		if _, err := ret.Write(b8s[:n]); err != nil {
			return "", fmt.Errorf("failed to write rune at position %d: %w", i, err)
		}
	}

	return ret.String(), nil
}
