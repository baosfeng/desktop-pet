package server

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/desktop-pet/petcore/internal/agent"
	"github.com/desktop-pet/petcore/internal/config"
	"github.com/desktop-pet/petcore/internal/core"
	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/fsm"
	"github.com/desktop-pet/petcore/internal/llm"
	_ "github.com/desktop-pet/petcore/internal/llm/mock"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/plugin"
	"github.com/desktop-pet/petcore/internal/tool"
)

func TestServer_Ping(t *testing.T) {
	stdin := strings.NewReader(`{"type":"cmd","id":"1","method":"ping","params":{}}` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, `"pong"`) {
		t.Errorf("expected pong in response, got: %s", out)
	}
}

func TestServer_Chat(t *testing.T) {
	// 使用快速回复以减少测试时间
	t.Setenv("MOCK_LLM_REPLY", "ok")
	stdin := strings.NewReader(`{"type":"cmd","id":"1","method":"chat","params":{"text":"hi"}}` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, `"done"`) {
		t.Errorf("expected done in response, got: %s", out)
	}
}

func TestServer_InvalidJSON(t *testing.T) {
	stdin := strings.NewReader(`not json` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, "invalid command") {
		t.Errorf("expected error message, got: %s", out)
	}
}

func TestServer_UnknownMethod(t *testing.T) {
	stdin := strings.NewReader(`{"type":"cmd","id":"1","method":"unknown"}` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, "unknown method") {
		t.Errorf("expected unknown method error, got: %s", out)
	}
}

func TestServer_GetStatus(t *testing.T) {
	stdin := strings.NewReader(`{"type":"cmd","id":"1","method":"get_status"}` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, "state") {
		t.Errorf("expected state in status, got: %s", out)
	}
}

func TestServer_VerifyAPIKey(t *testing.T) {
	stdin := strings.NewReader(`{"type":"cmd","id":"1","method":"verify_api_key","params":{"apiKey":"test-key","provider":"mock","baseUrl":"http://localhost","modelName":"mock-v1"}}` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, `"valid":true`) {
		t.Errorf("expected valid:true in response, got: %s", out)
	}
}

func TestServer_VerifyAPIKey_InvalidProvider(t *testing.T) {
	stdin := strings.NewReader(`{"type":"cmd","id":"1","method":"verify_api_key","params":{"apiKey":"key","provider":"nonexistent","baseUrl":"http://localhost","modelName":"m"}}` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, `"error"`) {
		t.Errorf("expected error response for unknown provider, got: %s", out)
	}
}

func TestSinkAdapter_Send(t *testing.T) {
	var stdout bytes.Buffer
	srv := New(nil, nil, &stdout)
	adapter := NewSinkAdapter(srv)

	err := adapter.Send(event.Event{
		Kind: event.EventPetSpeak,
		Data: map[string]string{"text": "hello"},
	})
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	if !strings.Contains(out, "pet.speak") {
		t.Errorf("expected pet.speak event, got: %s", out)
	}
}

func TestSinkAdapter_Close(t *testing.T) {
	var stdout bytes.Buffer
	srv := New(nil, nil, &stdout)
	adapter := NewSinkAdapter(srv)

	err := adapter.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestServer_ChatWithEmptyParams(t *testing.T) {
	// Empty text is handled gracefully (returns done without error)
	stdin := strings.NewReader(`{"type":"cmd","id":"1","method":"chat","params":{}}` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, `"done"`) {
		t.Errorf("expected done response for empty chat, got: %s", out)
	}
}

// ─── 测试辅助 ────────────────────────────────

func newTestServer(t *testing.T, stdin *strings.Reader, stdout *bytes.Buffer) *Server {
	t.Helper()

	provider, err := llm.NewProvider("mock", nil)
	if err != nil {
		t.Fatal(err)
	}
	machine := fsm.NewMockMachine(fsm.StateIdle)
	mem := memory.NewMockManager()
	toolReg := tool.NewRegistry()
	ag := agent.New(provider, agent.WithMemory(mem), agent.WithToolRegistry(toolReg))
	pluginReg := plugin.NewRegistry()
	cfg := config.DefaultConfig()

	eng := core.New(machine, ag, mem, pluginReg, toolReg, cfg, event.NoopSink{})
	srv := New(eng, stdin, stdout)

	sink := NewSinkAdapter(srv)
	ag.SetSink(sink)

	return srv
}

func TestServer_ChatInvalidParams(t *testing.T) {
	// JSON parse error for chat params (params is not an object)
	stdin := strings.NewReader(`{"type":"cmd","id":"1","method":"chat","params":"not-an-object"}` + "\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = srv.Run(ctx)

	out := stdout.String()
	if !strings.Contains(out, "invalid params") {
		t.Errorf("expected 'invalid params' error, got: %s", out)
	}
}

func TestServer_WriteJSON_HandlesNil(t *testing.T) {
	// writeJSON with nil input should not panic
	var stdout bytes.Buffer
	srv := New(nil, nil, &stdout)
	srv.writeJSON(nil) // nil should produce "null" which is valid JSON
	out := stdout.String()
	if out != "null\n" {
		t.Errorf("writeJSON(nil) = %q, want %q", out, "null\n")
	}
}

func TestServer_WriteJSON_WithChannelError(_ *testing.T) {
	// writeJSON with a value that cannot be marshaled should handle error gracefully
	var stdout bytes.Buffer
	srv := New(nil, nil, &stdout)
	// Marshal of a channel produces a specific error
	srv.writeJSON(make(chan int))
	// Should not panic, stdout should be empty since marshal fails
}

func TestServer_ChatWithEmptyLine(t *testing.T) {
	stdin := strings.NewReader("\n\n")
	var stdout bytes.Buffer

	srv := newTestServer(t, stdin, &stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_ = srv.Run(ctx)
	// Empty lines should be skipped without error
	out := stdout.String()
	if out != "" {
		t.Errorf("expected no output for empty lines, got: %s", out)
	}
}
