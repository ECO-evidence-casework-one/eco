package eco

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPDFCPUArgsAreOfflineReadOnlyAndConfigIndependent(t *testing.T) {
	for name, args := range map[string][]string{
		"relaxed": pdfcpuValidationArgs(`C:\work\evidence.pdf`, "relaxed"),
		"strict":  pdfcpuValidationArgs(`C:\work\evidence.pdf`, "strict"),
		"info":    pdfcpuInfoArgs(`C:\work\evidence.pdf`),
		"version": pdfcpuVersionArgs(),
	} {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "--offline") || !strings.Contains(joined, "--conf disable") {
			t.Fatalf("%s args are not forced offline/config-independent: %q", name, joined)
		}
		for _, forbidden := range []string{" optimize ", " merge ", " split ", " trim ", " rotate ", " watermark ", " stamp ", " encrypt ", " decrypt ", " create "} {
			if strings.Contains(" "+joined+" ", forbidden) {
				t.Fatalf("%s unexpectedly uses modifying pdfcpu command %q: %q", name, forbidden, joined)
			}
		}
	}
	if got := strings.Join(pdfcpuValidationArgs("x.pdf", "strict"), " "); !strings.Contains(got, "validate --mode strict") {
		t.Fatalf("strict validation args are wrong: %q", got)
	}
	if got := strings.Join(pdfcpuInfoArgs("x.pdf"), " "); !strings.Contains(got, "info --json") {
		t.Fatalf("info args are not JSON-only: %q", got)
	}
}

func TestOfflinePDFCPUEnvironmentRemovesProxyRoutes(t *testing.T) {
	env := offlinePDFCPUEnvironment([]string{
		"PATH=X",
		"HTTP_PROXY=http://proxy.invalid:8080",
		"HTTPS_PROXY=http://proxy.invalid:8080",
		"ALL_PROXY=socks5://proxy.invalid:1080",
	})
	joined := strings.Join(env, "\n")
	for _, forbidden := range []string{"HTTP_PROXY=", "HTTPS_PROXY=", "ALL_PROXY="} {
		if strings.Contains(strings.ToUpper(joined), forbidden) {
			t.Fatalf("proxy environment survived pdfcpu isolation: %q", joined)
		}
	}
	if !strings.Contains(joined, "PATH=X") || !strings.Contains(joined, "NO_PROXY=*") {
		t.Fatalf("expected environment controls are missing: %q", joined)
	}
}

func TestParsePDFCPUInfoUsesPinnedJSONShape(t *testing.T) {
	data := []byte(`{
		"header":{"version":"pdfcpu test"},
		"infos":[{
			"source":"reading.pdf",
			"version":"1.7",
			"pageCount":12,
			"title":"Bundle",
			"author":"Example",
			"subject":"Evidence",
			"producer":"Producer",
			"creator":"Creator",
			"creationDate":"D:20260101000000Z",
			"modificationDate":"D:20260102000000Z",
			"tagged":true,
			"hybrid":false,
			"linearized":true,
			"usingXRefStreams":true,
			"usingObjectStreams":true,
			"watermarked":false,
			"thumbnails":true,
			"form":true,
			"signatures":true,
			"appendOnly":true,
			"bookmarks":true,
			"names":true,
			"encrypted":true,
			"permissions":3900,
			"attachments":[{"fileName":"one.txt"},{"fileName":"two.txt"}],
			"futureField":"ignored for upstream compatibility"
		}]
	}`)
	// The test source above is a Go raw string; turn the readable escaped
	// representation into the actual JSON bytes expected from the CLI.
	data = bytes.ReplaceAll(data, []byte(`\"`), []byte(`"`))
	info, err := parsePDFCPUInfo(data)
	if err != nil {
		t.Fatal(err)
	}
	assessment := PDFAssessment{}
	assessment.applyInfo(info)
	if assessment.Version != "1.7" || assessment.PageCount != 12 || assessment.Title != "Bundle" || assessment.Author != "Example" {
		t.Fatalf("unexpected parsed PDF info: %+v", assessment)
	}
	if !assessment.Tagged || !assessment.Linearized || !assessment.Form || !assessment.Signatures || !assessment.Encrypted || assessment.AttachmentCount != 2 {
		t.Fatalf("PDF flags did not survive info parsing: %+v", assessment)
	}
	if assessment.Permissions != 3900 {
		t.Fatalf("permissions mismatch: %+v", assessment)
	}
}

func TestParsePDFCPUInfoRejectsMissingOrMultipleRecords(t *testing.T) {
	for i, data := range [][]byte{
		[]byte(`{"infos":[]}`),
		[]byte(`{"infos":[{"version":"1.7","pageCount":1},{"version":"1.7","pageCount":2}]}`),
		[]byte(`{"infos":[{"version":"","pageCount":1}]}`),
		[]byte(`not json`),
	} {
		data = bytes.ReplaceAll(data, []byte(`\"`), []byte(`"`))
		if _, err := parsePDFCPUInfo(data); err == nil {
			t.Fatalf("case %d: expected malformed pdfcpu info rejection", i+1)
		}
	}
}

func TestPDFCPULimitedCaptureFlagsOverflow(t *testing.T) {
	var buf bytes.Buffer
	capture := &pdfcpuLimitedCapture{buf: &buf, max: 5}
	if n, err := capture.Write([]byte("123456789")); err != nil || n != 9 {
		t.Fatalf("unexpected capture write: n=%d err=%v", n, err)
	}
	if !capture.overflow || buf.String() != "12345" {
		t.Fatalf("bounded capture failed: overflow=%v data=%q", capture.overflow, buf.String())
	}
}

func TestPDFCPUInfoEnvelopeRemainsForwardCompatible(t *testing.T) {
	data := []byte(`{"header":{"extra":true},"infos":[{"version":"2.0","pageCount":1,"newFlag":true}]}`)
	data = bytes.ReplaceAll(data, []byte(`\"`), []byte(`"`))
	var envelope pdfcpuInfoEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Infos) != 1 || envelope.Infos[0].Version != "2.0" {
		t.Fatalf("unexpected forward-compatible decode: %+v", envelope)
	}
}
