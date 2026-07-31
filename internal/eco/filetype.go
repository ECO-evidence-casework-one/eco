package eco

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Detection struct {
	Type          string
	ExtensionType string
	Mismatch      bool
	Dangerous     bool
	Warning       string
}

func DetectFile(path string) (Detection, error) {
	f, err := os.Open(path)
	if err != nil {
		return Detection{}, err
	}
	defer f.Close()
	buf := make([]byte, 4096)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return Detection{}, err
	}
	buf = buf[:n]
	ext := strings.ToLower(filepath.Ext(path))
	extType := typeFromExtension(ext)
	detected := detectBytes(buf)
	if detected == "zip" {
		if zt := detectZipSubtype(path); zt != "" {
			detected = zt
		}
	}
	if detected == "text" && extType != "unknown" && isTextFamily(extType) {
		detected = extType
	}
	if detected == "unknown" && extType != "unknown" && isTextFamily(extType) && looksText(buf) {
		detected = extType
	}
	dangerous := isDangerousType(detected) || isDangerousExtension(ext)
	mismatch := extType != "unknown" && detected != "unknown" && !compatibleTypes(extType, detected)
	warning := ""
	if dangerous {
		warning = "Executable, script, shortcut or program-like content is preserved but quarantined from automatic reading."
	} else if mismatch {
		warning = "The filename extension does not match the detected file content."
	}
	return Detection{Type: detected, ExtensionType: extType, Mismatch: mismatch, Dangerous: dangerous, Warning: warning}, nil
}

func detectBytes(b []byte) string {
	if len(b) >= 2 && b[0] == 'M' && b[1] == 'Z' {
		return "windows-executable"
	}
	if len(b) >= 4 && bytes.Equal(b[:4], []byte("%PDF")) {
		return "pdf"
	}
	if len(b) >= 8 && bytes.Equal(b[:8], []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}) {
		return "png"
	}
	if len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff {
		return "jpeg"
	}
	if len(b) >= 6 && (string(b[:6]) == "GIF87a" || string(b[:6]) == "GIF89a") {
		return "gif"
	}
	if len(b) >= 2 && string(b[:2]) == "BM" {
		return "bmp"
	}
	if len(b) >= 4 && (binary.LittleEndian.Uint32(b[:4]) == 0x002a4949 || binary.BigEndian.Uint32(b[:4]) == 0x4d4d002a) {
		return "tiff"
	}
	if len(b) >= 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WEBP" {
		return "webp"
	}
	if len(b) >= 4 && bytes.Equal(b[:4], []byte{'P', 'K', 0x03, 0x04}) {
		return "zip"
	}
	if len(b) >= 6 && bytes.Equal(b[:6], []byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}) {
		return "7z"
	}
	if len(b) >= 4 && bytes.Equal(b[:4], []byte{'R', 'a', 'r', '!'}) {
		return "rar"
	}
	if len(b) >= 8 && bytes.Equal(b[:8], []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}) {
		return "ole-compound"
	}
	if len(b) >= 4 && bytes.Equal(b[:4], []byte("{\\rt")) {
		return "rtf"
	}
	if len(b) >= 5 && bytes.Equal(bytes.ToLower(b[:5]), []byte("<?xml")) {
		return "xml"
	}
	if len(b) >= 5 && bytes.Equal(bytes.ToLower(bytes.TrimSpace(b)[:min(5, len(bytes.TrimSpace(b)))]), []byte("<html")) {
		return "html"
	}
	if len(b) >= 4 && bytes.Equal(b[:4], []byte("OggS")) {
		return "ogg"
	}
	if len(b) >= 12 && string(b[4:8]) == "ftyp" {
		return "mp4"
	}
	if len(b) >= 3 && string(b[:3]) == "ID3" {
		return "mp3"
	}
	if len(b) >= 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WAVE" {
		return "wav"
	}
	if looksText(b) {
		return "text"
	}
	return "unknown"
}

func detectZipSubtype(path string) string {
	if err := zipPreflight(path, 10000); err != nil {
		return "zip"
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "zip"
	}
	defer zr.Close()
	names := make(map[string]bool, len(zr.File))
	for _, f := range zr.File {
		n := strings.ToLower(strings.ReplaceAll(f.Name, "\\", "/"))
		names[n] = true
		if len(names) > 10000 {
			return "zip"
		}
	}
	switch {
	case names["word/document.xml"]:
		return "docx"
	case names["xl/workbook.xml"]:
		return "xlsx"
	case names["ppt/presentation.xml"]:
		return "pptx"
	case names["content.xml"] && names["mimetype"]:
		for _, f := range zr.File {
			if strings.ToLower(f.Name) == "mimetype" && f.UncompressedSize64 < 256 {
				r, err := f.Open()
				if err == nil {
					data, _ := io.ReadAll(io.LimitReader(r, 255))
					r.Close()
					s := string(data)
					switch {
					case strings.Contains(s, "opendocument.text"):
						return "odt"
					case strings.Contains(s, "opendocument.spreadsheet"):
						return "ods"
					case strings.Contains(s, "opendocument.presentation"):
						return "odp"
					}
				}
			}
		}
	}
	return "zip"
}

func typeFromExtension(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".jpe":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	case ".bmp":
		return "bmp"
	case ".tif", ".tiff":
		return "tiff"
	case ".webp":
		return "webp"
	case ".pdf":
		return "pdf"
	case ".txt", ".log":
		return "text"
	case ".md", ".markdown":
		return "markdown"
	case ".csv":
		return "csv"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".rtf":
		return "rtf"
	case ".docx":
		return "docx"
	case ".xlsx":
		return "xlsx"
	case ".pptx":
		return "pptx"
	case ".odt":
		return "odt"
	case ".ods":
		return "ods"
	case ".odp":
		return "odp"
	case ".eml":
		return "eml"
	case ".zip":
		return "zip"
	case ".7z":
		return "7z"
	case ".rar":
		return "rar"
	case ".exe", ".dll", ".com", ".scr", ".msi":
		return "windows-executable"
	case ".bat", ".cmd", ".ps1", ".vbs", ".js", ".jse", ".wsf", ".hta":
		return "script"
	case ".lnk", ".url":
		return "shortcut"
	case ".mp3":
		return "mp3"
	case ".wav":
		return "wav"
	case ".ogg", ".oga":
		return "ogg"
	case ".mp4", ".m4v", ".mov":
		return "mp4"
	default:
		return "unknown"
	}
}

func looksText(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	bad := 0
	for _, c := range b {
		if c == 0 {
			return false
		}
		if c < 0x09 || (c > 0x0d && c < 0x20) {
			bad++
		}
	}
	return float64(bad)/float64(len(b)) < 0.02
}

func isTextFamily(t string) bool {
	switch t {
	case "text", "markdown", "csv", "json", "xml", "html", "rtf", "eml":
		return true
	}
	return false
}

func isDangerousType(t string) bool {
	return t == "windows-executable" || t == "script" || t == "shortcut"
}
func isDangerousExtension(ext string) bool { return isDangerousType(typeFromExtension(ext)) }

func compatibleTypes(a, b string) bool {
	if a == b {
		return true
	}
	if isTextFamily(a) && isTextFamily(b) {
		return true
	}
	if a == "zip" && (b == "docx" || b == "xlsx" || b == "pptx" || b == "odt" || b == "ods" || b == "odp") {
		return false
	}
	if b == "zip" && (a == "docx" || a == "xlsx" || a == "pptx" || a == "odt" || a == "ods" || a == "odp") {
		return false
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func readPrefix(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(bufio.NewReader(f), max))
}

// zipPreflight reads the ZIP end-of-central-directory record before Go's ZIP
// reader allocates its entry table. This bounds ordinary ZIP entry counts and
// deliberately refuses ZIP64 for automatic inspection in this preview.
func zipPreflight(path string, maxEntries int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() < 22 {
		return io.ErrUnexpectedEOF
	}
	const maxEOCDSearch = int64(22 + 65535)
	readSize := info.Size()
	if readSize > maxEOCDSearch {
		readSize = maxEOCDSearch
	}
	if _, err = f.Seek(info.Size()-readSize, io.SeekStart); err != nil {
		return err
	}
	tail := make([]byte, readSize)
	if _, err = io.ReadFull(f, tail); err != nil {
		return err
	}
	sig := []byte{'P', 'K', 0x05, 0x06}
	idx := bytes.LastIndex(tail, sig)
	if idx < 0 || idx+22 > len(tail) {
		return errors.New("ZIP end record not found")
	}
	total := int(binary.LittleEndian.Uint16(tail[idx+10 : idx+12]))
	cdSize := binary.LittleEndian.Uint32(tail[idx+12 : idx+16])
	cdOffset := binary.LittleEndian.Uint32(tail[idx+16 : idx+20])
	if total == 0xffff || cdSize == 0xffffffff || cdOffset == 0xffffffff {
		return errors.New("ZIP64 automatic inspection is not enabled in this preview")
	}
	if total > maxEntries {
		return fmt.Errorf("archive has %d entries; safe automatic limit is %d", total, maxEntries)
	}
	if int64(cdOffset)+int64(cdSize) > info.Size() {
		return errors.New("ZIP central directory is outside the file")
	}
	return nil
}

func unsafeCompressionRatio(f *zip.File) bool {
	if f.UncompressedSize64 <= 1024*1024 {
		return false
	}
	if f.CompressedSize64 == 0 {
		return true
	}
	return f.UncompressedSize64/f.CompressedSize64 > 1000
}
