package githubfirstprobe

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/blevesearch/bleve/v2"
	"github.com/emersion/go-mbox"
	"github.com/emersion/go-message/mail"
	"github.com/gabriel-vasile/mimetype"
	"github.com/mholt/archives"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/xuri/excelize/v2"
)

func TestGitHubFirstStackWorksTogether(t *testing.T) {
	t.Run("wails-shell-options", func(t *testing.T) {
		app := &options.App{Title: "ECO GitHub-First Probe"}
		if app.Title == "" {
			t.Fatal("Wails options did not retain application title")
		}
	})

	t.Run("matter-search-is-memory-only", func(t *testing.T) {
		idx, err := bleve.NewMemOnly(bleve.NewIndexMapping())
		if err != nil {
			t.Fatal(err)
		}
		defer idx.Close()

		docs := map[string]map[string]any{
			"evidence:1": {"matter_id": "MATTER-A", "kind": "Evidence", "text": "The written warranty confirmation is still missing."},
			"issue:1":    {"matter_id": "MATTER-A", "kind": "Issue", "text": "Obtain warranty confirmation from the organisation."},
			"task:1":     {"matter_id": "MATTER-A", "kind": "Task", "text": "Follow up for written warranty confirmation."},
		}
		for id, doc := range docs {
			if err := idx.Index(id, doc); err != nil {
				t.Fatalf("index %s: %v", id, err)
			}
		}
		q := bleve.NewMatchQuery("warranty confirmation")
		res, err := idx.Search(bleve.NewSearchRequest(q))
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Hits) != 3 {
			t.Fatalf("expected 3 connected Matter search hits, got %d", len(res.Hits))
		}
	})

	t.Run("archive-identification", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, err := zw.Create("nested/evidence.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, "preserved synthetic evidence"); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}

		format, _, err := archives.Identify(context.Background(), "fixture.zip", bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := format.(archives.Extractor); !ok {
			t.Fatalf("identified archive is not extractable: %T", format)
		}
	})

	t.Run("rfc-mail-parsing", func(t *testing.T) {
		eml := "From: sender@example.test\r\nTo: recipient@example.test\r\nSubject: Warranty confirmation\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nPlease send the written confirmation.\r\n"
		mr, err := mail.CreateReader(strings.NewReader(eml))
		if err != nil {
			t.Fatal(err)
		}
		subject, err := mr.Header.Subject()
		if err != nil {
			t.Fatal(err)
		}
		if subject != "Warranty confirmation" {
			t.Fatalf("unexpected subject %q", subject)
		}
	})

	t.Run("mbox-container-parsing", func(t *testing.T) {
		mboxData := "From sender@example.test Thu Jan 01 00:00:00 2026\nSubject: One\n\nFirst\n\nFrom sender@example.test Fri Jan 02 00:00:00 2026\nSubject: Two\n\nSecond\n"
		r := mbox.NewReader(strings.NewReader(mboxData))
		count := 0
		for {
			msg, err := r.NextMessage()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.ReadAll(msg); err != nil {
				t.Fatal(err)
			}
			count++
		}
		if count != 2 {
			t.Fatalf("expected 2 messages, got %d", count)
		}
	})

	t.Run("xlsx-round-trip", func(t *testing.T) {
		book := excelize.NewFile()
		defer book.Close()
		if err := book.SetCellValue("Sheet1", "A1", "Warranty confirmation"); err != nil {
			t.Fatal(err)
		}
		buf, err := book.WriteToBuffer()
		if err != nil {
			t.Fatal(err)
		}
		reopened, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		defer reopened.Close()
		value, err := reopened.GetCellValue("Sheet1", "A1")
		if err != nil {
			t.Fatal(err)
		}
		if value != "Warranty confirmation" {
			t.Fatalf("unexpected cell value %q", value)
		}
	})

	t.Run("content-mime-detection", func(t *testing.T) {
		mt := mimetype.Detect([]byte("plain evidence text\n"))
		if mt == nil || mt.String() == "" {
			t.Fatal("MIME detector returned no classification")
		}
	})
}
