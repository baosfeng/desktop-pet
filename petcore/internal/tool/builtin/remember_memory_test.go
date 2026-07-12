package builtin

import (
	"context"
	"testing"

	"github.com/desktop-pet/petcore/internal/memory"
)

func TestRememberTool_WithMemory_PersistsFact(t *testing.T) {
	mem := memory.NewInMemoryManager()
	tool := NewRemember(mem)

	result, err := tool.Execute(context.Background(), `{"key":"coffee","value":"user likes black coffee","importance":7}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}

	// Verify memory was persisted
	val, err := mem.Recall("coffee")
	if err != nil {
		t.Fatalf("failed to recall: %v", err)
	}
	if val != "user likes black coffee" {
		t.Errorf("recalled value = %q, want %q", val, "user likes black coffee")
	}
}

func TestRememberTool_WithMemory_DefaultImportance(t *testing.T) {
	mem := memory.NewInMemoryManager()
	tool := NewRemember(mem)

	_, err := tool.Execute(context.Background(), `{"key":"test","value":"test value"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := mem.Recall("test")
	if err != nil {
		t.Fatal(err)
	}
	if val != "test value" {
		t.Errorf("recalled value = %q, want %q", val, "test value")
	}
}

func TestRememberTool_WithMemory_NilManager(t *testing.T) {
	tool := NewRemember(nil)
	result, err := tool.Execute(context.Background(), `{"key":"test","value":"test value"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestTimestamp_WithEnvVar(t *testing.T) {
	t.Setenv("MOCK_NOW_MS", "1234567890")
	ts := timestamp()
	if ts != 1234567890 {
		t.Errorf("timestamp() = %d, want %d", ts, 1234567890)
	}
}

func TestTimestamp_WithoutEnvVar(t *testing.T) {
	t.Setenv("MOCK_NOW_MS", "")
	ts := timestamp()
	if ts != 0 {
		t.Errorf("timestamp() = %d, want 0", ts)
	}
}

func TestTimestamp_InvalidEnvVar(t *testing.T) {
	t.Setenv("MOCK_NOW_MS", "not-a-number")
	ts := timestamp()
	if ts != 0 {
		t.Errorf("timestamp() with invalid = %d, want 0", ts)
	}
}
