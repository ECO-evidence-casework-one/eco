package eco

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
