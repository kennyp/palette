package palette

import (
	"reflect"
	"testing"

	"github.com/kennyp/palette/color"
)

func TestNew(t *testing.T) {
	name := "Test Palette"
	p := New(name)
	
	if p.Name != name {
		t.Errorf("New() name = %v, want %v", p.Name, name)
	}
	
	if p.Len() != 0 {
		t.Errorf("New() should create empty palette, got length %d", p.Len())
	}
	
	if !p.IsEmpty() {
		t.Errorf("New() should create empty palette")
	}
}

func TestNewWithColors(t *testing.T) {
	colors := []NamedColor{
		{Name: "Red", Color: color.NewRGB(255, 0, 0)},
		{Name: "Green", Color: color.NewRGB(0, 255, 0)},
	}
	
	p := NewWithColors("Test", colors...)
	
	if p.Name != "Test" {
		t.Errorf("NewWithColors() name = %v, want Test", p.Name)
	}
	
	if p.Len() != 2 {
		t.Errorf("NewWithColors() length = %d, want 2", p.Len())
	}
	
	if p.IsEmpty() {
		t.Errorf("NewWithColors() should not be empty")
	}
}

func TestAdd(t *testing.T) {
	p := New("Test")
	red := color.NewRGB(255, 0, 0)
	
	p.Add(red, "Red")
	
	if p.Len() != 1 {
		t.Errorf("Add() length = %d, want 1", p.Len())
	}
	
	got, err := p.Get(0)
	if err != nil {
		t.Errorf("Add() failed to retrieve color: %v", err)
	}
	
	if got.Name != "Red" {
		t.Errorf("Add() name = %v, want Red", got.Name)
	}
	
	if got.Color.ToRGB() != red {
		t.Errorf("Add() color = %v, want %v", got.Color, red)
	}
}

func TestAddColor(t *testing.T) {
	p := New("Test")
	red := color.NewRGB(255, 0, 0)
	
	p.AddColor(red)
	
	if p.Len() != 1 {
		t.Errorf("AddColor() length = %d, want 1", p.Len())
	}
	
	got, err := p.Get(0)
	if err != nil {
		t.Errorf("AddColor() failed to retrieve color: %v", err)
	}
	
	if got.Name != "" {
		t.Errorf("AddColor() should have empty name, got %v", got.Name)
	}
}

func TestRemove(t *testing.T) {
	p := New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	err := p.Remove(0)
	if err != nil {
		t.Errorf("Remove() error = %v", err)
	}
	
	if p.Len() != 1 {
		t.Errorf("Remove() length = %d, want 1", p.Len())
	}
	
	got, _ := p.Get(0)
	if got.Name != "Green" {
		t.Errorf("Remove() remaining color = %v, want Green", got.Name)
	}
	
	// Test out of bounds
	err = p.Remove(5)
	if err == nil {
		t.Errorf("Remove() should error on out of bounds index")
	}
}

func TestRemoveByName(t *testing.T) {
	p := New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	removed := p.RemoveByName("Red")
	if !removed {
		t.Errorf("RemoveByName() should return true when color found")
	}
	
	if p.Len() != 1 {
		t.Errorf("RemoveByName() length = %d, want 1", p.Len())
	}
	
	removed = p.RemoveByName("Blue")
	if removed {
		t.Errorf("RemoveByName() should return false when color not found")
	}
}

func TestGet(t *testing.T) {
	p := New("Test")
	red := color.NewRGB(255, 0, 0)
	p.Add(red, "Red")
	
	got, err := p.Get(0)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	
	if got.Name != "Red" {
		t.Errorf("Get() name = %v, want Red", got.Name)
	}
	
	// Test out of bounds
	_, err = p.Get(5)
	if err == nil {
		t.Errorf("Get() should error on out of bounds index")
	}
}

func TestGetByName(t *testing.T) {
	p := New("Test")
	red := color.NewRGB(255, 0, 0)
	p.Add(red, "Red")
	
	got, found := p.GetByName("Red")
	if !found {
		t.Errorf("GetByName() should find Red")
	}
	
	if got.Name != "Red" {
		t.Errorf("GetByName() name = %v, want Red", got.Name)
	}
	
	_, found = p.GetByName("Blue")
	if found {
		t.Errorf("GetByName() should not find Blue")
	}
}

func TestClear(t *testing.T) {
	p := New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	p.Clear()
	
	if !p.IsEmpty() {
		t.Errorf("Clear() should make palette empty")
	}
	
	if p.Len() != 0 {
		t.Errorf("Clear() length = %d, want 0", p.Len())
	}
}

func TestClone(t *testing.T) {
	p := New("Original")
	p.Description = "Original description"
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.SetMetadata("key", "value")
	
	clone := p.Clone()
	
	if clone.Name != p.Name {
		t.Errorf("Clone() name = %v, want %v", clone.Name, p.Name)
	}
	
	if clone.Description != p.Description {
		t.Errorf("Clone() description = %v, want %v", clone.Description, p.Description)
	}
	
	if clone.Len() != p.Len() {
		t.Errorf("Clone() length = %d, want %d", clone.Len(), p.Len())
	}
	
	// Test that metadata is copied
	if value, ok := clone.GetMetadata("key"); !ok || value != "value" {
		t.Errorf("Clone() should copy metadata")
	}
	
	// Test that modifying clone doesn't affect original
	clone.Add(color.NewRGB(0, 255, 0), "Green")
	if p.Len() == clone.Len() {
		t.Errorf("Clone() should be independent of original")
	}
}

func TestFilter(t *testing.T) {
	p := New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewCMYK(100, 0, 100, 0), "Green")
	p.Add(color.NewRGB(0, 0, 255), "Blue")
	
	filtered := p.Filter(func(c NamedColor) bool {
		return c.Color.ColorSpace() == "RGB"
	})
	
	if filtered.Len() != 2 {
		t.Errorf("Filter() length = %d, want 2", filtered.Len())
	}
	
	// Original should be unchanged
	if p.Len() != 3 {
		t.Errorf("Filter() should not modify original palette")
	}
}

func TestFilterByColorSpace(t *testing.T) {
	p := New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewCMYK(100, 0, 100, 0), "Green")
	p.Add(color.NewRGB(0, 0, 255), "Blue")
	
	rgbOnly := p.FilterByColorSpace("RGB")
	
	if rgbOnly.Len() != 2 {
		t.Errorf("FilterByColorSpace() length = %d, want 2", rgbOnly.Len())
	}
	
	cmykOnly := p.FilterByColorSpace("CMYK")
	
	if cmykOnly.Len() != 1 {
		t.Errorf("FilterByColorSpace() length = %d, want 1", cmykOnly.Len())
	}
}

func TestMap(t *testing.T) {
	p := New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	mapped := p.Map(func(c NamedColor) NamedColor {
		return NamedColor{
			Name: "Mapped " + c.Name,
			Color: c.Color,
		}
	})
	
	if mapped.Len() != p.Len() {
		t.Errorf("Map() length = %d, want %d", mapped.Len(), p.Len())
	}
	
	got, _ := mapped.Get(0)
	if got.Name != "Mapped Red" {
		t.Errorf("Map() name = %v, want Mapped Red", got.Name)
	}
	
	// Original should be unchanged
	orig, _ := p.Get(0)
	if orig.Name != "Red" {
		t.Errorf("Map() should not modify original palette")
	}
}

func TestConvertToColorSpace(t *testing.T) {
	p := New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewHSB(120, 100, 100), "Green")
	
	cmykPalette, err := p.ConvertToColorSpace("CMYK")
	if err != nil {
		t.Errorf("ConvertToColorSpace() error = %v", err)
	}
	
	if cmykPalette.Len() != p.Len() {
		t.Errorf("ConvertToColorSpace() length = %d, want %d", cmykPalette.Len(), p.Len())
	}
	
	// Check that all colors are now CMYK
	for i := 0; i < cmykPalette.Len(); i++ {
		c, _ := cmykPalette.Get(i)
		if c.Color.ColorSpace() != "CMYK" {
			t.Errorf("ConvertToColorSpace() color %d is %v, want CMYK", i, c.Color.ColorSpace())
		}
	}
	
	// Test unknown color space
	unknown, err := p.ConvertToColorSpace("XYZ")
	if err != nil {
		t.Errorf("ConvertToColorSpace() should handle unknown color spaces")
	}
	
	// Should keep original colors for unknown color space
	orig, _ := p.Get(0)
	conv, _ := unknown.Get(0)
	if orig.Color.ColorSpace() != conv.Color.ColorSpace() {
		t.Errorf("ConvertToColorSpace() should preserve colors for unknown color space")
	}
}

func TestString(t *testing.T) {
	p := New("Test Palette")
	p.Description = "A test palette"
	p.Add(color.NewRGB(255, 0, 0), "Red")
	
	str := p.String()
	
	if !contains(str, "Test Palette") {
		t.Errorf("String() should contain palette name")
	}
	
	if !contains(str, "A test palette") {
		t.Errorf("String() should contain description")
	}
	
	if !contains(str, "Red") {
		t.Errorf("String() should contain color names")
	}
	
	if !contains(str, "1 colors") {
		t.Errorf("String() should contain color count")
	}
}

func TestMetadata(t *testing.T) {
	p := New("Test")
	
	// Test setting and getting metadata
	p.SetMetadata("format", "test")
	p.SetMetadata("version", 42)
	
	if value, ok := p.GetMetadata("format"); !ok || value != "test" {
		t.Errorf("SetMetadata/GetMetadata failed for string value")
	}
	
	if value, ok := p.GetMetadata("version"); !ok || value != 42 {
		t.Errorf("SetMetadata/GetMetadata failed for int value")
	}
	
	if _, ok := p.GetMetadata("nonexistent"); ok {
		t.Errorf("GetMetadata should return false for nonexistent key")
	}
	
	// Test listing metadata keys
	keys := p.ListMetadataKeys()
	if len(keys) != 2 {
		t.Errorf("ListMetadataKeys() length = %d, want 2", len(keys))
	}
	
	// Keys should be sorted
	if !reflect.DeepEqual(keys, []string{"format", "version"}) {
		t.Errorf("ListMetadataKeys() = %v, want [format version]", keys)
	}
	
	// Test removing metadata
	p.RemoveMetadata("format")
	if _, ok := p.GetMetadata("format"); ok {
		t.Errorf("RemoveMetadata should remove the key")
	}
	
	keys = p.ListMetadataKeys()
	if len(keys) != 1 {
		t.Errorf("ListMetadataKeys() after removal length = %d, want 1", len(keys))
	}
}

func TestValidate(t *testing.T) {
	// Test valid palette
	p := New("Test")
	p.Add(color.NewRGB(255, 0, 0), "Red")
	p.Add(color.NewRGB(0, 255, 0), "Green")
	
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v for valid palette", err)
	}
	
	// Test empty name
	p2 := New("")
	if err := p2.Validate(); err == nil {
		t.Errorf("Validate() should error for empty name")
	}
	
	// Test duplicate names
	p3 := New("Test")
	p3.Add(color.NewRGB(255, 0, 0), "Red")
	p3.Add(color.NewRGB(0, 255, 0), "Red") // Duplicate name
	
	if err := p3.Validate(); err == nil {
		t.Errorf("Validate() should error for duplicate names")
	}
	
	// Test that empty names don't cause duplicate errors
	p4 := New("Test")
	p4.AddColor(color.NewRGB(255, 0, 0)) // Empty name
	p4.AddColor(color.NewRGB(0, 255, 0)) // Empty name
	
	if err := p4.Validate(); err != nil {
		t.Errorf("Validate() should not error for multiple empty names: %v", err)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkAdd(b *testing.B) {
	p := New("Benchmark")
	red := color.NewRGB(255, 0, 0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Add(red, "Red")
		p.Clear()
	}
}

func BenchmarkGet(b *testing.B) {
	p := New("Benchmark")
	for i := 0; i < 100; i++ {
		p.Add(color.NewRGB(uint8(i), 0, 0), "Color")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Get(i % 100)
	}
}

func BenchmarkClone(b *testing.B) {
	p := New("Benchmark")
	for i := 0; i < 100; i++ {
		p.Add(color.NewRGB(uint8(i), 0, 0), "Color")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Clone()
	}
}

func BenchmarkConvertToColorSpace(b *testing.B) {
	p := New("Benchmark")
	for i := 0; i < 100; i++ {
		p.Add(color.NewRGB(uint8(i), 0, 0), "Color")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.ConvertToColorSpace("CMYK")
	}
}