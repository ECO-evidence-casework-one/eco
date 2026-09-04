#!/usr/bin/env bash
set -euo pipefail

DONOR_COMMIT='b3c860c2375335b0bc6676c430107a553725991d'
DONOR_REPO='https://github.com/ledongthuc/pdf.git'
DONOR_DIR="${RUNNER_TEMP:-/tmp}/ledongthuc-pdf"
COMPAT_DIR="${RUNNER_TEMP:-/tmp}/ledongthuc-pdf-go123"
TARGET='third_party/ledongthuc_pdf'

rm -rf "$DONOR_DIR" "$COMPAT_DIR" "$TARGET"
git clone --filter=blob:none --no-checkout "$DONOR_REPO" "$DONOR_DIR"
git -C "$DONOR_DIR" fetch --depth 1 origin "$DONOR_COMMIT"
git -C "$DONOR_DIR" checkout --detach "$DONOR_COMMIT"
test "$(git -C "$DONOR_DIR" rev-parse HEAD)" = "$DONOR_COMMIT"
grep -q 'Redistribution and use in source and binary forms' "$DONOR_DIR/LICENSE"

cp -a "$DONOR_DIR" "$COMPAT_DIR"
rm -rf "$COMPAT_DIR/.git"
sed -i 's/^go 1\.24\.1$/go 1.23/' "$COMPAT_DIR/go.mod"
(
  cd "$COMPAT_DIR"
  go test ./...
)

mkdir -p "$TARGET"
for f in ascii85.go lex.go name.go page.go ps.go read.go text.go LICENSE README.md; do
  cp "$DONOR_DIR/$f" "$TARGET/$f"
done
cat > "$TARGET/go.mod" <<'EOF'
module github.com/ledongthuc/pdf

go 1.23
EOF
cat > "$TARGET/ECO_PROVENANCE.md" <<'EOF'
# ECO vendored PDF reader provenance

Upstream: `ledongthuc/pdf`
Upstream commit: `b3c860c2375335b0bc6676c430107a553725991d`
Upstream licence: BSD-3-Clause (`LICENSE` retained verbatim)
Upstream module directive at this commit: `go 1.24.1`
ECO compatibility adjustment: local vendored module declares `go 1.23` only; source files are otherwise copied from the exact upstream commit.

Qualification before integration proved the full upstream test suite and a known-text PDF extraction smoke test under Go 1.23.12. ECO's integration adds its own bounds, panic containment, per-page warnings and page-aware SourceSegment metadata around the donor parser.
EOF

go mod edit -require=github.com/ledongthuc/pdf@v0.0.0-20260903153007-b3c860c23753
go mod edit -replace=github.com/ledongthuc/pdf=./third_party/ledongthuc_pdf

python3 - <<'PY'
from pathlib import Path
p = Path('internal/eco/extract.go')
s = p.read_text(encoding='utf-8')
import_anchor='\t"unicode/utf8"\n)'
if import_anchor not in s:
    raise SystemExit('extract.go import anchor not found')
s = s.replace(import_anchor, '\t"unicode/utf8"\n\n\tpdf "github.com/ledongthuc/pdf"\n)', 1)
switch_anchor='\tcase "zip":\n\t\ttext, err = inspectZIP(path)\n\tdefault:'
if switch_anchor not in s:
    raise SystemExit('extract.go switch anchor not found')
s = s.replace(switch_anchor, '\tcase "zip":\n\t\ttext, err = inspectZIP(path)\n\tcase "pdf":\n\t\treturn extractPDFReadable(path)\n\tdefault:', 1)
func_anchor='func extractPlainFamily(path, typ string) (string, error) {'
if func_anchor not in s:
    raise SystemExit('extract.go function anchor not found')
pdf_func=r'''const (
\tmaxNativePDFPages      = 10000
\tmaxNativePDFSegments   = 20000
\tmaxNativePDFInputBytes = int64(512 * 1024 * 1024)
)

func extractPDFReadable(path string) (text string, segments []SourceSegment, warnings []string) {
\tdefer func() {
\t\tif recover() != nil {
\t\t\ttext = ""
\t\t\tsegments = nil
\t\t\twarnings = []string{"The original PDF was preserved, but native PDF parsing stopped safely after an unexpected parser failure."}
\t\t}
\t}()

\tinfo, err := os.Stat(path)
\tif err != nil {
\t\treturn "", nil, []string{"The original PDF was preserved, but native PDF extraction could not open the reading copy: " + err.Error()}
\t}
\tif !info.Mode().IsRegular() {
\t\treturn "", nil, []string{"The original PDF was preserved, but native PDF extraction requires a regular file reading copy."}
\t}
\tif info.Size() > maxNativePDFInputBytes {
\t\treturn "", nil, []string{"The original PDF was preserved, but native PDF extraction was skipped because the file exceeds ECO's 512 MiB parser safety bound."}
\t}

\tf, reader, err := pdf.Open(path)
\tif err != nil {
\t\treturn "", nil, []string{"The original PDF was preserved, but native PDF extraction could not open its structure: " + boundPDFExtractDiagnostic(err.Error())}
\t}
\tdefer f.Close()

\tpages := reader.NumPage()
\tif pages <= 0 {
\t\treturn "", nil, []string{"The original PDF was preserved, but no readable PDF pages were declared."}
\t}
\tif pages > maxNativePDFPages {
\t\treturn "", nil, []string{fmt.Sprintf("The original PDF was preserved, but native PDF extraction was skipped because it declares %d pages (limit %d).", pages, maxNativePDFPages)}
\t}

\tvar out strings.Builder
\tordinal := 1
\tbounded := false
\tfor pageIndex := 1; pageIndex <= pages; pageIndex++ {
\t\tpage := reader.Page(pageIndex)
\t\tpageText, pageErr := page.GetPlainText(nil)
\t\tif pageErr != nil {
\t\t\twarnings = append(warnings, fmt.Sprintf("PDF page %d native text extraction did not complete: %s", pageIndex, boundPDFExtractDiagnostic(pageErr.Error())))
\t\t\tcontinue
\t\t}
\t\tpageText = normalizeText(pageText)
\t\tif pageText == "" {
\t\t\tcontinue
\t\t}

\t\theader := fmt.Sprintf("Page %d\n", pageIndex)
\t\tseparator := ""
\t\tif out.Len() > 0 {
\t\t\tseparator = "\n\n"
\t\t}
\t\tremaining := int(maxExtractBytes) - out.Len() - len(separator) - len(header)
\t\tif remaining <= 0 {
\t\t\tbounded = true
\t\t\tbreak
\t\t}
\t\tif len(pageText) > remaining {
\t\t\tpageText = strings.ToValidUTF8(pageText[:remaining], "�")
\t\t\tbounded = true
\t\t}
\t\tout.WriteString(separator)
\t\tout.WriteString(header)
\t\tout.WriteString(pageText)

\t\tfor _, seg := range segmentText(pageText) {
\t\t\tif len(segments) >= maxNativePDFSegments {
\t\t\t\tbounded = true
\t\t\t\tbreak
\t\t\t}
\t\t\tseg.ID = fmt.Sprintf("SEG-PDF-%04d", ordinal)
\t\t\tseg.Ordinal = ordinal
\t\t\tseg.Page = pageIndex
\t\t\tseg.PageHint = fmt.Sprintf("Page %d", pageIndex)
\t\t\tseg.Origin = "pdf-native"
\t\t\tsegments = append(segments, seg)
\t\t\tordinal++
\t\t}
\t\tif bounded {
\t\t\tbreak
\t\t}
\t}
\tif bounded {
\t\twarnings = append(warnings, "Native PDF text extraction reached ECO's bounded text/segment limit; the preserved original remains authoritative.")
\t}
\ttext = normalizeText(out.String())
\tif text == "" && len(warnings) == 0 {
\t\twarnings = append(warnings, "The PDF contains no extractable native text. A registered local OCR path may be required for scanned pages.")
\t}
\treturn text, segments, warnings
}

func boundPDFExtractDiagnostic(text string) string {
\ttext = strings.TrimSpace(text)
\trunes := []rune(text)
\tif len(runes) > 512 {
\t\treturn string(runes[:512]) + "…"
\t}
\treturn text
}

'''
pdf_func = pdf_func.replace(r'\t', '\t')
s = s.replace(func_anchor, pdf_func + func_anchor, 1)
p.write_text(s, encoding='utf-8')
PY

cat > internal/eco/pdf_extract_test.go <<'EOF'
package eco

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractReadablePDFNativeText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "native.pdf")
	if err := os.WriteFile(path, makeNativeTextPDF("ECO PDF TEXT 123"), 0600); err != nil {
		t.Fatal(err)
	}
	text, segments, warnings := ExtractReadable(path, "pdf")
	if !strings.Contains(text, "ECO PDF TEXT 123") {
		t.Fatalf("missing known PDF text: %q", text)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(segments) == 0 {
		t.Fatal("expected page-aware PDF segments")
	}
	if segments[0].Page != 1 || segments[0].PageHint != "Page 1" || segments[0].Origin != "pdf-native" {
		t.Fatalf("unexpected PDF segment provenance: %+v", segments[0])
	}
}

func TestExtractReadableMalformedPDFFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog"), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, warnings := ExtractReadable(path, "pdf")
	if len(warnings) == 0 {
		t.Fatal("malformed PDF should produce a bounded warning")
	}
}

func makeNativeTextPDF(text string) []byte {
	text = strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)").Replace(text)
	stream := []byte("BT /F1 24 Tf 72 700 Td (" + text + ") Tj ET\n")
	objects := [][]byte{
		[]byte("<< /Type /Catalog /Pages 2 0 R >>"),
		[]byte("<< /Type /Pages /Kids [3 0 R] /Count 1 >>"),
		[]byte("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>"),
		[]byte(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream), stream)),
		[]byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"),
	}
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, object := range objects {
		offsets[i+1] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n", i+1)
		b.Write(object)
		b.WriteString("\nendobj\n")
	}
	xref := b.Len()
	b.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	return b.Bytes()
}
EOF

python3 - <<'PY'
from pathlib import Path
p = Path('THIRD_PARTY_NOTICES.md')
s = p.read_text(encoding='utf-8')
marker = '## ledongthuc/pdf — native PDF reader'
if marker not in s:
    s += '''\n\n## ledongthuc/pdf — native PDF reader\n\n- Upstream: `https://github.com/ledongthuc/pdf`\n- Exact source commit: `b3c860c2375335b0bc6676c430107a553725991d`\n- Licence: BSD-3-Clause; verbatim licence retained at `third_party/ledongthuc_pdf/LICENSE`.\n- ECO compatibility change: the vendored local module declares Go 1.23 instead of upstream's Go 1.24.1 module directive; qualified source files are otherwise copied from the exact upstream commit.\n- Purpose: bounded, page-aware native-text extraction from PDFs. Scanned/image-only PDFs still require a separately registered local OCR path.\n'''
    p.write_text(s, encoding='utf-8')
PY

mkdir -p docs/foss
cat > docs/foss/NATIVE_PDF_READER_INTEGRATION.md <<'EOF'
# Native PDF reader integration

Date: 2026-09-04

ECO vendors `ledongthuc/pdf` at exact commit `b3c860c2375335b0bc6676c430107a553725991d` under BSD-3-Clause to provide a small Go-native path for text-bearing PDFs.

The current upstream commit declares Go 1.24.1, but prior qualification proved the complete upstream test suite and known-text extraction under Go 1.23.12 after changing only that module metadata directive. ECO retains the exact source files and records the local Go 1.23 directive in the vendored module.

ECO wraps the donor with additional controls: regular-file and 512 MiB input bounds, 10,000-page and 20,000-segment bounds, panic containment, bounded diagnostics, page-level extraction, per-page warnings, and `Page` / `PageHint` / `Origin=pdf-native` segment provenance. A PDF with no native text is not falsely declared readable; the caller is told that registered local OCR may be required.

This integration deliberately reduces Docling from a critical native-PDF dependency. Docling source remains acquired for later advanced layout/table use, but its standard model pipeline uses separate Hugging Face model assets and is not required for baseline native PDF text extraction.
EOF

gofmt -w internal/eco/extract.go internal/eco/pdf_extract_test.go
go mod tidy
go test ./...
go vet ./...
python3 scripts/check_source_policy.py
git diff --check

git config user.name 'ECO GitHub integration bot'
git config user.email 'actions@users.noreply.github.com'
git add go.mod go.sum internal/eco/extract.go internal/eco/pdf_extract_test.go third_party/ledongthuc_pdf THIRD_PARTY_NOTICES.md docs/foss/NATIVE_PDF_READER_INTEGRATION.md
if git diff --cached --quiet; then
  echo 'No integration changes to commit.'
  exit 0
fi
git commit -m 'Integrate qualified native PDF text reader [pdf-reader-vendor]'
git push origin HEAD:integration/native-pdf-reader-20260904
