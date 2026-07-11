# 模块规划：AI Agent (petcore/internal/agent)

> **前置依赖：** llm（Provider 接口）、memory（Manager 接口）、tool（Registry 接口）
> **被以下模块依赖：** core（通过 Agent 接口）
> **对应 Phase：** Phase 1 — MVP

---

## 1. 模块定位

AI Agent 是对话管理的核心——连接 LLM、记忆系统和工具系统。使用 Pipeline 洋葱模型，每个处理阶段（Stage）可独立测试和替换。

## 2. 接口与边界

### 对外暴露

```go
type Agent interface {
    Run(ctx context.Context, req Request) error   // 执行一轮完整对话
    SetSink(sink event.Sink)                        // 设置事件输出
}

// Option 模式构造函数
func New(provider llm.Provider, opts ...Option) Agent
func WithMemory(m memory.Manager) Option
func WithToolRegistry(r tool.Registry) Option
```

### 内部 Stage 接口（可扩展）

```go
type Stage interface {
    Name() string
    Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error
}
```

### 内置 4 个 Stage

| Stage | 职责 | Phase |
|-------|------|-------|
| `PreProcessStage` | 消息标准化、敏感词过滤 | 2 |
| `MemoryStage` | 注入短期 + 长期记忆到上下文 | 2 |
| `LLMCallStage` | 调用 Provider.Stream()，处理流式回复和工具调用 | 1 |
| `PostProcessStage` | 提取事实写入记忆 | 2 |

### 依赖的外部接口

- `llm.Provider` — LLM 调用
- `memory.Manager` — 记忆读写
- `tool.Registry` — 工具执行
- `event.Sink` — 输出流式事件

## 3. 可拆卸性设计

| 机制 | 实现 |
|------|------|
| **接口隔离** | Agent 自身也是接口，可被 Mock 替换 |
| **Option 注入** | 依赖通过 `With*` 函数注入，非必需依赖可选 |
| **洋葱模型** | 每个 Stage 独立、可替换、可测试 |
| **无默认 Sink** | `SetSink` 显式设置，默认为 `NoopSink` |
| **Mock 串联** | 测试中使用 `llm.NewProvider("mock")` + `memory.NewMockManager()` |

## 4. 已完成的代码

| 文件 | 行数 | 内容 |
|------|------|------|
| `agent/agent.go` | ~105 行 | Agent 接口、Option 模式、agentImpl.Run |
| `agent/agent_test.go` | ~140 行 | Mock Provider 集成测试、Pipeline 顺序测试、无 Provider 错误测试 |
| `agent/pipeline.go` | ~140 行 | Stage 接口、Pipeline 洋葱模型、4 个内置 Stage |

## 5. 剩余工作任务

- [ ] **Phase 2**：`PreProcessStage` — 消息标准化和敏感词过滤
- [ ] **Phase 2**：`MemoryStage` — 从 memory 检索短期+长期记忆注入上下文
- [ ] **Phase 2**：`PostProcessStage` — 对话结束后提取事实写入 memory
- [ ] **Phase 4**：`LLMCallStage` — 处理工具调用循环（`ChunkToolCall`）

## 6. 验收标准

- [x] `go build ./internal/agent/...` 通过
- [x] `go test ./internal/agent/...` 全部通过
- [x] Agent 支持通过 Option 注入 memory 和 tool 依赖
- [x] Pipeline 串联 4 个 Stage 并能验证执行顺序
- [x] LLMCallStage 能调用 Provider.Stream() 并输出流式事件
