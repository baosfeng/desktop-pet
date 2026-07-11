// Package llm 提供 LLM Provider 接口和注册表。
//
// Provider 接口：
//
//	type Provider interface {
//	    Name() string
//	    Model() string
//	    Stream(ctx, Request) (<-chan Chunk, error)
//	    Chat(ctx, Request) (Response, error)
//	}
//
// 注册表模式：各 Provider 在 init() 中自注册，main.go 通过 blank import 引入。
//
// 内置 Provider：
//   - openai: OpenAI 兼容 API（DeepSeek / 硅基流动 / Ollama）
//   - mock:   本地 Mock，用于离线开发和测试
package llm

import (
	"context"
	"errors"
)

// Provider 是所有 LLM 服务商的统一接口。
type Provider interface {
	Name() string
	Model() string
	Stream(ctx context.Context, req Request) (<-chan Chunk, error)
	Chat(ctx context.Context, req Request) (Response, error)
}

// Request 是 LLM 调用请求参数。
type Request struct {
	Messages     []Message
	SystemPrompt string
	Tools        []Tool
	Temperature  float64
	MaxTokens    int
}

// Message 表示一条对话消息。
type Message struct {
	Role    string // system / user / assistant / tool
	Content string
	Images  []Image // 可选：多模态
}

// Image 表示多模态输入中的图片。
type Image struct {
	MIMEType string
	Data     []byte
	URL      string
}

// Chunk 是流式响应中的一个片段。
type Chunk struct {
	Type     ChunkType // Text / Reasoning / ToolCall / Usage / Done / Error
	Text     string
	ToolCall *ToolCall
	Usage    *Usage
	Error    error
}

// ChunkType 表示流式片段的类型。
type ChunkType int

// 预定义的 Chunk 类型常量。
const (
	ChunkText      ChunkType = iota // 文本片段
	ChunkReasoning                  // 推理过程
	ChunkToolCall                   // 工具调用
	ChunkUsage                      // token 用量
	ChunkDone                       // 流结束
	ChunkError                      // 错误
)

// Response 是非流式对话的完整响应。
type Response struct {
	Content   string
	ToolCalls []ToolCall
	Usage     *Usage
}

// Tool 是可供 LLM 调用的工具定义。
type Tool struct {
	Name        string
	Description string
	Schema      map[string]any
}

// ToolCall 是 LLM 发出的工具调用请求。
type ToolCall struct {
	ID      string
	Tool    string
	Args    string // JSON 字符串
	Waiting bool   // 是否等待执行结果
}

// Usage 表示一次请求的 token 用量。
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ProviderFactory 是 Provider 的构造函数。
type ProviderFactory func(cfg map[string]any) (Provider, error)

var registry = make(map[string]ProviderFactory)

// Register 注册一个 Provider 工厂函数。
// 通常在 Provider 的 init() 中调用。
func Register(name string, factory ProviderFactory) {
	if _, ok := registry[name]; ok {
		panic("llm: provider already registered: " + name)
	}
	registry[name] = factory
}

// NewProvider 根据名称和配置创建 Provider 实例。
func NewProvider(name string, cfg map[string]any) (Provider, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, errors.New("llm: unknown provider: " + name)
	}
	return factory(cfg)
}

// RegisteredProviders 返回所有已注册的 Provider 名称。
func RegisteredProviders() []string {
	providers := make([]string, 0, len(registry))
	for name := range registry {
		providers = append(providers, name)
	}
	return providers
}
