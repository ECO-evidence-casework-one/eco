package eco

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
)

func DecodeSupportedImage(data []byte) (image.Image, string, error) {
	if len(data) >= 2 && string(data[:2]) == "BM" {
		img, err := decodeBMP(data)
		return img, "bmp", err
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 160_000_000 {
		return nil, format, fmt.Errorf("image dimensions are unsafe or too large: %dx%d", cfg.Width, cfg.Height)
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	return img, format, err
}

func init() {
	image.RegisterFormat("jpeg", "\xff\xd8", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "GIF8", gif.Decode, gif.DecodeConfig)
}

func AssessImage(img image.Image) ImageAssessment {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	orientation := "Square"
	if w > h {
		orientation = "Landscape"
	} else if h > w {
		orientation = "Portrait"
	}

	maxSample := 800
	sx, sy := max(1, w/maxSample), max(1, h/maxSample)
	var count float64
	var sum, sumSq float64
	gray := make([][]float64, 0, h/sy+1)
	for y := b.Min.Y; y < b.Max.Y; y += sy {
		row := make([]float64, 0, w/sx+1)
		for x := b.Min.X; x < b.Max.X; x += sx {
			v := luminance(img.At(x, y))
			row = append(row, v)
			sum += v
			sumSq += v * v
			count++
		}
		gray = append(gray, row)
	}
	mean := 0.0
	std := 0.0
	if count > 0 {
		mean = sum / count
		variance := sumSq/count - mean*mean
		if variance > 0 {
			std = math.Sqrt(variance)
		}
	}

	var lapSum, lapSq, lapCount float64
	var edgeCount float64
	for y := 1; y+1 < len(gray); y++ {
		for x := 1; x+1 < len(gray[y]); x++ {
			v := -4*gray[y][x] + gray[y-1][x] + gray[y+1][x] + gray[y][x-1] + gray[y][x+1]
			lapSum += v
			lapSq += v * v
			lapCount++
			if math.Abs(v) > 18 {
				edgeCount++
			}
		}
	}
	blurVar := 0.0
	edgeDensity := 0.0
	if lapCount > 0 {
		lm := lapSum / lapCount
		blurVar = lapSq/lapCount - lm*lm
		edgeDensity = edgeCount / lapCount
	}

	glareRatio, shadowImbalance, borderInkRatio, probableDoublePage := imageVisionMetrics(img)
	skewCorrection, skewConfidence := EstimateSkewAngle(img)
	cropRect, cropConfidence := SuggestDocumentBounds(img)

	warnings := []string{}
	if w < 900 || h < 900 {
		warnings = append(warnings, "Resolution may be too low for dependable small-text reading.")
	}
	if mean < 55 {
		warnings = append(warnings, "The image is very dark and may need brightness or shadow correction.")
	}
	if mean > 225 {
		warnings = append(warnings, "The image is very bright and highlights may be washed out.")
	}
	if std < 25 {
		warnings = append(warnings, "Contrast is low; faint text may be difficult to read.")
	}
	if blurVar < 90 {
		warnings = append(warnings, "The image may be blurred. Check important wording against the original view.")
	}
	if glareRatio > 0.08 {
		warnings = append(warnings, "Bright glare may obscure part of the page. Compare important wording with the original photograph.")
	}
	if shadowImbalance > 55 {
		warnings = append(warnings, "Lighting is uneven across the image and may hide faint text.")
	}
	if borderInkRatio > 0.28 {
		warnings = append(warnings, "Dark content reaches the image edge. A page corner or wording may be cut off.")
	}
	if math.Abs(skewCorrection) >= 1.0 && skewConfidence >= 0.12 {
		warnings = append(warnings, fmt.Sprintf("The page appears skewed by about %.1f°. ECO can preview a non-destructive deskew correction.", -skewCorrection))
	}
	if probableDoublePage {
		warnings = append(warnings, "This may contain two photographed pages. Review whether the image should be split before OCR.")
	}
	if float64(w*h)/1_000_000 > 80 {
		warnings = append(warnings, "This is an unusually large image. ECO will use bounded preview processing.")
	}
	label := "Good local preview quality"
	if len(warnings) == 1 {
		label = "Check one quality warning"
	}
	if len(warnings) > 1 {
		label = fmt.Sprintf("Check %d quality warnings", len(warnings))
	}
	var crop *CropSuggestion
	if cropConfidence >= 0.45 && cropRect != img.Bounds() {
		crop = &CropSuggestion{Region: rectToNormalized(cropRect, img.Bounds()), Confidence: cropConfidence}
	}
	return ImageAssessment{
		Width: w, Height: h, Megapixels: float64(w*h) / 1_000_000,
		Brightness: mean, Contrast: std, BlurVariance: blurVar,
		Orientation: orientation, QualityLabel: label, Warnings: warnings,
		EdgeDensity: edgeDensity, PerceptualHash: DifferenceHash(img),
		SkewCorrectionDegrees: skewCorrection, SkewConfidence: skewConfidence,
		GlareRatio: glareRatio, ShadowImbalance: shadowImbalance, BorderInkRatio: borderInkRatio,
		ProbableDoublePage: probableDoublePage, SuggestedCrop: crop,
	}
}

func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	rf, gf, bf := float64(r>>8), float64(g>>8), float64(b>>8)
	return 0.2126*rf + 0.7152*gf + 0.0722*bf
}

func DifferenceHash(img image.Image) string {
	b := img.Bounds()
	var bits uint64
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			x1 := b.Min.X + (x*b.Dx())/9
			x2 := b.Min.X + ((x+1)*b.Dx())/9
			yy := b.Min.Y + ((2*y+1)*b.Dy())/16
			bits <<= 1
			if luminance(img.At(x1, yy)) > luminance(img.At(x2, yy)) {
				bits |= 1
			}
		}
	}
	return fmt.Sprintf("%016x", bits)
}

func HashDistance(a, b string) int {
	if len(a) != 16 || len(b) != 16 {
		return 65
	}
	var x, y uint64
	if _, err := fmt.Sscanf(a, "%x", &x); err != nil {
		return 65
	}
	if _, err := fmt.Sscanf(b, "%x", &y); err != nil {
		return 65
	}
	v := x ^ y
	n := 0
	for v != 0 {
		n++
		v &= v - 1
	}
	return n
}

func ApplyReadingMode(img image.Image, mode string) image.Image {
	if mode == "original" || mode == "" {
		return img
	}
	if mode == "adaptive" {
		return AdaptiveThreshold(img)
	}
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			l := uint8(math.Max(0, math.Min(255, luminance(img.At(b.Min.X+x, b.Min.Y+y)))))
			if mode == "contrast" {
				if l > 150 {
					l = 255
				} else {
					l = 0
				}
			}
			dst.SetRGBA(x, y, color.RGBA{l, l, l, 255})
		}
	}
	return dst
}

func RotateImage(img image.Image, degrees int) image.Image {
	degrees = ((degrees % 360) + 360) % 360
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if degrees == 0 {
		return img
	}
	if degrees == 180 {
		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(w-1-x, h-1-y, img.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	}
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if degrees == 90 {
				dst.Set(h-1-y, x, img.At(b.Min.X+x, b.Min.Y+y))
			} else {
				dst.Set(y, w-1-x, img.At(b.Min.X+x, b.Min.Y+y))
			}
		}
	}
	return dst
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func decodeBMP(data []byte) (image.Image, error) {
	if len(data) < 54 || string(data[:2]) != "BM" {
		return nil, errors.New("invalid BMP header")
	}
	offset := int(binary.LittleEndian.Uint32(data[10:14]))
	dibSize := int(binary.LittleEndian.Uint32(data[14:18]))
	if dibSize < 40 || len(data) < 14+dibSize {
		return nil, errors.New("unsupported BMP DIB header")
	}
	width := int(int32(binary.LittleEndian.Uint32(data[18:22])))
	rawHeight := int32(binary.LittleEndian.Uint32(data[22:26]))
	topDown := rawHeight < 0
	height := int(rawHeight)
	if height < 0 {
		height = -height
	}
	planes := binary.LittleEndian.Uint16(data[26:28])
	bits := binary.LittleEndian.Uint16(data[28:30])
	compression := binary.LittleEndian.Uint32(data[30:34])
	if width <= 0 || height <= 0 || int64(width)*int64(height) > 160_000_000 {
		return nil, errors.New("unsafe BMP dimensions")
	}
	if planes != 1 || compression != 0 || (bits != 24 && bits != 32) {
		return nil, fmt.Errorf("unsupported BMP format: planes=%d bits=%d compression=%d", planes, bits, compression)
	}
	rowBytes := ((width*int(bits) + 31) / 32) * 4
	need := offset + rowBytes*height
	if offset < 0 || need > len(data) {
		return nil, errors.New("truncated BMP pixel data")
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bytesPerPixel := int(bits / 8)
	for y := 0; y < height; y++ {
		srcY := height - 1 - y
		if topDown {
			srcY = y
		}
		row := data[offset+srcY*rowBytes : offset+(srcY+1)*rowBytes]
		for x := 0; x < width; x++ {
			i := x * bytesPerPixel
			b, g, r := row[i], row[i+1], row[i+2]
			a := byte(255)
			if bytesPerPixel == 4 {
				a = row[i+3]
				if a == 0 {
					a = 255
				}
			}
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return img, nil
}
