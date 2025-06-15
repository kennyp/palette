package colorswatch_test

import (
	"testing"
	
	"github.com/kennyp/palette/adobe/colorswatch"
)

func TestColorSpaceString(t *testing.T) {
	tests := map[colorswatch.ColorSpace]string{
		colorswatch.ColorSpaceRGB:       "RGB",
		colorswatch.ColorSpaceHSB:       "HSB",
		colorswatch.ColorSpaceCMYK:      "CMYK",
		colorswatch.ColorSpaceLab:       "Lab",
		colorswatch.ColorSpaceGrayscale: "Grayscale",
	}

	for colorSpace, expected := range tests {
		if got := colorSpace.String(); got != expected {
			t.Errorf("ColorSpace(%d).String() = %s, want %s", colorSpace, got, expected)
		}
	}
}

func TestConstants(t *testing.T) {
	if colorswatch.FileTypeMac != "8BCO" {
		t.Errorf("FileTypeMac = %s, want 8BCO", colorswatch.FileTypeMac)
	}
	
	if colorswatch.FileTypeWindows != "ACO" {
		t.Errorf("FileTypeWindows = %s, want ACO", colorswatch.FileTypeWindows)
	}
	
	if colorswatch.Version1 != 1 {
		t.Errorf("Version1 = %d, want 1", colorswatch.Version1)
	}
	
	if colorswatch.Version2 != 2 {
		t.Errorf("Version2 = %d, want 2", colorswatch.Version2)
	}
}

func TestNewColorSwatch(t *testing.T) {
	cs := &colorswatch.ColorSwatch{
		Version: colorswatch.Version1,
		Colors:  []*colorswatch.Color{},
	}
	
	if cs.Version != colorswatch.Version1 {
		t.Errorf("ColorSwatch Version = %d, want %d", cs.Version, colorswatch.Version1)
	}
	
	if cs.Colors == nil {
		t.Error("ColorSwatch Colors should not be nil")
	}
}

func TestNewColor(t *testing.T) {
	color := &colorswatch.Color{
		ColorSpace: colorswatch.ColorSpaceRGB,
		Values:     [4]uint16{255, 0, 0, 0}, // RGB + alpha
		Name:       "Test Red",
	}
	
	if color.ColorSpace != colorswatch.ColorSpaceRGB {
		t.Errorf("Color ColorSpace = %v, want %v", color.ColorSpace, colorswatch.ColorSpaceRGB)
	}
	
	if color.Name != "Test Red" {
		t.Errorf("Color Name = %s, want Test Red", color.Name)
	}
	
	expectedValues := [4]uint16{255, 0, 0, 0}
	if color.Values != expectedValues {
		t.Errorf("Color Values = %v, want %v", color.Values, expectedValues)
	}
}

func TestColorSpaceValidation(t *testing.T) {
	validColorSpaces := []colorswatch.ColorSpace{
		colorswatch.ColorSpaceRGB,
		colorswatch.ColorSpaceHSB,
		colorswatch.ColorSpaceCMYK,
		colorswatch.ColorSpaceLab,
		colorswatch.ColorSpaceGrayscale,
	}
	
	for _, cs := range validColorSpaces {
		if cs.String() == "" {
			t.Errorf("ColorSpace %d should have a string representation", cs)
		}
	}
}