package extractor

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	sec "github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
)

type Extractor struct {
	// MaxBytes allows overriding the default read limit for tests or special cases.
	MaxBytes int64
	// storageClient is optional, used for ExtractFromStorage
	storageClient sec.StorageClient
}

func NewExtractor() *Extractor {
	return &Extractor{MaxBytes: DefaultMaxBytes}
}

// NewExtractorWithStorage creates an extractor with storage support
func NewExtractorWithStorage(storageClient sec.StorageClient) *Extractor {
	return &Extractor{
		MaxBytes:      DefaultMaxBytes,
		storageClient: storageClient,
	}
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

// ExtractFromStorage implements sec.DataExtractor.
// Downloads a file from storage by object path and extracts its text content.
func (e *Extractor) ExtractFromStorage(ctx context.Context, objectPath string) (string, error) {
	if e.storageClient == nil {
		return "", fmt.Errorf("storage client is not configured")
	}

	if objectPath == "" {
		return "", fmt.Errorf("object path is required")
	}

	logger.Log.Debug("downloading file from storage", zap.String("object_path", objectPath))

	// Get signed URL
	signedURL, err := e.storageClient.SignURL(ctx, objectPath, 3600)
	if err != nil {
		logger.Log.Error("failed to get signed url", zap.Error(err), zap.String("object_path", objectPath))
		return "", fmt.Errorf("failed to get signed url: %w", err)
	}

	logger.Log.Debug("downloading from signed url", zap.String("object_path", objectPath))

	// Download file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		logger.Log.Error("failed to create download request", zap.Error(err), zap.String("signed_url", signedURL))
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("failed to download file", zap.Error(err), zap.String("object_path", objectPath))
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error("failed to close response body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("download failed with status", zap.Int("status", resp.StatusCode), zap.String("object_path", objectPath))
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Read all data into buffer
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("failed to read response body", zap.Error(err), zap.String("object_path", objectPath))
		return "", fmt.Errorf("failed to read downloaded file: %w", err)
	}

	logger.Log.Debug("file downloaded successfully", zap.String("object_path", objectPath), zap.Int("size", len(data)))

	// Extract text
	bufReader := bytes.NewReader(data)
	contentType := resp.Header.Get("Content-Type")
	result, err := e.ExtractText(ctx, bufReader, contentType)
	if err != nil {
		logger.Log.Error("failed to extract text", zap.Error(err), zap.String("object_path", objectPath))
		return "", fmt.Errorf("failed to extract text: %w", err)
	}

	logger.Log.Debug("text extracted successfully", zap.String("object_path", objectPath), zap.Int("length", len(result.Text)))

	return result.Text, nil
}
