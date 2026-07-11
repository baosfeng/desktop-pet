# 工具系统 + 插件系统 — 实现规范

> **子 Agent 任务 ID:** petcore-007-tool-plugin
> **包路径:** `petcore/internal/tool/` + `petcore/internal/plugin/`
> **依赖:** tool — 无；plugin — 无
> **被依赖:** agent（依赖 tool.Registry）、core（依赖两者）

---

## 1. 模块职责

**工具系统:** LLM 可调用的能力单元（web_search、remember 等），注册中心模式。
**插件系统:** 三层插件架构（L1 YAML / L2 JS / L3 MCP），管理生命周期。

## 2. 接口契约

### 2.1 工具系统

```go
// 单个工具
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args string) (string, error)
}

// 注册中心
type Registry interface {
    Register(tool Tool) error
    Execute(ctx context.Context, name, args string) (string, error)
    List() []Tool
}

func NewRegistry() Registry
```

### 2.2 插件系统

```go
type Type int
const (
    TypeL1YAML Type = iota  // 声明式 YAML 动作包
    TypeL2JS                // JS 脚本引擎
    TypeL3MCP               // MCP 子进程
)

type Plugin interface {
    Name() string
    Type() Type
    Start(ctx context.Context) error
    Stop() error
}

type Registry interface {
    Register(p Plugin) error
    List() []Plugin
    Get(name string) (Plugin, error)
    StartAll(ctx context.Context) error
    StopAll() error
}

func NewRegistry() Registry
```

### 2.3 Mock 实现

```go
// MockTool
type MockTool struct {
    MockName        string
    MockDescription string
    MockExecute     func(ctx context.Context, args string) (string, error)
}

// MockPlugin
type MockPlugin struct {
    MockName string
    MockType Type
    OnStart  func(ctx context.Context) error
    OnStop   func() error
}
```

## 3. 已实现内容

| 文件 | 状态 | 内容 |
|------|------|------|
| `tool/registry.go` | ✅ 完整 | Tool 接口、Registry 接口 + 实现、MockTool |
| `tool/registry_test.go` | ✅ 完整 | 注册/执行/未找到错误 |
| `plugin/registry.go` | ✅ 完整 | Plugin 接口、Registry 接口 + 实现、MockPlugin |
| `plugin/plugin_test.go` | ✅ 完整 | 类型字符串测试 |

## 4. 待实现

- [ ] **Phase 4**：内置工具 `builtin/web_search.go`
- [ ] **Phase 4**：MCP 协议客户端 `plugin/mcp_client.go`
- [ ] **Phase 4**：L1 YAML 加载器 `plugin/l1_yaml.go` + fsnotify
- [ ] **Phase 4**：L2 JS 引擎 `plugin/l2_script.go` + goja
- [ ] **Phase 4**：Agent 工具调用循环（`Agent.Run` 中处理 ChunkToolCall）

## 5. 子 Agent 验收命令

```bash
cd petcore
go build ./internal/tool/...
go test -v -count=1 ./internal/tool/...
go build ./internal/plugin/...
go test -v -count=1 ./internal/plugin/...
```

## 6. 被其他模块使用的方式

```go
// Tool
toolReg := tool.NewRegistry()
toolReg.Register(&tool.MockTool{MockName: "my-tool", MockExecute: func(ctx, args) (string, error) {
    return "result", nil
}})
result, _ := toolReg.Execute(ctx, "my-tool", `{"key":"val"}`)

// Plugin
pluginReg := plugin.NewRegistry()
pluginReg.Register(&plugin.MockPlugin{MockName: "greeter"})
pluginReg.StartAll(ctx)
pluginReg.StopAll()
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
