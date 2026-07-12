//nolint:errcheck // HTTP 测试 handler 中的 Encode/Write 错误在测试中可安全忽略
package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		_ = json.NewEncoder(w).Encode(resp)
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

		_, _ = w.Write([]byte("data: "))
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []choice{{Delta: delta{Content: "Hel"}, FinishReason: ""}},
		})
		_, _ = w.Write([]byte("\n\ndata: "))
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []choice{{Delta: delta{Content: "lo"}, FinishReason: ""}},
		})
		_, _ = w.Write([]byte("\n\ndata: "))
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []choice{{Delta: delta{Content: ""}, FinishReason: "stop"}},
		})
		_, _ = w.Write([]byte("\n\ndata: [DONE]\n\n"))
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
		_ = json.NewEncoder(w).Encode(chatResponse{
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

func TestStream_HTTPErrorResponse(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(chatResponse{
			Error: &apiError{Message: "Invalid API key"},
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
			break
		}
	}
	if !gotError {
		t.Error("expected error chunk from unauthorized stream")
	}
}

func TestStream_ToolCallsFinish(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("data: "))
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []choice{{
				Delta: delta{
					Content: "",
					ToolCalls: []toolCall{{
						ID:   "call_123",
						Type: "function",
						Function: toolCallFunc{
							Name:      "remember",
							Arguments: `{"key":"test","value":"val"}`,
						},
					}},
				},
				FinishReason: "",
			}},
		})
		_, _ = w.Write([]byte("\n\ndata: "))
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []choice{{Delta: delta{}, FinishReason: "tool_calls"}},
		})
		_, _ = w.Write([]byte("\n\ndata: [DONE]\n\n"))
	})
	defer srv.Close()

	ch, err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "call tool"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	foundToolCall := false
	for chunk := range ch {
		if chunk.Type == llm.ChunkToolCall && chunk.ToolCall != nil {
			foundToolCall = true
			if chunk.ToolCall.Tool != "remember" {
				t.Errorf("ToolCall.Tool = %q, want %q", chunk.ToolCall.Tool, "remember")
			}
		}
	}
	if !foundToolCall {
		t.Error("expected tool call chunk")
	}
}

func TestExtractToolCallID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"toolcall:abc123", "abc123"},
		{"hello", "tool_5"},
		{"toolcall:", ""},
		{"", "tool_0"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := extractToolCallID(tc.input)
			if !strings.HasPrefix(got, "tool_") && got != tc.want {
				t.Errorf("extractToolCallID(%q) = %q, want %q", tc.input, got, tc.want)
			}
			if strings.HasPrefix(tc.input, "toolcall:") && got != tc.want {
				t.Errorf("extractToolCallID(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestChat_NoChoices(t *testing.T) {
	srv, p := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(chatResponse{Choices: []choice{}})
	})
	defer srv.Close()

	_, err := p.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
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
