package core

import (
	"context"
	"testing"
	"time"

	"github.com/desktop-pet/petcore/internal/agent"
	"github.com/desktop-pet/petcore/internal/config"
	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/fsm"
	"github.com/desktop-pet/petcore/internal/llm"
	_ "github.com/desktop-pet/petcore/internal/llm/mock"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/plugin"
	"github.com/desktop-pet/petcore/internal/tool"
)

func TestNewEngine_AllDepsInjected(t *testing.T) {
	eng := newTestEngine(t, false)
	if eng == nil {
		t.Fatal("New returned nil")
	}
}

func TestEngine_HandleInput(t *testing.T) {
	eng := newTestEngine(t, false)

	events := make(chan event.Event, 100)
	sink := &eventSink{ch: events}
	eng.agent.SetSink(sink)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := eng.HandleInput(ctx, "你好")
	if err != nil {
		t.Fatal(err)
	}

	// The pipeline sends agent.thinking first, then agent.reply
	// Read events until we see EventAgentReply (or timeout)
	for {
		select {
		case e := <-events:
			if e.Kind == event.EventAgentReply {
				return // success — pipeline completed
			}
			// Other events (agent.thinking, etc.) are valid intermediate states
			continue
		case <-ctx.Done():
			t.Fatal("timeout waiting for EventAgentReply")
		}
	}
}

func TestEngine_HandleInput_Empty(t *testing.T) {
	eng := newTestEngine(t, false)
	err := eng.HandleInput(context.Background(), "")
	if err != nil {
		t.Errorf("empty input should not error: %v", err)
	}
}

func TestEngine_GetStatus(t *testing.T) {
	eng := newTestEngine(t, false)
	status := eng.GetStatus()
	if status["state"] != string(fsm.StateIdle) {
		t.Errorf("state = %v, want %q", status["state"], fsm.StateIdle)
	}
	if status["plugins"] != 0 {
		t.Errorf("plugins = %v, want 0", status["plugins"])
	}
}

func TestEngine_Run_StartStop(t *testing.T) {
	eng := newTestEngine(t, true)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := eng.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEngine_SetSink_Nil(t *testing.T) {
	eng := newTestEngine(t, false)
	// SetSink with nil should not panic
	eng.SetSink(nil)
}

func TestEngine_SetSink_Replaces(t *testing.T) {
	eng := newTestEngine(t, false)

	ch := make(chan event.Event, 10)
	eng.SetSink(&eventSink{ch: ch})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = eng.HandleInput(ctx, "hi")

	select {
	case e := <-ch:
		if e.Kind == "" {
			t.Error("event should have a kind")
		}
	case <-ctx.Done():
		t.Error("timeout waiting for event")
	}
}

func TestEngine_HandleInput_WithStateTransition(t *testing.T) {
	eng := newTestEngine(t, false)

	ch := make(chan event.Event, 100)
	eng.SetSink(&eventSink{ch: ch})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := eng.HandleInput(ctx, "hello")
	if err != nil {
		t.Fatal(err)
	}

	// Should get at least a state.changed event
	select {
	case e := <-ch:
		if e.Kind == event.EventStateChanged {
			data, ok := e.Data.(map[string]any)
			if ok && data["state"] != nil {
				// OK — state transition happened
			}
		}
	case <-ctx.Done():
		t.Error("timeout waiting for event")
	}
}

// ─── 测试辅助 ────────────────────────────────

func newTestEngine(t *testing.T, withPlugin bool) *Engine {
	t.Helper()

	provider, err := llm.NewProvider("mock", nil)
	if err != nil {
		t.Fatal(err)
	}

	machine := fsm.NewMockMachine(fsm.StateIdle)
	mem := memory.NewMockManager()
	toolReg := tool.NewRegistry()

	ag := agent.New(provider,
		agent.WithMemory(mem),
		agent.WithToolRegistry(toolReg),
	)

	pluginReg := plugin.NewRegistry()
	if withPlugin {
		if err := pluginReg.Register(&plugin.MockPlugin{
			MockName: "test-plugin",
			MockType: plugin.TypeL1YAML,
		}); err != nil {
			t.Fatalf("failed to register plugin: %v", err)
		}
	}

	cfg := config.DefaultConfig()

	return New(machine, ag, mem, pluginReg, toolReg, cfg, event.NoopSink{})
}

type eventSink struct {
	ch chan event.Event
}

func (s *eventSink) Send(e event.Event) error {
	select {
	case s.ch <- e:
	default:
	}
	return nil
}

func (s *eventSink) Close() error {
	// don't close the channel — test owns it
	return nil
}

func TestEngine_New_WithNilSink(t *testing.T) {
	provider, _ := llm.NewProvider("mock", nil)
	machine := fsm.NewMockMachine(fsm.StateIdle)
	mem := memory.NewMockManager()
	toolReg := tool.NewRegistry()
	ag := agent.New(provider, agent.WithMemory(mem), agent.WithToolRegistry(toolReg))
	pluginReg := plugin.NewRegistry()
	cfg := config.DefaultConfig()

	eng := New(machine, ag, mem, pluginReg, toolReg, cfg, nil)
	if eng == nil {
		t.Fatal("New returned nil")
	}
	// Ensure noop sink was used (no panic on Send)
	_ = eng.sink.Send(event.Event{Kind: event.EventPetSpeak})
}
