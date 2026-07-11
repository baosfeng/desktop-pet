# 记忆系统 — 实现规范

> **子 Agent 任务 ID:** petcore-005-memory
> **包路径:** `petcore/internal/memory/`
> **依赖:** 无（不依赖其他 PetCore 内部包）
> **被依赖:** agent（通过 Manager 接口）、core（通过 Manager 接口）

---

## 1. 模块职责

三层记忆系统的统一入口：L1 短期（内存环形缓冲区）、L2 核心（SQLite key-value，Phase 2 落地）、L3 长期（SQLite + FTS5，Phase 2 落地）。Phase 1 全部使用内存实现。

## 2. 接口契约

### 2.1 Manager 接口

```go
type Manager interface {
    // L1 短期记忆
    AddShortTerm(msg Message)
    GetShortTerm() []Message
    ClearShortTerm()

    // L2 核心记忆
    Remember(key, value string, importance int) error
    Recall(key string) (string, error)
    GetAllCore() (map[string]string, error)

    // L3 长期记忆
    Store(ctx context.Context, fact Fact) error
    Search(ctx context.Context, query string, limit int) ([]Fact, error)
}
```

### 2.2 数据结构

```go
type Message struct {
    Role    string  // system / user / assistant / tool
    Content string
}

type Fact struct {
    Key        string
    Value      string
    Category   string  // preference / fact / event
    Importance int     // 1-5
    Timestamp  int64
}
```

### 2.3 内存实现和 Mock

```go
func NewInMemoryManager() *InMemoryManager  // 全内存实现，可作生产使用
func NewMockManager() *InMemoryManager      // 同上，语义别名供测试使用
```

## 3. 已实现内容

| 文件 | 状态 | 内容 |
|------|------|------|
| `memory/manager.go` | ✅ 完整 | Manager 接口、InMemoryManager（三层全实现，sync.RWMutex 并发安全） |
| `memory/memory_test.go` | ✅ 完整 | CRUD 测试 + 并发安全测试 |

## 4. 待实现

- [ ] **Phase 2**：SQLite 持久化 `sqliteManager`（引入 `modernc.org/sqlite`）
- [ ] **Phase 2**：L1 短期滑动窗口裁剪（超过 `ShortTermSize` 自动丢弃）
- [ ] **Phase 2**：L3 长期 FTS5 全文检索

## 5. 子 Agent 验收命令

```bash
cd petcore
go build ./internal/memory/...
go test -v -count=1 ./internal/memory/...
```

## 6. 被其他模块使用的方式

```go
import "github.com/desktop-pet/petcore/internal/memory"

mem := memory.NewMockManager()
mem.AddShortTerm(memory.Message{Role: "user", Content: "你好"})
msgs := mem.GetShortTerm()

mem.Remember("name", "小明", 5)
val, _ := memory.Recall("name")  // "小明"

// 在 agent.Agent 中使用：
a := agent.New(provider, agent.WithMemory(mem))
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
