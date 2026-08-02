package extractor

import "errors"

var (
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrTooLarge          = errors.New("file exceeds size limit")
	ErrCorruptedFile     = errors.New("corrupted or unreadable file")
	ErrExtractionFailed  = errors.New("extraction failed")
)
