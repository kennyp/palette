package color

import (
	"testing"
)

func TestRGB(t *testing.T) {
	tests := map[string]struct {
		r, g, b uint8
		want string
	}{
		"Red":   {255, 0, 0, "RGB(255, 0, 0)"},
		"Green": {0, 255, 0, "RGB(0, 255, 0)"},
		"Blue":  {0, 0, 255, "RGB(0, 0, 255)"},
		"White": {255, 255, 255, "RGB(255, 255, 255)"},
		"Black": {0, 0, 0, "RGB(0, 0, 0)"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rgb := NewRGB(tt.r, tt.g, tt.b)
			if got := rgb.String(); got != tt.want {
				t.Errorf("RGB.String() = %v, want %v", got, tt.want)
			}
			if got := rgb.ColorSpace(); got != "RGB" {
				t.Errorf("RGB.ColorSpace() = %v, want RGB", got)
			}
			if got := rgb.ToRGB(); got != rgb {
				t.Errorf("RGB.ToRGB() = %v, want %v", got, rgb)
			}
		})
	}
}

func TestNewRGBFromFloat(t *testing.T) {
	tests := map[string]struct {
		r, g, b float64
		want RGB
	}{
		"Full intensity": {1.0, 0.0, 0.0, RGB{255, 0, 0}},
		"Half intensity": {0.5, 0.5, 0.5, RGB{128, 128, 128}},
		"Zero intensity": {0.0, 0.0, 0.0, RGB{0, 0, 0}},
		"Over range":     {1.5, -0.5, 0.0, RGB{255, 0, 0}}, // Should clamp
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewRGBFromFloat(tt.r, tt.g, tt.b)
			if got != tt.want {
				t.Errorf("NewRGBFromFloat(%v, %v, %v) = %v, want %v", tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}

func TestCMYK(t *testing.T) {
	tests := map[string]struct {
		c, m, y, k uint8
		want string
	}{
		"Cyan":    {100, 0, 0, 0, "CMYK(100%, 0%, 0%, 0%)"},
		"Magenta": {0, 100, 0, 0, "CMYK(0%, 100%, 0%, 0%)"},
		"Yellow":  {0, 0, 100, 0, "CMYK(0%, 0%, 100%, 0%)"},
		"Black":   {0, 0, 0, 100, "CMYK(0%, 0%, 0%, 100%)"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmyk := NewCMYK(tt.c, tt.m, tt.y, tt.k)
			if got := cmyk.String(); got != tt.want {
				t.Errorf("CMYK.String() = %v, want %v", got, tt.want)
			}
			if got := cmyk.ColorSpace(); got != "CMYK" {
				t.Errorf("CMYK.ColorSpace() = %v, want CMYK", got)
			}
			if got := cmyk.ToCMYK(); got != cmyk {
				t.Errorf("CMYK.ToCMYK() = %v, want %v", got, cmyk)
			}
		})
	}
}

func TestHSB(t *testing.T) {
	tests := map[string]struct {
		h uint16
		s, b uint8
		want string
	}{
		"Red":   {0, 100, 100, "HSB(0°, 100%, 100%)"},
		"Green": {120, 100, 100, "HSB(120°, 100%, 100%)"},
		"Blue":  {240, 100, 100, "HSB(240°, 100%, 100%)"},
		"White": {0, 0, 100, "HSB(0°, 0%, 100%)"},
		"Black": {0, 0, 0, "HSB(0°, 0%, 0%)"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hsb := NewHSB(tt.h, tt.s, tt.b)
			if got := hsb.String(); got != tt.want {
				t.Errorf("HSB.String() = %v, want %v", got, tt.want)
			}
			if got := hsb.ColorSpace(); got != "HSB" {
				t.Errorf("HSB.ColorSpace() = %v, want HSB", got)
			}
			if got := hsb.ToHSB(); got != hsb {
				t.Errorf("HSB.ToHSB() = %v, want %v", got, hsb)
			}
		})
	}
}

func TestLAB(t *testing.T) {
	tests := map[string]struct {
		l, a, b int8
		want string
	}{
		"White":    {100, 0, 0, "LAB(100, 0, 0)"},
		"Black":    {0, 0, 0, "LAB(0, 0, 0)"},
		"Mid gray": {50, 0, 0, "LAB(50, 0, 0)"},
		"Red-ish":  {50, 50, 25, "LAB(50, 50, 25)"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			lab := NewLAB(tt.l, tt.a, tt.b)
			if got := lab.String(); got != tt.want {
				t.Errorf("LAB.String() = %v, want %v", got, tt.want)
			}
			if got := lab.ColorSpace(); got != "LAB" {
				t.Errorf("LAB.ColorSpace() = %v, want LAB", got)
			}
			if got := lab.ToLAB(); got != lab {
				t.Errorf("LAB.ToLAB() = %v, want %v", got, lab)
			}
		})
	}
}

func TestRGBToCMYK(t *testing.T) {
	tests := map[string]struct {
		rgb RGB
		want CMYK
	}{
		"White": {RGB{255, 255, 255}, CMYK{0, 0, 0, 0}},
		"Black": {RGB{0, 0, 0}, CMYK{0, 0, 0, 100}},
		"Red":   {RGB{255, 0, 0}, CMYK{0, 100, 100, 0}},
		"Green": {RGB{0, 255, 0}, CMYK{100, 0, 100, 0}},
		"Blue":  {RGB{0, 0, 255}, CMYK{100, 100, 0, 0}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.rgb.ToCMYK()
			if got != tt.want {
				t.Errorf("RGB.ToCMYK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCMYKToRGB(t *testing.T) {
	tests := map[string]struct {
		cmyk CMYK
		want RGB
	}{
		"White": {CMYK{0, 0, 0, 0}, RGB{255, 255, 255}},
		"Black": {CMYK{0, 0, 0, 100}, RGB{0, 0, 0}},
		"Red":   {CMYK{0, 100, 100, 0}, RGB{255, 0, 0}},
		"Green": {CMYK{100, 0, 100, 0}, RGB{0, 255, 0}},
		"Blue":  {CMYK{100, 100, 0, 0}, RGB{0, 0, 255}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.cmyk.ToRGB()
			if got != tt.want {
				t.Errorf("CMYK.ToRGB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRGBToHSB(t *testing.T) {
	tests := map[string]struct {
		rgb RGB
		want HSB
	}{
		"Red":     {RGB{255, 0, 0}, HSB{0, 100, 100}},
		"Green":   {RGB{0, 255, 0}, HSB{120, 100, 100}},
		"Blue":    {RGB{0, 0, 255}, HSB{240, 100, 100}},
		"White":   {RGB{255, 255, 255}, HSB{0, 0, 100}},
		"Black":   {RGB{0, 0, 0}, HSB{0, 0, 0}},
		"Yellow":  {RGB{255, 255, 0}, HSB{60, 100, 100}},
		"Cyan":    {RGB{0, 255, 255}, HSB{180, 100, 100}},
		"Magenta": {RGB{255, 0, 255}, HSB{300, 100, 100}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.rgb.ToHSB()
			if got != tt.want {
				t.Errorf("RGB.ToHSB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHSBToRGB(t *testing.T) {
	tests := map[string]struct {
		hsb HSB
		want RGB
	}{
		"Red":     {HSB{0, 100, 100}, RGB{255, 0, 0}},
		"Green":   {HSB{120, 100, 100}, RGB{0, 255, 0}},
		"Blue":    {HSB{240, 100, 100}, RGB{0, 0, 255}},
		"White":   {HSB{0, 0, 100}, RGB{255, 255, 255}},
		"Black":   {HSB{0, 0, 0}, RGB{0, 0, 0}},
		"Yellow":  {HSB{60, 100, 100}, RGB{255, 255, 0}},
		"Cyan":    {HSB{180, 100, 100}, RGB{0, 255, 255}},
		"Magenta": {HSB{300, 100, 100}, RGB{255, 0, 255}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.hsb.ToRGB()
			if got != tt.want {
				t.Errorf("HSB.ToRGB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorConversionRoundTrip(t *testing.T) {
	// Test that RGB -> CMYK -> RGB preserves color reasonably well
	original := RGB{128, 64, 192}
	cmyk := original.ToCMYK()
	converted := cmyk.ToRGB()
	
	// Allow for some rounding error
	tolerance := uint8(2)
	if abs(int(original.R) - int(converted.R)) > int(tolerance) ||
		abs(int(original.G) - int(converted.G)) > int(tolerance) ||
		abs(int(original.B) - int(converted.B)) > int(tolerance) {
		t.Errorf("RGB->CMYK->RGB conversion lost precision: %v -> %v -> %v", original, cmyk, converted)
	}
}

func TestColorConversionHSB(t *testing.T) {
	// Test that RGB -> HSB -> RGB preserves color
	original := RGB{128, 64, 192}
	hsb := original.ToHSB()
	converted := hsb.ToRGB()
	
	// Allow for some rounding error
	tolerance := uint8(2)
	if abs(int(original.R) - int(converted.R)) > int(tolerance) ||
		abs(int(original.G) - int(converted.G)) > int(tolerance) ||
		abs(int(original.B) - int(converted.B)) > int(tolerance) {
		t.Errorf("RGB->HSB->RGB conversion lost precision: %v -> %v -> %v", original, hsb, converted)
	}
}

func TestLABConversion(t *testing.T) {
	// Test LAB conversion with known values
	white := RGB{255, 255, 255}
	lab := white.ToLAB()
	
	// White should have high L value and near-zero a,b
	if lab.L < 90 {
		t.Errorf("White RGB should convert to high L value, got L=%d", lab.L)
	}
	if abs(int(lab.A)) > 5 || abs(int(lab.B)) > 5 {
		t.Errorf("White RGB should convert to near-zero a,b values, got a=%d, b=%d", lab.A, lab.B)
	}
	
	// Test round trip
	converted := lab.ToRGB()
	tolerance := uint8(5) // LAB conversion has more precision loss
	if abs(int(white.R) - int(converted.R)) > int(tolerance) ||
		abs(int(white.G) - int(converted.G)) > int(tolerance) ||
		abs(int(white.B) - int(converted.B)) > int(tolerance) {
		t.Errorf("RGB->LAB->RGB conversion lost too much precision: %v -> %v -> %v", white, lab, converted)
	}
}

func TestClampingFunctions(t *testing.T) {
	tests := map[string]struct {
		value, min, max, want float64
	}{
		"Within range": {0.5, 0.0, 1.0, 0.5},
		"Below minimum": {-0.5, 0.0, 1.0, 0.0},
		"Above maximum": {1.5, 0.0, 1.0, 1.0},
		"At minimum": {0.0, 0.0, 1.0, 0.0},
		"At maximum": {1.0, 0.0, 1.0, 1.0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clamp(%v, %v, %v) = %v, want %v", tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestNewCMYKValidation(t *testing.T) {
	// Test that values are clamped to valid ranges
	cmyk := NewCMYK(150, 200, 50, 75) // Over 100%
	want := CMYK{100, 100, 50, 75}   // Should be clamped
	if cmyk != want {
		t.Errorf("NewCMYK should clamp values, got %v, want %v", cmyk, want)
	}
}

func TestNewHSBValidation(t *testing.T) {
	// Test that hue wraps around
	hsb := NewHSB(400, 150, 150) // Over ranges
	want := HSB{40, 100, 100}    // Should wrap/clamp
	if hsb != want {
		t.Errorf("NewHSB should wrap/clamp values, got %v, want %v", hsb, want)
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests
func BenchmarkRGBToCMYK(b *testing.B) {
	rgb := RGB{128, 64, 192}
	for b.Loop() {
		_ = rgb.ToCMYK()
	}
}

func BenchmarkRGBToHSB(b *testing.B) {
	rgb := RGB{128, 64, 192}
	for b.Loop() {
		_ = rgb.ToHSB()
	}
}

func BenchmarkRGBToLAB(b *testing.B) {
	rgb := RGB{128, 64, 192}
	for b.Loop() {
		_ = rgb.ToLAB()
	}
}

func BenchmarkLABToRGB(b *testing.B) {
	lab := LAB{50, 20, -30}
	for b.Loop() {
		_ = lab.ToRGB()
	}
}