# 配置管理 — 实现规范

> **子 Agent 任务 ID:** petcore-002-config
> **包路径:** `petcore/internal/config/`
> **依赖:** 无（零内部依赖，仅 Go 标准库）
> **被依赖:** core（持有 `*config.Config`）、memory（通过 Config.Memory 配置）

---

## 1. 模块职责

负责加载和管理 PetCore 的 TOML 配置文件。API Key 通过环境变量读取，不硬编码。

## 2. 接口契约

### 2.1 核心函数

```go
func Load(path string) (*Config, error)      // 从 TOML 文件加载，文件不存在返回默认
func DefaultConfig() *Config                 // 返回默认配置（每次独立副本）

type Config struct {
    LLM     LLMConfig     `toml:"llm"`
    Agent   AgentConfig   `toml:"agent"`
    Window  WindowConfig  `toml:"window"`
    Update  UpdateConfig  `toml:"update"`
    Memory  MemoryConfig  `toml:"memory"`
    Plugin  PluginConfig  `toml:"plugin"`
}
```

### 2.2 API Key 安全读取

```go
func (c *LLMConfig) APIKey() string  // 从环境变量 c.APIKeyEnv 读取
```

### 2.3 关键默认值

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `LLM.Provider` | `"mock"` | Phase 1 默认使用 Mock |
| `LLM.Temperature` | `0.7` | |
| `Agent.SystemPrompt` | `"You are a cute desktop pet."` | |
| `Agent.MaxToolTurns` | `10` | |
| `Memory.ShortTermSize` | `20` | L1 短期记忆轮数 |
| `Window.AlwaysOnTop` | `true` | |
| `Plugin.ActionsDir` | `~/.desktop-pet/actions` | |
| `Update.AutoCheck` | `true` | |

## 3. 已实现内容

| 文件 | 状态 | 内容 |
|------|------|------|
| `config/config.go` | ✅ 完整 | 结构体定义、DefaultConfig、Load（骨架）、APIKey |
| `config/config_test.go` | ✅ 完整 | 默认值测试、防变异测试、FileNotExist、API Key |

## 4. 待实现

- [ ] **Load() 中的 TOML 解析** — 引入 `BurntSushi/toml` 解析文件内容
- [ ] **Validate()** — 校验必填项（如 API Key 环境变量是否设置）

## 5. 子 Agent 验收命令

```bash
cd petcore
go build ./internal/config/...
go test -v -count=1 ./internal/config/...
```

## 6. 被其他模块使用的方式

```go
cfg := config.DefaultConfig()
// 或
cfg, err := config.Load("~/.desktop-pet/config.toml")

// 读取 API Key（从环境变量）
key := cfg.LLM.APIKey()
```
