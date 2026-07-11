# 核心引擎 + Sidecar 通信 — 实现规范

> **子 Agent 任务 ID:** petcore-008-core-server
> **包路径:** `petcore/internal/core/` + `petcore/internal/server/` + `petcore/cmd/petcore/main.go`
> **依赖:** event、config、fsm、llm、memory、agent、tool、plugin（全部内部模块）
> **被依赖:** 无（这是最上层的组装层）

---

## 1. 模块职责

**核心引擎** (`core`) 是所有子模块的组装者——构造函数注入全部依赖，绝不 import 具体子模块实现。
**Sidecar 通信** (`server`) 是 PetCore 与父进程（Tauri Rust 壳）之间的 stdin/stdout JSON 通信层。
**main.go** 是双模式入口（sidecar 模式 + CLI 模式）。

## 2. 接口契约

### 2.1 Core Engine

```go
func New(
    fsm fsm.Machine,
    ag agent.Agent,
    mem memory.Manager,
    pl plugin.Registry,
    tl tool.Registry,
    cfg *config.Config,
    sink event.Sink,
) *Engine

func (e *Engine) Run(ctx context.Context) error          // 启动引擎（阻塞）
func (e *Engine) HandleInput(ctx context.Context, text string) error  // 处理用户输入
func (e *Engine) GetStatus() map[string]any               // 获取状态快照
```

### 2.2 Server（sidecar 通信）

```go
func New(engine *core.Engine, reader io.Reader, writer io.Writer) *Server
func (s *Server) Run(ctx context.Context) error  // 启动通信循环（阻塞）
```

### 2.3 Wire 协议

**请求（stdin）：**
```json
{"type":"cmd","id":"1","method":"chat","params":{"text":"你好"}}
{"type":"cmd","id":"2","method":"get_status","params":{}}
{"type":"cmd","id":"3","method":"ping","params":{}}
```

**响应（stdout）：**
```json
{"type":"resp","id":"1","result":{"done":true}}
{"type":"event","event":"pet.speak","data":{"text":"你好~"}}
{"type":"resp","id":"2","result":{"state":"idle","plugins":0}}
```

### 2.4 Server 支持的 Sidecar 命令

| 命令 | 处理函数 | 说明 |
|------|---------|------|
| `ping` | 返回 `{"pong":"ok"}` | 健康检查 |
| `chat` | 调用 `engine.HandleInput(text)` | 对话 |
| `get_status` | 调用 `engine.GetStatus()` | 状态查询 |

### 2.5 SinkAdapter

```go
func NewSinkAdapter(s *Server) *SinkAdapter  // 将 Server JSON 输出适配为 Sink 接口
// 实现 event.Sink 接口，PetCore 事件通过 sidecar 推送给父进程
```

### 2.6 main.go 入口

```go
// 双模式启动
petcore --mode sidecar    // 默认：stdin/stdout JSON 协议
petcore --cli              // 交互式对话（Phase 2 实现）
petcore --config /path/to/config.toml  // 指定配置文件
```

## 3. 依赖注入链路（在 buildEngine 中组装）

```
config.Load() → *config.Config
    │
llm.NewProvider(cfg.LLM.Provider, ...) → llm.Provider
    │
fsm.NewMockMachine(StateIdle) → fsm.Machine
    │
memory.NewInMemoryManager() → memory.Manager
    │
tool.NewRegistry() → tool.Registry
    │
agent.New(provider, WithMemory(mem), WithToolRegistry(toolReg)) → agent.Agent
    │
plugin.NewRegistry() → plugin.Registry
    │
core.New(machine, agent, mem, pluginReg, toolReg, cfg, sink) → *core.Engine
    │
server.New(engine, os.Stdin, os.Stdout) → *server.Server
```

## 4. 已实现内容

| 文件 | 状态 | 内容 |
|------|------|------|
| `core/engine.go` | ✅ 完整 | Engine 结构体、New、Run、HandleInput、GetStatus |
| `core/engine_test.go` | ✅ 完整 | 依赖注入测试、HandleInput 集成测试、Run 启停 |
| `server/server.go` | ✅ 完整 | Server、Wire 协议、SinkAdapter |
| `server/server_test.go` | ✅ 完整 | ping/chat/status/无效命令测试 |
| `cmd/petcore/main.go` | ✅ 完整 | 双模式入口、buildEngine 组装、信号处理 |

## 5. 待实现

- [ ] Sidecar 模式中将 SinkAdapter 注入 Engine（当前 `main.go:114` 有 TODO）
- [ ] CLI 模式（`runCLI` 函数，Phase 2 用 bubbletea 实现）
- [ ] Engine 主事件循环（`core/engine.go:77` 有 TODO）

## 6. 子 Agent 验收命令

```bash
cd petcore
go build ./internal/core/...
go test -v -count=1 ./internal/core/...
go build ./internal/server/...
go test -v -count=1 ./internal/server/...
go build ./cmd/petcore/...

# 端到端测试（sidecar 模式）
echo '{"type":"cmd","id":"1","method":"ping","params":{}}' | go run ./cmd/petcore/
# 应输出: {"type":"resp","id":"1","result":{"pong":"ok"}}
```

## 7. 测试辅助函数（供其他模块参考）

```go
// newTestEngine 组装所有 Mock 依赖
func newTestEngine(t *testing.T, withPlugin bool) *Engine {
    t.Helper()
    provider, _ := llm.NewProvider("mock", nil)
    machine := fsm.NewMockMachine(fsm.StateIdle)
    mem := memory.NewMockManager()
    toolReg := tool.NewRegistry()
    ag := agent.New(provider, agent.WithMemory(mem), agent.WithToolRegistry(toolReg))
    pluginReg := plugin.NewRegistry()
    cfg := config.DefaultConfig()
    return New(machine, ag, mem, pluginReg, toolReg, cfg, event.NoopSink{})
}
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
