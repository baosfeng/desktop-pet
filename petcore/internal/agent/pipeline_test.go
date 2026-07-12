package agent

import (
	"context"
	"testing"

	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/feature"
	"github.com/desktop-pet/petcore/internal/llm"
	_ "github.com/desktop-pet/petcore/internal/llm/mock"
	"github.com/desktop-pet/petcore/internal/memory"
)

func TestPreProcessStage_FiltersEmptyMessages(t *testing.T) {
	s := &PreProcessStage{}
	pCtx := &pipelineCtx{
		Request: Request{
			Messages: []llm.Message{
				{Role: "user", Content: "hello"},
				{Role: "user", Content: ""},
				{Role: "assistant", Content: "hi"},
			},
		},
		Sink: event.NoopSink{},
	}

	err := s.Process(context.Background(), pCtx, func(_ context.Context, _ *pipelineCtx) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(pCtx.Request.Messages) != 2 {
		t.Errorf("expected 2 messages after filtering, got %d", len(pCtx.Request.Messages))
	}
}

func TestMemoryStage_InjectsMemory(t *testing.T) {
	mem := memory.NewInMemoryManager()
	mem.AddShortTerm(memory.Message{Role: "user", Content: "previous chat"})
	_ = mem.Remember("favorite", "pizza", 5)

	flags := feature.New(map[string]bool{"agent.memory_stage": true})
	flags.RegisterDefaults()

	s := &MemoryStage{memory: mem, flags: flags}
	pCtx := &pipelineCtx{
		Request: Request{
			Messages:     []llm.Message{{Role: "user", Content: "hi"}},
			SystemPrompt: "You are a pet.",
		},
		Sink: event.NoopSink{},
	}

	var capturedPrompt string
	err := s.Process(context.Background(), pCtx, func(_ context.Context, ctx *pipelineCtx) error {
		capturedPrompt = ctx.SystemPrompt
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if capturedPrompt == "You are a pet." {
		t.Error("expected memory to be injected into system prompt")
	}
	if !contains(capturedPrompt, "previous chat") {
		t.Error("expected short-term memory in system prompt")
	}
	if !contains(capturedPrompt, "favorite") {
		t.Error("expected core memory in system prompt")
	}
}

func TestMemoryStage_DisabledByFlag(t *testing.T) {
	mem := memory.NewInMemoryManager()
	mem.AddShortTerm(memory.Message{Role: "user", Content: "test"})

	flags := feature.New(map[string]bool{})
	flags.RegisterDefaults()

	s := &MemoryStage{memory: mem, flags: flags}
	pCtx := &pipelineCtx{
		Request: Request{
			Messages:     []llm.Message{{Role: "user", Content: "hi"}},
			SystemPrompt: "You are a pet.",
		},
		Sink: event.NoopSink{},
	}

	var capturedPrompt string
	err := s.Process(context.Background(), pCtx, func(_ context.Context, ctx *pipelineCtx) error {
		capturedPrompt = ctx.SystemPrompt
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if capturedPrompt != "You are a pet." {
		t.Error("expected no memory injection when flag is disabled")
	}
}

func TestPostProcessStage_SavesToMemory(t *testing.T) {
	mem := memory.NewInMemoryManager()
	s := &PostProcessStage{memory: mem, sink: event.NoopSink{}}
	pCtx := &pipelineCtx{
		Request: Request{
			Messages: []llm.Message{
				{Role: "user", Content: "hello"},
				{Role: "assistant", Content: "hi there"},
			},
		},
		Sink: event.NoopSink{},
	}

	err := s.Process(context.Background(), pCtx, func(_ context.Context, _ *pipelineCtx) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	shortTerm := mem.GetShortTerm()
	if len(shortTerm) != 2 {
		t.Fatalf("expected 2 messages in short-term, got %d", len(shortTerm))
	}
	if shortTerm[0].Content != "hello" {
		t.Errorf("message 0 content = %q, want %q", shortTerm[0].Content, "hello")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
