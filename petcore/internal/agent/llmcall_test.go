package agent

import (
	"context"
	"testing"
	"time"

	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/feature"
	"github.com/desktop-pet/petcore/internal/llm"
	_ "github.com/desktop-pet/petcore/internal/llm/mock"
	"github.com/desktop-pet/petcore/internal/tool"
)

func TestLLMCallStage_ToolCallCycle(t *testing.T) {
	t.Setenv("MOCK_LLM_TOOL_CALL", "true")
	t.Setenv("MOCK_LLM_REPLY", "")
	provider, err := llm.NewProvider("mock", nil)
	if err != nil {
		t.Fatal(err)
	}

	toolReg := tool.NewRegistry()
	// Register a test tool
	toolReg.Register(&tool.MockTool{
		MockName:        "test_tool",
		MockDescription: "A test tool",
		MockExecute: func(_ context.Context, _ string) (string, error) {
			return "tool result", nil
		},
	})

	flags := feature.New(map[string]bool{"agent.tool_loop": true})
	flags.RegisterDefaults()

	sink := &eventChannelSink{ch: make(chan event.Event, 100)}
	stage := &LLMCallStage{
		provider: provider,
		sink:     sink,
		tool:     toolReg,
		flags:    flags,
	}

	pCtx := &pipelineCtx{
		Request: Request{
			Messages:     []llm.Message{{Role: "user", Content: "use a tool"}},
			SystemPrompt: "You can use tools.",
		},
		Sink: sink,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = stage.Process(ctx, pCtx, func(_ context.Context, _ *pipelineCtx) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLLMCallStage_ToolLoopDisabled(t *testing.T) {
	t.Setenv("MOCK_LLM_REPLY", "no tools")
	provider, _ := llm.NewProvider("mock", nil)

	flags := feature.New(map[string]bool{})
	flags.RegisterDefaults()

	sink := &eventChannelSink{ch: make(chan event.Event, 100)}
	stage := &LLMCallStage{
		provider: provider,
		sink:     sink,
		flags:    flags,
	}

	pCtx := &pipelineCtx{
		Request: Request{
			Messages: []llm.Message{{Role: "user", Content: "hi"}},
		},
		Sink: sink,
	}

	err := stage.Process(context.Background(), pCtx, func(_ context.Context, _ *pipelineCtx) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
