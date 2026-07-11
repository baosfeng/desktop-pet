# 模块规划：核心引擎 + Sidecar 通信 (petcore/internal/core + server)

> **前置依赖：** event、fsm、config、llm、memory、agent、tool、plugin（全部）
> **被以下模块依赖：** cmd/petcore/main.go（入口组装）
> **对应 Phase：** Phase 1 — MVP

---

## 1. 模块定位

**核心引擎** (`core`) 是所有子模块的组装者——通过构造函数注入所有依赖，绝不 import 具体子模块实现。它协调 FSM 状态转换、Agent 对话处理、事件输出。

**Sidecar 通信** (`server`) 是 PetCore 与父进程（Tauri Rust 壳）之间的通信层。通过 stdin/stdout JSON 协议接收命令、返回响应、推送事件。

## 2. 接口与边界

### Core Engine

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

func (e *Engine) Run(ctx context.Context) error          // 启动引擎主循环（阻塞）
func (e *Engine) HandleInput(ctx context.Context, text string) error  // 处理用户输入
func (e *Engine) GetStatus() map[string]any               // 获取状态快照
```

### Server

```go
func New(engine *core.Engine, reader io.Reader, writer io.Writer) *Server
func (s *Server) Run(ctx context.Context) error  // 启动通信循环（阻塞）

// Wire 协议
type Command struct { Type, ID, string; Method string; Params json.RawMessage }
type WireEvent struct { Type, Event string; Data any }

// SinkAdapter — 将 Server 输出适配为 Sink 接口
func NewSinkAdapter(s *Server) *SinkAdapter
```

### 支持的 Sidecar 命令

| 命令 | 参数 | 响应 | Phase |
|------|------|------|-------|
| `ping` | 无 | `{"pong":"ok"}` | 1 |
| `chat` | `{"text":"..."}` | `{"done":true}` + 流式事件 | 1 |
| `get_status` | 无 | `{"state":"...","plugins":N}` | 1 |

## 3. 可拆卸性设计

| 机制 | 实现 |
|------|------|
| **构造函数注入全部依赖** | `core.New(fsm, agent, memory, plugin, tool, config, sink)` |
| **不 import 具体子模块** | Engine 字段全是接口类型 |
| **NoopSink 默认** | Sink 为 nil 时使用 `event.NoopSink{}` |
| **SinkAdapter** | Server 将 JSON 输出适配为 Sink 接口 |
| **独立测试** | `newTestEngine()` 组装所有 Mock 依赖 |
| **Server 可独立测试** | 用 `strings.NewReader` / `bytes.Buffer` 替代真实 stdin/stdout |

### 依赖注入链路

```
cmd/petcore/main.go
  ├── config.Load() → *config.Config
  ├── llm.NewProvider() → llm.Provider
  ├── fsm.NewMockMachine() → fsm.Machine
  ├── memory.NewInMemoryManager() → memory.Manager
  ├── tool.NewRegistry() → tool.Registry
  ├── agent.New(provider, WithMemory, WithToolRegistry) → agent.Agent
  ├── plugin.NewRegistry() → plugin.Registry
  └── core.New(machine, agent, mem, plugin, tool, cfg, sink) → *core.Engine
```

所有模块在 `main.go` 中组装，没有隐式的全局依赖或 init() 顺序问题。

## 4. 已完成的代码

| 文件 | 行数 | 内容 |
|------|------|------|
| `core/engine.go` | ~120 行 | Engine 结构体、New、Run、HandleInput、GetStatus |
| `core/engine_test.go` | ~135 行 | 依赖注入测试、HandleInput 集成测试、Run 启停 |
| `server/server.go` | ~166 行 | Server、Command/Response/WireEvent 协议、SinkAdapter |
| `server/server_test.go` | ~90 行 | ping/chat/get_status/无效命令测试 |
| `cmd/petcore/main.go` | ~125 行 | 双模式入口（sidecar + CLI）、buildEngine 组装函数 |

## 5. 剩余工作任务

- [ ] **Phase 1**：Sidecar 模式下将 SinkAdapter 注入 Engine（当前 TODO 标记）
- [ ] **Phase 2**：实现 Engine 主事件循环（定时任务、空闲检测等）
- [ ] **Phase 2**：实现 CLI 模式（bubbletea TUI）
- [ ] **Phase 2**：Engine 增加 graceful shutdown 事件发送
- [ ] **Phase 2**：Sidecar 协议增加 `list_plugins`、`reload_plugin` 命令

## 6. 验收标准

- [x] `go build ./internal/core/...` 通过
- [x] `go build ./internal/server/...` 通过
- [x] `go build ./cmd/petcore/...` 通过
- [x] `go test ./internal/core/...` 全部通过（依赖注入 + HandleInput + Run 启停）
- [x] `go test ./internal/server/...` 全部通过（ping/chat/status/错误处理）
- [x] Engine 构造函数注入全部 7 个依赖
- [x] Server 支持 sidecar JSON 协议（cmd/resp/event 三层）
- [x] `SinkAdapter` 将 JSON 输出适配为 Sink 接口
