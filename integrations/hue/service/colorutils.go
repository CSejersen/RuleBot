package service

import "math"

// Convert from Hue XY color space to approximate RGB (normalized 0–1)
func XYToRGB(x, y, brightness float64) (r, g, b float64) {
	if y == 0 {
		return 0, 0, 0
	}

	z := 1.0 - x - y
	Y := brightness
	X := (Y / y) * x
	Z := (Y / y) * z

	// Convert to RGB using Wide Gamut D65 conversion
	r = X*1.612 - Y*0.203 - Z*0.302
	g = -X*0.509 + Y*1.412 + Z*0.066
	b = X*0.026 - Y*0.072 + Z*0.962

	// Clamp and gamma correct
	r, g, b = gammaCorrect(r), gammaCorrect(g), gammaCorrect(b)
	return clamp01(r), clamp01(g), clamp01(b)
}

// Convert approximate RGB (0–1) to XY
func RGBToXY(r, g, b float64) (x, y float64) {
	// Inverse gamma correction
	r, g, b = inverseGamma(r), inverseGamma(g), inverseGamma(b)

	X := r*0.664511 + g*0.154324 + b*0.162028
	Y := r*0.283881 + g*0.668433 + b*0.047685
	Z := r*0.000088 + g*0.072310 + b*0.986039

	if X+Y+Z == 0 {
		return 0, 0
	}
	x = X / (X + Y + Z)
	y = Y / (X + Y + Z)
	return
}

// RGB → HSV (all in 0–1, hue returned in degrees)
func RGBToHSV(r, g, b float64) (h, s, v float64) {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	v = max
	if max != 0 {
		s = delta / max
	} else {
		s, h = 0, 0
		return
	}

	switch {
	case delta == 0:
		h = 0
	case max == r:
		h = 60 * math.Mod(((g-b)/delta), 6)
	case max == g:
		h = 60 * (((b - r) / delta) + 2)
	case max == b:
		h = 60 * (((r - g) / delta) + 4)
	}

	if h < 0 {
		h += 360
	}
	return
}

// HSV → RGB (hue in degrees)
func HSVToRGB(h, s, v float64) (r, g, b float64) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return r + m, g + m, b + m
}

// --- helpers ---
func gammaCorrect(v float64) float64 {
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

func inverseGamma(v float64) float64 {
	if v > 0.04045 {
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return v / 12.92
}

func clamp01(v float64) float64 {
	return math.Min(1.0, math.Max(0.0, v))
}
