# FSM 状态机 — 实现规范

> **子 Agent 任务 ID:** petcore-003-fsm
> **包路径:** `petcore/internal/fsm/`
> **依赖:** event（仅用到 event.Type 常量）
> **被依赖:** core（通过 Machine 接口）

---

## 1. 模块职责

定义宠物的 4 态行为状态机及其转换规则。纯逻辑模块，无 IO、无 goroutine。

## 2. 接口契约

### 2.1 状态常量

```go
type State string

const (
    StateIdle        State = "idle"         // 待机
    StateAttention   State = "attention"    // 关注
    StateInteraction State = "interaction"  // 交互
    StateSpeaking    State = "speaking"     // 说话
)
```

### 2.2 Machine 接口

```go
type Machine interface {
    Current() State
    Transition(evt event.Type) error
    OnTransition(fn func(from, to State))
}
```

### 2.3 转换规则表

```
idle ──(EventStateChanged)──→ attention
idle ──(EventPetSpeak)──→ speaking
attention ──(EventStateChanged)──→ interaction
attention ──(EventPetSpeak)──→ speaking
attention ──(EventError)──→ idle
interaction ──(EventAgentReply)──→ speaking
interaction ──(EventError)──→ idle
speaking ──(EventStateChanged)──→ idle
```

### 2.4 工具函数

```go
func IsValidTransition(from State, evt event.Type) bool  // 检查转换是否合法
func TransitionsFrom(s State) []event.Type                // 查询某状态的所有可用事件
```

### 2.5 错误类型

```go
type ErrTransitionNotAllowed struct {
    From State
    Evt  event.Type
}
func (e *ErrTransitionNotAllowed) Error() string
```

## 3. 已实现内容

| 文件 | 状态 | 内容 |
|------|------|------|
| `fsm/machine.go` | ✅ 完整 | State 常量、Machine 接口、transitionRules 表、MockMachine |
| `fsm/fsm_test.go` | ✅ 完整 | 13 条转换规则全覆盖 |

## 4. 待实现

- [ ] **真实的 Transition()** — 当前 MockMachine 的 Transition 是 no-op
- [ ] **Transition 中调用 onFn** — 状态变化后触发 `OnTransition` 注册的回调

## 5. 子 Agent 验收命令

```bash
cd petcore
go build ./internal/fsm/...
go test -v -count=1 ./internal/fsm/...
```

## 6. 被其他模块 Mock 的方式

```go
import "github.com/desktop-pet/petcore/internal/fsm"

machine := fsm.NewMockMachine(fsm.StateIdle)
// machine.Transition() 是 no-op，适合测试
```
