package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Video settings
const (
	Width  = 1280
	Height = 720
	FPS    = 30
)

// Audio settings
const (
	SampleRate = 44100
	FFTSize    = 2048
)

// Visualization settings
const (
	NumBars      = 64   // Number of bars
	BarWidth     = 12   // Width of each bar
	BarGap       = 8    // Gap between bars
	CenterGap    = 100  // Gap between top and bottom bar sections
	MaxBarHeight = 0.50 // Maximum bar height as fraction of available space

	// Logarithmic frequency binning constants
	MinFreqHz = 40.0    // Lower bound of human hearing
	MaxFreqHz = 18000.0 // Upper bound before Nyquist rolloff

	// Spike prevention constant
	MaxAudioJump = 2.2 // Maximum allowed jump between frames to prevent visual artifacts

	// Velocity-based peak physics constants (vibeviz-style)
	SmoothFactor      = 0.4   // EMA smoothing factor for bar heights
	SmoothPeak        = 0.028 // Peak smoothing factor
	PeakThreshold     = 0.008 // Threshold for detecting new peaks
	SettleThreshold   = 0.012 // Threshold for settling state
	PeakVelocityBase  = 0.09  // Base velocity for new peaks
	PeakVelocityScale = 0.065 // Velocity scale factor based on magnitude
	PeakDecelFactor   = 0.08  // Peak deceleration factor
	PeakPowExp        = 1.3   // Exponent for velocity-based decay
	PeakFriction      = 0.92  // Friction factor for velocity damping
)

// CAVA algorithm constants
// These values are derived from the CAVA audio visualiser project
// https://github.com/karlstav/cava
const (
	Framerate      = 30.0
	NoiseReduction = 0.77  // CAVA default integral smoothing
	FallAccel      = 0.028 // CAVA gravity acceleration constant

	// Gravity modifier formula: pow(GravityFramerateRef/Framerate, GravityExponent) * GravityBase / NoiseReduction
	// This scales bar fall speed based on framerate deviation from CAVA's reference 60fps
	GravityFramerateRef = 60.0 // CAVA reference framerate for gravity calculation
	GravityExponent     = 2.5  // Exponent for framerate scaling
	GravityBase         = 1.54 // Base gravity multiplier
	GravityMin          = 1.0  // Minimum gravity modifier (floor)

	// Auto-sensitivity adjustment constants
	// These control dynamic gain adjustment based on peak detection
	SensitivityDecay   = 0.985 // Multiplier when overshoot detected (1.5% reduction per frame)
	SensitivityGrowth  = 1.002 // Multiplier when no overshoot (0.2% increase per frame)
	SensitivityMin     = 0.05  // Minimum sensitivity floor
	SensitivityMax     = 2.0   // Maximum sensitivity ceiling
	OvershootThreshold = 1.0   // Threshold for soft knee compression
)

// Appearance - Visual styling configuration
// Note: Future customization support will allow users to override these defaults.
// Embedded assets are currently located in internal/renderer/assets/
const (
	// Bar colors (RGB values for visualization bars)
	BarColorR = 164
	BarColorG = 0
	BarColorB = 0

	// Text/UI colors (RGB values for title overlay and framing lines)
	// Brand yellow #F8B31D - used for title text, framing lines, and thumbnail text
	TextColorR = 248
	TextColorG = 179
	TextColorB = 29

	// Embedded asset paths (relative to internal/renderer/assets/)
	// Background image: bg.png - scaled to video resolution (1280x720)
	// Thumbnail image: thumb.png - used as base for thumbnail generation
	BackgroundImageAsset = "assets/bg.png"
	ThumbnailImageAsset  = "assets/thumb.png"

	// Embedded font paths (relative to internal/renderer/assets/)
	// Video title font: Poppins-Regular.ttf - used for video overlay text
	// Thumbnail font: Poppins-Bold.ttf - used for thumbnail generation
	VideoTitleFontAsset = "assets/Poppins-Regular.ttf"
	ThumbnailFontAsset  = "assets/Poppins-Bold.ttf"

	// Thumbnail layout
	ThumbnailMargin              = 30  // Margin in pixels from edges for thumbnail text
	ThumbnailTextRotationDegrees = 3.0 // Rotation angle for thumbnail text (degrees, clockwise)

	// Video overlay
	FramingLineHeight = 4 // Height in pixels of framing lines above/below center gap
)

// Theme types for visualization
type ThemeType string

const (
	ThemeDefault   ThemeType = "default"
	ThemeSynthwave ThemeType = "synthwave"
)

// Synthwave theme HSV constants
const (
	SynthwaveHueStart = 0.83 // Purple start
	SynthwaveHueMid1  = 0.80 // Purple-Pink transition
	SynthwaveHueMid2  = 0.66 // Pink
	SynthwaveHueEnd   = 0.50 // Orange-Cyan end
)

// RuntimeConfig holds optional runtime overrides for customization
// When fields are nil/empty, the defaults from constants above are used
type RuntimeConfig struct {
	// Optional color overrides (RGB values 0-255)
	BarColorR *uint8
	BarColorG *uint8
	BarColorB *uint8

	TextColorR *uint8
	TextColorG *uint8
	TextColorB *uint8

	// Optional image path overrides
	BackgroundImagePath string
	ThumbnailImagePath  string

	// Theme selection
	Theme ThemeType
}

// GetBarColor returns the bar color RGB values (uses override or default)
func (c *RuntimeConfig) GetBarColor() (r, g, b uint8) {
	if c.BarColorR != nil && c.BarColorG != nil && c.BarColorB != nil {
		return *c.BarColorR, *c.BarColorG, *c.BarColorB
	}
	return BarColorR, BarColorG, BarColorB
}

// GetTextColor returns the text color RGB values (uses override or default)
func (c *RuntimeConfig) GetTextColor() (r, g, b uint8) {
	if c.TextColorR != nil && c.TextColorG != nil && c.TextColorB != nil {
		return *c.TextColorR, *c.TextColorG, *c.TextColorB
	}
	return TextColorR, TextColorG, TextColorB
}

// GetBackgroundImagePath returns the background image path (uses override or default embedded asset)
func (c *RuntimeConfig) GetBackgroundImagePath() string {
	if c.BackgroundImagePath != "" {
		return c.BackgroundImagePath
	}
	return BackgroundImageAsset
}

// GetThumbnailImagePath returns the thumbnail image path (uses override or default embedded asset)
func (c *RuntimeConfig) GetThumbnailImagePath() string {
	if c.ThumbnailImagePath != "" {
		return c.ThumbnailImagePath
	}
	return ThumbnailImageAsset
}

// ParseHexColor parses a hex color string (#RRGGBB or RRGGBB) and returns RGB values
func ParseHexColor(hex string) (r, g, b uint8, err error) {
	// Remove leading # if present
	hex = strings.TrimPrefix(hex, "#")

	// Validate length
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color format: must be 6 characters (RRGGBB)")
	}

	// Parse RGB components
	var rgb uint64
	rgb, err = strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %w", err)
	}

	r = uint8((rgb >> 16) & 0xFF)
	g = uint8((rgb >> 8) & 0xFF)
	b = uint8(rgb & 0xFF)

	return r, g, b, nil
}

// HSVToRGB converts HSV color values to RGB
// h: 0.0-1.0, s: 0.0-1.0, v: 0.0-1.0
// Returns: r, g, b as 0-255 uint8 values
func HSVToRGB(h, s, v float64) (r, g, b uint8) {
	var c, x, m float64
	var rFloat, gFloat, bFloat float64

	if s == 0 {
		rFloat, gFloat, bFloat = v, v, v
	} else {
		c = v * (1 - math.Abs(2*math.Mod(h, 1)-1))
		x = c * (1 - math.Abs(math.Mod(h*6, 2)-1))
		m = v - c

		switch {
		case h < 1/6:
			rFloat, gFloat, bFloat = c, x, 0
		case h < 2/6:
			rFloat, gFloat, bFloat = x, c, 0
		case h < 3/6:
			rFloat, gFloat, bFloat = 0, c, x
		case h < 4/6:
			rFloat, gFloat, bFloat = 0, x, c
		case h < 5/6:
			rFloat, gFloat, bFloat = x, 0, c
		default:
			rFloat, gFloat, bFloat = c, 0, x
		}
	}

	return uint8((rFloat + m) * 255), uint8((gFloat + m) * 255), uint8((bFloat + m) * 255)
}

// GetSynthwaveColor returns RGB color for Synthwave theme based on bar position and intensity
func GetSynthwaveColor(barIndex, numBars int, intensity float64) (r, g, b uint8) {
	// Clamp inputs to valid ranges to avoid math errors
	barIndex := barIndex
	if barIndex < 0 {
		barIndex = 0
	}
	if barIndex >= numBars {
		barIndex = numBars - 1
	}

	t := float64(barIndex) / float64(numBars-1)
	var hue float64
	
	// Calculate hue based on position
	if t < 0.3 {
		hue = SynthwaveHueStart + t*(SynthwaveHueMid1-SynthwaveHueStart)/0.3
	} else if t < 0.7 {
		hue = SynthwaveHueMid1 + (t-0.3)*(SynthwaveHueMid2-SynthwaveHueMid1)/0.4
	} else {
		hue = SynthwaveHueMid2 + (t-0.7)*(SynthwaveHueEnd-SynthwaveHueMid2)/0.3
	}

	// Adjust saturation and value based on intensity for dynamic appearance
	saturation := 0.72 + 0.23*math.Pow(intensity, 0.8)
	value := 0.63 + 0.33*math.Pow(intensity, 0.5)

	return HSVToRGB(hue, saturation, value)
}

	// Adjust saturation and value based on intensity for dynamic appearance
	saturation := 0.72 + 0.23*math.Pow(intensity, 0.8)
	value := 0.63 + 0.33*math.Pow(intensity, 0.5)

	return HSVToRGB(hue, saturation, value)
}
