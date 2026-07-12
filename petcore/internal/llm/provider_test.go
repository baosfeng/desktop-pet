package llm

import (
	"context"
	"testing"
)

func TestRegisterAndNewProvider(t *testing.T) {
	// Register a test provider
	Register("test_provider", func(_ map[string]any) (Provider, error) {
		return &mockProvider{name: "test_provider"}, nil
	})

	provider, err := NewProvider("test_provider", nil)
	if err != nil {
		t.Fatal(err)
	}
	if provider.Name() != "test_provider" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "test_provider")
	}
}

func TestNewProvider_Unknown(t *testing.T) {
	_, err := NewProvider("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRegisteredProviders(t *testing.T) {
	// Register a check provider within this test
	Register("check_reg", func(_ map[string]any) (Provider, error) {
		return &mockProvider{name: "check_reg"}, nil
	})
	names := RegisteredProviders()
	found := false
	for _, n := range names {
		if n == "check_reg" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'check_reg' in registered providers")
	}
}

func TestRegister_DuplicatePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()

	Register("dup_provider", func(_ map[string]any) (Provider, error) {
		return &mockProvider{name: "dup"}, nil
	})
	Register("dup_provider", func(_ map[string]any) (Provider, error) {
		return &mockProvider{name: "dup"}, nil
	})
}

// ─── Mock Provider ──────────────────────────

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Model() string { return "test" }
func (m *mockProvider) Stream(_ context.Context, _ Request) (<-chan Chunk, error) {
	ch := make(chan Chunk, 1)
	ch <- Chunk{Type: ChunkDone}
	close(ch)
	return ch, nil
}
func (m *mockProvider) Chat(_ context.Context, _ Request) (Response, error) {
	return Response{Content: "mock response"}, nil
}
