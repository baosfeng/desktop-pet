package tool

import (
	"context"
	"testing"
)

func TestNewRegistry_Empty(t *testing.T) {
	r := NewRegistry()
	if list := r.List(); len(list) != 0 {
		t.Errorf("expected empty registry, got %d tools", len(list))
	}
}

func TestRegistry_RegisterAndList(t *testing.T) {
	r := NewRegistry()
	tool := &MockTool{
		MockName:        "echo",
		MockDescription: "echo back input",
		MockExecute: func(_ context.Context, args string) (string, error) {
			return args, nil
		},
	}
	if err := r.Register(tool); err != nil {
		t.Fatal(err)
	}
	list := r.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(list))
	}
	if list[0].Name() != "echo" {
		t.Errorf("tool name = %q, want %q", list[0].Name(), "echo")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	tool := &MockTool{MockName: "dup"}
	_ = r.Register(tool)
	err := r.Register(tool)
	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestRegistry_Execute(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(&MockTool{
		MockName: "greet",
		MockExecute: func(_ context.Context, args string) (string, error) {
			return "Hello, " + args, nil
		},
	})
	result, err := r.Execute(context.Background(), "greet", "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello, Alice" {
		t.Errorf("result = %q, want %q", result, "Hello, Alice")
	}
}

func TestRegistry_ExecuteNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute(context.Background(), "nonexistent", "")
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestRegistryInterface(t *testing.T) {
	var _ Registry = (*builtinRegistry)(nil)
}
