package eco

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	pdf "github.com/ledongthuc/pdf"
)

const maxExtractBytes = int64(24 * 1024 * 1024)

var (
	tagRE        = regexp.MustCompile(`(?s)<[^>]+>`)
	scriptRE     = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>`)
	rtfControlRE = regexp.MustCompile(`\\[a-zA-Z]+-?\d* ?|[{}]`)
	spaceRE      = regexp.MustCompile(`[ \t\x0b\f]+`)
	newlineRE    = regexp.MustCompile(`\n{3,}`)
)

func ExtractReadable(path, typ string) (string, []SourceSegment, []string) {
	warnings := []string{}
	var text string
	var err error
	switch typ {
	case "text", "markdown", "csv", "json", "xml", "html", "rtf":
		text, err = extractPlainFamily(path, typ)
	case "docx":
		text, err = extractDOCX(path)
	case "xlsx":
		text, err = extractXLSX(path)
	case "pptx":
		text, err = extractPPTX(path)
	case "odt", "ods", "odp":
		text, err = extractOpenDocument(path)
	case "eml":
		text, err = extractEML(path)
	case "zip":
		text, err = inspectZIP(path)
	case "pdf":
		return extractPDFReadable(path)
	default:
		return "", nil, warnings
	}
	if err != nil {
		warnings = append(warnings, "The original was preserved, but safe text extraction did not complete: "+err.Error())
		return "", nil, warnings
	}
	text = normalizeText(text)
	if len(text) > int(maxExtractBytes) {
		text = text[:maxExtractBytes]
		warnings = append(warnings, "Extracted text was bounded to 24 MiB for this preview.")
	}
	return text, segmentText(text), warnings
}

const (
	maxNativePDFPages      = 10000
	maxNativePDFSegments   = 20000
	maxNativePDFInputBytes = int64(512 * 1024 * 1024)
)

func extractPDFReadable(path string) (text string, segments []SourceSegment, warnings []string) {
	defer func() {
		if recover() != nil {
			text = ""
			segments = nil
			warnings = []string{"The original PDF was preserved, but native PDF parsing stopped safely after an unexpected parser failure."}
		}
	}()

	info, err := os.Stat(path)
	if err != nil {
		return "", nil, []string{"The original PDF was preserved, but native PDF extraction could not open the reading copy: " + err.Error()}
	}
	if !info.Mode().IsRegular() {
		return "", nil, []string{"The original PDF was preserved, but native PDF extraction requires a regular file reading copy."}
	}
	if info.Size() > maxNativePDFInputBytes {
		return "", nil, []string{"The original PDF was preserved, but native PDF extraction was skipped because the file exceeds ECO's 512 MiB parser safety bound."}
	}

	f, reader, err := pdf.Open(path)
	if err != nil {
		return "", nil, []string{"The original PDF was preserved, but native PDF extraction could not open its structure: " + boundPDFExtractDiagnostic(err.Error())}
	}
	defer f.Close()

	pages := reader.NumPage()
	if pages <= 0 {
		return "", nil, []string{"The original PDF was preserved, but no readable PDF pages were declared."}
	}
	if pages > maxNativePDFPages {
		return "", nil, []string{fmt.Sprintf("The original PDF was preserved, but native PDF extraction was skipped because it declares %d pages (limit %d).", pages, maxNativePDFPages)}
	}

	var out strings.Builder
	ordinal := 1
	bounded := false
	for pageIndex := 1; pageIndex <= pages; pageIndex++ {
		page := reader.Page(pageIndex)
		pageText, pageErr := page.GetPlainText(nil)
		if pageErr != nil {
			warnings = append(warnings, fmt.Sprintf("PDF page %d native text extraction did not complete: %s", pageIndex, boundPDFExtractDiagnostic(pageErr.Error())))
			continue
		}
		pageText = normalizeText(pageText)
		if pageText == "" {
			continue
		}

		header := fmt.Sprintf("Page %d\n", pageIndex)
		separator := ""
		if out.Len() > 0 {
			separator = "\n\n"
		}
		remaining := int(maxExtractBytes) - out.Len() - len(separator) - len(header)
		if remaining <= 0 {
			bounded = true
			break
		}
		if len(pageText) > remaining {
			pageText = strings.ToValidUTF8(pageText[:remaining], "�")
			bounded = true
		}
		out.WriteString(separator)
		out.WriteString(header)
		out.WriteString(pageText)

		for _, seg := range segmentText(pageText) {
			if len(segments) >= maxNativePDFSegments {
				bounded = true
				break
			}
			seg.ID = fmt.Sprintf("SEG-PDF-%04d", ordinal)
			seg.Ordinal = ordinal
			seg.Page = pageIndex
			seg.PageHint = fmt.Sprintf("Page %d", pageIndex)
			seg.Origin = "pdf-native"
			segments = append(segments, seg)
			ordinal++
		}
		if bounded {
			break
		}
	}
	if bounded {
		warnings = append(warnings, "Native PDF text extraction reached ECO's bounded text/segment limit; the preserved original remains authoritative.")
	}
	text = normalizeText(out.String())
	if text == "" && len(warnings) == 0 {
		warnings = append(warnings, "The PDF contains no extractable native text. A registered local OCR path may be required for scanned pages.")
	}
	return text, segments, warnings
}

func boundPDFExtractDiagnostic(text string) string {
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) > 512 {
		return string(runes[:512]) + "…"
	}
	return text
}

func extractPlainFamily(path, typ string) (string, error) {
	data, err := readFileBounded(path, maxExtractBytes)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(data) {
		data = []byte(strings.ToValidUTF8(string(data), "�"))
	}
	s := string(data)
	switch typ {
	case "html":
		s = scriptRE.ReplaceAllString(s, " ")
		s = tagRE.ReplaceAllString(s, "\n")
		s = html.UnescapeString(s)
	case "rtf":
		s = strings.ReplaceAll(s, "\\par", "\n")
		s = rtfControlRE.ReplaceAllString(s, " ")
	case "json":
		var v any
		if json.Unmarshal(data, &v) == nil {
			b, _ := json.MarshalIndent(v, "", "  ")
			s = string(b)
		}
	case "csv":
		r := csv.NewReader(strings.NewReader(s))
		r.FieldsPerRecord = -1
		var b strings.Builder
		for rows := 0; rows < 50000; rows++ {
			rec, e := r.Read()
			if e == io.EOF {
				break
			}
			if e != nil {
				return s, nil
			}
			b.WriteString(strings.Join(rec, " | "))
			b.WriteByte('\n')
		}
		s = b.String()
	}
	return s, nil
}

func zipRead(path, name string, limit int64) ([]byte, error) {
	if err := zipPreflight(path, 10000); err != nil {
		return nil, err
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if strings.EqualFold(strings.ReplaceAll(f.Name, "\\", "/"), name) {
			if unsafeCompressionRatio(f) {
				return nil, fmt.Errorf("archive entry has an unsafe compression ratio")
			}
			if int64(f.UncompressedSize64) > limit {
				return nil, fmt.Errorf("archive entry too large")
			}
			r, e := f.Open()
			if e != nil {
				return nil, e
			}
			defer r.Close()
			return io.ReadAll(io.LimitReader(r, limit+1))
		}
	}
	return nil, os.ErrNotExist
}

func xmlText(data []byte, paragraphElements map[string]bool) string {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var b strings.Builder
	depth := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if paragraphElements[t.Name.Local] && b.Len() > 0 {
				b.WriteByte('\n')
			}
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s != "" {
				if b.Len() > 0 && lastByte(&b) != '\n' {
					b.WriteByte(' ')
				}
				b.WriteString(s)
			}
		case xml.EndElement:
			if paragraphElements[t.Name.Local] {
				b.WriteByte('\n')
			}
			depth--
		}
		_ = depth
	}
	return b.String()
}

func lastByte(b *strings.Builder) byte {
	s := b.String()
	if len(s) == 0 {
		return 0
	}
	return s[len(s)-1]
}

func extractDOCX(path string) (string, error) {
	data, err := zipRead(path, "word/document.xml", maxExtractBytes)
	if err != nil {
		return "", err
	}
	return xmlText(data, map[string]bool{"p": true, "tr": true, "tbl": true}), nil
}

func extractPPTX(path string) (string, error) {
	if err := zipPreflight(path, 10000); err != nil {
		return "", err
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	type slide struct {
		name string
		data []byte
	}
	slides := []slide{}
	for _, f := range zr.File {
		n := strings.ToLower(strings.ReplaceAll(f.Name, "\\", "/"))
		if strings.HasPrefix(n, "ppt/slides/slide") && strings.HasSuffix(n, ".xml") {
			if unsafeCompressionRatio(f) {
				continue
			}
			if f.UncompressedSize64 > uint64(maxExtractBytes) {
				continue
			}
			r, e := f.Open()
			if e != nil {
				continue
			}
			d, _ := io.ReadAll(io.LimitReader(r, maxExtractBytes))
			r.Close()
			slides = append(slides, slide{n, d})
		}
	}
	sort.Slice(slides, func(i, j int) bool { return naturalLess(slides[i].name, slides[j].name) })
	var b strings.Builder
	for i, s := range slides {
		fmt.Fprintf(&b, "Slide %d\n", i+1)
		b.WriteString(xmlText(s.data, map[string]bool{"p": true}))
		b.WriteString("\n\n")
	}
	return b.String(), nil
}

func naturalLess(a, b string) bool { return extractNumber(a) < extractNumber(b) }
func extractNumber(s string) int {
	n := 0
	seen := false
	for _, r := range s {
		if unicode.IsDigit(r) {
			seen = true
			n = n*10 + int(r-'0')
		} else if seen {
			break
		}
	}
	return n
}

func extractXLSX(path string) (string, error) {
	if err := zipPreflight(path, 10000); err != nil {
		return "", err
	}
	shared := []string{}
	if d, err := zipRead(path, "xl/sharedStrings.xml", maxExtractBytes); err == nil {
		dec := xml.NewDecoder(bytes.NewReader(d))
		var cur strings.Builder
		inSI := false
		for {
			tok, e := dec.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				break
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "si" {
					inSI = true
					cur.Reset()
				}
			case xml.CharData:
				if inSI {
					cur.Write([]byte(t))
				}
			case xml.EndElement:
				if t.Name.Local == "si" {
					shared = append(shared, strings.TrimSpace(cur.String()))
					inSI = false
				}
			}
		}
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	type sh struct {
		name string
		data []byte
	}
	sheets := []sh{}
	for _, f := range zr.File {
		n := strings.ToLower(strings.ReplaceAll(f.Name, "\\", "/"))
		if strings.HasPrefix(n, "xl/worksheets/sheet") && strings.HasSuffix(n, ".xml") {
			if unsafeCompressionRatio(f) {
				continue
			}
			if f.UncompressedSize64 > uint64(maxExtractBytes) {
				continue
			}
			r, e := f.Open()
			if e != nil {
				continue
			}
			d, _ := io.ReadAll(io.LimitReader(r, maxExtractBytes))
			r.Close()
			sheets = append(sheets, sh{n, d})
		}
	}
	sort.Slice(sheets, func(i, j int) bool { return naturalLess(sheets[i].name, sheets[j].name) })
	var out strings.Builder
	for si, s := range sheets {
		fmt.Fprintf(&out, "Sheet %d\n", si+1)
		dec := xml.NewDecoder(bytes.NewReader(s.data))
		cellType := ""
		cellRef := ""
		inV := false
		var value strings.Builder
		for {
			tok, e := dec.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				break
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "c" {
					cellType = ""
					cellRef = ""
					for _, a := range t.Attr {
						if a.Name.Local == "t" {
							cellType = a.Value
						}
						if a.Name.Local == "r" {
							cellRef = a.Value
						}
					}
				}
				if t.Name.Local == "v" || t.Name.Local == "t" {
					inV = true
					value.Reset()
				}
			case xml.CharData:
				if inV {
					value.Write([]byte(t))
				}
			case xml.EndElement:
				if t.Name.Local == "v" || t.Name.Local == "t" {
					v := strings.TrimSpace(value.String())
					if cellType == "s" {
						idx, _ := strconv.Atoi(v)
						if idx >= 0 && idx < len(shared) {
							v = shared[idx]
						}
					}
					if v != "" {
						if cellRef != "" {
							out.WriteString(cellRef + ": ")
						}
						out.WriteString(v)
						out.WriteByte('\n')
					}
					inV = false
				}
			}
		}
		out.WriteByte('\n')
	}
	return out.String(), nil
}

func extractOpenDocument(path string) (string, error) {
	d, err := zipRead(path, "content.xml", maxExtractBytes)
	if err != nil {
		return "", err
	}
	return xmlText(d, map[string]bool{"p": true, "h": true, "table-row": true}), nil
}

func extractEML(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	msg, err := mail.ReadMessage(bufio.NewReader(io.LimitReader(f, maxExtractBytes)))
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, h := range []string{"From", "To", "Cc", "Date", "Subject", "Message-Id"} {
		if v := msg.Header.Get(h); v != "" {
			fmt.Fprintf(&b, "%s: %s\n", h, v)
		}
	}
	b.WriteByte('\n')
	ct := msg.Header.Get("Content-Type")
	med, params, _ := mime.ParseMediaType(ct)
	if strings.HasPrefix(med, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			p, e := mr.NextPart()
			if e == io.EOF {
				break
			}
			if e != nil {
				break
			}
			pct, _, _ := mime.ParseMediaType(p.Header.Get("Content-Type"))
			if strings.HasPrefix(pct, "text/plain") || strings.HasPrefix(pct, "text/html") {
				d, _ := io.ReadAll(io.LimitReader(p, maxExtractBytes))
				s := string(d)
				if strings.Contains(pct, "html") {
					s = tagRE.ReplaceAllString(scriptRE.ReplaceAllString(s, " "), "\n")
				}
				b.WriteString(s)
				b.WriteByte('\n')
			}
		}
	} else {
		d, _ := io.ReadAll(io.LimitReader(msg.Body, maxExtractBytes))
		b.Write(d)
	}
	return b.String(), nil
}

func inspectZIP(path string) (string, error) {
	if err := zipPreflight(path, 10000); err != nil {
		return "", err
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	if len(zr.File) > 10000 {
		return "", fmt.Errorf("archive has more than 10,000 entries")
	}
	var total uint64
	var b strings.Builder
	fmt.Fprintf(&b, "ZIP archive: %d entries\n", len(zr.File))
	for i, f := range zr.File {
		if unsafeCompressionRatio(f) {
			return "", fmt.Errorf("archive entry %q has an unsafe compression ratio", f.Name)
		}
		total += f.UncompressedSize64
		if total > 2*1024*1024*1024 {
			return "", fmt.Errorf("archive expands beyond 2 GiB safety limit")
		}
		n := strings.ReplaceAll(f.Name, "\\", "/")
		clean := filepath.Clean(n)
		if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
			fmt.Fprintf(&b, "[unsafe path] %s\n", n)
			continue
		}
		if i < 5000 {
			fmt.Fprintf(&b, "%s | %s\n", n, HumanBytes(int64(f.UncompressedSize64)))
		}
	}
	if len(zr.File) > 5000 {
		b.WriteString("Entry list bounded to first 5,000 items.\n")
	}
	return b.String(), nil
}

func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = spaceRE.ReplaceAllString(s, " ")
	s = newlineRE.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

func segmentText(text string) []SourceSegment {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	paras := strings.Split(text, "\n")
	out := []SourceSegment{}
	var b strings.Builder
	ord := 1
	flush := func() {
		s := strings.TrimSpace(b.String())
		if s != "" {
			if len(s) > 2200 {
				s = s[:2200]
			}
			out = append(out, SourceSegment{ID: fmt.Sprintf("SEG-%04d", ord), Ordinal: ord, Text: s})
			ord++
		}
		b.Reset()
	}
	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			flush()
			continue
		}
		if b.Len()+len(p) > 1800 {
			flush()
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(p)
	}
	flush()
	return out
}
