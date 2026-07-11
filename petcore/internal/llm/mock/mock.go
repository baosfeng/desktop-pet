// Package mock 提供本地 Mock LLM Provider，用于离线开发和测试。
//
// 使用方式：
//
//	import "github.com/desktop-pet/petcore/internal/llm/mock"
//	// mock 在 init() 中自动注册为 "mock" Provider
//
//	provider, _ := llm.NewProvider("mock", nil)
//	stream, _ := provider.Stream(ctx, req)
//
// Mock Provider 支持通过环境变量自定义回复：
//   - MOCK_LLM_REPLY: 设置固定回复内容（默认英文占位）
//   - MOCK_LLM_DELAY: 模拟延迟，单位毫秒（默认 0）
//   - MOCK_LLM_SHOW_REASONING: 设为 "true" 显示推理过程（默认 false）
package mock

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/desktop-pet/petcore/internal/llm"
)

func init() {
	llm.Register("mock", func(_ map[string]any) (llm.Provider, error) {
		return &Provider{}, nil
	})
}

// Provider 是一个本地模拟 LLM Provider。
type Provider struct{}

func (p *Provider) Name() string { return "mock" }

func (p *Provider) Model() string { return "mock-v1" }

func (p *Provider) Stream(_ context.Context, req llm.Request) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 10)

	go func() {
		defer close(ch)

		if d := getMockDelay(); d > 0 {
			time.Sleep(d)
		}

		reply := os.Getenv("MOCK_LLM_REPLY")
		if reply == "" {
			reply = "[mock] Hello! I am your desktop pet."
		}

		if os.Getenv("MOCK_LLM_SHOW_REASONING") == "true" {
			ch <- llm.Chunk{Type: llm.ChunkReasoning, Text: "[mock] thinking..."}
		}

		for _, r := range reply {
			ch <- llm.Chunk{Type: llm.ChunkText, Text: string(r)}
			time.Sleep(10 * time.Millisecond)
		}

		msgCount := len(req.Messages)
		runeCount := len([]rune(reply))

		ch <- llm.Chunk{Type: llm.ChunkUsage, Usage: &llm.Usage{
			PromptTokens:     msgCount,
			CompletionTokens: runeCount,
			TotalTokens:      msgCount + runeCount,
		}}
		ch <- llm.Chunk{Type: llm.ChunkDone}
	}()

	return ch, nil
}

func (p *Provider) Chat(_ context.Context, req llm.Request) (llm.Response, error) {
	reply := os.Getenv("MOCK_LLM_REPLY")
	if reply == "" {
		reply = "[mock] Hello! I am your desktop pet."
	}

	msgCount := len(req.Messages)
	runeCount := len([]rune(reply))

	return llm.Response{
		Content: reply,
		Usage: &llm.Usage{
			PromptTokens:     msgCount,
			CompletionTokens: runeCount,
			TotalTokens:      msgCount + runeCount,
		},
	}, nil
}

func getMockDelay() time.Duration {
	s := os.Getenv("MOCK_LLM_DELAY")
	if s == "" {
		return 0
	}
	ms, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

// Ensure Provider implements llm.Provider.
var _ llm.Provider = (*Provider)(nil)
