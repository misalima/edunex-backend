package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGroqClientAnalyzeReturnsStructuredResultFromOutputText(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("expected /responses, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("expected Authorization header with test key, got %q", got)
		}

		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if _, ok := reqBody["input"].(string); !ok {
			t.Fatalf("expected input in request body")
		}
		rf, ok := reqBody["response_format"].(map[string]any)
		if !ok {
			t.Fatalf("expected response_format in request body")
		}
		if rf["type"] != "json_object" {
			t.Fatalf("expected response_format.type json_object, got %v", rf["type"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "output_text": "{\"metadata\":{\"title\":\"Plano de Frações\",\"subject\":\"Matemática\",\"grade_level\":\"6º ano\",\"objectives\":[\"Compreender frações\"],\"bncc_skills\":[\"Matemática - resolução de problemas com frações\"]},\"analysis\":{\"pedagogical_feedback\":\"## Pontos Fortes\\n- Objetivos claros\\n\\n## Pontos a Melhorar\\n- Diferenciar níveis\",\"alignment_score\":82,\"suggestions\":[\"Adicionar avaliação formativa\"],\"missing_elements\":[\"Critérios de avaliação\"]}}"
}`))
	}))
	defer ts.Close()

	client := NewGroqClient("test-key", "test-model")
	client.baseURL = ts.URL

	result, err := client.Analyze(context.Background(), "Plano de aula de exemplo")
	if err != nil {
		t.Fatalf("Analyze returned unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil AnalysisResult")
	}
	if result.Metadata.Title != "Plano de Frações" {
		t.Fatalf("expected title Plano de Frações, got %q", result.Metadata.Title)
	}
	if result.Analysis.AlignmentScore != 82 {
		t.Fatalf("expected alignment score 82, got %d", result.Analysis.AlignmentScore)
	}
	if !strings.Contains(result.Analysis.PedagogicalFeedback, "Pontos Fortes") {
		t.Fatalf("expected markdown feedback, got %q", result.Analysis.PedagogicalFeedback)
	}
}

func TestGroqClientAnalyzeReturnsErrRateLimitedOn429(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limit"}`))
	}))
	defer ts.Close()

	client := NewGroqClient("test-key", "test-model")
	client.baseURL = ts.URL

	_, err := client.Analyze(context.Background(), "Plano de aula")
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestGroqClientAnalyzeReturnsErrAIUnavailableOn500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal"}`))
	}))
	defer ts.Close()

	client := NewGroqClient("test-key", "test-model")
	client.baseURL = ts.URL

	_, err := client.Analyze(context.Background(), "Plano de aula")
	if !errors.Is(err, ErrAIUnavailable) {
		t.Fatalf("expected ErrAIUnavailable, got %v", err)
	}
}

func TestGroqClientAnalyzeReturnsErrorWhenAPIKeyIsEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer" && got != "Bearer " {
			t.Fatalf("expected empty bearer token, got %q", got)
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer ts.Close()

	client := &GroqClient{
		APIKey:        "",
		Model:         "test-model",
		MaxInputChars: defaultMaxInputChars,
		Timeout:       defaultTimeout,
		httpClient:    &http.Client{},
		baseURL:       ts.URL,
	}

	_, err := client.Analyze(context.Background(), "Plano de aula")
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestGroqClientAnalyzeReturnsErrInvalidResponseOnInvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"output_text":`))
	}))
	defer ts.Close()

	client := NewGroqClient("test-key", "test-model")
	client.baseURL = ts.URL

	_, err := client.Analyze(context.Background(), "Plano de aula")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestTruncateAtWordBoundaryReturnsWholeTextWhenInputIsShorter(t *testing.T) {
	input := "texto curto"
	out := truncateAtWordBoundary(input, 20)
	if out != input {
		t.Fatalf("expected %q, got %q", input, out)
	}
}

func TestTruncateAtWordBoundaryCutsAtWordBoundary(t *testing.T) {
	input := "um dois tres quatro"
	out := truncateAtWordBoundary(input, 11)
	if out != "um dois" {
		t.Fatalf("expected %q, got %q", "um dois", out)
	}
}

func TestTruncateAtWordBoundaryDoesNotBreakUTF8Runes(t *testing.T) {
	input := "ação incrível"
	out := truncateAtWordBoundary(input, 4)
	if out != "ação" {
		t.Fatalf("expected %q, got %q", "ação", out)
	}
	if !utf8.ValidString(out) {
		t.Fatalf("expected valid UTF-8 output, got %q", out)
	}
}
