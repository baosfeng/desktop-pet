package builtin

import (
	"context"
	"testing"
)

func TestRememberTool_Name(t *testing.T) {
	tool := NewRemember(nil)
	if tool.Name() != "remember" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "remember")
	}
}

func TestRememberTool_Description_NotEmpty(t *testing.T) {
	tool := NewRemember(nil)
	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestRememberTool_Execute_Valid(t *testing.T) {
	tool := NewRemember(nil)
	result, err := tool.Execute(context.Background(), `{"key":"coffee","value":"user likes black coffee","importance":7}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRememberTool_Execute_MissingKey(t *testing.T) {
	tool := NewRemember(nil)
	_, err := tool.Execute(context.Background(), `{"value":"test","importance":5}`)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestRememberTool_Execute_MissingValue(t *testing.T) {
	tool := NewRemember(nil)
	_, err := tool.Execute(context.Background(), `{"key":"test","importance":5}`)
	if err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestRememberTool_Execute_DefaultImportance(t *testing.T) {
	tool := NewRemember(nil)
	result, err := tool.Execute(context.Background(), `{"key":"test","value":"test value"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRememberTool_Execute_InvalidJSON(t *testing.T) {
	tool := NewRemember(nil)
	_, err := tool.Execute(context.Background(), `not-json`)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
