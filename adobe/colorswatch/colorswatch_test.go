package colorswatch

import (
	"testing"
)

func TestColorSpaceString(t *testing.T) {
	tests := map[ColorSpace]string{
		ColorSpaceRGB:       "RGB",
		ColorSpaceHSB:       "HSB",
		ColorSpaceCMYK:      "CMYK",
		ColorSpaceLab:       "Lab",
		ColorSpaceGrayscale: "Grayscale",
	}

	for colorSpace, expected := range tests {
		if got := colorSpace.String(); got != expected {
			t.Errorf("ColorSpace(%d).String() = %s, want %s", colorSpace, got, expected)
		}
	}
}

func TestConstants(t *testing.T) {
	if FileTypeMac != "8BCO" {
		t.Errorf("FileTypeMac = %s, want 8BCO", FileTypeMac)
	}
	
	if FileTypeWindows != "ACO" {
		t.Errorf("FileTypeWindows = %s, want ACO", FileTypeWindows)
	}
	
	if Version1 != 1 {
		t.Errorf("Version1 = %d, want 1", Version1)
	}
	
	if Version2 != 2 {
		t.Errorf("Version2 = %d, want 2", Version2)
	}
}

func TestNewColorSwatch(t *testing.T) {
	cs := &ColorSwatch{
		Version: Version1,
		Colors:  []*Color{},
	}
	
	if cs.Version != Version1 {
		t.Errorf("ColorSwatch Version = %d, want %d", cs.Version, Version1)
	}
	
	if cs.Colors == nil {
		t.Error("ColorSwatch Colors should not be nil")
	}
}

func TestNewColor(t *testing.T) {
	color := &Color{
		ColorSpace: ColorSpaceRGB,
		Values:     [4]uint16{255, 0, 0, 0}, // RGB + alpha
		Name:       "Test Red",
	}
	
	if color.ColorSpace != ColorSpaceRGB {
		t.Errorf("Color ColorSpace = %v, want %v", color.ColorSpace, ColorSpaceRGB)
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
	validColorSpaces := []ColorSpace{
		ColorSpaceRGB,
		ColorSpaceHSB,
		ColorSpaceCMYK,
		ColorSpaceLab,
		ColorSpaceGrayscale,
	}
	
	for _, cs := range validColorSpaces {
		if cs.String() == "" {
			t.Errorf("ColorSpace %d should have a string representation", cs)
		}
	}
}