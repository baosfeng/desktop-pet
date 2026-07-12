package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/desktop-pet/petcore/internal/llm"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Provider) {
	t.Helper()
	srv := httptest.NewServer(handler)
	p := &Provider{
		model:   "gpt-4o-mini",
		baseURL: srv.URL,
		apiKey:  "test-key",
	}
	return srv, p
}

func TestProvider_Name(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	if p.Name() != "openai" {
		t.Errorf("Name() = %q, want %q", p.Name(), "openai")
	}
}

func TestProvider_Model(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	if p.Model() != "gpt-4o-mini" {
		t.Errorf("Model() = %q, want %q", p.Model(), "gpt-4o-mini")
	}
}

func TestChat_NonStreaming(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing auth header")
		}

		resp := chatResponse{
			Choices: []choice{
				{
					Message: &chatMessage{
						Role:    "assistant",
						Content: "Hello!",
					},
					FinishReason: "stop",
				},
			},
			Usage: &llm.Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	resp, err := p.Chat(context.Background(), llm.Request{
		Messages:     []llm.Message{{Role: "user", Content: "hi"}},
		SystemPrompt: "You are a pet.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "Hello!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello!")
	}
}

func TestStream_ReturnsTextChunks(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		// SSE format
		w.Write([]byte("data: "))
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []choice{{Delta: delta{Content: "Hel"}, FinishReason: ""}},
		})
		w.Write([]byte("\n\ndata: "))
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []choice{{Delta: delta{Content: "lo"}, FinishReason: ""}},
		})
		w.Write([]byte("\n\ndata: "))
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []choice{{Delta: delta{Content: ""}, FinishReason: "stop"}},
		})
		w.Write([]byte("\n\ndata: [DONE]\n\n"))
	})
	defer srv.Close()

	ch, err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	var texts []string
	for chunk := range ch {
		if chunk.Type == llm.ChunkText {
			texts = append(texts, chunk.Text)
		}
	}

	joined := ""
	for _, t := range texts {
		joined += t
	}
	if joined != "Hello" {
		t.Errorf("streamed text = %q, want %q", joined, "Hello")
	}
}

func TestChat_HTTPError(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(chatResponse{
			Error: &apiError{Message: "Invalid API key"},
		})
	})
	defer srv.Close()

	_, err := p.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error on unauthorized")
	}
}

func TestInit_WithFactory(t *testing.T) {
	provider, err := llm.NewProvider("openai", map[string]any{
		"model":    "gpt-4",
		"base_url": "http://localhost:8080/v1",
		"api_key":  "sk-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if provider.Model() != "gpt-4" {
		t.Errorf("Model = %q, want %q", provider.Model(), "gpt-4")
	}
}

func TestInit_MissingAPIKey(t *testing.T) {
	_, err := llm.NewProvider("openai", map[string]any{
		"model": "gpt-4",
	})
	if err == nil {
		t.Fatal("expected error when API key is missing")
	}
}
