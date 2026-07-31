package eco

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	maxOCRWords        = 250_000
	maxOCRLines        = 50_000
	maxOCRWordText     = 4_096
	maxOCRLineText     = 32_768
	maxOCRWarningText  = 4_096
	maxOCRIdentityText = 200
	maxOCRLanguageText = 100
)

var sha256TextPattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

// ParseOCRTSV converts coordinate-bearing Tesseract-compatible TSV output
// into source segments and a reviewable OCR receipt. It does not claim that
// OCR is correct: every line retains confidence and its exact image region.
func ParseOCRTSV(tsv, engine, engineVersion, language, sourceSHA256 string, imageWidth, imageHeight int) (OCRReceipt, []SourceSegment, error) {
	if imageWidth <= 0 || imageHeight <= 0 {
		return OCRReceipt{}, nil, errors.New("OCR image dimensions must be positive")
	}
	if strings.TrimSpace(engine) == "" || len([]rune(engine)) > maxOCRIdentityText {
		return OCRReceipt{}, nil, errors.New("OCR engine identity is required and must be bounded")
	}
	if len([]rune(engineVersion)) > maxOCRIdentityText || len([]rune(language)) > maxOCRLanguageText {
		return OCRReceipt{}, nil, errors.New("OCR engine version or language is unbounded")
	}
	if !sha256TextPattern.MatchString(sourceSHA256) {
		return OCRReceipt{}, nil, errors.New("OCR source SHA-256 is required")
	}
	type lineKey struct {
		page, block, paragraph, line int
	}
	type lineBuild struct {
		words                    []OCRWord
		left, top, right, bottom int
		confSum                  float64
		confN                    int
	}
	lines := map[lineKey]*lineBuild{}
	order := make([]lineKey, 0)
	wordCount := 0
	scanner := bufio.NewScanner(strings.NewReader(tsv))
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	row := 0
	for scanner.Scan() {
		row++
		text := scanner.Text()
		if row == 1 && strings.HasPrefix(strings.ToLower(text), "level\t") {
			continue
		}
		cols := strings.Split(text, "\t")
		if len(cols) < 12 {
			continue
		}
		level, _ := strconv.Atoi(cols[0])
		if level != 5 {
			continue
		}
		wordText := strings.TrimSpace(strings.Join(cols[11:], "	"))
		if wordText == "" {
			continue
		}
		if len([]rune(wordText)) > maxOCRWordText {
			wordText = string([]rune(wordText)[:maxOCRWordText])
		}
		if wordCount >= maxOCRWords {
			return OCRReceipt{}, nil, errors.New("OCR output exceeds the safe word limit")
		}
		page, _ := strconv.Atoi(cols[1])
		block, _ := strconv.Atoi(cols[2])
		paragraph, _ := strconv.Atoi(cols[3])
		line, _ := strconv.Atoi(cols[4])
		left, e1 := strconv.Atoi(cols[6])
		top, e2 := strconv.Atoi(cols[7])
		width, e3 := strconv.Atoi(cols[8])
		height, e4 := strconv.Atoi(cols[9])
		conf, e5 := strconv.ParseFloat(cols[10], 64)
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil || width <= 0 || height <= 0 || math.IsNaN(conf) || math.IsInf(conf, 0) {
			continue
		}
		wordRect := image.Rect(left, top, left+width, top+height).Intersect(image.Rect(0, 0, imageWidth, imageHeight))
		if wordRect.Empty() {
			continue
		}
		left, top, width, height = wordRect.Min.X, wordRect.Min.Y, wordRect.Dx(), wordRect.Dy()
		region := rectToNormalized(wordRect, image.Rect(0, 0, imageWidth, imageHeight))
		word := OCRWord{Text: wordText, Confidence: clampFloat(conf/100, 0, 1), Region: region, Page: max(1, page)}
		key := lineKey{max(1, page), block, paragraph, line}
		lb := lines[key]
		if lb == nil {
			if len(order) >= maxOCRLines {
				return OCRReceipt{}, nil, errors.New("OCR output exceeds the safe line limit")
			}
			lb = &lineBuild{left: left, top: top, right: left + width, bottom: top + height}
			lines[key] = lb
			order = append(order, key)
		}
		lb.words = append(lb.words, word)
		if left < lb.left {
			lb.left = left
		}
		if top < lb.top {
			lb.top = top
		}
		if left+width > lb.right {
			lb.right = left + width
		}
		if top+height > lb.bottom {
			lb.bottom = top + height
		}
		if conf >= 0 {
			lb.confSum += clampFloat(conf/100, 0, 1)
			lb.confN++
		}
		wordCount++
	}
	if err := scanner.Err(); err != nil {
		return OCRReceipt{}, nil, err
	}
	if wordCount == 0 {
		receipt := OCRReceipt{Engine: engine, EngineVersion: engineVersion, Language: language, Status: "no-text", SourceSHA256: sourceSHA256, CreatedAt: time.Now().UTC()}
		if err := ValidateOCRReceipt(receipt); err != nil {
			return OCRReceipt{}, nil, err
		}
		return receipt, nil, nil
	}

	receipt := OCRReceipt{Engine: engine, EngineVersion: engineVersion, Language: language, Status: "ready", SourceSHA256: sourceSHA256, CreatedAt: time.Now().UTC()}
	segments := make([]SourceSegment, 0, len(order))
	var allConf float64
	var allN int
	for i, key := range order {
		lb := lines[key]
		parts := make([]string, 0, len(lb.words))
		for _, w := range lb.words {
			parts = append(parts, w.Text)
			receipt.Words = append(receipt.Words, w)
		}
		conf := 0.0
		if lb.confN > 0 {
			conf = lb.confSum / float64(lb.confN)
			allConf += lb.confSum
			allN += lb.confN
		}
		lineRect := image.Rect(lb.left, lb.top, lb.right, lb.bottom).Intersect(image.Rect(0, 0, imageWidth, imageHeight))
		if lineRect.Empty() {
			continue
		}
		region := rectToNormalized(lineRect, image.Rect(0, 0, imageWidth, imageHeight))
		lineText := strings.Join(parts, " ")
		if len([]rune(lineText)) > maxOCRLineText {
			lineText = string([]rune(lineText)[:maxOCRLineText])
		}
		line := OCRLine{Text: lineText, Confidence: conf, Region: region, Page: key.page, Words: append([]OCRWord(nil), lb.words...)}
		receipt.Lines = append(receipt.Lines, line)
		regionCopy := region
		segments = append(segments, SourceSegment{ID: fmt.Sprintf("OCR-%d-%d", key.page, i+1), Ordinal: i + 1, Text: lineText, PageHint: fmt.Sprintf("Image page %d", key.page), Page: key.page, Region: &regionCopy, Origin: "ocr", Confidence: conf})
	}
	if allN > 0 {
		receipt.AverageConfidence = allConf / float64(allN)
	}
	if receipt.AverageConfidence < 0.75 {
		receipt.Warnings = append(receipt.Warnings, "OCR confidence is below 75%. Check material wording against the exact image regions.")
	}
	if err := ValidateOCRReceipt(receipt); err != nil {
		return OCRReceipt{}, nil, err
	}
	if err := validateOCRSegments(receipt, segments); err != nil {
		return OCRReceipt{}, nil, err
	}
	return receipt, segments, nil
}

func ValidateOCRReceipt(r OCRReceipt) error {
	if strings.TrimSpace(r.Engine) == "" || len([]rune(r.Engine)) > maxOCRIdentityText {
		return errors.New("OCR engine identity is missing or unbounded")
	}
	if len([]rune(r.EngineVersion)) > maxOCRIdentityText || len([]rune(r.Language)) > maxOCRLanguageText {
		return errors.New("OCR engine version or language is unbounded")
	}
	if !sha256TextPattern.MatchString(r.SourceSHA256) {
		return errors.New("OCR source SHA-256 is missing or invalid")
	}
	if r.Status != "ready" && r.Status != "no-text" && r.Status != "failed" {
		return errors.New("unsupported OCR status")
	}
	if r.CreatedAt.IsZero() {
		return errors.New("OCR receipt creation time is missing")
	}
	if math.IsNaN(r.AverageConfidence) || math.IsInf(r.AverageConfidence, 0) || r.AverageConfidence < 0 || r.AverageConfidence > 1 {
		return errors.New("OCR average confidence is invalid")
	}
	if len(r.Words) > maxOCRWords || len(r.Lines) > maxOCRLines || len(r.Warnings) > 100 {
		return errors.New("OCR receipt exceeds safe collection limits")
	}
	for _, warning := range r.Warnings {
		if len([]rune(warning)) > maxOCRWarningText {
			return errors.New("OCR warning text is unbounded")
		}
	}
	validateWord := func(w OCRWord) error {
		if strings.TrimSpace(w.Text) == "" || len([]rune(w.Text)) > maxOCRWordText {
			return errors.New("OCR word text is empty or unbounded")
		}
		if !w.Region.Valid() {
			return errors.New("OCR word region is invalid")
		}
		if w.Page < 1 {
			return errors.New("OCR word page is invalid")
		}
		if math.IsNaN(w.Confidence) || math.IsInf(w.Confidence, 0) || w.Confidence < 0 || w.Confidence > 1 {
			return errors.New("OCR word confidence is invalid")
		}
		return nil
	}
	for _, w := range r.Words {
		if err := validateWord(w); err != nil {
			return err
		}
	}
	nestedWords := 0
	for _, line := range r.Lines {
		if strings.TrimSpace(line.Text) == "" || len([]rune(line.Text)) > maxOCRLineText {
			return errors.New("OCR line text is empty or unbounded")
		}
		if !line.Region.Valid() || line.Page < 1 || math.IsNaN(line.Confidence) || math.IsInf(line.Confidence, 0) || line.Confidence < 0 || line.Confidence > 1 {
			return errors.New("OCR line coordinates, page or confidence are invalid")
		}
		nestedWords += len(line.Words)
		if nestedWords > maxOCRWords {
			return errors.New("OCR nested word collection exceeds the safe limit")
		}
		for _, w := range line.Words {
			if err := validateWord(w); err != nil {
				return err
			}
			if w.Page != line.Page {
				return errors.New("OCR word page does not match its line")
			}
		}
	}
	if r.Status == "ready" && len(r.Lines) == 0 {
		return errors.New("ready OCR receipt has no lines")
	}
	if r.Status != "ready" && (len(r.Words) != 0 || len(r.Lines) != 0) {
		return errors.New("non-ready OCR receipt contains recognised text")
	}
	return nil
}

func validateOCRSegments(receipt OCRReceipt, segments []SourceSegment) error {
	if receipt.Status != "ready" {
		if len(segments) != 0 {
			return errors.New("non-ready OCR receipt cannot add source segments")
		}
		return nil
	}
	if len(segments) != len(receipt.Lines) || len(segments) > maxOCRLines {
		return errors.New("OCR source segments do not match the validated receipt")
	}
	seen := make(map[string]struct{}, len(segments))
	for i, seg := range segments {
		if strings.TrimSpace(seg.ID) == "" || len([]rune(seg.ID)) > maxOCRIdentityText {
			return errors.New("OCR source segment ID is missing or unbounded")
		}
		if _, exists := seen[seg.ID]; exists {
			return errors.New("OCR source segment IDs are not unique")
		}
		seen[seg.ID] = struct{}{}
		if seg.Ordinal < 1 || seg.Page < 1 || seg.Origin != "ocr" {
			return errors.New("OCR source segment metadata is invalid")
		}
		if seg.Region == nil || !seg.Region.Valid() || strings.TrimSpace(seg.Text) == "" || len([]rune(seg.Text)) > maxOCRLineText {
			return errors.New("OCR source segment is missing bounded text coordinates")
		}
		if math.IsNaN(seg.Confidence) || math.IsInf(seg.Confidence, 0) || seg.Confidence < 0 || seg.Confidence > 1 {
			return errors.New("OCR source segment confidence is invalid")
		}
		line := receipt.Lines[i]
		if seg.Text != line.Text || seg.Page != line.Page || math.Abs(seg.Confidence-line.Confidence) > 0.000001 || !regionsEqual(*seg.Region, line.Region) {
			return errors.New("OCR source segment diverges from its validated receipt line")
		}
	}
	return nil
}

func regionsEqual(a, b NormalizedRegion) bool {
	return math.Abs(a.X-b.X) <= 0.000001 && math.Abs(a.Y-b.Y) <= 0.000001 && math.Abs(a.Width-b.Width) <= 0.000001 && math.Abs(a.Height-b.Height) <= 0.000001
}
