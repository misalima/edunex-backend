package ai

import "errors"

var (
	// ErrAIUnavailable indicates the AI provider is unavailable or returned an unexpected server error.
	ErrAIUnavailable = errors.New("ai provider unavailable")
	// ErrInvalidResponse indicates the provider response could not be parsed or did not contain generated text.
	ErrInvalidResponse = errors.New("invalid ai response")
	// ErrRateLimited indicates the provider rejected the request due to rate limiting.
	ErrRateLimited = errors.New("ai provider rate limited")
)
