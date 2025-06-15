package color

import (
	"fmt"
	"math"
)

// Color represents a color in any color space.
type Color interface {
	// String returns a string representation of the color.
	String() string

	// ColorSpace returns the name of the color space.
	ColorSpace() string

	// ToRGB converts the color to RGB color space.
	ToRGB() RGB

	// ToCMYK converts the color to CMYK color space.
	ToCMYK() CMYK

	// ToLAB converts the color to LAB color space.
	ToLAB() LAB

	// ToHSB converts the color to HSB color space.
	ToHSB() HSB
}

// RGB represents a color in RGB color space.
type RGB struct {
	R, G, B uint8
}

// NewRGB creates a new RGB color with validation.
func NewRGB(r, g, b uint8) RGB {
	return RGB{R: r, G: g, B: b}
}

// NewRGBFromFloat creates a new RGB color from float64 values (0.0-1.0).
func NewRGBFromFloat(r, g, b float64) RGB {
	return RGB{
		R: uint8(math.Round(clamp(r, 0, 1) * 255)),
		G: uint8(math.Round(clamp(g, 0, 1) * 255)),
		B: uint8(math.Round(clamp(b, 0, 1) * 255)),
	}
}

func (c RGB) String() string {
	return fmt.Sprintf("RGB(%d, %d, %d)", c.R, c.G, c.B)
}

func (c RGB) ColorSpace() string {
	return "RGB"
}

func (c RGB) ToRGB() RGB {
	return c
}

func (c RGB) ToCMYK() CMYK {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	k := 1.0 - math.Max(r, math.Max(g, b))
	if k == 1.0 {
		return CMYK{C: 0, M: 0, Y: 0, K: 100}
	}

	cy := (1.0 - r - k) / (1.0 - k)
	mg := (1.0 - g - k) / (1.0 - k)
	ye := (1.0 - b - k) / (1.0 - k)

	return CMYK{
		C: uint8(math.Round(cy * 100)),
		M: uint8(math.Round(mg * 100)),
		Y: uint8(math.Round(ye * 100)),
		K: uint8(math.Round(k * 100)),
	}
}

func (c RGB) ToLAB() LAB {
	// Convert RGB to XYZ first
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	// Apply gamma correction
	if r > 0.04045 {
		r = math.Pow((r+0.055)/1.055, 2.4)
	} else {
		r = r / 12.92
	}

	if g > 0.04045 {
		g = math.Pow((g+0.055)/1.055, 2.4)
	} else {
		g = g / 12.92
	}

	if b > 0.04045 {
		b = math.Pow((b+0.055)/1.055, 2.4)
	} else {
		b = b / 12.92
	}

	// Convert to XYZ using sRGB matrix
	x := r*0.4124564 + g*0.3575761 + b*0.1804375
	y := r*0.2126729 + g*0.7151522 + b*0.0721750
	z := r*0.0193339 + g*0.1191920 + b*0.9503041

	// Convert XYZ to LAB
	// D65 illuminant
	xn := 0.95047
	yn := 1.00000
	zn := 1.08883

	x = x / xn
	y = y / yn
	z = z / zn

	fx := labF(x)
	fy := labF(y)
	fz := labF(z)

	l := 116*fy - 16
	a := 500 * (fx - fy)
	b2 := 200 * (fy - fz)

	return LAB{
		L: int8(math.Round(clamp(l, 0, 100))),
		A: int8(math.Round(clamp(a, -128, 127))),
		B: int8(math.Round(clamp(b2, -128, 127))),
	}
}

func (c RGB) ToHSB() HSB {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	var h, s float64
	brightness := max

	if delta == 0 {
		h = 0
		s = 0
	} else {
		s = delta / max

		switch max {
		case r:
			h = 60 * (math.Mod((g-b)/delta, 6))
		case g:
			h = 60 * ((b-r)/delta + 2)
		case b:
			h = 60 * ((r-g)/delta + 4)
		}

		if h < 0 {
			h += 360
		}
	}

	return HSB{
		H: uint16(math.Round(h)),
		S: uint8(math.Round(s * 100)),
		B: uint8(math.Round(brightness * 100)),
	}
}

// CMYK represents a color in CMYK color space.
type CMYK struct {
	C, M, Y, K uint8 // 0-100
}

// NewCMYK creates a new CMYK color with validation.
func NewCMYK(c, m, y, k uint8) CMYK {
	return CMYK{
		C: clampUint8(c, 0, 100),
		M: clampUint8(m, 0, 100),
		Y: clampUint8(y, 0, 100),
		K: clampUint8(k, 0, 100),
	}
}

func (c CMYK) String() string {
	return fmt.Sprintf("CMYK(%d%%, %d%%, %d%%, %d%%)", c.C, c.M, c.Y, c.K)
}

func (c CMYK) ColorSpace() string {
	return "CMYK"
}

func (c CMYK) ToRGB() RGB {
	cy := float64(c.C) / 100.0
	mg := float64(c.M) / 100.0
	ye := float64(c.Y) / 100.0
	k := float64(c.K) / 100.0

	r := 255 * (1 - cy) * (1 - k)
	g := 255 * (1 - mg) * (1 - k)
	b := 255 * (1 - ye) * (1 - k)

	return RGB{
		R: uint8(math.Round(r)),
		G: uint8(math.Round(g)),
		B: uint8(math.Round(b)),
	}
}

func (c CMYK) ToCMYK() CMYK {
	return c
}

func (c CMYK) ToLAB() LAB {
	return c.ToRGB().ToLAB()
}

func (c CMYK) ToHSB() HSB {
	return c.ToRGB().ToHSB()
}

// LAB represents a color in LAB color space.
type LAB struct {
	L    int8 // 0-100
	A, B int8 // -128 to 127
}

// NewLAB creates a new LAB color with validation.
func NewLAB(l, a, b int8) LAB {
	return LAB{
		L: clampInt8(l, 0, 100),
		A: a,
		B: b,
	}
}

func (c LAB) String() string {
	return fmt.Sprintf("LAB(%d, %d, %d)", c.L, c.A, c.B)
}

func (c LAB) ColorSpace() string {
	return "LAB"
}

func (c LAB) ToRGB() RGB {
	// Convert LAB to XYZ first
	l := float64(c.L)
	a := float64(c.A)
	b := float64(c.B)

	fy := (l + 16) / 116
	fx := a/500 + fy
	fz := fy - b/200

	x := labFInv(fx)
	y := labFInv(fy)
	z := labFInv(fz)

	// D65 illuminant
	xn := 0.95047
	yn := 1.00000
	zn := 1.08883

	x = x * xn
	y = y * yn
	z = z * zn

	// Convert XYZ to RGB
	r := x*3.2404542 + y*-1.5371385 + z*-0.4985314
	g := x*-0.9692660 + y*1.8760108 + z*0.0415560
	b2 := x*0.0556434 + y*-0.2040259 + z*1.0572252

	// Apply gamma correction
	if r > 0.0031308 {
		r = 1.055*math.Pow(r, 1/2.4) - 0.055
	} else {
		r = 12.92 * r
	}

	if g > 0.0031308 {
		g = 1.055*math.Pow(g, 1/2.4) - 0.055
	} else {
		g = 12.92 * g
	}

	if b2 > 0.0031308 {
		b2 = 1.055*math.Pow(b2, 1/2.4) - 0.055
	} else {
		b2 = 12.92 * b2
	}

	return RGB{
		R: uint8(math.Round(clamp(r, 0, 1) * 255)),
		G: uint8(math.Round(clamp(g, 0, 1) * 255)),
		B: uint8(math.Round(clamp(b2, 0, 1) * 255)),
	}
}

func (c LAB) ToCMYK() CMYK {
	return c.ToRGB().ToCMYK()
}

func (c LAB) ToLAB() LAB {
	return c
}

func (c LAB) ToHSB() HSB {
	return c.ToRGB().ToHSB()
}

// HSB represents a color in HSB (HSV) color space.
type HSB struct {
	H uint16 // 0-359
	S uint8  // 0-100
	B uint8  // 0-100
}

// NewHSB creates a new HSB color with validation.
func NewHSB(h uint16, s, b uint8) HSB {
	return HSB{
		H: h % 360,
		S: clampUint8(s, 0, 100),
		B: clampUint8(b, 0, 100),
	}
}

func (c HSB) String() string {
	return fmt.Sprintf("HSB(%d°, %d%%, %d%%)", c.H, c.S, c.B)
}

func (c HSB) ColorSpace() string {
	return "HSB"
}

func (c HSB) ToRGB() RGB {
	h := float64(c.H)
	s := float64(c.S) / 100.0
	v := float64(c.B) / 100.0

	chroma := v * s
	x := chroma * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - chroma

	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = chroma, x, 0
	case h < 120:
		r, g, b = x, chroma, 0
	case h < 180:
		r, g, b = 0, chroma, x
	case h < 240:
		r, g, b = 0, x, chroma
	case h < 300:
		r, g, b = x, 0, chroma
	default:
		r, g, b = chroma, 0, x
	}

	return RGB{
		R: uint8(math.Round((r + m) * 255)),
		G: uint8(math.Round((g + m) * 255)),
		B: uint8(math.Round((b + m) * 255)),
	}
}

func (c HSB) ToCMYK() CMYK {
	return c.ToRGB().ToCMYK()
}

func (c HSB) ToLAB() LAB {
	return c.ToRGB().ToLAB()
}

func (c HSB) ToHSB() HSB {
	return c
}

// Helper functions

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampUint8(value, min, max uint8) uint8 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampInt8(value, min, max int8) int8 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func labF(t float64) float64 {
	if t > 0.008856 {
		return math.Pow(t, 1.0/3.0)
	}
	return 7.787*t + 16.0/116.0
}

func labFInv(t float64) float64 {
	if t > 0.206893 {
		return math.Pow(t, 3)
	}
	return (t - 16.0/116.0) / 7.787
}
