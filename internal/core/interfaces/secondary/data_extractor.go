package secondary

import (
	"context"
	"io"
)

// ExtractionResult contains the normalized text extracted from a file and
// lightweight metadata that may be useful for downstream processing.
type ExtractionResult struct {
	Text       string // plain text ready to be sent to the AI pipeline
	Pages      int    // number of pages (when known, e.g. PDF)
	SourceType string // detected source type: "pdf" | "docx" | "text" | ...
}

// DataExtractor is a port for extracting text content from arbitrary binary data.
// Implementations might handle PDFs, Word documents, plain text, or other formats.
type DataExtractor interface {
	// ExtractText extracts textual content from the provided reader. The
	// contentType parameter can be used to guide the extraction (for example
	// "application/pdf" or "text/plain"). It returns an ExtractionResult or
	// an error if extraction fails.
	ExtractText(ctx context.Context, r io.Reader, contentType string) (*ExtractionResult, error)
	// ExtractFromStorage downloads and extracts text content from a storage object path
	ExtractFromStorage(ctx context.Context, objectPath string) (string, error)
}
