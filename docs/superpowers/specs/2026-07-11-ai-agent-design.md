# AI Agent — 实现规范

> **子 Agent 任务 ID:** petcore-006-agent
> **包路径:** `petcore/internal/agent/`
> **依赖:** llm（Provider 接口）、memory（Manager 接口）、tool（Registry 接口）、event（Sink 接口）
> **被依赖:** core（通过 Agent 接口）

---

## 1. 模块职责

AI Agent 是对话管理的核心。使用 Pipeline 洋葱模型串联多个处理 Stage，每个 Stage 可独立测试和替换。

## 2. 接口契约

### 2.1 Agent 接口

```go
type Agent interface {
    Run(ctx context.Context, req Request) error   // 执行一轮完整对话
    SetSink(sink event.Sink)                       // 设置事件输出
}

type Request struct {
    Messages     []llm.Message
    SystemPrompt string
}

// Option 模式构造函数
func New(provider llm.Provider, opts ...Option) Agent
func WithMemory(m memory.Manager) Option
func WithToolRegistry(r tool.Registry) Option
```

### 2.2 Pipeline Stage 接口

```go
type Stage interface {
    Name() string
    Process(ctx context.Context, pCtx *pipelineCtx, next StageFunc) error
}

type StageFunc func(ctx context.Context, pCtx *pipelineCtx) error
```

### 2.3 内置 Stage

| Stage | 文件 | 状态 | 职责 |
|-------|------|------|------|
| `PreProcessStage` | `agent/pipeline.go:52-59` | ✅ 骨架 | 消息标准化（Phase 2 实现过滤） |
| `MemoryStage` | `agent/pipeline.go:62-71` | ✅ 骨架 | 注入记忆到上下文（Phase 2 实现） |
| `LLMCallStage` | `agent/pipeline.go:74-124` | ✅ 完整 | 调用 Provider.Stream()，输出流式事件 |
| `PostProcessStage` | `agent/pipeline.go:127-137` | ✅ 骨架 | 事实提取写入（Phase 2 实现） |

### 2.4 LLMCallStage 执行流程

```
LLMCallStage.Process:
  1. 发送 agent.thinking (status=true) 事件
  2. 调用 provider.Stream(ctx, req)
  3. 遍历 stream:
     - ChunkText → 发送 agent.reply {text, done:false}
     - ChunkToolCall → (Phase 4) 执行工具调用
     - ChunkDone → 发送 agent.reply {done:true}
  4. 发送 agent.thinking (status=false) 事件
  5. 调用 next(ctx, pCtx) 进入下一 Stage
```

## 3. 已实现内容

| 文件 | 状态 | 内容 |
|------|------|------|
| `agent/agent.go` | ✅ 完整 | Agent 接口、Option、agentImpl.Run |
| `agent/agent_test.go` | ✅ 完整 | Mock Provider 集成测试、Pipeline 顺序测试 |
| `agent/pipeline.go` | ✅ 完整 | Stage 接口、Pipeline 洋葱模型、4 个内置 Stage |

## 4. 待实现

- [ ] **PreProcessStage** — 敏感词过滤、消息长度限制
- [ ] **MemoryStage** — 从 memory 检索短期+长期记忆注入 SystemPrompt
- [ ] **PostProcessStage** — 对话后提取事实写入 memory
- [ ] **LLMCallStage** — Phase 4 的工具调用循环

## 5. 子 Agent 验收命令

```bash
cd petcore
go build ./internal/agent/...
go test -v -count=1 ./internal/agent/...
MOCK_LLM_REPLY="你好" go test -v -count=1 -run TestAgent_Run ./internal/agent/...
```

## 6. 被其他模块使用的方式

```go
import "github.com/desktop-pet/petcore/internal/agent"
import _ "github.com/desktop-pet/petcore/internal/llm/mock"

provider, _ := llm.NewProvider("mock", nil)
mem := memory.NewMockManager()
toolReg := tool.NewRegistry()

ag := agent.New(provider,
    agent.WithMemory(mem),
    agent.WithToolRegistry(toolReg),
)

// 在 core.Engine 中：
eng := core.New(machine, ag, mem, pluginReg, toolReg, cfg, sink)
```

---

## 验收标准

> 以下验收标准适用于所有子 Agent 和人工任务。缺少任何一项即不视为完成。

- [ ] `go build ./...` 通过
- [ ] `golangci-lint run ./...` 无 error
- [ ] `go test -race ./...` 全部通过
- [ ] 新增代码覆盖率 ≥60%（增量覆盖率）
- [ ] 边界情况已覆盖（空输入 / 错误输入 / 并发场景）
- [ ] 测试使用 Mock 而非真实外部服务（LLM / 网络 / 文件系统）
- [ ] 编译时接口合规性检查：`var _ Interface = (*Impl)(nil)`
