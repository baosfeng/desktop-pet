// Package tool 提供工具定义、注册中心和调度能力。
//
// 工具是 LLM 可以调用的能力单元（如 web_search、remember、desktop_ctrl）。
// 内置工具通过 Register() 注册，MCP 工具通过 MCP 客户端自动发现。
package tool

import (
	"context"
	"fmt"
	"sync"
)

// Tool 是单个工具的接口定义。
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args string) (string, error)
}

// Registry 是工具注册中心的接口。
type Registry interface {
	Register(tool Tool) error
	Execute(ctx context.Context, name, args string) (string, error)
	List() []Tool
}

// Ensure builtinRegistry implements Registry.
var _ Registry = (*builtinRegistry)(nil)

// builtinRegistry 是 Registry 的内存实现。
type builtinRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry 创建一个空的工具注册中心。
func NewRegistry() Registry {
	return &builtinRegistry{
		tools: make(map[string]Tool),
	}
}

func (r *builtinRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := tool.Name()
	if _, ok := r.tools[name]; ok {
		return fmt.Errorf("tool: already registered: %s", name)
	}
	r.tools[name] = tool
	return nil
}

func (r *builtinRegistry) Execute(ctx context.Context, name, args string) (string, error) {
	r.mu.RLock()
	tool, ok := r.tools[name]
	r.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("tool: not found: %s", name)
	}
	return tool.Execute(ctx, args)
}

func (r *builtinRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

// ─── Mock Tool 用于测试 ──────────────────────

// MockTool 是一个可配置的 mock 工具，供其他模块测试使用。
type MockTool struct {
	MockName        string
	MockDescription string
	MockExecute     func(ctx context.Context, args string) (string, error)
}

func (m *MockTool) Name() string                             { return m.MockName }
func (m *MockTool) Description() string                      { return m.MockDescription }
func (m *MockTool) Execute(ctx context.Context, args string) (string, error) { return m.MockExecute(ctx, args) }
