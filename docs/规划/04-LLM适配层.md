# 模块规划：LLM 适配层 (petcore/internal/llm)

> **前置依赖：** 无（不依赖其他 PetCore 内部包）
> **被以下模块依赖：** agent（通过 Provider 接口）
> **对应 Phase：** Phase 1 — MVP

---

## 1. 模块定位

LLM 适配层是 PetCore 与大语言模型的桥梁。采用 **Provider 接口 + 注册表模式**，各 LLM 服务商在各自包的 `init()` 中自注册，主程序通过 blank import 启用。

## 2. 接口与边界

### 对外暴露

```go
// Provider — 所有 LLM 服务商的统一接口
type Provider interface {
    Name() string
    Model() string
    Stream(ctx, Request) (<-chan Chunk, error)
    Chat(ctx, Request) (Response, error)
}

// 注册表
func Register(name string, factory ProviderFactory)    // 各 Provider 在 init() 中调用
func NewProvider(name string, cfg) (Provider, error)  // 按名称创建
func RegisteredProviders() []string                    // 查询已注册

// Request / Response / Chunk — 统一的数据结构
type Request struct {
    Messages     []Message
    SystemPrompt string
    Tools        []Tool
    Temperature  float64
    MaxTokens    int
}
```

### 内置 Provider

| Provider | 文件 | 状态 | 说明 |
|----------|------|------|------|
| `mock` | `llm/mock/mock.go` | ✅ 完整实现 | 环境变量控制回复/延迟/推理过程 |
| `openai` | `llm/openai/openai.go` | ⏳ 骨架 | `init()` 已注册，调用返回 `not yet implemented` |
| `ollama` | `llm/ollama/ollama.go` | ⏳ 骨架 | 同上 |

## 3. 可拆卸性设计

| 机制 | 实现 |
|------|------|
| **接口隔离** | `Provider` 接口仅 4 个方法，极简 |
| **注册表模式** | `init()` 自注册，新增 Provider 只需新建包 + blank import |
| **零依赖层** | `llm` 基础包不依赖任何 PetCore 内部模块 |
| **Mock Provider** | `llm/mock` 完整实现，环境变量可控，所有模块测试都依赖它 |
| **编译时检查** | `var _ llm.Provider = (*Provider)(nil)` |

## 4. 已完成的代码

| 文件 | 行数 | 内容 |
|------|------|------|
| `llm/provider.go` | ~135 行 | Provider 接口、Request/Response/Chunk/Message 类型、注册表 |
| `llm/mock/mock.go` | ~110 行 | Mock Provider，环境变量控制回复/延迟/推理展示 |
| `llm/openai/openai.go` | ~18 行 | OpenAI 注册骨架 |
| `llm/ollama/ollama.go` | ~18 行 | Ollama 注册骨架 |

## 5. 剩余工作任务

- [ ] **Phase 1**：实现 OpenAI Provider 的 `Stream()` 方法（使用 `sashabaranov/go-openai` 或标准 HTTP）
- [ ] **Phase 1**：实现 Ollama Provider 的 `Stream()` 方法
- [ ] **Phase 2**：为 OpenAI/Ollama Provider 编写单元测试

## 6. 验收标准

- [x] `go build ./internal/llm/...` 通过
- [x] `go test ./internal/llm/...` 全部通过
- [x] Provider 接口定义完整（Name/Model/Stream/Chat）
- [x] 注册表支持新增 Provider 无需修改核心代码
- [x] Mock Provider 可通过环境变量配置
- [x] OpenAI/Ollama 骨架已注册（调用时返回明确错误）
