package extractor

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestDetectDocxAndExtract(t *testing.T) {
	// create an in-memory docx (zip) with word/document.xml
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	_, err = io.WriteString(f, `<?xml version="1.0" encoding="UTF-8"?><w:document><w:body><w:p><w:r><w:t>Hello</w:t></w:r></w:p><w:p><w:r><w:t>World</w:t></w:r></w:p></w:body></w:document>`)
	if err != nil {
		t.Fatalf("write xml: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	ex := NewExtractor()
	res, err := ex.ExtractText(context.Background(), bytes.NewReader(buf.Bytes()), "")
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}
	if res.SourceType != "docx" {
		t.Fatalf("expected docx, got %s", res.SourceType)
	}
	if !strings.Contains(res.Text, "Hello") {
		t.Fatalf("expected 'Hello' in text, got: %s", res.Text)
	}
	if !strings.Contains(res.Text, "World") {
		t.Fatalf("expected 'World' in text, got: %s", res.Text)
	}
	if !strings.Contains(res.Text, "\n\n") {
		t.Fatalf("expected paragraph breaks in text, got: %q", res.Text)
	}
}

func TestDetectPdfByHeaderOnly(t *testing.T) {
	data := []byte("%PDF-1.4\n%\u00e2\u00e3\u00cf\u00d3\n1 0 obj\n<< /Type /Catalog >>")
	ex := NewExtractor()
	res, err := ex.ExtractText(context.Background(), bytes.NewReader(data), "")
	if res != nil {
		t.Fatalf("expected no result, got %+v", res)
	}
	if !errors.Is(err, ErrCorruptedFile) {
		t.Fatalf("expected ErrCorruptedFile, got %v", err)
	}
}

func TestUnsupportedFormat(t *testing.T) {
	ex := NewExtractor()
	_, err := ex.ExtractText(context.Background(), bytes.NewReader([]byte{0x00, 0x01, 0x02, 0x03}), "")
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestTooLarge(t *testing.T) {
	ex := &Extractor{MaxBytes: 10}
	data := bytes.Repeat([]byte("a"), 20)
	_, err := ex.ExtractText(context.Background(), bytes.NewReader(data), "text/plain")
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestCorruptedDocx(t *testing.T) {
	ex := NewExtractor()
	data := []byte{0x50, 0x4b, 0x03, 0x04, 0x00}
	_, err := ex.ExtractText(context.Background(), bytes.NewReader(data), "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if !errors.Is(err, ErrCorruptedFile) {
		t.Fatalf("expected ErrCorruptedFile, got %v", err)
	}
}
