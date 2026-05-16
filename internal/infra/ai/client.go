package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	secondary "github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
)

const (
	groqBaseURL          = "https://api.groq.com/openai/v1"
	defaultModel         = "llama-3.1-8b-instant"
	defaultMaxInputChars = 12000
	defaultTimeout       = 30 * time.Second
)

// GroqClient is an AI provider implementation backed by the Groq Responses API.
type GroqClient struct {
	APIKey        string
	Model         string
	MaxInputChars int
	Timeout       time.Duration

	httpClient *http.Client
	baseURL    string
}

var _ secondary.AIProvider = (*GroqClient)(nil)

// NewGroqClient builds a GroqClient with default timeout and input limits.
func NewGroqClient(apiKey, model string) *GroqClient {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		panic("ai.NewGroqClient: apiKey is required")
	}

	if strings.TrimSpace(model) == "" {
		model = defaultModel
	}

	c := &GroqClient{
		APIKey:        apiKey,
		Model:         strings.TrimSpace(model),
		MaxInputChars: defaultMaxInputChars,
		Timeout:       defaultTimeout,
		baseURL:       groqBaseURL,
	}
	// Timeout is enforced by request context in Analyze; the client is intentionally
	// created without a global Timeout to keep context as the single timeout source.
	c.httpClient = &http.Client{}
	return c
}

// Analyze sends lesson plan text to Groq and returns the structured analysis result.
func (c *GroqClient) Analyze(ctx context.Context, text string) (*secondary.AnalysisResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}

	maxChars := c.MaxInputChars
	if maxChars <= 0 {
		maxChars = defaultMaxInputChars
	}

	trimmed := truncateAtWordBoundary(text, maxChars)
	prompt := buildPrompt(trimmed)

	payload := map[string]any{
		"model": c.Model,
		"input": prompt,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIUnavailable, err)
	}

	requestCtx := ctx
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		requestCtx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}

	endpoint := strings.TrimRight(c.baseURL, "/") + "/responses"
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIUnavailable, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctxErr := requestCtx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("%w: %v", ErrAIUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if ctxErr := requestCtx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("%w: %v", ErrAIUnavailable, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf("groq authentication failed (status %d): check GROQ_API_KEY and model permissions", resp.StatusCode)
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("%w: status 429", ErrRateLimited)
	default:
		return nil, fmt.Errorf("%w: status %d body=%s", ErrAIUnavailable, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	result, err := parseAnalysisResult(respBody)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseAnalysisResult(body []byte) (*secondary.AnalysisResult, error) {
	var direct secondary.AnalysisResult
	if err := json.Unmarshal(body, &direct); err == nil {
		if isAnalysisResultPopulated(&direct) {
			return &direct, nil
		}
	}

	candidate, err := extractResponseText(body)
	if err != nil {
		return nil, err
	}

	var parsed secondary.AnalysisResult
	if err := json.Unmarshal([]byte(candidate), &parsed); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	if !isAnalysisResultPopulated(&parsed) {
		return nil, ErrInvalidResponse
	}

	return &parsed, nil
}

func extractResponseText(body []byte) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	if direct, ok := raw["output_text"].(string); ok && strings.TrimSpace(direct) != "" {
		return strings.TrimSpace(direct), nil
	}

	output, ok := raw["output"].([]any)
	if !ok {
		return "", ErrInvalidResponse
	}

	var chunks []string
	for _, item := range output {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		content, ok := obj["content"].([]any)
		if !ok {
			continue
		}
		for _, contentItem := range content {
			contentObj, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := contentObj["text"].(string); ok && strings.TrimSpace(text) != "" {
				chunks = append(chunks, strings.TrimSpace(text))
			}
		}
	}

	if len(chunks) == 0 {
		return "", ErrInvalidResponse
	}
	return strings.Join(chunks, "\n"), nil
}

func isAnalysisResultPopulated(result *secondary.AnalysisResult) bool {
	if result == nil {
		return false
	}
	if strings.TrimSpace(result.Analysis.PedagogicalFeedback) != "" {
		return true
	}
	if len(result.Analysis.Suggestions) > 0 || len(result.Analysis.MissingElements) > 0 {
		return true
	}
	if len(result.Metadata.Objectives) > 0 || len(result.Metadata.BNCCSkills) > 0 {
		return true
	}
	if strings.TrimSpace(result.Metadata.Title) != "" || strings.TrimSpace(result.Metadata.Subject) != "" || strings.TrimSpace(result.Metadata.GradeLevel) != "" {
		return true
	}
	return false
}

func truncateAtWordBoundary(input string, maxChars int) string {
	if maxChars <= 0 {
		return ""
	}

	runes := []rune(input)
	if len(runes) <= maxChars {
		return input
	}

	prefix := runes[:maxChars]
	cut := -1
	for i := len(prefix) - 1; i >= 0; i-- {
		if unicode.IsSpace(prefix[i]) {
			cut = i
			break
		}
	}
	if cut <= 0 {
		return strings.TrimSpace(string(prefix))
	}
	return strings.TrimSpace(string(prefix[:cut]))
}
