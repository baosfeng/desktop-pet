package ollama

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
		model:   "qwen2.5:7b",
		baseURL: srv.URL,
	}
	return srv, p
}

func TestProvider_Name(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	if p.Name() != "ollama" {
		t.Errorf("Name() = %q, want %q", p.Name(), "ollama")
	}
}

func TestChat_NonStreaming(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: &ollamaMessage{
				Role:    "assistant",
				Content: "Hello from Ollama!",
			},
			Done: true,
		})
	})
	defer srv.Close()

	resp, err := p.Chat(context.Background(), llm.Request{
		Messages:     []llm.Message{{Role: "user", Content: "hi"}},
		SystemPrompt: "You are a pet.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "Hello from Ollama!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from Ollama!")
	}
}

func TestStream_ReturnsTextChunks(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: &ollamaMessage{Content: "Hel"},
			Done:    false,
		})
		w.Write([]byte("\n"))
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: &ollamaMessage{Content: "lo"},
			Done:    false,
		})
		w.Write([]byte("\n"))
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Done: true,
		})
		w.Write([]byte("\n"))
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
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Error: "internal error",
		})
	})
	defer srv.Close()

	_, err := p.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInit_WithFactory(t *testing.T) {
	provider, err := llm.NewProvider("ollama", map[string]any{
		"model":    "llama3.1:8b",
		"base_url": "http://localhost:11434",
	})
	if err != nil {
		t.Fatal(err)
	}
	if provider.Model() != "llama3.1:8b" {
		t.Errorf("Model = %q, want %q", provider.Model(), "llama3.1:8b")
	}
}

func TestStream_HTTPError(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Error: "server error",
		})
	})
	defer srv.Close()

	ch, err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	gotError := false
	for chunk := range ch {
		if chunk.Type == llm.ChunkError {
			gotError = true
			_ = chunk.Error.Error() // ensure Error() is called
		}
	}
	if !gotError {
		t.Error("expected error chunk from stream")
	}
}

func TestStream_NonJSONError(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("upstream error"))
	})
	defer srv.Close()

	ch, err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	gotError := false
	for chunk := range ch {
		if chunk.Type == llm.ChunkError {
			gotError = true
		}
	}
	if !gotError {
		t.Error("expected error chunk from non-JSON error")
	}
}

func TestChat_NonJSONError(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("upstream error"))
	})
	defer srv.Close()

	_, err := p.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProvider_httpClientLazy(t *testing.T) {
	p := &Provider{}
	c1 := p.httpClient()
	c2 := p.httpClient()
	if c1 != c2 {
		t.Error("httpClient should return cached instance")
	}
}
