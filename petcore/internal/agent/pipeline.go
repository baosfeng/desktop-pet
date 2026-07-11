// Package agent — Pipeline 洋葱模型。
package agent

import (
	"context"
	"fmt"

	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/llm"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/tool"
)

// Stage 是 Pipeline 中的一个处理阶段。
type Stage interface {
	Name() string
	Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error
}

// StageFunc 是下一阶段的调用函数。
type StageFunc func(ctx context.Context, pCtx *pipelineCtx) error

// Pipeline 将多个 Stage 串联成洋葱模型。
type Pipeline struct {
	stages []Stage
}

// NewPipeline 创建一个由给定 stages 组成的管线。
// stages 按顺序执行，每个 stage 调用 next 进入下一个 stage。
func NewPipeline(stages ...Stage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Process 按洋葱模型执行所有 Stage。
func (p *Pipeline) Process(ctx context.Context, pCtx *pipelineCtx) error {
	return p.build(0)(ctx, pCtx)
}

func (p *Pipeline) build(index int) StageFunc {
	return func(ctx context.Context, pCtx *pipelineCtx) error {
		if index >= len(p.stages) {
			return nil
		}
		stage := p.stages[index]
		return stage.Process(ctx, pCtx, p.build(index+1))
	}
}

// ─── 内置 Stage ──────────────────────────────

// PreProcessStage 消息标准化、过滤。
type PreProcessStage struct{}

// Name 返回阶段名称。
func (s *PreProcessStage) Name() string { return "PreProcess" }

// Process 执行消息标准化、过滤等预处理。
func (s *PreProcessStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	// TODO: Phase 2 — 敏感词过滤、消息长度限制
	return next(ctx, pCtx)
}

// MemoryStage 注入相关记忆到上下文。
type MemoryStage struct {
	memory memory.Manager
}

// Name 返回阶段名称。
func (s *MemoryStage) Name() string { return "Memory" }

// Process 注入相关记忆到上下文。
func (s *MemoryStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	// TODO: Phase 2 — 从 memory 检索短期 + 长期记忆，注入到 SystemPrompt/Messages
	return next(ctx, pCtx)
}

// LLMCallStage 调用 LLM Provider，处理流式回复和工具调用。
type LLMCallStage struct {
	provider llm.Provider
	sink     event.Sink
	tool     tool.Registry
}

// Name 返回阶段名称。
func (s *LLMCallStage) Name() string { return "LLMCall" }

// Process 调用 LLM Provider，处理流式回复和工具调用。
func (s *LLMCallStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	// 构建 LLM Request
	req := llm.Request{
		Messages:     pCtx.Request.Messages,
		SystemPrompt: pCtx.SystemPrompt,
	}

	// 发送思考事件
	_ = pCtx.Sink.Send(event.Event{
		Kind: event.EventAgentThinking,
		Data: map[string]bool{"status": true},
		Meta: event.NewMeta("agent", ""),
	})

	stream, err := s.provider.Stream(ctx, req)
	if err != nil {
		return fmt.Errorf("agent: LLM stream error: %w", err)
	}

	for chunk := range stream {
		switch chunk.Type {
		case llm.ChunkText:
			_ = pCtx.Sink.Send(event.Event{
				Kind: event.EventAgentReply,
				Data: map[string]any{"text": chunk.Text, "done": false},
			})
		case llm.ChunkToolCall:
			// TODO: Phase 4 — 执行工具调用并继续
		case llm.ChunkDone:
			_ = pCtx.Sink.Send(event.Event{
				Kind: event.EventAgentReply,
				Data: map[string]any{"done": true},
			})
		}
	}

	_ = pCtx.Sink.Send(event.Event{
		Kind: event.EventAgentThinking,
		Data: map[string]bool{"status": false},
	})

	return next(ctx, pCtx)
}

// PostProcessStage 后处理：事实提取等。
type PostProcessStage struct {
	memory memory.Manager
	sink   event.Sink
}

// Name 返回阶段名称。
func (s *PostProcessStage) Name() string { return "PostProcess" }

// Process 执行后处理：事实提取等。
func (s *PostProcessStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	// TODO: Phase 2 — 提取新事实并写入记忆
	return next(ctx, pCtx)
}
