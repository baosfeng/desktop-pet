# LLM 适配层 — 实现规范

> **子 Agent 任务 ID:** petcore-004-llm
> **包路径:** `petcore/internal/llm/`
> **依赖:** 无（基础包不依赖其他 PetCore 内部模块）
> **子包:** `llm/mock/`、`llm/openai/`、`llm/ollama/`
> **被依赖:** agent（通过 Provider 接口）

---

## 1. 模块职责

LLM 适配层是 PetCore 与大语言模型之间的桥梁。采用 Provider 接口 + 注册表模式，各 Provider 在各自包的 `init()` 中自注册。

## 2. 接口契约

### 2.1 Provider 接口

```go
type Provider interface {
    Name() string
    Model() string
    Stream(ctx context.Context, req Request) (<-chan Chunk, error)
    Chat(ctx context.Context, req Request) (Response, error)
}
```

### 2.2 数据结构

```go
type Request struct {
    Messages     []Message
    SystemPrompt string
    Tools        []Tool
    Temperature  float64
    MaxTokens    int
}

type Message struct {
    Role    string   // system / user / assistant / tool
    Content string
    Images  []Image  // 可选：多模态
}

type Chunk struct {
    Type     ChunkType  // Text / Reasoning / ToolCall / Usage / Done / Error
    Text     string
    ToolCall *ToolCall
    Usage    *Usage
    Error    error
}

type Response struct {
    Content   string
    ToolCalls []ToolCall
    Usage     *Usage
}
```

### 2.3 注册表

```go
func Register(name string, factory ProviderFactory)  // 在 init() 中调用
func NewProvider(name string, cfg map[string]any) (Provider, error)
func RegisteredProviders() []string
```

## 3. 内置 Provider 清单

| Provider | 包路径 | 状态 | 说明 |
|----------|--------|------|------|
| `mock` | `llm/mock/mock.go` | ✅ 完整 | init() 自注册，环境变量 `MOCK_LLM_REPLY`/`_DELAY`/`_SHOW_REASONING` 控制行为 |
| `openai` | `llm/openai/openai.go` | ⏳ 骨架 | init() 已注册，返回 `not yet implemented` |
| `ollama` | `llm/ollama/ollama.go` | ⏳ 骨架 | init() 已注册，返回 `not yet implemented` |

## 4. 已实现内容

| 文件 | 状态 | 内容 |
|------|------|------|
| `llm/provider.go` | ✅ 完整 | Provider 接口、所有数据结构、注册表 |
| `llm/mock/mock.go` | ✅ 完整 | Mock Provider 完整实现 |

## 5. 待实现

- [ ] **OpenAI Provider** — 实现 `Stream()` 和 `Chat()`，使用 `sashabaranov/go-openai` 或标准 HTTP
- [ ] **Ollama Provider** — 实现 `Stream()` 和 `Chat()`，调用 Ollama REST API

## 6. Mock Provider 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MOCK_LLM_REPLY` | `"[mock] Hello!"` | 固定回复内容 |
| `MOCK_LLM_DELAY` | `0` | 模拟延迟（毫秒） |
| `MOCK_LLM_SHOW_REASONING` | `false` | 设为 `"true"` 显示推理过程 |

## 7. 子 Agent 验收命令

```bash
cd petcore
go build ./internal/llm/...
go test -v -count=1 ./internal/llm/...        # provider.go 无测试文件，仅验证编译
go test -v -count=1 ./internal/llm/mock/...   # mock 包测试
```

## 8. 被其他模块使用的方式

```go
import "github.com/desktop-pet/petcore/internal/llm"
import _ "github.com/desktop-pet/petcore/internal/llm/mock"  // blank import 注册

provider, err := llm.NewProvider("mock", nil)
// provider.Name() == "mock"
// provider.Model() == "mock-v1"

// 在 agent.Agent 中使用：
a := agent.New(provider, ...)
```
