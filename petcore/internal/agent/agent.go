// Package agent 提供 AI Agent 和 Pipeline 管线。
//
// Agent 是对话管理的核心，连接 LLM Provider、记忆系统和工具系统。
// Pipeline 使用洋葱模型，每个 Stage 可独立测试和替换。
package agent

import (
	"context"
	"fmt"

	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/llm"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/tool"
)

// Agent 是 AI Agent 的主接口。
type Agent interface {
	// Run 执行一轮完整的 AI 对话处理。
	// 包含：上下文构建 → LLM 调用 → 工具执行循环 → 后处理。
	Run(ctx context.Context, req Request) error

	// SetSink 设置事件输出 Sink（Agent 通过 Sink 发送流式回复等事件）。
	SetSink(sink event.Sink)
}

// Request 是 Agent 的输入请求。
type Request struct {
	Messages    []llm.Message
	SystemPrompt string
}

// Option 是 Agent 的构造函数选项。
type Option func(*agentImpl)

// WithMemory 注入记忆系统。
func WithMemory(m memory.Manager) Option {
	return func(a *agentImpl) { a.memory = m }
}

// WithToolRegistry 注入工具注册中心。
func WithToolRegistry(r tool.Registry) Option {
	return func(a *agentImpl) { a.tool = r }
}

// New 创建一个新的 Agent 实例。
// provider 是必需的，其他依赖通过 Option 注入。
func New(provider llm.Provider, opts ...Option) Agent {
	a := &agentImpl{
		provider: provider,
		sink:     event.NoopSink{},
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// ─── Agent 实现 ──────────────────────────────

type agentImpl struct {
	provider llm.Provider
	memory   memory.Manager
	tool     tool.Registry
	sink     event.Sink
	maxTurns int
}

func (a *agentImpl) SetSink(sink event.Sink) {
	if sink == nil {
		sink = event.NoopSink{}
	}
	a.sink = sink
}

func (a *agentImpl) Run(ctx context.Context, req Request) error {
	if a.provider == nil {
		return fmt.Errorf("agent: no LLM provider configured")
	}

	// Pipeline 各阶段串联
	pipeline := NewPipeline(
		&PreProcessStage{},
		&MemoryStage{memory: a.memory},
		&LLMCallStage{provider: a.provider, sink: a.sink, tool: a.tool},
		&PostProcessStage{memory: a.memory, sink: a.sink},
	)

	// 构造内部上下文
	pCtx := &pipelineCtx{
		Request: req,
		Sink:    a.sink,
	}

	return pipeline.Process(ctx, pCtx)
}

// ─── Pipeline 上下文 ─────────────────────────

type pipelineCtx struct {
	Request
	Sink event.Sink
}

func (c *pipelineCtx) Messages() []llm.Message {
	return c.Request.Messages
}
