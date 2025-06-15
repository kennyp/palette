package colorbook

import (
	"testing"
)

func TestBookIDString(t *testing.T) {
	tests := map[BookID]string{
		BookIDANPA:                  "ANPA",
		BookIDFocoltone:             "Focoltone",
		BookIDPantoneCoated:         "PantoneCoated",
		BookIDPantoneProcess:        "PantoneProcess",
		BookIDPantoneProSlim:        "PantoneProSlim",
		BookIDPantoneUncoated:       "PantoneUncoated",
		BookIDToyo:                  "Toyo",
		BookIDTrumatch:              "Trumatch",
		BookIDHKSE:                  "HKSE",
		BookIDHKSK:                  "HKSK",
		BookIDHKSN:                  "HKSN",
		BookIDHKSZ:                  "HKSZ",
		BookIDDIC:                   "DIC",
		BookIDPantonePastelCoated:   "PantonePastelCoated",
		BookIDPantonePastelUncoated: "PantonePastelUncoated",
		BookIDPantoneMetallic:       "PantoneMetallic",
	}

	for bookID, expected := range tests {
		if got := bookID.String(); got != expected {
			t.Errorf("BookID(%d).String() = %s, want %s", bookID, got, expected)
		}
	}
}

func TestColorTypeString(t *testing.T) {
	tests := map[ColorType]string{
		ColorTypeRGB:  "RGB",
		ColorTypeCMYK: "CMYK",
		ColorTypeLab:  "Lab",
	}

	for colorType, expected := range tests {
		if got := colorType.String(); got != expected {
			t.Errorf("ColorType(%d).String() = %s, want %s", colorType, got, expected)
		}
	}
}

func TestConstants(t *testing.T) {
	if FileType != "8BCB" {
		t.Errorf("FileType = %s, want 8BCB", FileType)
	}
	
	if DefaultVersion != 1 {
		t.Errorf("DefaultVersion = %d, want 1", DefaultVersion)
	}
}

func TestNewColorBook(t *testing.T) {
	cb := &ColorBook{
		ID:            BookIDPantoneCoated,
		Version:       DefaultVersion,
		ColorType:     ColorTypeRGB,
		Title:         "Test Color Book",
		Prefix:        "TEST",
		Postfix:       "",
		Description:   "A test color book",
		ColorsPerPage: 10,
		KeyColorPage:  1,
		Colors:        []*Color{},
	}
	
	if cb.Version != DefaultVersion {
		t.Errorf("ColorBook Version = %d, want %d", cb.Version, DefaultVersion)
	}
	
	if cb.ID != BookIDPantoneCoated {
		t.Errorf("ColorBook ID = %v, want %v", cb.ID, BookIDPantoneCoated)
	}
	
	if cb.ColorType != ColorTypeRGB {
		t.Errorf("ColorBook ColorType = %v, want %v", cb.ColorType, ColorTypeRGB)
	}
}

func TestNewColor(t *testing.T) {
	color := &Color{
		Name:       "Test Red",
		Key:        [6]byte{'T', 'R', 0, 0, 0, 0},
		Components: [3]byte{255, 0, 0},
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