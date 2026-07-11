# 桌面 AI 宠物 — 设计规范

> 一个像 QQ 宠物一样在桌面上陪伴你，接入 AI 大模型、越用越懂你的桌面伴侣。
> 最终目标：通过自然对话控制电脑。

## 一、核心原则与技术选型

### 核心原则

| 原则 | 说明 |
|------|------|
| **隐私优先** | 所有记忆、配置、数据存储在本地，不上传任何云端。仅对话文本发送给 LLM API |
| **更新即体验** | 小白用户唯一接触新功能的途径是「检查更新」。更新必须静默、增量、一键完成 |
| **渐进增强** | Phase 1 先跑通最小闭环，后续逐步叠加记忆、情感、工具、语音、桌面控制 |
| **内核-壳分离** | PetCore (Go 内核) + Tauri (Rust 壳)，多壳复用（CLI + Desktop + Server） |
| **可扩展** | LLM 适配层、三层插件系统、MCP 工具协议，预留所有未来能力接口 |

### 技术选型

| 层 | 技术 | 理由 |
|----|------|------|
| 桌面容器 | **Tauri v2** (Rust) | 系统原生 WebView，安装包 ~15MB，内存 ~40MB，透明窗口 + 鼠标穿透原生支持 |
| Live2D 渲染 | PixiJS 8 + pixi-live2d-display | 社区最成熟的 Live2D Web 渲染方案，WebView 内运行 |
| 前端 UI | React 18 + TypeScript + Vite | 聊天气泡、设置面板、组件化开发 |
| 后端内核 | **Go PetCore** (CGO_ENABLED=0) | 单静态二进制 (~10MB)，多壳复用 (CLI + Desktop + Server) |
| LLM 适配 | Provider 接口 + 注册表模式 | 统一接入 OpenAI/DeepSeek/硅基流动/Ollama 等 |
| 记忆存储 | SQLite + FTS5（modernc.org/sqlite） | 纯 Go 实现，零 CGO 依赖，渐进引入向量检索 |
| Agent 框架 | Pipeline 洋葱模型（参考 NyaDeskPet + Reasonix） | 职责分离清晰，Stage 可插拔 |
| 工具协议 | MCP (Model Context Protocol) | 行业标准工具扩展协议 |
| 通信协议 | Tauri IPC (invoke + event) + sidecar stdin/stdout JSON | 零网络开销，三层通信 |
| 自动更新 | tauri-plugin-updater | GitHub Releases 增量更新 |

---

## 二、系统架构

### 2.1 总体架构

```
┌──────────────────────────────────────────────────────────────────────────┐
│  DesktopPet.app  (~15-25MB)  [Tauri v2 + WebView + Go PetCore]           │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │  Tauri Rust 壳                                                        │  │
│  │  ┌──────────────┐ ┌────────────┐ ┌──────────────┐ ┌──────────────┐ │  │
│  │  │ 窗口管理器    │ │ 系统托盘    │ │ Sidecar 管理  │ │ 自动更新     │ │  │
│  │  │ (透明/穿透/  │ │ (菜单图标/ │ │ (spawn/stop/ │ │ (tauri-plugin │ │  │
│  │  │  置顶/多屏)  │ │  右键菜单) │ │  健康检查)   │ │  -updater)   │ │  │
│  │  └──────────────┘ └────────────┘ └──────┬───────┘ └──────────────┘ │  │
│  │                                          │ stdin/stdout JSON        │  │
│  │  ┌───────────────────────────────────────▼────────────────────────┐  │  │
│  │  │  事件转发: Rust 读取 sidecar stdout → 转为 Tauri event 推给前端 │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  └──────────────────────┬──────────────────────────────────────────────┘  │
│                         │ Tauri IPC (invoke + event)                     │
│  ┌──────────────────────▼──────────────────────────────────────────────┐  │
│  │  React 前端 (WebView)                                               │  │
│  │  ┌────────────────┐ ┌────────────┐ ┌────────────┐ ┌──────────────┐ │  │
│  │  │ Live2D 渲染    │ │ 聊天气泡   │ │ 设置面板   │ │ bridge.ts    │ │  │
│  │  │ (PixiJS 8 +   │ │ (打字机    │ │ (LLM配置/  │ │ (Tauri IPC   │ │  │
│  │  │  pixi-live2d) │ │  效果+MD) │ │  人设/记忆) │ │  封装)       │ │  │
│  │  └────────────────┘ └────────────┘ └────────────┘ └──────────────┘ │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                           │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │  Go PetCore (sidecar 子进程, CGO_ENABLED=0, ~10MB)                    │  │
│  │  ┌──────┐ ┌─────────┐ ┌────────┐ ┌──────┐ ┌────────┐ ┌──────┐     │  │
│  │  │ 状态  │ │ AI      │ │ 记忆   │ │ LLM  │ │ 插件   │ │ 工具  │     │  │
│  │  │ 引擎  │ │ Agent   │ │ 系统   │ │ 适配 │ │ L1/L2 │ │ 系统  │     │  │
│  │  │ (FSM)│ │(Pipeline)│ │ (3层)  │ │(Prov)│ │  /L3  │ │ (MCP) │     │  │
│  │  └──────┘ └─────────┘ └────────┘ └──────┘ └────────┘ └──────┘     │  │
│  │  ┌──────┐ ┌──────────────────────────────────────────────────────┐  │  │
│  │  │ CLI  │ │ 事件总线 (Event Bus): 内核发出事件，壳层通过 Sink 消费  │  │  │
│  │  │ Shell│ └──────────────────────────────────────────────────────┘  │  │
│  │  └──────┘                                                           │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│  本地数据目录 (~/.desktop-pet/)                                          │
│  ├── config.toml     # 配置（API Key / 人设 / 窗口设置）                  │
│  ├── memory.db       # SQLite 结构化记忆 + FTS5 全文检索                   │
│  ├── actions/        # L1 YAML 动作包（fsnotify 热加载）                  │
│  ├── scripts/        # L2 JS 脚本插件（fsnotify 热加载）                  │
│  ├── logs/           # 运行日志                                          │
│  └── models/         # Live2D 模型文件                                    │
└──────────────────────────────────────────────────────────────────────────┘
```

### 2.2 通信拓扑

| 链路 | 协议 | 方向 | 说明 |
|------|------|------|------|
| 前端 ↔ Tauri 壳 | Tauri IPC | 双向 | JS `invoke()` 调用 Rust 命令，`listen()` 接收事件 |
| Tauri 壳 ↔ Go PetCore | stdin/stdout JSON 行 | 双向 | Tauri 通过 sidecar 模式管理子进程，管道通信 |
| Go PetCore → 前端（透传） | Tauri event | 单向 | Rust 收到 sidecar 事件后 `app.emit()` 推给前端 |

### 2.3 数据目录

所有用户数据存储在 `~/.desktop-pet/`，独立于 `.app` 之外。更新或重装应用不影响用户数据。

---

## 三、Tauri 壳层

### 3.1 窗口管理器

**职责：** 透明无边框窗口、鼠标穿透切换、置顶、多显示器跟随、位置记忆

**核心接口：**

```
// Tauri 命令 (前端通过 invoke 调用)
toggle_clickthrough(enabled: bool)    // 切换鼠标穿透
get_window_position() → {x, y}       // 获取窗口位置
set_window_position(x, y)            // 设置窗口位置
```

**鼠标穿透状态切换：**

```
待机 → ignoreCursorEvents = true
  │ 鼠标进入窗口区域
  ▼
ignoreCursorEvents = false → 宠物转向鼠标
  │
  ├── 点击/拖拽 → 交互状态
  └── 超时(5s) → 恢复穿透 → 待机
```

**Tauri 配置：**

```json
{
  "app": {
    "windows": [{
      "label": "pet",
      "transparent": true,
      "decorations": false,
      "alwaysOnTop": true,
      "focus": false
    }]
  }
}
```

### 3.2 系统托盘

**职责：** 菜单栏常驻图标、右键菜单唤出/隐藏/退出/设置

**菜单结构：**

```
[宠物图标]
├── 显示/隐藏宠物
├── 对话
├── ─────────
├── 设置
│   ├── LLM 配置
│   ├── 角色人设
│   └── 模型选择
├── 检查更新
├── ─────────
└── 退出
```

**快捷键：** `tauri-plugin-global-shortcut` 注册 `Ctrl+Shift+P` 唤出/隐藏。

### 3.3 Sidecar 管理

**职责：** 管理 Go PetCore 子进程的启动、停止、健康检查和自动重启

**生命周期：**

```
Tauri 启动
  │
  ├── 从 binaries/petcore-{arch} 找到二进制
  ├── spawn 子进程 (stdin/stdout pipe)
  │
  ├── 启动健康检查 (每 5 秒发送 "ping")
  ├── 子进程崩溃 → 自动重启 (最多 3 次)
  │
  └── Tauri 关闭 → 发送 shutdown → 等待退出
```

### 3.4 自动更新

**职责：** 启动时和运行时每 6 小时检查更新，后台静默下载，一键重启

**实现：** `tauri-plugin-updater`，GitHub Releases 增量包。

**配置：**

```json
{
  "plugins": {
    "updater": {
      "active": true,
      "endpoints": ["https://github.com/{owner}/{repo}/releases/latest/download/updater.json"],
      "pubkey": "..."
    }
  }
}
```

### 3.5 事件转发

**职责：** 从 sidecar stdout 读取 JSON 事件 → 解析 → 通过 `app.emit()` 推送给前端

```
// Rust 端事件转发循环
loop {
    let line = read_line(&mut sidecar.stdout);
    let event: WireEvent = serde_json::from_str(&line)?;
    app.emit("pet:event", event);
}
```

---

## 四、前端 (React + Live2D)

### 4.1 Live2D 渲染

- **渲染引擎：** PixiJS 8 + pixi-live2d-display
- **模型格式：** Live2D Cubism 4 (.model3.json)
- **动画控制：** 由 PetCore 状态机驱动，前端被动播放

### 4.2 聊天气泡

- 浮动气泡，跟随宠物位置
- 打字机效果逐字显示（streaming text chunk）
- 支持 Markdown 格式
- 显示情感标签和表情符号

### 4.3 设置面板

- LLM 配置：API Key / Base URL / 模型名
- 角色人设：性格描述 / 说话风格
- 记忆管理：查看 / 编辑 / 清除
- 宠物外观：Live2D 模型选择
- 通用：开机自启 / 透明度 / 音量

### 4.4 Tauri IPC 桥接

**文件：** `frontend/src/lib/bridge.ts`

```
// 调用命令
invoke('chat', { text: string })        // 发送对话
invoke('get_status') → PetStatus        // 获取状态
invoke('query_memory', { query })       // 查询记忆
invoke('get_config') → Config           // 获取配置
invoke('set_config', { key, value })    // 修改配置

// 监听事件
listen('pet:event', handler)            // 接收宠物事件流
```

**事件处理：**

```typescript
listen('pet:event', (e) => {
    switch (e.payload.kind) {
        case 'pet.speak':      // 显示聊天气泡
        case 'pet.action':     // 播放 Live2D 动画
        case 'pet.emotion':    // 更新表情
        case 'agent.thinking': // 显示思考状态
        case 'state.changed':  // 更新状态机
    }
});
```

---

## 五、PetCore 内核

### 5.1 核心引擎 + 状态机

**文件：** `petcore/internal/core/`, `petcore/internal/fsm/`

**引擎主循环：**

```
启动
  │
  ├── 加载配置 (config.toml)
  ├── 连接 SQLite (memory.db)
  ├── 启动事件循环
  │
  ├── [goroutine] sidecar server: 监听 stdin, 处理命令
  ├── [goroutine] 插件监控: fsnotify 监听 actions/scripts/ 变更
  │
  └── 事件循环: 接收内部事件 → 状态转换 → 通过 Sink 输出
```

**5 态 FSM：**

```
                    ┌──────────┐
          ┌────────►│  待机     │◄──────────────┐
          │         │ (穿透开)  │                │
          │         │ (随机动作) │               │
          │         └────┬─────┘                │
          │              │                      │
      对话结束         检测到用户               超时
          │              │                      │
          │         ┌────▼─────┐                │
          ├────────►│  关注     │────────────────┤
          │         │ (穿透关)  │                │
          │         │ (转向鼠标) │                │
          │         └────┬─────┘                │
          │              │                      │
          │         用户对话/点击                 │
          │              │                      │
          │         ┌────▼─────┐                │
          │         │  交互     │                │
          │         │ (对话/   │────思考中───────┤
          │         │  表情动画)│ (等待 LLM 回复) │
          │         └────┬─────┘                │
          │              │                      │
          │         LLM 流式回复                  │
          │              │                      │
          │         ┌────▼─────┐                │
          │         │  说话     │                │
          │         │ (打字机  │────────────────┤
          │         │  口型同步)│                │
          │         └──────────┘                │
          └─────────────────────────────────────┘
```

**状态 → 窗口行为映射：**

| 状态 | 鼠标穿透 | 置顶 | Live2D 动画 |
|------|---------|------|-------------|
| 待机 | ✅ 开启 | ✅ | 随机小动作（眨眼/晃头） |
| 关注 | ❌ 关闭 | ✅ | 转向鼠标方向 |
| 交互 | ❌ 关闭 | ✅ | 表情随对话变化 |
| 说话 | ❌ 关闭 | ✅ | 口型同步 |

### 5.2 AI Agent + Pipeline

**文件：** `petcore/internal/agent/`

**Agent 主循环：**

```
用户消息 → 构建上下文 → 调用 Provider.Stream()
  │
  ├── Text Chunk  → 实时推送前端 (打字机效果)
  ├── Tool Call   → 执行工具 → 结果注入 → 继续
  └── Done        → 后处理 → 提取事实 → 写入记忆
```

**Pipeline 洋葱模型：**

```
输入 → PreProcess → Memory → Context → LLMCall → PostProcess → Response → 输出
                                           │
                                     工具执行结果注入 (递归)
```

| Stage | 职责 |
|-------|------|
| PreProcess | 消息标准化、敏感词过滤、日志 |
| Memory | 注入 L1 短期 + L2 核心 + L3 长期记忆 |
| Context | 构建 system prompt + 工具定义 |
| LLMCall | 调用 Provider.Stream()，流式解析 |
| PostProcess | 提取新事实 → 写入记忆 |
| Response | 触发宠物动作/情绪 |

### 5.3 LLM 适配层

**文件：** `petcore/internal/llm/`

**Provider 接口：**

```go
type Provider interface {
    Name() string
    Model() string
    Stream(ctx, Request) (<-chan Chunk, error)
    Chat(ctx, Request) (Response, error)
}
```

**注册表模式：**

```go
func init() {
    Register("openai", func(cfg) (Provider, error) {
        return &OpenAIProvider{cfg}, nil
    })
    Register("ollama", func(cfg) (Provider, error) {
        return &OllamaProvider{cfg}, nil
    })
}
```

**通过 blank import 触发注册：**

```go
import (
    _ "petcore/internal/llm/openai"
    _ "petcore/internal/llm/ollama"
)
```

### 5.4 记忆系统

**文件：** `petcore/internal/memory/`

**三层架构：**

| 层级 | 存储 | 生命周期 | 容量 | 内容 |
|------|------|---------|------|------|
| L1 短期 | 内存（环形缓冲区） | 当前会话 | 最近 20 轮 | 对话上下文 |
| L2 核心 | SQLite（key-value） | 永久 | 不限 | 用户姓名/偏好/事实 |
| L3 长期 | SQLite + FTS5 | 长期（可衰减） | 不限 | 对话摘要/历史事件 |

**工作流：**

```
用户说话 → L1 短期追加
  │
  ├── LLM 发现新信息 → remember 工具调用
  │   ├── 结构化事实 → L2 核心 (SQLite)
  │   └── 事件性事实 → L3 长期 (FTS5)
  │
  └── 下次对话
      ├── 读取 L2 → 用户偏好
      ├── 检索 L3 → FTS5 关键词 + 语义匹配
      └── 融合 → 构建 context
```

**提取策略：**
- **主动提取（推荐）：** LLM 在对话中调用 `remember` 工具写入
- **后台异步：** 对话结束后 Extractor 分析对话内容，去重后写入

### 5.5 插件系统

**文件：** `petcore/internal/plugin/`

**三层架构：**

```
L1: YAML 动作包 (actions/*.yaml)
    声明式配置，定义 触发器 + 动作序列
    热加载: fsnotify 监控，文件变更自动重载
    适用: 非开发者用户

L2: JS 脚本 (scripts/*.js)
    引擎: goja (纯 Go JS 引擎, 无需 V8)
    沙箱: 限制文件/网络访问
    热加载: fsnotify 监控
    适用: 进阶用户

L3: MCP 工具 (外部子进程)
    协议: stdio JSON-RPC
    隔离: 独立进程 + 权限声明
    适用: 开发者 / 生态扩展
```

### 5.6 工具系统

**文件：** `petcore/internal/tool/`

- **注册中心：** `Register(name, fn)` 注册内置工具
- **MCP 发现：** `MCPClient.ListTools()` 从 L3 插件自动发现
- **调度：** Agent 工具调用循环中自动选择和执行

**内置工具：**

| 工具名 | 说明 | 阶段 |
|--------|------|------|
| `web_search` | 联网搜索 | Phase 4 |
| `remember` | 记住用户事实 | Phase 1 |
| `desktop_ctrl` | 桌面控制（打开应用/文件操作） | Phase 6 |

### 5.7 CLI Shell

**文件：** `petcore/cmd/petcore/main.go`（CLI 模式）

**技术：** bubbletea (Go TUI 框架)

**命令：**

| 命令 | 说明 |
|------|------|
| `petcore chat` | 交互式对话（REPL） |
| `petcore run "..."` | 单次对话 |
| `petcore status` | 查看宠物状态 |
| `petcore memory query "..."` | 查询记忆 |
| `petcore memory list` | 列出所有记忆 |
| `petcore config set k=v` | 修改配置 |
| `petcore plugin list` | 列出已加载插件 |

### 5.8 配置管理

**文件：** `petcore/internal/config/`

**格式：** TOML

**路径：** `~/.desktop-pet/config.toml`

**核心配置项：**

```toml
[llm]
provider = "openai"
model = "deepseek-v4-flash"
base_url = "https://api.deepseek.com"
api_key_env = "DEEPSEEK_API_KEY"   # 从环境变量读取

[agent]
system_prompt = "你是一只可爱的猫娘..."
temperature = 0.7
max_tool_turns = 10

[window]
always_on_top = true
transparent = true
opacity = 1.0

[update]
auto_check = true
channel = "stable"

[telemetry]
enabled = false
```

---

## 六、通信协议

### 6.1 统一事件格式

所有跨层通信使用统一 `WireEvent` 结构：

```go
type WireEvent struct {
    Kind string          `json:"kind"`       // 事件类型
    Data json.RawMessage `json:"data"`       // 事件 payload
    Meta EventMeta       `json:"meta"`       // 元信息
}

type EventMeta struct {
    Timestamp int64  `json:"ts"`
    Source    string `json:"source"`   // core / plugin / user
    SessionID string `json:"session_id"`
}
```

### 6.2 事件类型全表

| Kind | 方向 | Data 结构 | 说明 |
|------|------|--------|------|
| `state.changed` | PetCore → 前端 | `{ from, to }` | 状态机切换 |
| `pet.speak` | PetCore → 前端 | `{ text, emotion, animation }` | 宠物说话 |
| `pet.action` | PetCore → 前端 | `{ action, duration }` | 播放 Live2D 动作 |
| `pet.emotion` | PetCore → 前端 | `{ mood, intensity }` | 情绪变化 |
| `agent.thinking` | PetCore → 前端 | `{ status }` | AI 思考中/完成 |
| `agent.reply` | PetCore → 前端 | `{ text, done }` | 流式回复片段 |
| `memory.updated` | PetCore → 前端 | `{ summary }` | 记忆更新通知 |
| `error` | PetCore → 前端 | `{ code, message }` | 错误通知 |

### 6.3 Sidecar 命令协议

**Tauri → PetCore (stdin)：**

```json
{"type":"cmd","id":"1","method":"chat","params":{"text":"你好"}}
{"type":"cmd","id":"2","method":"get_status","params":{}}
{"type":"cmd","id":"3","method":"ping","params":{}}
```

**PetCore → Tauri (stdout)：**

```json
{"type":"resp","id":"1","result":{"done":true}}
{"type":"event","event":"pet.speak","data":{"text":"你好~","emotion":"happy"}}
{"type":"event","event":"pet.action","data":{"action":"wave","duration":2000}}
{"type":"error","id":"2","code":"NOT_READY","message":"pet not initialized"}
```

---

## 七、分阶段实施路线

### Phase 1: MVP — 看得见、聊得来

- [ ] Tauri 透明窗口 + 基础 Live2D 显示
- [ ] 宠物状态机（待机/关注/交互/说话）
- [ ] 聊天气泡 UI + Tauri IPC 通信
- [ ] Go PetCore 引擎 + sidecar 通信
- [ ] LLM 适配层 + Provider 注册表
- [ ] 配置面板（API Key / 模型选择）
- [ ] 本地 SQLite 记忆（L1 短期 + L2 核心）
- [ ] 基础拖拽 + 右键菜单 + 系统托盘

**不做的：** 语音、情感、L3 长期记忆、工具、插件

### ✅ 代码质量基线（贯穿所有 Phase）

以下质量要求随代码编写逐步落地，不单独规划 Phase：

| 要求 | 适用语言 | 期望 Phase | 说明 |
|------|---------|-----------|------|
| **错误追踪 (Sentry)** | Rust + TS | Phase 1 接入 | 桌面应用崩溃报告标准，opt-in 用户授权 |
| **核心数据结构模糊测试** | Go (`go test -fuzz`) | Phase 1 | 对 FSM、记忆系统做 fuzz testing |
| **单元测试覆盖率 ≥60%** | Go + Rust + TS | Phase 1 | CI 已配置阈值门禁 |
| **e2e 集成测试** | TS (Playwright) | Phase 2 | 模拟用户操作流程 |
| **基准回归检查** | Go (`benchstat`) + Rust (`criterion`) | Phase 2 | 核心路径性能不退化 |
| **国际化 (i18n)** | TS (`react-intl` / `i18next`) | Phase 3 | 界面文字外部化 |
| **无障碍 (a11y)** | TS (`eslint-plugin-jsx-a11y`) | Phase 3 | 键盘导航 + 屏幕阅读器 |
| **性能预算（体积/内存）** | 构建时检查 | Phase 2 | 打包体积不超预算 |

### Phase 2: 记忆系统完整化

- [ ] L3 长期记忆：FTS5 全文检索
- [ ] 自动摘要 + 事实提取（L2+L3）
- [ ] 记忆编辑 UI（查看/修改/删除）
- [ ] 记忆演进机制（提取/归纳/遗忘/修正）

### Phase 3: 情感与主动行为

- [ ] 情绪系统（心情值 + 6 种情绪标签）
- [ ] Live2D 表情联动
- [ ] 主动问候/关心/回忆
- [ ] 每日日记/小结

### Phase 4: 工具与扩展

- [ ] MCP 协议接入 + 联网搜索工具
- [ ] L1 YAML 动作包 + fsnotify 热加载
- [ ] L2 JS 脚本插件 + goja 引擎
- [ ] L3 MCP 子进程插件
- [ ] 插件市场基础结构

### Phase 5: 语音交互

- [ ] 语音输入（Whisper / Vosk）
- [ ] 语音输出（TTS）
- [ ] Live2D 口型同步
- [ ] 语音打断

### Phase 6: 桌面控制 🚀

- [ ] 通过对话打开应用
- [ ] 文件操作
- [ ] 浏览器控制
- [ ] 日程管理
- [ ] 权限控制（可审计、可撤销）

---

## 八、打包与发布

### 8.1 用户视角

```
1. 下载 DesktopPet.dmg (~25MB) → 拖入 Applications
2. 双击打开 → 宠物出现在桌面上 🎉
3. 右键托盘 → 设置 → 填入 API Key
4. 点击宠物 → 开始聊天
5. 右键 → 「检查更新」→ 自动下载 → 一键重启
```

### 8.2 构建流程

```bash
# 1. 编译 Go PetCore (CGO_ENABLED=0)
cd petcore
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
  go build -o ../src-tauri/binaries/petcore-aarch64-apple-darwin ./cmd/petcore/

# 2. 构建前端
cd frontend
pnpm build

# 3. 打包 .app（自动包含 petcore + 前端）
cd ../src-tauri
cargo tauri build

# 产物: src-tauri/target/release/bundle/dmg/DesktopPet-1.0.0.dmg (~25MB)
```

### 8.3 体积预算

| 组件 | 大小 |
|------|------|
| Tauri 壳 (Rust + WebView) | ~10-15 MB |
| Go PetCore 静态二进制 | ~8-12 MB |
| 前端编译产物 (React + PixiJS) | ~2-5 MB |
| Live2D 基础模型 | ~3-5 MB |
| **.app 首次下载** | **~23-37 MB** |

### 8.4 更新策略

| 策略 | 实现 |
|------|------|
| 自动检查 | 每次启动 + 运行时每 6 小时 |
| 自动下载 | 检测到更新后立即后台静默下载 |
| 安装时机 | 用户确认后重启安装（不强制） |
| 更新日志 | 从 GitHub Release body 获取 |
| 网络容错 | 下载失败 → 下次重试 |
| 断点续传 | tauri-plugin-updater 原生支持 |
| 启动崩溃回滚 | 首次启动标记成功，下次未成功自动回退 |

---

## 九、隐私与安全

| 层面 | 措施 |
|------|------|
| 数据传输 | 前端 ←→ PetCore：本机进程内 IPC，不经过外网 |
| 外部 API | 仅对话文本发送给 LLM API（可配置） |
| 本地存储 | 所有数据存储在 `~/.desktop-pet/`，不上传 |
| API Key | 本地存储，仅用于调用 LLM 服务 |
| 桌面控制 | 需要用户明确授权，所有操作可审计、可撤销 |

---

## 十、参考项目

| 项目 | ⭐ | 参考价值 |
|------|----|---------|
| [Reasonix](https://github.com/esengine/DeepSeek-Reasonix) | 26,600 | Go 内核 + 多壳架构，Provider 注册表模式 |
| [bongo-cat-next](https://github.com/liwenka1/bongo-cat-next) | 130 | Tauri + Live2D 桌宠，透明窗口 + 鼠标穿透 |
| [my-neuro](https://github.com/morettt/my-neuro) | 1,303 | AI 桌宠后端，记忆系统 + MCP 集成 |
| [NyaDeskPet](https://github.com/gameswu/NyaDeskPet) | 9 | Pipeline Agent 架构，插件 WebSocket 扩展 |
| [Desktop-pet](https://github.com/git2968/Desktop-pet) | 4 | MCP 工具扩展，Agent Skills 系统 |
| [MemOS](https://github.com/MemTensor/MemOS) | — | 面向 LLM Agent 的记忆操作系统 |
| [Espanso](https://github.com/espanso/espanso) | 14,100 | 热重载 YAML 包管理器，L1 插件参考 |
| [Open-LLM-VTuber](https://github.com/Open-LLM-VTuber/Open-LLM-VTuber) | 12,466 | 语音管线 + Live2D 口型同步 |
| [Glances](https://github.com/nicolargo/glances) | 33,100 | 核心-UI 分离参考，多 Shell 复用内核 |

---

## 附录：关键设计决策

| 决策 | 选择 | 备选 | 理由 |
|------|------|------|------|
| 架构 | Tauri v2 + Go PetCore 内核分离 | Electron + Python / Tauri + Rust | 安装包 ~15MB，内存 ~40MB，Go 生态成熟 |
| 桌面壳 | Tauri v2 (Rust) | Electron / Wails | 原生鼠标穿透 + 自动更新 |
| 后端语言 | Go (CGO_ENABLED=0) | Python / Rust | 单静态二进制，多壳复用 |
| 可视化 | Live2D via PixiJS | Spritesheet / VRM 3D | QQ 宠物风格 |
| AI 协议 | Provider 接口 + 注册表 | LangChain / 自定义 | 最广泛兼容性 |
| 工具协议 | MCP | 自研工具系统 | 行业标准 |
| 记忆存储 | SQLite + FTS5 | ChromaDB / Qdrant | 纯 Go 零 CGO 依赖 |
| 打包方式 | Tauri bundle + Go cross-compile | PyInstaller / electron-builder | 零系统依赖 |
| 更新机制 | tauri-plugin-updater | electron-updater | 原生集成，增量更新 |
