// Package colorbook provides types for reading and writing Adobe Color Book files.
//
// Spec is implemented per the [documentation].
//
// [documentation]: https://www.adobe.com/devnet-apps/photoshop/fileformatashtml/#50577411_pgfId-1066780
package colorbook

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	FileType              = "8BCB" // Signature and filetype for a Color Book file.
	DefaultVersion uint16 = 1      // Default ersion of Adobe Color Book format.
)

//go:generate go tool stringer -type=BookID -trimprefix=BookID
type BookID uint16 // Unique ID for a ColorBook

const (
	BookIDANPA                  BookID = 3000
	BookIDFocoltone             BookID = 3001
	BookIDPantoneCoated         BookID = 3002
	BookIDPantoneProcess        BookID = 3003
	BookIDPantoneProSlim        BookID = 3004
	BookIDPantoneUncoated       BookID = 3005
	BookIDToyo                  BookID = 3006
	BookIDTrumatch              BookID = 3007
	BookIDHKSE                  BookID = 3008
	BookIDHKSK                  BookID = 3009
	BookIDHKSN                  BookID = 3010
	BookIDHKSZ                  BookID = 3011
	BookIDDIC                   BookID = 3012
	BookIDPantonePastelCoated   BookID = 3020
	BookIDPantonePastelUncoated BookID = 3021
	BookIDPantoneMetallic       BookID = 3022
)

//go:generate go tool stringer -type=ColorType -trimprefix=ColorType
type ColorType uint16

const (
	ColorTypeRGB  ColorType = 0
	ColorTypeCMYK ColorType = 2
	ColorTypeLab  ColorType = 7
)

type Color struct {
	Name       string  `json:"name"`
	Key        [6]byte `json:"key"`
	Components [4]byte `json:"components"` // Changed to 4 bytes to support CMYK
}

type ColorBook struct {
	ID            BookID    `json:"book_id"`
	Version       uint16    `json:"version"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Prefix        string    `json:"prefix"`
	Postfix       string    `json:"postfix"`
	ColorsPerPage uint16    `json:"colors_per_page"`
	KeyColorPage  uint16    `json:"key_color_page"`
	ColorType     ColorType `json:"color_type"`
	Colors        []*Color  `json:"colors"`
}

func (b *ColorBook) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write signature
	if _, err := buf.WriteString(FileType); err != nil {
		return nil, fmt.Errorf("failed to write signature: %w", err)
	}

	// Write version
	if err := binary.Write(buf, binary.BigEndian, b.Version); err != nil {
		return nil, fmt.Errorf("failed to write version: %w", err)
	}

	// Write book ID
	if err := binary.Write(buf, binary.BigEndian, b.ID); err != nil {
		return nil, fmt.Errorf("failed to write book ID: %w", err)
	}

	// Write strings
	if err := writeString(buf, b.Title); err != nil {
		return nil, fmt.Errorf("failed to write title: %w", err)
	}

	if err := writeString(buf, b.Prefix); err != nil {
		return nil, fmt.Errorf("failed to write prefix: %w", err)
	}

	if err := writeString(buf, b.Postfix); err != nil {
		return nil, fmt.Errorf("failed to write postfix: %w", err)
	}

	if err := writeString(buf, b.Description); err != nil {
		return nil, fmt.Errorf("failed to write description: %w", err)
	}

	// Write number of colors
	numColors := uint16(len(b.Colors))
	if err := binary.Write(buf, binary.BigEndian, numColors); err != nil {
		return nil, fmt.Errorf("failed to write number of colors: %w", err)
	}

	// Write colors per page
	if err := binary.Write(buf, binary.BigEndian, b.ColorsPerPage); err != nil {
		return nil, fmt.Errorf("failed to write colors per page: %w", err)
	}

	// Write key color page
	if err := binary.Write(buf, binary.BigEndian, b.KeyColorPage); err != nil {
		return nil, fmt.Errorf("failed to write key color page: %w", err)
	}

	// Write color type
	if err := binary.Write(buf, binary.BigEndian, b.ColorType); err != nil {
		return nil, fmt.Errorf("failed to write color type: %w", err)
	}

	// Write colors
	for i, color := range b.Colors {
		if err := writeString(buf, color.Name); err != nil {
			return nil, fmt.Errorf("failed to write color %d name: %w", i, err)
		}

		if _, err := buf.Write(color.Key[:]); err != nil {
			return nil, fmt.Errorf("failed to write color %d key: %w", i, err)
		}

		// Write components based on color type
		switch b.ColorType {
		case ColorTypeRGB, ColorTypeLab:
			// RGB and LAB use 3 components
			components := [3]byte{color.Components[0], color.Components[1], color.Components[2]}
			if err := binary.Write(buf, binary.BigEndian, components); err != nil {
				return nil, fmt.Errorf("failed to write color %d components: %w", i, err)
			}
		case ColorTypeCMYK:
			// CMYK uses 4 components
			if err := binary.Write(buf, binary.BigEndian, color.Components); err != nil {
				return nil, fmt.Errorf("failed to write color %d components: %w", i, err)
			}
		default:
			return nil, fmt.Errorf("unsupported color type for writing components: %v", b.ColorType)
		}
	}

	return buf.Bytes(), nil
}

func (b *ColorBook) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)

	signature := make([]byte, 4)
	if _, err := io.ReadFull(buf, signature); err != nil {
		return err
	}

	slog.Debug("verifying signature", slog.String("signature", string(signature)))
	if FileType != string(signature) {
		return errors.New("invalid file type")
	}

	if err := binary.Read(buf, binary.BigEndian, &b.Version); err != nil {
		return err
	}

	slog.Debug("verifying version", slog.Int("version", int(b.Version)))
	if b.Version != DefaultVersion {
		return fmt.Errorf("version %d not supported", b.Version)
	}

	if err := binary.Read(buf, binary.BigEndian, &b.ID); err != nil {
		return fmt.Errorf("failed to parse book id (%w)", err)
	}

	if b.Title, err = readString(slog.With("field", "Title"), buf); err != nil {
		return err
	}

	if b.Prefix, err = readString(slog.With("field", "prefix"), buf); err != nil {
		return err
	}

	if b.Postfix, err = readString(slog.With("field", "postfix"), buf); err != nil {
		return err
	}

	if b.Description, err = readString(slog.With("field", "description"), buf); err != nil {
		return err
	}

	var numColors uint16
	if err := binary.Read(buf, binary.BigEndian, &numColors); err != nil {
		return fmt.Errorf("failed to parse number of colors (%w)", err)
	}

	if err := binary.Read(buf, binary.BigEndian, &b.ColorsPerPage); err != nil {
		return fmt.Errorf("failed to parse colors per page (%w)", err)
	}

	if err := binary.Read(buf, binary.BigEndian, &b.KeyColorPage); err != nil {
		return fmt.Errorf("failed to parse key color page (%w)", err)
	}

	if err := binary.Read(buf, binary.BigEndian, &b.ColorType); err != nil {
		return fmt.Errorf("failed to parse color type (%w)", err)
	}

	if !(b.ColorType == ColorTypeRGB || b.ColorType == ColorTypeCMYK || b.ColorType == ColorTypeLab) {
		return fmt.Errorf("unexpected color type %v", b.ColorType)
	}

	b.Colors = make([]*Color, numColors)
	for i := range b.Colors {
		c := &Color{}

		if c.Name, err = readString(slog.With("color", i, "field", "name"), buf); err != nil {
			return err
		}

		if _, err := io.ReadFull(buf, c.Key[:]); err != nil {
			return fmt.Errorf("failed to read key for color %d (%w)", i, err)
		}

		// Read components based on color type
		switch b.ColorType {
		case ColorTypeRGB, ColorTypeLab:
			// RGB and LAB use 3 components
			var components [3]byte
			if err := binary.Read(buf, binary.BigEndian, &components); err != nil {
				return fmt.Errorf("failed to read components for color %d (%w)", i, err)
			}
			c.Components = [4]byte{components[0], components[1], components[2], 0}
		case ColorTypeCMYK:
			// CMYK uses 4 components
			if err := binary.Read(buf, binary.BigEndian, &c.Components); err != nil {
				return fmt.Errorf("failed to read components for color %d (%w)", i, err)
			}
		default:
			return fmt.Errorf("unsupported color type for reading components: %v", b.ColorType)
		}

		slog.Debug("parsed color", slog.Int("index", i), slog.Any("color", c))

		b.Colors[i] = c
	}

	return nil
}

func readString(log *slog.Logger, r io.Reader) (string, error) {
	var length uint32

	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", fmt.Errorf("failed to read length (%w)", err)
	}

	log.Debug("read length", slog.Any("length", length))

	s := make([]byte, 2*length)
	if _, err := io.ReadFull(r, s); err != nil {
		return "", fmt.Errorf("failed to read bytes (%w)", err)
	}

	ret := &bytes.Buffer{}

	u16s := make([]uint16, 1)
	b8s := make([]byte, 4)
	for i, l := 0, len(s); i < l; i += 2 {
		u16s[0] = uint16(s[i+1]) + (uint16(s[i]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8s, r[0])
		if _, err := ret.Write(b8s[:n]); err != nil {
			return "", fmt.Errorf("failed to write rune %d (%w)", i, err)
		}
	}

	return ret.String(), nil
}

func writeString(w io.Writer, s string) error {
	runes := []rune(s)
	length := uint32(len(runes))

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

	return nil
}
