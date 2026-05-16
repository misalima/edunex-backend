package extractor

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	sec "github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
)

type Extractor struct {
	// MaxBytes allows overriding the default read limit for tests or special cases.
	MaxBytes int64
}

func NewExtractor() *Extractor {
	return &Extractor{MaxBytes: DefaultMaxBytes}
}

// ExtractText implements sec.DataExtractor.
func (e *Extractor) ExtractText(ctx context.Context, r io.Reader, contentType string) (*sec.ExtractionResult, error) {
	// read up to MaxBytes+1 to detect oversize
	maxBytes := e.MaxBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}

	lr := &io.LimitedReader{R: r, N: maxBytes + 1}
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrTooLarge
	}

	// decide type: 1) trusted contentType 2) detect PDF 3) detect DOCX (zip+word/document.xml)
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct != "" {
		if strings.Contains(ct, "pdf") {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return extractPDF(ctx, data)
		}
		// For DOCX rely on the official mime (wordprocessingml). Checking for
		// the literal "docx" in the content-type is unreliable and unnecessary.
		if strings.Contains(ct, "wordprocessingml") {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return extractDOCX(ctx, data)
		}
	}

	// detect PDF by header
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return extractPDF(ctx, data)
	}

	// detect zip (DOCX)
	if len(data) >= 4 && bytes.Equal(data[:4], []byte{0x50, 0x4b, 0x03, 0x04}) {
		// open zip from buffer and check for word/document.xml
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err == nil {
			for _, f := range zr.File {
				if f.Name == "word/document.xml" {
					if err := ctx.Err(); err != nil {
						return nil, err
					}
					// Note: we reopen the zip inside extractDOCX. This means
					// we parse the in-memory bytes twice (once here to detect
					// the presence of document.xml and again inside
					// extractDOCX). For the MVP this is acceptable because
					// the file is buffered in memory (max 10MB). If needed we
					// can optimize by creating the zip.Reader once and passing
					// it to extractDOCX.
					return extractDOCX(ctx, data)
				}
			}
		}
	}

	// fallback: if content looks like plain text
	detected := http.DetectContentType(data)
	// For the MVP be conservative: accept only explicit text/plain as a
	// fallback. Other text/* types (html, csv) are rejected to avoid edge
	// cases where non-lesson-plan files slip through.
	if strings.HasPrefix(detected, "text/plain") {
		txt := normalizeText(string(data))
		return &sec.ExtractionResult{Text: txt, Pages: 0, SourceType: "text"}, nil
	}

	return nil, ErrUnsupportedFormat
}
