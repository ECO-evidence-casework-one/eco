package eco

import (
	"errors"
	"image"
	"image/color"
	"math"
	"sort"
)

// PointF is a floating-point image coordinate used by document transforms.
type PointF struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Quad describes a photographed page in clockwise order: top-left,
// top-right, bottom-right and bottom-left.
type Quad struct {
	TopLeft     PointF `json:"top_left"`
	TopRight    PointF `json:"top_right"`
	BottomRight PointF `json:"bottom_right"`
	BottomLeft  PointF `json:"bottom_left"`
}

// CropSuggestion records a non-destructive page-boundary suggestion.
type CropSuggestion struct {
	Region     NormalizedRegion `json:"region"`
	Confidence float64          `json:"confidence"`
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func lumaByte(c color.Color) uint8 {
	return uint8(clampFloat(luminance(c), 0, 255))
}

func downsampleGray(img image.Image, maxDim int) *image.Gray {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return image.NewGray(image.Rect(0, 0, 1, 1))
	}
	scale := 1.0
	if w > maxDim || h > maxDim {
		scale = math.Min(float64(maxDim)/float64(w), float64(maxDim)/float64(h))
	}
	dw := max(1, int(math.Round(float64(w)*scale)))
	dh := max(1, int(math.Round(float64(h)*scale)))
	dst := image.NewGray(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		sy := b.Min.Y + minInt(h-1, int(float64(y)*float64(h)/float64(dh)))
		for x := 0; x < dw; x++ {
			sx := b.Min.X + minInt(w-1, int(float64(x)*float64(w)/float64(dw)))
			dst.SetGray(x, y, color.Gray{Y: lumaByte(img.At(sx, sy))})
		}
	}
	return dst
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func rectToNormalized(r, full image.Rectangle) NormalizedRegion {
	if full.Dx() <= 0 || full.Dy() <= 0 {
		return NormalizedRegion{}
	}
	return NormalizedRegion{
		X:      clampFloat(float64(r.Min.X-full.Min.X)/float64(full.Dx()), 0, 1),
		Y:      clampFloat(float64(r.Min.Y-full.Min.Y)/float64(full.Dy()), 0, 1),
		Width:  clampFloat(float64(r.Dx())/float64(full.Dx()), 0, 1),
		Height: clampFloat(float64(r.Dy())/float64(full.Dy()), 0, 1),
	}
}

func normalizedToRect(n NormalizedRegion, full image.Rectangle) image.Rectangle {
	if full.Dx() <= 0 || full.Dy() <= 0 {
		return image.Rectangle{}
	}
	x0 := full.Min.X + int(math.Round(clampFloat(n.X, 0, 1)*float64(full.Dx())))
	y0 := full.Min.Y + int(math.Round(clampFloat(n.Y, 0, 1)*float64(full.Dy())))
	x1 := full.Min.X + int(math.Round(clampFloat(n.X+n.Width, 0, 1)*float64(full.Dx())))
	y1 := full.Min.Y + int(math.Round(clampFloat(n.Y+n.Height, 0, 1)*float64(full.Dy())))
	if x1 <= x0 || y1 <= y0 {
		return image.Rectangle{}
	}
	return image.Rect(x0, y0, x1, y1).Intersect(full)
}

// SuggestDocumentBounds finds a likely page rectangle without changing the
// original image. It is deliberately conservative: low-confidence results
// return the full image rather than silently removing evidence.
func SuggestDocumentBounds(img image.Image) (image.Rectangle, float64) {
	full := img.Bounds()
	if full.Dx() < 40 || full.Dy() < 40 {
		return full, 0
	}
	g := downsampleGray(img, 600)
	b := g.Bounds()
	border := max(2, minInt(b.Dx(), b.Dy())/30)
	borderValues := make([]float64, 0, 2*border*(b.Dx()+b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if x < border || x >= b.Dx()-border || y < border || y >= b.Dy()-border {
				borderValues = append(borderValues, float64(g.GrayAt(x, y).Y))
			}
		}
	}
	if len(borderValues) == 0 {
		return full, 0
	}
	sort.Float64s(borderValues)
	bg := borderValues[len(borderValues)/2]
	var variance float64
	for _, v := range borderValues {
		d := v - bg
		variance += d * d
	}
	std := math.Sqrt(variance / float64(len(borderValues)))
	threshold := math.Max(18, 2.2*std)

	col := make([]float64, b.Dx())
	row := make([]float64, b.Dy())
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if math.Abs(float64(g.GrayAt(x, y).Y)-bg) >= threshold {
				col[x]++
				row[y]++
			}
		}
	}
	colNeed := math.Max(3, float64(b.Dy())*0.08)
	rowNeed := math.Max(3, float64(b.Dx())*0.08)
	left, right := -1, -1
	for x, v := range col {
		if v >= colNeed {
			if left < 0 {
				left = x
			}
			right = x
		}
	}
	top, bottom := -1, -1
	for y, v := range row {
		if v >= rowNeed {
			if top < 0 {
				top = y
			}
			bottom = y
		}
	}
	if left < 0 || top < 0 || right <= left || bottom <= top {
		return full, 0
	}
	marginX := max(2, b.Dx()/50)
	marginY := max(2, b.Dy()/50)
	left = max(0, left-marginX)
	top = max(0, top-marginY)
	right = minInt(b.Dx()-1, right+marginX)
	bottom = minInt(b.Dy()-1, bottom+marginY)
	areaRatio := float64((right-left+1)*(bottom-top+1)) / float64(b.Dx()*b.Dy())
	if areaRatio < 0.18 {
		return full, 0
	}

	// Compare activity inside and outside the proposed page.
	var inside, outside float64
	var insideN, outsideN float64
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			d := math.Abs(float64(g.GrayAt(x, y).Y) - bg)
			if x >= left && x <= right && y >= top && y <= bottom {
				inside += d
				insideN++
			} else {
				outside += d
				outsideN++
			}
		}
	}
	insideMean := inside / math.Max(1, insideN)
	outsideMean := outside / math.Max(1, outsideN)
	separation := (insideMean - outsideMean) / math.Max(20, insideMean)
	confidence := clampFloat(0.35+separation+math.Min(0.25, areaRatio/3), 0, 1)
	if areaRatio > 0.97 || confidence < 0.45 {
		return full, confidence
	}

	sx := float64(full.Dx()) / float64(b.Dx())
	sy := float64(full.Dy()) / float64(b.Dy())
	r := image.Rect(
		full.Min.X+int(math.Floor(float64(left)*sx)),
		full.Min.Y+int(math.Floor(float64(top)*sy)),
		full.Min.X+int(math.Ceil(float64(right+1)*sx)),
		full.Min.Y+int(math.Ceil(float64(bottom+1)*sy)),
	).Intersect(full)
	if r.Empty() {
		return full, 0
	}
	return r, confidence
}

// CropImage creates a derived view. It never changes the source image.
func CropImage(img image.Image, r image.Rectangle) image.Image {
	r = r.Intersect(img.Bounds())
	if r.Empty() {
		return img
	}
	dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			dst.Set(x, y, img.At(r.Min.X+x, r.Min.Y+y))
		}
	}
	return dst
}

func sampleBilinear(img image.Image, x, y float64, bg color.Color) color.Color {
	b := img.Bounds()
	if x < float64(b.Min.X) || y < float64(b.Min.Y) || x > float64(b.Max.X-1) || y > float64(b.Max.Y-1) {
		return bg
	}
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := minInt(b.Max.X-1, x0+1)
	y1 := minInt(b.Max.Y-1, y0+1)
	fx, fy := x-float64(x0), y-float64(y0)
	c00 := color.RGBAModel.Convert(img.At(x0, y0)).(color.RGBA)
	c10 := color.RGBAModel.Convert(img.At(x1, y0)).(color.RGBA)
	c01 := color.RGBAModel.Convert(img.At(x0, y1)).(color.RGBA)
	c11 := color.RGBAModel.Convert(img.At(x1, y1)).(color.RGBA)
	mix := func(a, b, c, d uint8) uint8 {
		v := (1-fx)*(1-fy)*float64(a) + fx*(1-fy)*float64(b) + (1-fx)*fy*float64(c) + fx*fy*float64(d)
		return uint8(clampFloat(math.Round(v), 0, 255))
	}
	return color.RGBA{mix(c00.R, c10.R, c01.R, c11.R), mix(c00.G, c10.G, c01.G, c11.G), mix(c00.B, c10.B, c01.B, c11.B), mix(c00.A, c10.A, c01.A, c11.A)}
}

// RotateImageAngle creates a non-destructive arbitrary-angle reading view.
// Positive angles rotate clockwise in screen coordinates.
func RotateImageAngle(img image.Image, degrees float64) image.Image {
	if math.Abs(degrees) < 0.01 {
		return img
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	rad := degrees * math.Pi / 180
	cosA, sinA := math.Cos(rad), math.Sin(rad)
	cx, cy := float64(w-1)/2, float64(h-1)/2
	bg := color.RGBA{255, 255, 255, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := float64(x)-cx, float64(y)-cy
			// Inverse map destination to source.
			sx := cosA*dx + sinA*dy + cx + float64(b.Min.X)
			sy := -sinA*dx + cosA*dy + cy + float64(b.Min.Y)
			dst.Set(x, y, sampleBilinear(img, sx, sy, bg))
		}
	}
	return dst
}

func projectionScore(img image.Image) float64 {
	g := downsampleGray(img, 420)
	threshold := otsuThreshold(g)
	rows := make([]float64, g.Bounds().Dy())
	var total float64
	for y := 0; y < g.Bounds().Dy(); y++ {
		for x := 0; x < g.Bounds().Dx(); x++ {
			if g.GrayAt(x, y).Y < threshold {
				rows[y]++
				total++
			}
		}
	}
	if total < float64(g.Bounds().Dx()*g.Bounds().Dy())*0.002 {
		return 0
	}
	mean := total / float64(len(rows))
	var variance float64
	for _, v := range rows {
		d := v - mean
		variance += d * d
	}
	return variance / float64(len(rows)) / math.Max(1, mean)
}

// EstimateSkewAngle returns the correction angle that most strongly aligns
// text-like horizontal structures. It is bounded to +/- 8 degrees.
func EstimateSkewAngle(img image.Image) (float64, float64) {
	small := downsampleGray(img, 420)
	type candidate struct{ angle, score float64 }
	candidates := make([]candidate, 0, 33)
	for i := -16; i <= 16; i++ {
		angle := float64(i) * 0.5
		score := projectionScore(RotateImageAngle(small, angle))
		candidates = append(candidates, candidate{angle, score})
	}
	best := candidates[0]
	scores := make([]float64, 0, len(candidates))
	for _, c := range candidates {
		scores = append(scores, c.score)
		if c.score > best.score {
			best = c
		}
	}
	if best.score <= 0 {
		return 0, 0
	}
	sort.Float64s(scores)
	median := scores[len(scores)/2]
	confidence := clampFloat((best.score-median)/math.Max(best.score, 1), 0, 1)
	if confidence < 0.08 || math.Abs(best.angle) < 0.26 {
		return 0, confidence
	}
	return best.angle, confidence
}

func otsuThreshold(g *image.Gray) uint8 {
	var hist [256]float64
	var total float64
	for _, v := range g.Pix {
		hist[v]++
		total++
	}
	if total == 0 {
		return 128
	}
	var sum float64
	for i, h := range hist {
		sum += float64(i) * h
	}
	var sumB, wB, maxVar float64
	threshold := 128
	for i, h := range hist {
		wB += h
		if wB == 0 {
			continue
		}
		wF := total - wB
		if wF == 0 {
			break
		}
		sumB += float64(i) * h
		mB := sumB / wB
		mF := (sum - sumB) / wF
		between := wB * wF * (mB - mF) * (mB - mF)
		if between > maxVar {
			maxVar = between
			threshold = i
		}
	}
	return uint8(threshold)
}

func AdaptiveThreshold(img image.Image) image.Image {
	g := downsampleGray(img, 800)
	threshold := otsuThreshold(g)
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			v := uint8(255)
			if lumaByte(img.At(b.Min.X+x, b.Min.Y+y)) <= threshold {
				v = 0
			}
			dst.SetRGBA(x, y, color.RGBA{v, v, v, 255})
		}
	}
	return dst
}

func distance(a, b PointF) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func pointFinite(p PointF) bool {
	return !math.IsNaN(p.X) && !math.IsNaN(p.Y) && !math.IsInf(p.X, 0) && !math.IsInf(p.Y, 0)
}

func cross(a, b, c PointF) float64 {
	return (b.X-a.X)*(c.Y-b.Y) - (b.Y-a.Y)*(c.X-b.X)
}

func validatePerspectiveQuad(q Quad, bounds image.Rectangle) error {
	points := []PointF{q.TopLeft, q.TopRight, q.BottomRight, q.BottomLeft}
	if bounds.Empty() {
		return errors.New("source image is empty")
	}
	for _, p := range points {
		if !pointFinite(p) || p.X < float64(bounds.Min.X) || p.Y < float64(bounds.Min.Y) || p.X > float64(bounds.Max.X-1) || p.Y > float64(bounds.Max.Y-1) {
			return errors.New("perspective-correction point is outside the source image")
		}
	}
	var sign float64
	for i := 0; i < 4; i++ {
		c := cross(points[i], points[(i+1)%4], points[(i+2)%4])
		if math.Abs(c) < 0.5 {
			return errors.New("perspective-correction quadrilateral is degenerate")
		}
		if sign == 0 {
			sign = math.Copysign(1, c)
		} else if math.Copysign(1, c) != sign {
			return errors.New("perspective-correction points are self-intersecting or unordered")
		}
	}
	var area float64
	for i := 0; i < 4; i++ {
		a, b := points[i], points[(i+1)%4]
		area += a.X*b.Y - b.X*a.Y
	}
	if math.Abs(area)/2 < 4 {
		return errors.New("perspective-correction quadrilateral is too small")
	}
	return nil
}

// PerspectiveCorrect maps a photographed quadrilateral to a flat rectangle.
// The interpolation is bounded and deterministic. The original is untouched.
func PerspectiveCorrect(img image.Image, q Quad) (image.Image, error) {
	if err := validatePerspectiveQuad(q, img.Bounds()); err != nil {
		return nil, err
	}
	w := int(math.Round((distance(q.TopLeft, q.TopRight) + distance(q.BottomLeft, q.BottomRight)) / 2))
	h := int(math.Round((distance(q.TopLeft, q.BottomLeft) + distance(q.TopRight, q.BottomRight)) / 2))
	if w < 2 || h < 2 || int64(w)*int64(h) > 100_000_000 {
		return nil, errors.New("unsafe or invalid perspective-correction dimensions")
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{255, 255, 255, 255}
	for y := 0; y < h; y++ {
		v := float64(y) / float64(max(1, h-1))
		for x := 0; x < w; x++ {
			u := float64(x) / float64(max(1, w-1))
			// Bilinear map through the source quadrilateral.
			sx := (1-u)*(1-v)*q.TopLeft.X + u*(1-v)*q.TopRight.X + u*v*q.BottomRight.X + (1-u)*v*q.BottomLeft.X
			sy := (1-u)*(1-v)*q.TopLeft.Y + u*(1-v)*q.TopRight.Y + u*v*q.BottomRight.Y + (1-u)*v*q.BottomLeft.Y
			dst.Set(x, y, sampleBilinear(img, sx, sy, bg))
		}
	}
	return dst, nil
}

func imageVisionMetrics(img image.Image) (glareRatio, shadowImbalance, borderInkRatio float64, probableDoublePage bool) {
	g := downsampleGray(img, 500)
	w, h := g.Bounds().Dx(), g.Bounds().Dy()
	if w <= 0 || h <= 0 {
		return
	}
	var glare, total float64
	quadSum := [4]float64{}
	quadN := [4]float64{}
	borderW := max(1, minInt(w, h)/30)
	var borderInk, borderN float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := float64(g.GrayAt(x, y).Y)
			total++
			if v >= 248 {
				glare++
			}
			qi := 0
			if x >= w/2 {
				qi++
			}
			if y >= h/2 {
				qi += 2
			}
			quadSum[qi] += v
			quadN[qi]++
			if x < borderW || x >= w-borderW || y < borderW || y >= h-borderW {
				borderN++
				if v < 80 {
					borderInk++
				}
			}
		}
	}
	glareRatio = glare / math.Max(1, total)
	mins, maxs := 255.0, 0.0
	for i := range quadSum {
		m := quadSum[i] / math.Max(1, quadN[i])
		mins = math.Min(mins, m)
		maxs = math.Max(maxs, m)
	}
	shadowImbalance = maxs - mins
	borderInkRatio = borderInk / math.Max(1, borderN)
	if float64(w)/float64(h) > 1.35 {
		strip := max(1, w/30)
		centerStart := max(0, w/2-strip)
		centerEnd := minInt(w, w/2+strip)
		var center, left, right float64
		var cn, ln, rn float64
		for y := h / 10; y < h-h/10; y++ {
			for x := 0; x < w; x++ {
				v := float64(g.GrayAt(x, y).Y)
				switch {
				case x >= centerStart && x < centerEnd:
					center += v
					cn++
				case x < w/3:
					left += v
					ln++
				case x >= 2*w/3:
					right += v
					rn++
				}
			}
		}
		centerMean := center / math.Max(1, cn)
		sideMean := (left/math.Max(1, ln) + right/math.Max(1, rn)) / 2
		probableDoublePage = centerMean > sideMean+7
	}
	return
}

// BoundedPreviewImage limits decoded preview memory while preserving the
// encrypted original at full resolution. It uses bilinear sampling and never
// replaces the evidence object.
func BoundedPreviewImage(img image.Image, maxPixels int64) image.Image {
	b := img.Bounds()
	pixels := int64(b.Dx()) * int64(b.Dy())
	if maxPixels <= 0 || pixels <= maxPixels || b.Dx() <= 0 || b.Dy() <= 0 {
		return img
	}
	scale := math.Sqrt(float64(maxPixels) / float64(pixels))
	w := max(1, int(math.Floor(float64(b.Dx())*scale)))
	h := max(1, int(math.Floor(float64(b.Dy())*scale)))
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{255, 255, 255, 255}
	for y := 0; y < h; y++ {
		sy := float64(b.Min.Y) + (float64(y)+0.5)*float64(b.Dy())/float64(h) - 0.5
		for x := 0; x < w; x++ {
			sx := float64(b.Min.X) + (float64(x)+0.5)*float64(b.Dx())/float64(w) - 0.5
			dst.Set(x, y, sampleBilinear(img, sx, sy, bg))
		}
	}
	return dst
}
