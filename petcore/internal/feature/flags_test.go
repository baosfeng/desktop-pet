package feature

import (
	"testing"
)

func TestNew_Empty(t *testing.T) {
	f := New(nil)
	if f == nil {
		t.Fatal("New returned nil")
	}
	if f.IsEnabled("anything") {
		t.Error("empty flags should return false")
	}
}

func TestIsEnabled(t *testing.T) {
	f := New(map[string]bool{"foo": true, "bar": false})

	if !f.IsEnabled("foo") {
		t.Error("foo should be enabled")
	}
	if f.IsEnabled("bar") {
		t.Error("bar should be disabled")
	}
	if f.IsEnabled("unknown") {
		t.Error("unknown flag should be disabled")
	}
}

func TestAllEnabled(t *testing.T) {
	f := New(map[string]bool{"a": true, "b": false, "c": true})
	enabled := f.AllEnabled()

	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled flags, got %d", len(enabled))
	}
	if !enabled["a"] {
		t.Error("a should be enabled")
	}
	if enabled["b"] {
		t.Error("b should not be enabled")
	}
}

func TestRegister(t *testing.T) {
	f := New(map[string]bool{"existing": true})

	// 已有值，不覆盖
	f.Register("existing", false)
	if !f.IsEnabled("existing") {
		t.Error("Register should not overwrite existing value")
	}

	// 新值，使用默认值
	f.Register("new_flag", false)
	if f.IsEnabled("new_flag") {
		t.Error("new_flag should default to false")
	}
}

func TestRegisterDefaults(t *testing.T) {
	f := New(map[string]bool{FlagMemoryStage: true})
	f.RegisterDefaults()

	if !f.IsEnabled(FlagMemoryStage) {
		t.Error("FlagMemoryStage should be enabled (from config)")
	}
	if f.IsEnabled(FlagToolLoop) {
		t.Error("FlagToolLoop should default to false")
	}
}

func TestString_NonEmpty(t *testing.T) {
	f := New(map[string]bool{"test": true})
	s := f.String()
	if len(s) == 0 {
		t.Error("String() should not be empty")
	}
}

func TestRegisteredFlags_AllPresent(t *testing.T) {
	flags := RegisteredFlags()
	expected := []string{FlagMemoryStage, FlagToolLoop, FlagL1YAML, FlagL2JS, FlagL3MCP}
	for _, name := range expected {
		if _, ok := flags[name]; !ok {
			t.Errorf("missing registered flag: %s", name)
		}
	}
}

var _ = (*Flags)(nil)
