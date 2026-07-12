// Package ollama 提供 Ollama 本地 LLM 的 Provider 实现。
//
// 在 init() 中自动注册为 "ollama" Provider。
// 调用 Ollama 的本地 REST API（默认 http://localhost:11434）。
//
// 配置方式（通过 llm.NewProvider 传入 cfg map）：
//   - model:    模型名称（默认 qwen2.5:7b）
//   - base_url: Ollama 服务地址（默认 http://localhost:11434）
package ollama

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
	llm.Register("ollama", func(cfg map[string]any) (llm.Provider, error) {
		p := &Provider{
			model:   "qwen2.5:7b",
			baseURL: "http://localhost:11434",
		}
		if cfg != nil {
			if v, ok := cfg["model"].(string); ok && v != "" {
				p.model = v
			}
			if v, ok := cfg["base_url"].(string); ok && v != "" {
				p.baseURL = strings.TrimRight(v, "/")
			}
		}
		return p, nil
	})
}

// Provider 是 Ollama 的 LLM Provider。
type Provider struct {
	model   string
	baseURL string
	client  *http.Client
}

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "ollama" }

// Model 返回当前模型名称。
func (p *Provider) Model() string { return p.model }

// httpClient 返回带超时的 HTTP 客户端（懒初始化）。
func (p *Provider) httpClient() *http.Client {
	if p.client == nil {
		p.client = &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				ResponseHeaderTimeout: 60 * time.Second,
			},
		}
	}
	return p.client
}

// ─── Ollama API 结构 ─────────────────────────

// ollamaChatRequest 是 Ollama 聊天 API 的请求体。
// 参考: https://github.com/ollama/ollama/blob/main/docs/api.md
type ollamaChatRequest struct {
	Model       string           `json:"model"`
	Messages    []ollamaMessage  `json:"messages"`
	Stream      bool             `json:"stream"`
	Temperature float64          `json:"temperature,omitempty"`
	Tools       []llm.Tool       `json:"tools,omitempty"`
}

type ollamaMessage struct {
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	ToolCalls []toolCall  `json:"tool_calls,omitempty"`
}

type toolCall struct {
	Function toolCallFunc `json:"function"`
}

type toolCallFunc struct {
	Name      string `json:"name"`
	Arguments any    `json:"arguments"`
}

// ollamaChatResponse 是 Ollama 聊天 API 的流式响应片段。
type ollamaChatResponse struct {
	Model       string         `json:"model"`
	CreatedAt   string         `json:"created_at"`
	Message     *ollamaMessage `json:"message,omitempty"`
	Done        bool           `json:"done"`
	DoneReason  string         `json:"done_reason,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// ─── Provider 接口实现 ───────────────────────

// Stream 执行流式对话并返回结果片段通道。
func (p *Provider) Stream(ctx context.Context, req llm.Request) (<-chan llm.Chunk, error) {
	oReq := p.buildRequest(req, true)
	body, err := json.Marshal(oReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.httpClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: http request: %w", err)
	}

	ch := make(chan llm.Chunk, 32)

	go func() {
		defer close(ch)
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			p.parseError(ch, httpResp)
			return
		}

		scanner := bufio.NewScanner(httpResp.Body)
		scanner.Buffer(make([]byte, 1024*64), 1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var oResp ollamaChatResponse
			if err := json.Unmarshal([]byte(line), &oResp); err != nil {
				ch <- llm.Chunk{
					Type:  llm.ChunkError,
					Error: fmt.Errorf("ollama: parse response: %w", err),
				}
				return
			}

			if oResp.Error != "" {
				ch <- llm.Chunk{
					Type:  llm.ChunkError,
					Error: fmt.Errorf("ollama: %s", oResp.Error),
				}
				return
			}

			if oResp.Message != nil {
				// 文本内容
				if oResp.Message.Content != "" {
					ch <- llm.Chunk{Type: llm.ChunkText, Text: oResp.Message.Content}
				}
				// 工具调用
				if len(oResp.Message.ToolCalls) > 0 {
					for _, tc := range oResp.Message.ToolCalls {
						argsJSON, _ := json.Marshal(tc.Function.Arguments)
						ch <- llm.Chunk{
							Type: llm.ChunkToolCall,
							ToolCall: &llm.ToolCall{
								ID:   fmt.Sprintf("ollama_%d", time.Now().UnixNano()),
								Tool: tc.Function.Name,
								Args: string(argsJSON),
							},
						}
					}
				}
			}

			if oResp.Done {
				ch <- llm.Chunk{Type: llm.ChunkDone}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- llm.Chunk{
				Type:  llm.ChunkError,
				Error: fmt.Errorf("ollama: scan: %w", err),
			}
		}
	}()

	return ch, nil
}

// Chat 执行非流式对话并返回完整响应。
func (p *Provider) Chat(ctx context.Context, req llm.Request) (llm.Response, error) {
	oReq := p.buildRequest(req, false)
	body, err := json.Marshal(oReq)
	if err != nil {
		return llm.Response{}, fmt.Errorf("ollama: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return llm.Response{}, fmt.Errorf("ollama: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.httpClient().Do(httpReq)
	if err != nil {
		return llm.Response{}, fmt.Errorf("ollama: http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return llm.Response{}, fmt.Errorf("ollama: read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return llm.Response{}, fmt.Errorf("ollama: HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	var oResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &oResp); err != nil {
		return llm.Response{}, fmt.Errorf("ollama: parse response: %w", err)
	}

	if oResp.Error != "" {
		return llm.Response{}, fmt.Errorf("ollama: %s", oResp.Error)
	}

	result := llm.Response{
		Content: "",
	}

	if oResp.Message != nil {
		result.Content = oResp.Message.Content
		for _, tc := range oResp.Message.ToolCalls {
			argsJSON, _ := json.Marshal(tc.Function.Arguments)
			result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
				ID:   fmt.Sprintf("ollama_%d", time.Now().UnixNano()),
				Tool: tc.Function.Name,
				Args: string(argsJSON),
			})
		}
	}

	return result, nil
}

// ─── 内部辅助方法 ────────────────────────────

func (p *Provider) buildRequest(req llm.Request, stream bool) ollamaChatRequest {
	oReq := ollamaChatRequest{
		Model:       p.model,
		Stream:      stream,
		Temperature: req.Temperature,
	}

	if req.SystemPrompt != "" {
		oReq.Messages = append(oReq.Messages, ollamaMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	for _, msg := range req.Messages {
		oReq.Messages = append(oReq.Messages, ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if len(req.Tools) > 0 {
		oReq.Tools = req.Tools
	}

	return oReq
}

func (p *Provider) parseError(ch chan<- llm.Chunk, resp *http.Response) {
	body, _ := io.ReadAll(resp.Body)
	errMsg := fmt.Sprintf("ollama: HTTP %d", resp.StatusCode)
	// Try to parse Ollama error response
	var oResp ollamaChatResponse
	if json.Unmarshal(body, &oResp) == nil && oResp.Error != "" {
		errMsg = fmt.Sprintf("ollama: %s", oResp.Error)
	}
	ch <- llm.Chunk{
		Type:  llm.ChunkError,
		Error: fmt.Errorf("%s", errMsg),
	}
}
