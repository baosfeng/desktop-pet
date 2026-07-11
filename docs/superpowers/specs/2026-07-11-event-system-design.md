# 事件系统 — 实现规范

> **子 Agent 任务 ID:** petcore-001-event
> **包路径:** `petcore/internal/event/`
> **依赖:** 无（零内部依赖，仅 Go 标准库 `time`）
> **被依赖:** fsm、core、server、所有 Shell 层

---

## 1. 模块职责

定义 PetCore 中所有事件类型和事件消费接口。是整个架构最基础的模块。

## 2. 接口契约

### 2.1 事件类型（8 种）

```go
package event

type Type string

const (
    EventStateChanged  Type = "state.changed"    // 状态机切换
    EventPetSpeak      Type = "pet.speak"         // 宠物说话
    EventPetAction     Type = "pet.action"        // 宠物动作
    EventPetEmotion    Type = "pet.emotion"       // 情绪变化
    EventAgentThinking Type = "agent.thinking"    // AI 思考中
    EventAgentReply    Type = "agent.reply"       // AI 回复片段
    EventMemoryUpdated Type = "memory.updated"    // 记忆更新
    EventError         Type = "error"             // 错误
)
```

### 2.2 核心结构

```go
type Event struct {
    Kind Type   // 事件类型
    Data any    // 事件负载
    Meta Meta   // 元信息
}

type Meta struct {
    Timestamp int64   // Unix 毫秒时间戳
    Source    string  // core / plugin / user
    SessionID string
}

func NewMeta(source, sessionID string) Meta  // 创建带当前时间戳的 Meta
```

### 2.3 Sink 接口（供 Shell 层实现）

```go
type Sink interface {
    Send(event Event) error  // 发送事件
    Close() error            // 关闭 Sink
}
```

### 2.4 NoopSink（默认空实现）

```go
type NoopSink struct{}

func (NoopSink) Send(Event) error { return nil }
func (NoopSink) Close() error     { return nil }
```

## 3. 已实现内容

所有代码已完成，无需额外实现。

| 文件 | 状态 |
|------|------|
| `event/event.go` | ✅ 完整 |
| `event/event_test.go` | ✅ 完整 |

## 4. 子 Agent 验收命令

```bash
cd petcore
go build ./internal/event/...
go test -v -count=1 ./internal/event/...
go vet ./internal/event/...
```

## 5. 被其他模块 Mock 的方式

其他模块测试中通过实现 `Sink` 接口来模拟事件消费：

```go
// 全局 NoopSink（忽略所有事件）
sink := event.NoopSink{}

// Channel-backed Sink（断言事件）
ch := make(chan event.Event, 100)
sink := &eventSink{ch: ch}

type eventSink struct{ ch chan event.Event }
func (s *eventSink) Send(e event.Event) error { s.ch <- e; return nil }
func (s *eventSink) Close() error             { return nil }
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
