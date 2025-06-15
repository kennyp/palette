package types_test

import (
	"testing"
	
	"github.com/kennyp/palette/adobe/types"
)

func TestUnicodeString(t *testing.T) {
	tests := []string{
		"",
		"Hello",
		"Test Color",
		"PANTONE 186 C",
		"Color with spaces",
		"Unicode: ñáéíóú",
	}

	for _, original := range tests {
		t.Run(original, func(t *testing.T) {
			us := types.UnicodeString(original)
			
			// Test String() method
			if us.String() != original {
				t.Errorf("UnicodeString.String() = %s, want %s", us.String(), original)
			}
			
			// Test MarshalBinary
			data, err := us.MarshalBinary()
			if err != nil {
				t.Errorf("UnicodeString.MarshalBinary() failed: %v", err)
			}
			
			// For non-empty strings, we should get some data
			if original != "" && len(data) == 0 {
				t.Error("MarshalBinary should produce data for non-empty string")
			}
			
			// Data length should be even (UTF-16 uses 2 bytes per code unit)
			if len(data)%2 != 0 {
				t.Errorf("MarshalBinary data length should be even, got %d", len(data))
			}
		})
	}
}

func TestUnicodeStringBasicFunctionality(t *testing.T) {
	// Test basic functionality
	us := types.UnicodeString("Test")
	
	if us.String() != "Test" {
		t.Errorf("String() = %s, want Test", us.String())
	}
	
	data, err := us.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() failed: %v", err)
	}
	
	// "Test" should be 4 UTF-16 code units = 8 bytes
	expectedLength := 8
	if len(data) != expectedLength {
		t.Errorf("MarshalBinary() data length = %d, want %d", len(data), expectedLength)
	}
}