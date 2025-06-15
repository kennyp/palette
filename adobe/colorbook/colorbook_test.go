package colorbook_test

import (
	"testing"
	
	"github.com/kennyp/palette/adobe/colorbook"
)

func TestBookIDString(t *testing.T) {
	tests := map[colorbook.BookID]string{
		colorbook.BookIDANPA:                  "ANPA",
		colorbook.BookIDFocoltone:             "Focoltone",
		colorbook.BookIDPantoneCoated:         "PantoneCoated",
		colorbook.BookIDPantoneProcess:        "PantoneProcess",
		colorbook.BookIDPantoneProSlim:        "PantoneProSlim",
		colorbook.BookIDPantoneUncoated:       "PantoneUncoated",
		colorbook.BookIDToyo:                  "Toyo",
		colorbook.BookIDTrumatch:              "Trumatch",
		colorbook.BookIDHKSE:                  "HKSE",
		colorbook.BookIDHKSK:                  "HKSK",
		colorbook.BookIDHKSN:                  "HKSN",
		colorbook.BookIDHKSZ:                  "HKSZ",
		colorbook.BookIDDIC:                   "DIC",
		colorbook.BookIDPantonePastelCoated:   "PantonePastelCoated",
		colorbook.BookIDPantonePastelUncoated: "PantonePastelUncoated",
		colorbook.BookIDPantoneMetallic:       "PantoneMetallic",
	}

	for bookID, expected := range tests {
		if got := bookID.String(); got != expected {
			t.Errorf("BookID(%d).String() = %s, want %s", bookID, got, expected)
		}
	}
}

func TestColorTypeString(t *testing.T) {
	tests := map[colorbook.ColorType]string{
		colorbook.ColorTypeRGB:  "RGB",
		colorbook.ColorTypeCMYK: "CMYK",
		colorbook.ColorTypeLab:  "Lab",
	}

	for colorType, expected := range tests {
		if got := colorType.String(); got != expected {
			t.Errorf("ColorType(%d).String() = %s, want %s", colorType, got, expected)
		}
	}
}

func TestConstants(t *testing.T) {
	if colorbook.FileType != "8BCB" {
		t.Errorf("FileType = %s, want 8BCB", colorbook.FileType)
	}
	
	if colorbook.DefaultVersion != 1 {
		t.Errorf("DefaultVersion = %d, want 1", colorbook.DefaultVersion)
	}
}

func TestNewColorBook(t *testing.T) {
	cb := &colorbook.ColorBook{
		ID:            colorbook.BookIDPantoneCoated,
		Version:       colorbook.DefaultVersion,
		ColorType:     colorbook.ColorTypeRGB,
		Title:         "Test Color Book",
		Prefix:        "TEST",
		Postfix:       "",
		Description:   "A test color book",
		ColorsPerPage: 10,
		KeyColorPage:  1,
		Colors:        []*colorbook.Color{},
	}
	
	if cb.Version != colorbook.DefaultVersion {
		t.Errorf("ColorBook Version = %d, want %d", cb.Version, colorbook.DefaultVersion)
	}
	
	if cb.ID != colorbook.BookIDPantoneCoated {
		t.Errorf("ColorBook ID = %v, want %v", cb.ID, colorbook.BookIDPantoneCoated)
	}
	
	if cb.ColorType != colorbook.ColorTypeRGB {
		t.Errorf("ColorBook ColorType = %v, want %v", cb.ColorType, colorbook.ColorTypeRGB)
	}
}

func TestNewColor(t *testing.T) {
	color := &colorbook.Color{
		Name:       "Test Red",
		Key:        [6]byte{'T', 'R', 0, 0, 0, 0},
		Components: [4]byte{255, 0, 0, 0},
	}
	
	if color.Name != "Test Red" {
		t.Errorf("Color Name = %s, want Test Red", color.Name)
	}
	
	if color.Key[0] != 'T' || color.Key[1] != 'R' {
		t.Errorf("Color Key = %v, want ['T' 'R' 0 0 0 0]", color.Key)
	}
	
	if color.Components[0] != 255 || color.Components[1] != 0 || color.Components[2] != 0 {
		t.Errorf("Color Components = %v, want [255 0 0]", color.Components)
	}
}