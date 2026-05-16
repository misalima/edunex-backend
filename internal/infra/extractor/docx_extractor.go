package extractor

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"strings"

	sec "github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
)

// extractDOCX extracts text from a DOCX file provided as a byte slice.
// It looks for word/document.xml inside the DOCX ZIP and parses XML,
// preserving paragraph breaks and runs of text.
func extractDOCX(ctx context.Context, data []byte) (*sec.ExtractionResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, ErrCorruptedFile
	}

	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return nil, ErrUnsupportedFormat
	}

	rc, err := docFile.Open()
	if err != nil {
		return nil, ErrCorruptedFile
	}
	defer func() { _ = rc.Close() }()

	// Read content
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, ErrCorruptedFile
	}

	// Basic XML parsing: look for <w:p> (paragraph) and <w:t> (text)
	decoder := xml.NewDecoder(bytes.NewReader(b))
	var sb strings.Builder
	var inText bool

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, ErrCorruptedFile
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := t.Name.Local
			if local == "p" {
				if sb.Len() > 0 {
					sb.WriteString("\n\n")
				}
			}
			if local == "br" {
				sb.WriteString("\n")
			}
			if local == "t" {
				inText = true
			}
		case xml.EndElement:
			local := t.Name.Local
			if local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				sb.WriteString(string(t))
			}
		}
	}

	out := normalizeText(sb.String())
	return &sec.ExtractionResult{Text: out, Pages: 0, SourceType: "docx"}, nil
}
