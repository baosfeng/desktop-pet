package plugin

import (
	"context"
	"testing"
)

func TestTypeString(t *testing.T) {
	tests := []struct {
		t    Type
		want string
	}{
		{TypeL1YAML, "L1/yaml"},
		{TypeL2JS, "L2/js"},
		{TypeL3MCP, "L3/mcp"},
		{Type(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("Type.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewRegistry_Empty(t *testing.T) {
	r := NewRegistry()
	if list := r.List(); len(list) != 0 {
		t.Errorf("expected empty registry, got %d plugins", len(list))
	}
}

func TestRegistry_RegisterAndList(t *testing.T) {
	r := NewRegistry()
	p := &MockPlugin{MockName: "greeter", MockType: TypeL1YAML}
	if err := r.Register(p); err != nil {
		t.Fatal(err)
	}
	list := r.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(list))
	}
	if list[0].Name() != "greeter" {
		t.Errorf("plugin name = %q, want %q", list[0].Name(), "greeter")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	p := &MockPlugin{MockName: "dup"}
	_ = r.Register(p)
	err := r.Register(p)
	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(&MockPlugin{MockName: "test"})
	p, err := r.Get("test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "test" {
		t.Errorf("got name %q, want %q", p.Name(), "test")
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Error("expected error for unknown plugin")
	}
}

func TestRegistry_StartAll_StopAll(t *testing.T) {
	r := NewRegistry()
	started := false
	stopped := false

	_ = r.Register(&MockPlugin{
		MockName: "test",
		OnStart: func(_ context.Context) error {
			started = true
			return nil
		},
		OnStop: func() error {
			stopped = true
			return nil
		},
	})

	ctx := context.Background()
	if err := r.StartAll(ctx); err != nil {
		t.Fatal(err)
	}
	if !started {
		t.Error("expected plugin to start")
	}

	if err := r.StopAll(); err != nil {
		t.Fatal(err)
	}
	if !stopped {
		t.Error("expected plugin to stop")
	}
}

func TestRegistryInterface(_ *testing.T) {
	var _ Registry = (*builtinRegistry)(nil)
}
