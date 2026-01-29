package renderer

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/linuxmatters/jivefire/internal/config"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	_ "image/jpeg"
	_ "image/png"
)

// getThumbnailTextColor returns the text color for thumbnail (uses runtime config or default)
func getThumbnailTextColor(runtimeConfig *config.RuntimeConfig) color.RGBA {
	return color.RGBA{
		R: config.ThumbnailTextColorR,
		G: config.ThumbnailTextColorG,
		B: config.ThumbnailTextColorB,
		A: 255,
	}
}

func getThumbnailShadowColor() color.RGBA {
	return color.RGBA{R: 0, G: 0, B: 0, A: config.ThumbnailShadowAlpha}
}

func getThumbnailGlowColor() color.RGBA {
	return color.RGBA{R: config.TextColorR, G: config.TextColorG, B: config.TextColorB, A: config.ThumbnailGlowAlpha}
}

// GenerateThumbnail creates a YouTube thumbnail with the title text overlaid
// The thumbnail is the same resolution as the video (1280x720)
func GenerateThumbnail(outputPath string, title string, runtimeConfig *config.RuntimeConfig) error {
	// Load the thumbnail background image
	thumbImg, err := loadThumbnailBackground(runtimeConfig)
	if err != nil {
		return fmt.Errorf("failed to load thumbnail background: %w", err)
	}

	// Load the bold font for thumbnail
	fontData, err := embeddedAssets.ReadFile(config.ThumbnailFontAsset)
	if err != nil {
		return fmt.Errorf("failed to load bold font: %w", err)
	}

	parsedFont, err := truetype.Parse(fontData)
	if err != nil {
		return fmt.Errorf("failed to parse font: %w", err)
	}

	// Split title into 2 lines
	line1, line2 := splitTitle(title)

	// Find the largest font size that fits within constraints
	boxWidth := config.ThumbnailBoxRight - config.ThumbnailBoxLeft
	boxHeight := config.ThumbnailBoxBottom - config.ThumbnailBoxTop
	fontSize := findOptimalFontSize(parsedFont, line1, line2, boxWidth, boxHeight)

	// Create font face with optimal size
	face := truetype.NewFace(parsedFont, &truetype.Options{
		Size: fontSize,
		DPI:  72,
	})
	defer face.Close()

	// Draw the text on the thumbnail
	drawThumbnailText(thumbImg, face, line1, line2, runtimeConfig)

	// Save the thumbnail
	if err := saveThumbnail(thumbImg, outputPath); err != nil {
		return fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return nil
}

// loadThumbnailBackground loads and scales the thumbnail background (from custom path or embedded asset)
func loadThumbnailBackground(runtimeConfig *config.RuntimeConfig) (*image.RGBA, error) {
	imagePath := runtimeConfig.GetThumbnailImagePath()

	var data []byte
	var err error

	// Check if using custom image path or embedded asset
	if runtimeConfig.ThumbnailImagePath != "" {
		// Load from filesystem
		data, err = os.ReadFile(imagePath)
	} else {
		// Prefer on-disk assets for local overrides, fallback to embedded assets
		diskPath := filepath.Join("internal", "renderer", imagePath)
		if _, statErr := os.Stat(diskPath); statErr == nil {
			data, err = os.ReadFile(diskPath)
		} else {
			data, err = embeddedAssets.ReadFile(imagePath)
		}
	}

	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Check if scaling is needed
	bounds := img.Bounds()
	if bounds.Dx() == config.Width && bounds.Dy() == config.Height {
		// Already correct size, just convert to RGBA
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
		return rgba, nil
	}

	// Scale to video resolution using the same method as LoadBackgroundImage
	dst := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Src, nil)
	return dst, nil
}

// splitTitle splits the title into 2 roughly equal lines
func splitTitle(title string) (string, string) {
	words := strings.Fields(title)
	if len(words) == 0 {
		return "", ""
	}
	if len(words) == 1 {
		return words[0], ""
	}

	// Split at the midpoint for roughly equal lines
	mid := len(words) / 2
	line1 := strings.Join(words[:mid], " ")
	line2 := strings.Join(words[mid:], " ")

	return line1, line2
}

// findOptimalFontSize finds the largest font size that fits within the title box.
func findOptimalFontSize(parsedFont *truetype.Font, line1, line2 string, boxWidth, boxHeight int) float64 {
	maxWidth := boxWidth - (2 * config.ThumbnailBoxPadding)
	maxHeight := boxHeight - (2 * config.ThumbnailBoxPadding)

	// Start with a large size and reduce until it fits
	for size := 150.0; size > 10.0; size -= 2.0 {
		face := truetype.NewFace(parsedFont, &truetype.Options{
			Size: size,
			DPI:  72,
		})

		// Measure both lines
		width1, bounds1 := measureText(face, line1)
		width2, bounds2 := measureText(face, line2)

		face.Close()

		// Check if both lines fit within width constraint
		if width1 > maxWidth || width2 > maxWidth {
			continue
		}

		// Calculate line spacing (50% of font size for more vertical spacing)
		lineSpacing := int(size * 0.5)

		// Height of each line (from top to bottom of glyphs)
		height1 := (bounds1.Max.Y - bounds1.Min.Y).Ceil()
		height2 := (bounds2.Max.Y - bounds2.Min.Y).Ceil()

		// Calculate where line 2 bottom would be:
		// Line 1 top: margin
		// Line 1 bottom: margin + height1
		// Line 2 top: margin + height1 + lineSpacing
		// Line 2 bottom: margin + height1 + lineSpacing + height2
		// Check if text block fits within the box
		totalHeight := height1 + lineSpacing + height2
		if width1 <= maxWidth && width2 <= maxWidth && totalHeight <= maxHeight {
			return size
		}
	}

	return 10.0 // Minimum fallback size
}

// measureText returns the width and actual bounds of rendered text
// Returns width, and the bounds rectangle (Min.Y is negative for ascent, Max.Y is positive for descent)
func measureText(face font.Face, text string) (int, fixed.Rectangle26_6) {
	d := &font.Drawer{Face: face}
	bounds, _ := d.BoundString(text)
	width := (bounds.Max.X - bounds.Min.X).Ceil()
	return width, bounds
}

// drawThumbnailText draws the title text centred within the thumbnail box.
func drawThumbnailText(img *image.RGBA, face font.Face, line1, line2 string, runtimeConfig *config.RuntimeConfig) {
	// Measure text dimensions - bounds.Min.Y is negative (ascent), bounds.Max.Y is positive (descent)
	width1, bounds1 := measureText(face, line1)
	width2, bounds2 := measureText(face, line2)

	// Calculate line spacing (35% of font size for tighter stacking)
	metrics := face.Metrics()
	fontSize := float64(metrics.Height) / 64.0 // Convert from fixed.Int26_6 to float64
	lineSpacing := int(fontSize * 0.35)

	// Calculate the height of each line (from visual top to visual bottom)
	height1 := (bounds1.Max.Y - bounds1.Min.Y).Ceil()
	height2 := (bounds2.Max.Y - bounds2.Min.Y).Ceil()

	// Calculate total text block dimensions
	maxWidth := width1
	if width2 > maxWidth {
		maxWidth = width2
	}
	totalHeight := height1 + lineSpacing + height2

	// Position text block centred within the thumbnail box.
	boxLeft := config.ThumbnailBoxLeft
	boxTop := config.ThumbnailBoxTop
	boxRight := config.ThumbnailBoxRight
	boxBottom := config.ThumbnailBoxBottom
	boxWidth := boxRight - boxLeft
	boxHeight := boxBottom - boxTop
	centerY := boxTop + boxHeight/2
	line1VisualTop := centerY - totalHeight/2
	line1BaselineY := line1VisualTop - bounds1.Min.Y.Ceil()

	line2VisualTop := line1VisualTop + height1 + lineSpacing
	line2BaselineY := line2VisualTop - bounds2.Min.Y.Ceil()

	// Draw a soft glow layer first, then shadow, then text.
	glowImg := image.NewRGBA(img.Bounds())
	drawCenteredLineWithColor(glowImg, face, line1, boxWidth, boxLeft, line1BaselineY, getThumbnailGlowColor())
	drawCenteredLineWithColor(glowImg, face, line2, boxWidth, boxLeft, line2BaselineY, getThumbnailGlowColor())
	blurredGlow := blurRGBA(glowImg, config.ThumbnailGlowBlurRadius)
	draw.Draw(img, img.Bounds(), blurredGlow, image.Point{}, draw.Over)

	drawCenteredLineWithColor(img, face, line1, boxWidth, boxLeft, line1BaselineY, getThumbnailShadowColor())
	drawCenteredLineWithColor(img, face, line2, boxWidth, boxLeft, line2BaselineY, getThumbnailShadowColor())
	drawCenteredLineWithColor(img, face, line1, boxWidth, boxLeft, line1BaselineY, getThumbnailTextColor(runtimeConfig))
	drawCenteredLineWithColor(img, face, line2, boxWidth, boxLeft, line2BaselineY, getThumbnailTextColor(runtimeConfig))
}

// drawCenteredLineWithColor draws a line of text centred on the target image.
func drawCenteredLineWithColor(img *image.RGBA, face font.Face, text string, boxWidth, boxLeft, baselineY int, textColor color.RGBA) {
	if text == "" {
		return
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	// Measure text width
	bounds, _ := d.BoundString(text)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil()

	// Center horizontally
	x := boxLeft + (boxWidth-textWidth)/2

	d.Dot = freetype.Pt(x, baselineY)
	d.DrawString(text)
}

func blurRGBA(src *image.RGBA, radius int) *image.RGBA {
	if radius <= 0 {
		return src
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	stride := src.Stride
	tmp := image.NewRGBA(bounds)
	dst := image.NewRGBA(bounds)

	// Horizontal pass.
	for y := 0; y < height; y++ {
		row := y * stride
		var sumR, sumG, sumB, sumA int
		window := 0
		for x := -radius; x <= radius; x++ {
			ix := x
			if ix < 0 {
				ix = 0
			} else if ix >= width {
				ix = width - 1
			}
			idx := row + ix*4
			sumR += int(src.Pix[idx])
			sumG += int(src.Pix[idx+1])
			sumB += int(src.Pix[idx+2])
			sumA += int(src.Pix[idx+3])
			window++
		}
		for x := 0; x < width; x++ {
			idx := row + x*4
			tmp.Pix[idx] = uint8(sumR / window)
			tmp.Pix[idx+1] = uint8(sumG / window)
			tmp.Pix[idx+2] = uint8(sumB / window)
			tmp.Pix[idx+3] = uint8(sumA / window)

			left := x - radius
			right := x + radius + 1
			if left >= 0 {
				lidx := row + left*4
				sumR -= int(src.Pix[lidx])
				sumG -= int(src.Pix[lidx+1])
				sumB -= int(src.Pix[lidx+2])
				sumA -= int(src.Pix[lidx+3])
				window--
			}
			if right < width {
				ridx := row + right*4
				sumR += int(src.Pix[ridx])
				sumG += int(src.Pix[ridx+1])
				sumB += int(src.Pix[ridx+2])
				sumA += int(src.Pix[ridx+3])
				window++
			}
		}
	}

	// Vertical pass.
	for x := 0; x < width; x++ {
		var sumR, sumG, sumB, sumA int
		window := 0
		for y := -radius; y <= radius; y++ {
			iy := y
			if iy < 0 {
				iy = 0
			} else if iy >= height {
				iy = height - 1
			}
			idx := iy*stride + x*4
			sumR += int(tmp.Pix[idx])
			sumG += int(tmp.Pix[idx+1])
			sumB += int(tmp.Pix[idx+2])
			sumA += int(tmp.Pix[idx+3])
			window++
		}
		for y := 0; y < height; y++ {
			idx := y*stride + x*4
			dst.Pix[idx] = uint8(sumR / window)
			dst.Pix[idx+1] = uint8(sumG / window)
			dst.Pix[idx+2] = uint8(sumB / window)
			dst.Pix[idx+3] = uint8(sumA / window)

			top := y - radius
			bottom := y + radius + 1
			if top >= 0 {
				tidx := top*stride + x*4
				sumR -= int(tmp.Pix[tidx])
				sumG -= int(tmp.Pix[tidx+1])
				sumB -= int(tmp.Pix[tidx+2])
				sumA -= int(tmp.Pix[tidx+3])
				window--
			}
			if bottom < height {
				bidx := bottom*stride + x*4
				sumR += int(tmp.Pix[bidx])
				sumG += int(tmp.Pix[bidx+1])
				sumB += int(tmp.Pix[bidx+2])
				sumA += int(tmp.Pix[bidx+3])
				window++
			}
		}
	}

	return dst
}

// saveThumbnail saves the thumbnail image to a PNG file
func saveThumbnail(img *image.RGBA, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return png.Encode(outFile, img)
}
