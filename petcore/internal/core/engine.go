// Package core 是 PetCore 的核心引擎。
//
// Engine 通过构造函数注入所有依赖（fsm, agent, memory, plugin, tool, sink, config），
// 绝不 import 具体子模块的实现。这是整个架构模块化程度的关键。
package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/desktop-pet/petcore/internal/agent"
	"github.com/desktop-pet/petcore/internal/config"
	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/fsm"
	"github.com/desktop-pet/petcore/internal/llm"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/plugin"
	"github.com/desktop-pet/petcore/internal/tool"
)

// Engine 是 PetCore 的核心引擎，协调所有子模块。
type Engine struct {
	fsm    fsm.Machine
	agent  agent.Agent
	memory memory.Manager
	plugin plugin.Registry
	tool   tool.Registry
	cfg    *config.Config
	sink   event.Sink
	log    *slog.Logger
}

// New 创建一个新的 Engine 实例。
// 所有依赖通过参数注入——这是保证模块可拆卸性的关键设计。
func New(
	fsm fsm.Machine,
	ag agent.Agent,
	mem memory.Manager,
	pl plugin.Registry,
	tl tool.Registry,
	cfg *config.Config,
	sink event.Sink,
) *Engine {
	if sink == nil {
		sink = event.NoopSink{}
	}

	return &Engine{
		fsm:    fsm,
		agent:  ag,
		memory: mem,
		plugin: pl,
		tool:   tl,
		cfg:    cfg,
		sink:   sink,
		log:    slog.Default(),
	}
}

// Run 启动引擎主循环。
// 这是一个阻塞调用，应放在 goroutine 中运行。
func (e *Engine) Run(ctx context.Context) error {
	// 启动所有插件
	if err := e.plugin.StartAll(ctx); err != nil {
		return fmt.Errorf("engine: plugin start: %w", err)
	}

	// 发送启动事件
	_ = e.sink.Send(event.Event{
		Kind: event.EventStateChanged,
		Data: map[string]any{"state": string(e.fsm.Current())},
		Meta: event.NewMeta("core", ""),
	})

	// 主事件循环（骨架）
	// TODO: Phase 2 — 实现完整的事件循环（接收 stdin 命令、定时任务等）
	<-ctx.Done()

	// 清理
	_ = e.plugin.StopAll()
	_ = e.sink.Close()

	return nil
}

// HandleInput 处理用户输入（对话消息）。
func (e *Engine) HandleInput(ctx context.Context, text string) error {
	if text == "" {
		return nil
	}

	// 状态转换：当前状态 → 交互状态（如果合法）
	if err := e.fsm.Transition(event.EventStateChanged); err != nil {
		// 非法的状态转换不阻塞对话
		e.log.Debug("state transition skipped", "error", err)
	}

	// 发送状态变更事件
	_ = e.sink.Send(event.Event{
		Kind: event.EventStateChanged,
		Data: map[string]any{"state": string(e.fsm.Current())},
		Meta: event.NewMeta("core", ""),
	})

	// 委托给 Agent 处理
	return e.agent.Run(ctx, agent.Request{
		Messages:     []llm.Message{{Role: "user", Content: text}},
		SystemPrompt: e.cfg.Agent.SystemPrompt,
	})
}

// GetStatus 返回引擎当前状态。
func (e *Engine) GetStatus() map[string]any {
	return map[string]any{
		"state":   string(e.fsm.Current()),
		"plugins": len(e.plugin.List()),
		"tools":   len(e.tool.List()),
	}
}
