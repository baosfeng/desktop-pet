// Package openai 提供 OpenAI 兼容 API 的 LLM Provider。
//
// 支持所有 OpenAI 兼容格式的 API 服务商，包括但不限于：
//   - OpenAI（默认）
//   - DeepSeek（配置 base_url=https://api.deepseek.com/v1，或使用 "deepseek" provider）
//   - 硅基流动 / 通义千问 / Groq / Ollama（兼容模式）等任意兼容 API
//
// 使用 "openai" provider 时，通过 base_url 指向不同服务商即可切换。
// 使用 "deepseek" provider 时，默认指向 DeepSeek 官方 API（也支持自定义 base_url）。
//
// 在 init() 中自动注册为 "openai"，deepseek 在同一包中注册。
//
// 配置方式（通过 llm.NewProvider 传入 cfg map）：
//   - model:    模型名称（默认 gpt-4o-mini）
//   - base_url: API 地址（默认 https://api.openai.com/v1）
//   - api_key:  API Key（优先从 LLMConfig.APIKeyEnv 环境变量读取）
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/desktop-pet/petcore/internal/llm"
)

func init() {
	llm.Register("openai", func(cfg map[string]any) (llm.Provider, error) {
		p := &Provider{
			model:   "gpt-4o-mini",
			baseURL: "https://api.openai.com/v1",
		}
		if cfg != nil {
			if v, ok := cfg["model"].(string); ok && v != "" {
				p.model = v
			}
			if v, ok := cfg["base_url"].(string); ok && v != "" {
				p.baseURL = strings.TrimRight(v, "/")
			}
			if v, ok := cfg["api_key"].(string); ok && v != "" {
				p.apiKey = v
			}
		}
		if p.apiKey == "" {
			return nil, fmt.Errorf("openai: API key is required (set PETCORE_API_KEY env or configure api_key_env)")
		}
		return p, nil
	})
}

// Provider 是 OpenAI 兼容 API 的 LLM Provider。
type Provider struct {
	model   string
	baseURL string
	apiKey  string
	client  *http.Client
}

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "openai" }

// Model 返回当前模型名称。
func (p *Provider) Model() string { return p.model }

// httpClient 返回带超时的 HTTP 客户端（懒初始化）。
func (p *Provider) httpClient() *http.Client {
	if p.client == nil {
		p.client = &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				ResponseHeaderTimeout: 30 * time.Second,
			},
		}
	}
	return p.client
}

// ─── 请求/响应结构 ───────────────────────────

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Tools       []llm.Tool    `json:"tools,omitempty"`
}

type chatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function toolCallFunc `json:"function"`
}

type toolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatResponse struct {
	Choices []choice   `json:"choices"`
	Usage   *llm.Usage `json:"usage,omitempty"`
	Error   *apiError  `json:"error,omitempty"`
}

type choice struct {
	Index        int          `json:"index"`
	Delta        delta        `json:"delta,omitempty"`
	Message      *chatMessage `json:"message,omitempty"`
	FinishReason string       `json:"finish_reason"`
}

type delta struct {
	Role             string     `json:"role,omitempty"`
	Content          string     `json:"content,omitempty"`
	ReasoningContent string     `json:"reasoning_content,omitempty"` // DeepSeek reasoning model
	ToolCalls        []toolCall `json:"tool_calls,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// ─── Provider 接口实现 ───────────────────────

// Stream 执行流式对话并返回结果片段通道。
//
//nolint:cyclop
func (p *Provider) Stream(ctx context.Context, req llm.Request) (<-chan llm.Chunk, error) {
	cReq := p.buildRequest(req, true)
	body, err := json.Marshal(cReq)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai: create request: %w", err)
	}
	p.setHeaders(httpReq)

	httpResp, err := p.httpClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: http request: %w", err)
	}

	ch := make(chan llm.Chunk, 32)

	go func() {
		defer close(ch)
		defer func() { _ = httpResp.Body.Close() }()

		if httpResp.StatusCode != http.StatusOK {
			p.parseError(ch, httpResp)
			return
		}

		scanner := bufio.NewScanner(httpResp.Body)
		// SSE lines can be long
		scanner.Buffer(make([]byte, 1024*64), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			// [DONE] 信号
			if data == "[DONE]" {
				ch <- llm.Chunk{Type: llm.ChunkDone}
				return
			}

			var sseResp chatResponse
			if err := json.Unmarshal([]byte(data), &sseResp); err != nil {
				ch <- llm.Chunk{
					Type:  llm.ChunkError,
					Error: fmt.Errorf("openai: parse SSE: %w", err),
				}
				return
			}

			if sseResp.Error != nil {
				ch <- llm.Chunk{
					Type:  llm.ChunkError,
					Error: fmt.Errorf("openai: API error: %s", sseResp.Error.Message),
				}
				return
			}

			if len(sseResp.Choices) == 0 {
				continue
			}
			c := sseResp.Choices[0]

			// 文本增量
			if c.Delta.Content != "" {
				ch <- llm.Chunk{Type: llm.ChunkText, Text: c.Delta.Content}
			}

			// 推理过程（DeepSeek reasoning model 专用字段）
			if c.Delta.ReasoningContent != "" {
				ch <- llm.Chunk{Type: llm.ChunkReasoning, Text: c.Delta.ReasoningContent}
			}

			// 工具调用增量（仅第一个工具调用片段触发）
			if len(c.Delta.ToolCalls) > 0 {
				tc := c.Delta.ToolCalls[0]
				if tc.Function.Name != "" || tc.Function.Arguments != "" {
					ch <- llm.Chunk{
						Type: llm.ChunkToolCall,
						ToolCall: &llm.ToolCall{
							ID:   tc.ID,
							Tool: tc.Function.Name,
							Args: tc.Function.Arguments,
						},
					}
				}
			}

			// Token 用量（通常只在最后一个 chunk 出现）
			if sseResp.Usage != nil {
				ch <- llm.Chunk{Type: llm.ChunkUsage, Usage: sseResp.Usage}
			}

			// Finish reason
			if c.FinishReason == "stop" {
				ch <- llm.Chunk{Type: llm.ChunkDone}
				return
			}
			if c.FinishReason == "tool_calls" {
				ch <- llm.Chunk{Type: llm.ChunkDone}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- llm.Chunk{
				Type:  llm.ChunkError,
				Error: fmt.Errorf("openai: scan SSE: %w", err),
			}
		}
	}()

	return ch, nil
}

// Chat 执行非流式对话并返回完整响应。
//
//nolint:cyclop
func (p *Provider) Chat(ctx context.Context, req llm.Request) (llm.Response, error) {
	cReq := p.buildRequest(req, false)
	body, err := json.Marshal(cReq)
	if err != nil {
		return llm.Response{}, fmt.Errorf("openai: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return llm.Response{}, fmt.Errorf("openai: create request: %w", err)
	}
	p.setHeaders(httpReq)

	httpResp, err := p.httpClient().Do(httpReq)
	if err != nil {
		return llm.Response{}, fmt.Errorf("openai: http request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return llm.Response{}, fmt.Errorf("openai: read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		var apiErr chatResponse
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != nil {
			return llm.Response{}, fmt.Errorf("openai: API error: %s", apiErr.Error.Message)
		}
		return llm.Response{}, fmt.Errorf("openai: HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	var resp chatResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return llm.Response{}, fmt.Errorf("openai: parse response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return llm.Response{}, fmt.Errorf("openai: no choices in response")
	}

	c := resp.Choices[0]
	result := llm.Response{
		Content: "",
		Usage:   resp.Usage,
	}

	if c.Message != nil {
		result.Content = c.Message.Content
		for _, tc := range c.Message.ToolCalls {
			result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
				ID:   tc.ID,
				Tool: tc.Function.Name,
				Args: tc.Function.Arguments,
			})
		}
	}

	return result, nil
}

// ─── 内部辅助方法 ────────────────────────────

func (p *Provider) buildRequest(req llm.Request, stream bool) chatRequest {
	cReq := chatRequest{
		Model:       p.model,
		Stream:      stream,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	// System prompt
	if req.SystemPrompt != "" {
		cReq.Messages = append(cReq.Messages, chatMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	// Messages
	for _, msg := range req.Messages {
		m := chatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		// Tool call results
		if msg.Role == "tool" {
			m.ToolCallID = extractToolCallID(msg.Content)
		}
		cReq.Messages = append(cReq.Messages, m)
	}

	// Tools
	if len(req.Tools) > 0 {
		cReq.Tools = req.Tools
	}

	return cReq
}

func (p *Provider) setHeaders(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.apiKey)
	r.Header.Set("Accept", "text/event-stream")
}

func (p *Provider) parseError(ch chan<- llm.Chunk, resp *http.Response) {
	body, _ := io.ReadAll(resp.Body)
	var apiErr chatResponse
	errMsg := fmt.Sprintf("openai: HTTP %d", resp.StatusCode)
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != nil {
		errMsg = fmt.Sprintf("openai: API error (%d): %s", resp.StatusCode, apiErr.Error.Message)
	}
	ch <- llm.Chunk{
		Type:  llm.ChunkError,
		Error: fmt.Errorf("%s", errMsg),
	}
}

// extractToolCallID 从 tool role 消息中提取 tool_call_id（简单的占位逻辑）。
// 实际使用中，前端应在消息体中携带 tool_call_id。
func extractToolCallID(content string) string {
	// 如果 content 以 toolcall: 开头，提取后面的 ID
	if strings.HasPrefix(content, "toolcall:") {
		return strings.TrimPrefix(content, "toolcall:")
	}
	// 默认用 content 的简短哈希作为 ID
	return fmt.Sprintf("tool_%d", len(content))
}
