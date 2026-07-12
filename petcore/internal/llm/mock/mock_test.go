package mock

import (
	"os"
	"testing"
	"time"

	"github.com/desktop-pet/petcore/internal/llm"
)

func TestGetMockDelay_Default(t *testing.T) {
	os.Unsetenv("MOCK_LLM_DELAY")
	d := getMockDelay()
	if d != 0 {
		t.Errorf("getMockDelay() = %v, want 0", d)
	}
}

func TestGetMockDelay_WithValue(t *testing.T) {
	t.Setenv("MOCK_LLM_DELAY", "50")
	d := getMockDelay()
	if d != 50*time.Millisecond {
		t.Errorf("getMockDelay() = %v, want 50ms", d)
	}
}

func TestGetMockDelay_InvalidValue(t *testing.T) {
	t.Setenv("MOCK_LLM_DELAY", "not-a-number")
	d := getMockDelay()
	if d != 0 {
		t.Errorf("getMockDelay() with invalid = %v, want 0", d)
	}
}

func TestProvider_NameAndModel(t *testing.T) {
	p := &Provider{}
	if p.Name() != "mock" {
		t.Errorf("Name() = %q, want %q", p.Name(), "mock")
	}
	if p.Model() != "mock-v1" {
		t.Errorf("Model() = %q, want %q", p.Model(), "mock-v1")
	}
}

func TestChat_DefaultReply(t *testing.T) {
	os.Unsetenv("MOCK_LLM_REPLY")
	p := &Provider{}
	resp, err := p.Chat(nil, llm.Request{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response")
	}
}
