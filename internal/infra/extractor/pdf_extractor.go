package extractor

import (
	"bytes"
	"context"
	"fmt"
	"io"

	pdf "github.com/ledongthuc/pdf"
	sec "github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
)

// extractPDF uses github.com/ledongthuc/pdf to extract plain text from PDF
// bytes. It opens the PDF directly from an in-memory reader to avoid
// temporary files and improve concurrency.
func extractPDF(ctx context.Context, data []byte) (*sec.ExtractionResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// open directly from memory
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptedFile, err)
	}
	numPages := r.NumPage()

	var buf bytes.Buffer
	// r.GetPlainText returns an io.Reader with the text content
	pr, err := r.GetPlainText()
	if err != nil {
		// fallback: try to read per page
		for i := 1; i <= numPages; i++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			p := r.Page(i)
			if p.V.IsNull() {
				continue
			}
			contentStr, err := p.GetPlainText(nil)
			if err != nil {
				continue
			}
			buf.WriteString(contentStr)
			// use a simple, LLM-friendly page separator
			buf.WriteString("\n---\n")
		}
		text := normalizeText(buf.String())
		return &sec.ExtractionResult{Text: text, Pages: numPages, SourceType: "pdf"}, nil
	}

	// Respect ctx during the copy: run copy in a goroutine and select on ctx.Done().
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(&buf, pr)
		done <- err
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
		}
	}

	text := normalizeText(buf.String())
	return &sec.ExtractionResult{Text: text, Pages: numPages, SourceType: "pdf"}, nil
}
