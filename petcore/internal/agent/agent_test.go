package agent

import (
	"context"
	"testing"

	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/llm"
	_ "github.com/desktop-pet/petcore/internal/llm/mock"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/tool"
)

func TestNewAgent_WithMockProvider(t *testing.T) {
	provider, err := llm.NewProvider("mock", nil)
	if err != nil {
		t.Fatal(err)
	}
	mem := memory.NewMockManager()
	toolReg := tool.NewRegistry()

	a := New(provider, WithMemory(mem), WithToolRegistry(toolReg))
	if a == nil {
		t.Fatal("New returned nil")
	}
}

func TestAgent_Run_WithMockProvider(t *testing.T) {
	t.Setenv("MOCK_LLM_REPLY", "hi")
	provider, err := llm.NewProvider("mock", nil)
	if err != nil {
		t.Fatal(err)
	}
	mem := memory.NewInMemoryManager()
	toolReg := tool.NewRegistry()

	a := New(provider, WithMemory(mem), WithToolRegistry(toolReg))

	events := make(chan event.Event, 100)
	sink := &eventChannelSink{ch: events}
	a.SetSink(sink)

	err = a.Run(context.Background(), Request{
		Messages:     []llm.Message{{Role: "user", Content: "你好"}},
		SystemPrompt: "You are a cute desktop pet.",
	})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-events:
		if e.Kind == "" {
			t.Error("received empty event kind")
		}
	default:
		t.Error("expected at least one event from agent run")
	}
}

func TestPipeline_StagesExecuteInOrder(t *testing.T) {
	var order []string

	p := NewPipeline(
		&recordingStage{name: "A", order: &order},
		&recordingStage{name: "B", order: &order},
		&recordingStage{name: "C", order: &order},
	)

	ctx := context.Background()
	pCtx := &pipelineCtx{
		Request: Request{Messages: []llm.Message{{Role: "user", Content: "hi"}}},
		Sink:    event.NoopSink{},
	}

	err := p.Process(ctx, pCtx)
	if err != nil {
		t.Fatal(err)
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(order))
	}
	if order[0] != "A" || order[1] != "B" || order[2] != "C" {
		t.Errorf("stage order = %v, want [A B C]", order)
	}
}

func TestPipeline_EmptyStages(t *testing.T) {
	p := NewPipeline()
	err := p.Process(context.Background(), &pipelineCtx{
		Request: Request{},
		Sink:    event.NoopSink{},
	})
	if err != nil {
		t.Errorf("empty pipeline should not error: %v", err)
	}
}

func TestAgent_NoProvider(t *testing.T) {
	a := New(nil)
	err := a.Run(context.Background(), Request{})
	if err == nil {
		t.Error("expected error when no provider")
	}
}

// ─── 测试辅助 ────────────────────────────────

type recordingStage struct {
	name  string
	order *[]string
}

func (s *recordingStage) Name() string { return s.name }

func (s *recordingStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	*s.order = append(*s.order, s.name)
	return next(ctx, pCtx)
}

type eventChannelSink struct {
	ch chan event.Event
}

func (s *eventChannelSink) Send(e event.Event) error {
	select {
	case s.ch <- e:
	default:
	}
	return nil
}

func (s *eventChannelSink) Close() error {
	return nil
}

var _ Agent = (*agentImpl)(nil)
