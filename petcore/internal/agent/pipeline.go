// Package agent — Pipeline 洋葱模型。
package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/feature"
	"github.com/desktop-pet/petcore/internal/llm"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/tool"
)

// ─── Stage 定义 ──────────────────────────────

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

// ─── PreProcessStage ─────────────────────────

// PreProcessStage 消息标准化、过滤。
type PreProcessStage struct{}

// Name 返回阶段名称。
func (s *PreProcessStage) Name() string { return "PreProcess" }

// Process 执行消息标准化、过滤等预处理。
func (s *PreProcessStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	// 过滤空消息
	var filtered []llm.Message
	for _, msg := range pCtx.Request.Messages {
		if msg.Content != "" || hasToolCalls(msg) {
			filtered = append(filtered, msg)
		}
	}
	pCtx.Request.Messages = filtered

	return next(ctx, pCtx)
}

func hasToolCalls(_ llm.Message) bool {
	return false // Placeholder: Phase 4+ tool call message tracking
}

// ─── MemoryStage ─────────────────────────────

// MemoryStage 注入相关记忆到上下文。
type MemoryStage struct {
	memory memory.Manager
	flags  *feature.Flags
}

// Name 返回阶段名称。
func (s *MemoryStage) Name() string { return "Memory" }

// Process 从记忆系统检索相关信息并注入到 system prompt。
func (s *MemoryStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	if s.memory == nil || s.flags == nil || !s.flags.IsEnabled(feature.FlagMemoryStage) {
		return next(ctx, pCtx)
	}

	// 1. 注入短期记忆（最近对话）
	shortTerm := s.memory.GetShortTerm()
	if len(shortTerm) > 0 {
		memBlock := "\n\n## 近期对话记录\n"
		for _, msg := range shortTerm {
			role := "用户"
			if msg.Role == "assistant" {
				role = "你"
			}
			memBlock += fmt.Sprintf("%s: %s\n", role, msg.Content)
		}
		pCtx.SystemPrompt += memBlock
	}

	// 2. 注入核心记忆（用户已知事实）
	coreMem, _ := s.memory.GetAllCore()
	if len(coreMem) > 0 {
		factBlock := "\n\n## 你已知的关于用户的事实\n"
		for key, val := range coreMem {
			factBlock += fmt.Sprintf("- %s: %s\n", key, val)
		}
		pCtx.SystemPrompt += factBlock
	}

	return next(ctx, pCtx)
}

// ─── LLMCallStage ────────────────────────────

// LLMCallStage 调用 LLM Provider，处理流式回复和工具调用。
type LLMCallStage struct {
	provider llm.Provider
	sink     event.Sink
	tool     tool.Registry
	flags    *feature.Flags
}

// Name 返回阶段名称。
func (s *LLMCallStage) Name() string { return "LLMCall" }

// Process 调用 LLM Provider，处理流式回复和工具调用循环。
//
//nolint:cyclop
func (s *LLMCallStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	maxTurns := 10
	enableToolLoop := s.flags != nil && s.flags.IsEnabled(feature.FlagToolLoop)
	messages := pCtx.Request.Messages

	for turn := 0; turn < maxTurns; turn++ {
		// 构建 LLM Request
		req := llm.Request{
			Messages:     messages,
			SystemPrompt: pCtx.SystemPrompt,
		}

		// 如果开启了工具循环且存在已注册的工具，注入工具定义
		if enableToolLoop && s.tool != nil {
			tools := s.tool.List()
			for _, t := range tools {
				req.Tools = append(req.Tools, llm.Tool{
					Name:        t.Name(),
					Description: t.Description(),
					Schema:      nil, // TODO: Phase 4 — 完善的 schema
				})
			}
		}

		// 发送思考事件
		if err := pCtx.Sink.Send(event.Event{
			Kind: event.EventAgentThinking,
			Data: map[string]bool{"status": true},
			Meta: event.NewMeta("agent", ""),
		}); err != nil {
			slog.Default().Error("pipeline: send thinking event failed", "error", err)
		}

		stream, err := s.provider.Stream(ctx, req)
		if err != nil {
			return fmt.Errorf("agent: LLM stream error: %w", err)
		}

		var collectedToolCalls []llm.ToolCall
		hasTools := false

		for chunk := range stream {
			switch chunk.Type {
			case llm.ChunkText:
				if err := pCtx.Sink.Send(event.Event{
					Kind: event.EventAgentReply,
					Data: map[string]any{"text": chunk.Text, "done": false},
				}); err != nil {
					slog.Default().Error("pipeline: send reply chunk failed", "error", err)
				}
				hasTools = false
			case llm.ChunkToolCall:
				if chunk.ToolCall != nil {
					collectedToolCalls = append(collectedToolCalls, *chunk.ToolCall)
					hasTools = true
				}
			case llm.ChunkUsage:
				if err := pCtx.Sink.Send(event.Event{
					Kind: event.EventMemoryUpdated,
					Data: map[string]any{"usage": chunk.Usage},
				}); err != nil {
					slog.Default().Error("pipeline: send usage event failed", "error", err)
				}
			case llm.ChunkError:
				if err := pCtx.Sink.Send(event.Event{
					Kind: event.EventError,
					Data: map[string]string{"error": chunk.Error.Error()},
				}); err != nil {
					slog.Default().Error("pipeline: send error event failed", "error", err)
				}
				return chunk.Error
			case llm.ChunkDone:
				// Stream 结束
			}
		}

		if err := pCtx.Sink.Send(event.Event{
			Kind: event.EventAgentThinking,
			Data: map[string]bool{"status": false},
		}); err != nil {
			slog.Default().Error("pipeline: send thinking-done event failed", "error", err)
		}

		// 没有工具调用，或工具循环未启用 → 结束
		if !hasTools || len(collectedToolCalls) == 0 || !enableToolLoop {
			break
		}

		// 执行工具调用并收集结果
		for _, tc := range collectedToolCalls {
			if s.tool == nil {
				continue
			}
			result, err := s.tool.Execute(ctx, tc.Tool, tc.Args)
			if err != nil {
				result = fmt.Sprintf("tool %q error: %v", tc.Tool, err)
			}

			// 将工具结果作为新消息添加到对话中
			messages = append(messages,
				llm.Message{Role: "assistant", Content: fmt.Sprintf("调用工具: %s", tc.Tool)},
				llm.Message{Role: "tool", Content: fmt.Sprintf("toolcall:%s\n%s", tc.ID, result)},
			)
		}

		// 清理工具调用，继续下一轮 LLM 调用
		collectedToolCalls = nil
	}

	// 更新 pCtx 中的消息列表供后续阶段使用
	pCtx.Request.Messages = messages

	return next(ctx, pCtx)
}

// ─── PostProcessStage ────────────────────────

// PostProcessStage 后处理：将对话写入短期记忆。
type PostProcessStage struct {
	memory memory.Manager
	sink   event.Sink
	flags  *feature.Flags
}

// Name 返回阶段名称。
func (s *PostProcessStage) Name() string { return "PostProcess" }

// Process 将对话保存到短期记忆，并发送记忆更新事件。
func (s *PostProcessStage) Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error {
	if s.memory == nil {
		return next(ctx, pCtx)
	}

	// 将用户消息和 AI 回复写入短期记忆
	for _, msg := range pCtx.Request.Messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			s.memory.AddShortTerm(memory.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 发送记忆更新事件
	if err := pCtx.Sink.Send(event.Event{
		Kind: event.EventMemoryUpdated,
		Data: map[string]any{
			"short_term_count": len(s.memory.GetShortTerm()),
		},
		Meta: event.NewMeta("agent", ""),
	}); err != nil {
		slog.Default().Error("pipeline: send memory-updated event failed", "error", err)
	}

	return next(ctx, pCtx)
}
