from pathlib import Path

ROOT = Path('.')

# Add MBOX type detection without changing binary sniffing: mbox is a text
# container whose extension identifies the format and whose inner messages are
# parsed/validated by the pinned donor plus net/mail.
p = ROOT / 'internal/eco/filetype.go'
s = p.read_text(encoding='utf-8')
s = s.replace('\tcase ".eml":\n\t\treturn "eml"\n', '\tcase ".eml":\n\t\treturn "eml"\n\tcase ".mbox", ".mbx":\n\t\treturn "mbox"\n', 1)
s = s.replace('\tcase "text", "markdown", "csv", "json", "xml", "html", "rtf", "eml":\n', '\tcase "text", "markdown", "csv", "json", "xml", "html", "rtf", "eml", "mbox":\n', 1)
p.write_text(s, encoding='utf-8')

p = ROOT / 'internal/eco/extract.go'
s = p.read_text(encoding='utf-8')
import_anchor = '\t"unicode/utf8"\n)'
if import_anchor not in s:
    raise SystemExit('extract import anchor not found')
s = s.replace(import_anchor, '\t"unicode/utf8"\n\n\tmbox "github.com/emersion/go-mbox"\n)', 1)

switch_anchor = '\tcase "eml":\n\t\ttext, err = extractEML(path)\n\tcase "zip":\n'
if switch_anchor not in s:
    raise SystemExit('extract switch anchor not found')
s = s.replace(switch_anchor, '\tcase "eml":\n\t\ttext, err = extractEML(path)\n\tcase "mbox":\n\t\ttext, err = extractMBOX(path)\n\tcase "zip":\n', 1)

start = s.index('func extractEML(path string) (string, error) {')
end = s.index('\nfunc inspectZIP(path string) (string, error) {', start)
old = s[start:end]
new = r'''const (
	maxMBOXScanBytes    = int64(256 * 1024 * 1024)
	maxMBOXMessageBytes = int64(8 * 1024 * 1024)
	maxMBOXMessages     = 10000
)

func extractEML(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return extractMailReader(io.LimitReader(f, maxExtractBytes))
}

func extractMailReader(r io.Reader) (string, error) {
	msg, err := mail.ReadMessage(bufio.NewReader(r))
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
					s = html.UnescapeString(s)
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

func extractMBOX(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	reader := mbox.NewReader(io.LimitReader(f, maxMBOXScanBytes))
	var out strings.Builder
	messages := 0
	for messages < maxMBOXMessages {
		messageReader, nextErr := reader.NextMessage()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			if messages == 0 {
				return "", nextErr
			}
			fmt.Fprintf(&out, "[Mailbox parsing stopped after message %d: %s]\n", messages, boundMailboxDiagnostic(nextErr.Error()))
			break
		}
		messages++
		raw, readErr := io.ReadAll(io.LimitReader(messageReader, maxMBOXMessageBytes+1))
		if readErr != nil {
			fmt.Fprintf(&out, "Message %d\n[Message could not be read safely: %s]\n\n", messages, boundMailboxDiagnostic(readErr.Error()))
			continue
		}
		if int64(len(raw)) > maxMBOXMessageBytes {
			fmt.Fprintf(&out, "Message %d\n[Message exceeds ECO's 8 MiB automatic-reading limit and was not decoded.]\n\n", messages)
			continue
		}
		text, parseErr := extractMailReader(bytes.NewReader(raw))
		if parseErr != nil {
			fmt.Fprintf(&out, "Message %d\n[Message headers/body could not be decoded safely: %s]\n\n", messages, boundMailboxDiagnostic(parseErr.Error()))
			continue
		}
		fmt.Fprintf(&out, "Message %d\n%s\n\n", messages, strings.TrimSpace(text))
		if out.Len() >= int(maxExtractBytes) {
			out.WriteString("[Mailbox readable preview reached ECO's 24 MiB extracted-text bound.]\n")
			break
		}
	}
	if messages >= maxMBOXMessages {
		fmt.Fprintf(&out, "[Mailbox preview reached the %d-message automatic-reading limit.]\n", maxMBOXMessages)
	}
	if info.Size() > maxMBOXScanBytes {
		fmt.Fprintf(&out, "[Mailbox is larger than ECO's %s automatic scan window; later content was not read in this preview.]\n", HumanBytes(maxMBOXScanBytes))
	}
	if messages == 0 && strings.TrimSpace(out.String()) == "" {
		return "", fmt.Errorf("mailbox contains no readable mbox messages")
	}
	return out.String(), nil
}

func boundMailboxDiagnostic(text string) string {
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) > 384 {
		return string(runes[:384]) + "…"
	}
	return text
}
'''
s = s[:start] + new + s[end:]
p.write_text(s, encoding='utf-8')

# Regression and seed-corpus fuzz target. Normal `go test` executes the seed
# corpus; dedicated qualification can run mutations for a bounded time.
test = r'''package eco

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testMBOX = `From alice@example.test Thu Jan  1 00:00:01 2026
From: Alice <alice@example.test>
To: Bob <bob@example.test>
Date: Thu, 01 Jan 2026 00:00:01 +0000
Subject: First message
Message-ID: <first@example.test>
Content-Type: text/plain; charset=utf-8

Hello Bob.
>From this body line should be unescaped by the mbox reader.

From bob@example.test Thu Jan  1 00:01:01 2026
From: Bob <bob@example.test>
To: Alice <alice@example.test>
Date: Thu, 01 Jan 2026 00:01:01 +0000
Subject: Second message
Message-ID: <second@example.test>
Content-Type: text/html; charset=utf-8

<html><body><p>Hello Alice.</p></body></html>
`

func TestDetectAndExtractMBOX(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mailbox.mbox")
	if err := os.WriteFile(path, []byte(testMBOX), 0600); err != nil {
		t.Fatal(err)
	}
	detection, err := DetectFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if detection.Type != "mbox" || detection.ExtensionType != "mbox" || detection.Mismatch {
		t.Fatalf("unexpected MBOX detection: %+v", detection)
	}
	text, segments, warnings := ExtractReadable(path, "mbox")
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	for _, want := range []string{"Message 1", "Subject: First message", "Hello Bob.", "From this body line", "Message 2", "Subject: Second message", "Hello Alice."} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in extracted mailbox text: %q", want, text)
		}
	}
	if len(segments) == 0 {
		t.Fatal("MBOX extraction did not create source segments")
	}
}

func TestMalformedMBOXFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.mbox")
	if err := os.WriteFile(path, []byte("Subject: missing mbox separator\n\nbody\n"), 0600); err != nil {
		t.Fatal(err)
	}
	text, _, warnings := ExtractReadable(path, "mbox")
	if text != "" || len(warnings) == 0 {
		t.Fatalf("malformed mailbox did not fail closed: text=%q warnings=%v", text, warnings)
	}
}

func FuzzMBOXReader(f *testing.F) {
	f.Add([]byte(testMBOX))
	f.Add([]byte("From a@example.test Thu Jan  1 00:00:01 2026\nFrom: a@example.test\nSubject: x\n\nbody\n"))
	f.Add([]byte("not an mbox"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		path := filepath.Join(t.TempDir(), "fuzz.mbox")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		text, _ := extractMBOX(path)
		if len(text) > int(maxExtractBytes)+8192 {
			t.Fatalf("MBOX parser returned unbounded text: %d bytes", len(text))
		}
	})
}
'''
(ROOT / 'internal/eco/mbox_extract_test.go').write_text(test, encoding='utf-8')

doc = '''# MBOX reader adaptation\n\nDate: 2026-09-04\n\nECO uses `emersion/go-mbox` (MIT) at exact acquired commit `1345da99f1254a23f517ffdc979f92359442473d` for MBOX message framing instead of inventing a mailbox parser. The source was acquired in FOSS donor Wave 2 with source-archive SHA-256 `b96b0ef7939de0fbe93557e7f8228f23ef484452e1909b74dc788415e7ab0566`.\n\nECO remains responsible for evidence-specific safety around the donor: the mailbox is streamed rather than loaded wholesale; automatic scanning is limited to 256 MiB, 10,000 messages and 8 MiB per individual message; extracted readable output remains under ECO's 24 MiB text bound; parser diagnostics are bounded; and malformed mailboxes fail closed while the preserved original remains authoritative.\n\nEach framed message is passed through ECO's existing `net/mail` / MIME readable-text path, so MBOX and EML share the same header/body handling rather than duplicating MIME logic. The feature is read-only and introduces no network behavior. A Go fuzz target preserves hostile-input coverage for the MBOX boundary parser, while normal tests execute its seed corpus deterministically.\n'''
(ROOT / 'docs/foss/MBOX_READER_ADAPTATION.md').write_text(doc, encoding='utf-8')

# Add the exact donor to third-party notices if not already present.
p = ROOT / 'THIRD_PARTY_NOTICES.md'
s = p.read_text(encoding='utf-8')
if 'emersion/go-mbox' not in s:
    s += '''\n\n## emersion/go-mbox — MBOX message framing\n\n- Upstream: `https://github.com/emersion/go-mbox`\n- Exact acquired commit: `1345da99f1254a23f517ffdc979f92359442473d`\n- Licence: MIT.\n- ECO use: bounded, read-only MBOX message framing before ECO's existing MIME/email extraction.\n'''
    p.write_text(s, encoding='utf-8')
