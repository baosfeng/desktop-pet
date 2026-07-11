// Package plugin 提供三层插件系统。
//
// Plugin 接口统一 L1（YAML 动作包）、L2（JS 脚本）、L3（MCP 子进程）。
// Registry 管理插件的注册、发现、生命周期。
package plugin

import (
	"context"
	"fmt"
	"sync"
)

// Type 表示插件类型。
type Type int

const (
	TypeL1YAML Type = iota // 声明式 YAML 动作包
	TypeL2JS               // JS 脚本引擎
	TypeL3MCP              // MCP 子进程
)

func (t Type) String() string {
	switch t {
	case TypeL1YAML:
		return "L1/yaml"
	case TypeL2JS:
		return "L2/js"
	case TypeL3MCP:
		return "L3/mcp"
	default:
		return "unknown"
	}
}

// Plugin 是单个插件的接口。
type Plugin interface {
	Name() string
	Type() Type
	Start(ctx context.Context) error
	Stop() error
}

// Registry 是插件注册中心的接口。
type Registry interface {
	Register(p Plugin) error
	List() []Plugin
	Get(name string) (Plugin, error)
	StartAll(ctx context.Context) error
	StopAll() error
}

// Ensure builtinRegistry implements Registry.
var _ Registry = (*builtinRegistry)(nil)

type builtinRegistry struct {
	mu   sync.RWMutex
	plugins map[string]Plugin
}

// NewRegistry 创建一个空的插件注册中心。
func NewRegistry() Registry {
	return &builtinRegistry{
		plugins: make(map[string]Plugin),
	}
}

func (r *builtinRegistry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := p.Name()
	if _, ok := r.plugins[name]; ok {
		return fmt.Errorf("plugin: already registered: %s", name)
	}
	r.plugins[name] = p
	return nil
}

func (r *builtinRegistry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		out = append(out, p)
	}
	return out
}

func (r *builtinRegistry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin: not found: %s", name)
	}
	return p, nil
}

func (r *builtinRegistry) StartAll(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.plugins {
		if err := p.Start(ctx); err != nil {
			return fmt.Errorf("plugin: start %s: %w", p.Name(), err)
		}
	}
	return nil
}

func (r *builtinRegistry) StopAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.plugins {
		if err := p.Stop(); err != nil {
			return fmt.Errorf("plugin: stop %s: %w", p.Name(), err)
		}
	}
	return nil
}

// ─── Mock Plugin 用于测试 ────────────────────

// MockPlugin 是一个可配置的 mock 插件，供其他模块测试使用。
type MockPlugin struct {
	MockName string
	MockType Type
	OnStart  func(ctx context.Context) error
	OnStop   func() error
}

func (m *MockPlugin) Name() string                     { return m.MockName }
func (m *MockPlugin) Type() Type                       { return m.MockType }
func (m *MockPlugin) Start(ctx context.Context) error {
	if m.OnStart != nil {
		return m.OnStart(ctx)
	}
	return nil
}
func (m *MockPlugin) Stop() error {
	if m.OnStop != nil {
		return m.OnStop()
	}
	return nil
}
