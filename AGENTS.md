# desktop-pet — 桌面 AI 宠物

> 一个像 QQ 宠物一样在桌面上陪伴你，接入 AI 大模型、越用越懂你的桌面伴侣。
> 最终目标：通过自然对话控制电脑。

> ⚠️ **强制执行规则：代码修改前必须先读文档**
>
> 在修改任何代码前，必须按以下顺序执行，**禁止跳过**：
>
> 1. **定位** — 根据用户需求的关键词，在下方「操作前必读规范」表格中找到对应文档
> 2. **读取** — 用 `read_file` 读取对应的模块文档或规范文档
> 3. **理解** — 确认核心入口类路径和关键流程
> 4. **编码** — 理解完整上下文后，再开始修改代码
>
> ❌ **禁止在完成上述步骤前**直接执行 `explore`、`read_only_task` 或 `task`。
> ❌ **若知识图谱工具已安装则禁止使用 explore：** 如果当前环境已安装代码图/代码索引类 MCP 工具（如 `search_graph`、`search_code`、`trace_path`、`query_graph` 等），**禁止使用 `explore` 进行任何代码检索**，必须且只能使用知识图谱类工具。
>    - 项目尚未建立索引 → 先调用索引初始化功能建立索引（如 `index_repository`）
>    - 索引已存在 → 根据代码变更情况自行判断是否需要刷新
> ❌ **未安装此类工具时也禁止直接 explore：** 必须先读取已有文档（文档已标注精确源码路径）；文档中找不到所需信息时，才可 fallback 到 `explore`/`grep` 等工具。
>
> ⚠️ **AI 权限规则：所有副作用操作必须先 ask 用户**
>
> 在执行以下操作前，**必须使用 `ask` 工具以选择题形式征求用户意见**，不得用自然语言提问代替，不得擅自执行：
>
> - **git 提交/推送** — 展示变更清单后用 ask 确认
> - **删除文件/目录** — 用 ask 确认后再执行
> - **覆盖/替换已有内容** — 用 ask 确认后再执行
>
> **豁免规则：** 同一会话内，用户通过 `ask` 确认过某类操作后，后续同类操作自动授权（提交豁免 ≠ 删除豁免）。
>
> ⚠️ **任务验收标准规则：所有任务必须附带可验证的验收标准**
>
> 每一项任务（无论是子 Agent 还是主 Agent 执行）在完成时必须提供以下证据：
> 1. **编译通过** — `go build ./...` / `cargo build` / `pnpm build`
> 2. **测试通过** — `go test -race ./...` / `cargo test` / `pnpm test -- --run`
> 3. **Lint 通过** — 对应语言 linter 无 error
> 4. **覆盖率满足 ≥60%** — 增量代码覆盖率证明
> 5. **边界情况覆盖** — 空输入、错误输入、并发场景的测试用例
> 6. **Mock 独立** — 测试不依赖真实外部服务
>
> 缺少验收标准的任务不视为完成。详细规范见[质量基线](docs/开发指南/质量基线.md)。

## 📁 项目结构

```
desktop-pet/
├── src-tauri/            ★ Tauri v2 Rust 壳
│   ├── src/
│   │   ├── main.rs       # 入口 + 命令注册
│   │   ├── commands.rs   # Tauri 命令（前端可调用）
│   │   ├── sidecar.rs    # Go PetCore 子进程管理
│   │   ├── tray.rs       # 系统托盘
│   │   └── window.rs     # 窗口管理（透明/穿透/置顶）
│   ├── Cargo.toml
│   └── tauri.conf.json
├── frontend/             ★ React + PixiJS + Live2D 前端
│   ├── src/
│   │   ├── components/   # React 组件
│   │   ├── hooks/        # Tauri IPC hooks
│   │   └── lib/bridge.ts # Tauri IPC 封装
│   └── package.json
├── petcore/              ★ Go PetCore 内核
│   ├── cmd/petcore/main.go  # 入口（sidecar + CLI 双模式）
│   └── internal/
│       ├── core/        # 状态引擎 + 事件循环
│       ├── agent/       # AI Agent (Pipeline)
│       ├── memory/      # 记忆系统（3层）
│       ├── llm/         # LLM 适配层（Provider 注册表）
│       ├── plugin/      # 插件系统（L1 YAML / L2 JS / L3 MCP）
│       ├── tool/        # 工具系统
│       ├── event/       # 事件系统
│       ├── fsm/         # 状态机
│       ├── server/      # sidecar 通信
│       └── config/      # 配置加载
├── docs/                 ★ 项目文档
│   ├── 索引.md
│   ├── 参考/
│   │   ├── 优秀开源项目参考.md
│   │   ├── 架构规划方案.md
│   │   └── 迁移规划.md
│   ├── 前端/
│   ├── 后端/
│   ├── 开发指南/
│   └── 概览/
├── .github/                # GitHub 社区 + CI 配置
│   ├── CODE_OF_CONDUCT.md  # 贡献者公约
│   ├── CONTRIBUTING.md     # 贡献指南
│   ├── PULL_REQUEST_TEMPLATE.md
│   ├── ISSUE_TEMPLATE/     # Bug / Feature 模板
│   ├── dependabot.yml      # 自动依赖更新
│   └── workflows/
│       └── ci.yml          # Lint → Test → Build 流水线
│       ├── codeql.yml       # CodeQL 安全分析
│       └── release.yml      # 发布流水线（4 平台矩阵）
├── .devcontainer/           # 容器化开发环境
│   ├── devcontainer.json
│   ├── setup.sh
│   └── post-create.sh
├── scripts/               # 构建/开发脚本
│   ├── dev.sh             # 开发环境一键启动
│   └── build.sh           # 生产构建（Go + 前端 + Tauri）
├── Makefile               # 统一开发命令入口
├── CHANGELOG.md            # 版本变更日志
├── lefthook.yml           # Pre-commit 钩子（lint + secret 扫描）
├── AGENTS.md
```

## 📂 文档索引

| 当你需要... | 必须先阅读 |
|------------|-----------|
| 了解项目整体设计 | → **[设计规范](docs/superpowers/specs/2026-07-09-desktop-pet-design.md)** |
| 修改前端代码（React/PixiJS/Live2D） | → **[前端文档索引](docs/前端/索引.md)** |
| 修改 Tauri 壳代码（Rust） | → **[前端文档索引](docs/前端/索引.md)** |
| 修改 PetCore 内核代码（Go） | → **[后端文档索引](docs/后端/索引.md)** |
| 了解架构规划 | → **[架构规划方案](docs/参考/架构规划方案.md)** |
| 了解迁移计划 | → **[迁移规划](docs/参考/迁移规划.md)** |
| 查看参考项目 | → **[优秀开源项目参考](docs/参考/优秀开源项目参考.md)** |
| 创建/修改/删除文档 | → **[文档规范](docs/开发指南/文档规范.md)** |
| 提交代码到 Git | → **[提交规范](docs/开发指南/提交规范.md)** |
| 搭建开发环境 | → **[快速上手](docs/概览/快速上手.md)** |
| 排查已知问题 | → **[踩坑记录](docs/踩坑/)** |
| 了解代码规范/Lint 规则 | → **[代码规范](docs/开发指南/代码规范.md)** |
| 了解质量基线 + 任务验收标准规范 | → **[质量基线](docs/开发指南/质量基线.md)** |
| 了解贡献方式 | → `.github/CONTRIBUTING.md` |
| 提交 Bug/Feature | → `.github/ISSUE_TEMPLATE/` |
| 使用开发/构建脚本 | → `scripts/dev.sh` / `scripts/build.sh` |
| 使用 Makefile 命令 | → `make help`（`make dev` / `make test` / `make lint`）|
| 配置 Pre-commit 钩子 | → `lefthook.yml`（`brew install lefthook && lefthook install`） |

> ⚠️ **语法最新性规则：每次代码改动后，必须保证对应语言的 linter 通过**
> - Go: `cd petcore && golangci-lint run ./...`
> - Rust: `cd src-tauri && cargo clippy -- -D warnings && cargo fmt --check`
> - TypeScript: `cd frontend && pnpm lint && pnpm format:check`
> - 这些检查已配置在 CI（`.github/workflows/ci.yml`）和 pre-commit（`lefthook.yml`）中自动执行
>
> ⚠️ **代码质量基线（详见[质量基线](docs/开发指南/质量基线.md)）**
> - **Phase 1**: Sentry 错误追踪、FSM 模糊测试、覆盖率 ≥60%
> - **Phase 2**: e2e 测试、基准回归检查、性能预算（体积/内存）
> - **Phase 3+**: i18n 国际化（react-intl）、a11y 无障碍（jsx-a11y）

## 功能模块

### 前端（桌面端）
| 模块 | 说明 | 业务关键词 |
|------|------|-----------|
| 窗口管理器 (Tauri) | 透明无边框窗口、鼠标穿透、拖拽 | 窗口、透明、置顶、拖拽 |
| Live2D 渲染 | Live2D 模型加载、动画播放 | Live2D、动画、表情、口型同步 |
| 宠物状态机 | 待机→关注→交互→说话 | 状态机、行为、AI联动 |
| 聊天气泡 | 浮动聊天气泡 UI、打字机效果 | 聊天、气泡、对话、消息 |
| 设置面板 | LLM 配置、角色人设、记忆管理 | 设置、配置、人设、API Key |
| Tauri IPC 桥接 | 通过 Tauri IPC 与 PetCore 通信 | IPC、invoke、事件监听 |

### 后端（Go PetCore 内核）
| 模块 | 说明 | 业务关键词 |
|------|------|-----------|
| 核心引擎 | 状态机 + 事件循环 | 引擎、状态、事件 |
| AI Agent | 对话管理、LLM 调用、Pipeline 管线 | AI、Agent、对话、LLM |
| 记忆系统 | 三层记忆：短期/核心/长期 | 记忆、检索、长期记忆 |
| LLM 适配层 | Provider 注册表，OpenAI/Ollama | LLM、模型适配、Provider |
| 工具系统 | MCP 客户端、工具注册、联网搜索 | MCP、工具、联网、桌面控制 |
| 插件系统 | L1 YAML 动作包 / L2 JS 脚本 / L3 MCP | 插件、热加载、扩展 |
| CLI Shell | bubbletea 终端界面 | CLI、调试、管理 |

### Tauri 壳（Rust）
| 模块 | 说明 | 业务关键词 |
|------|------|-----------|
| 窗口管理器 | 透明窗口、鼠标穿透、置顶、多屏 | 窗口、透明、穿透 |
| 系统托盘 | 菜单栏图标、常驻后台 | 托盘、后台、唤出 |
| sidecar 管理 | Go PetCore 子进程生命周期 | sidecar、子进程、spawn |
| 自动更新 | tauri-plugin-updater | 更新、增量、静默 |
| 事件转发 | Rust ↔ JS 事件桥接 | 事件、IPC、桥接 |

## 关键文档链接

- → [设计规范](docs/superpowers/specs/2026-07-09-desktop-pet-design.md)
- → [前端文档索引](docs/前端/索引.md)
- → [后端文档索引](docs/后端/索引.md)
- → [架构规划方案](docs/参考/架构规划方案.md)
- → [迁移规划](docs/参考/迁移规划.md)
- → [优秀开源项目参考](docs/参考/优秀开源项目参考.md)
- → [代码规范](docs/开发指南/代码规范.md)
- → [文档规范](docs/开发指南/文档规范.md)
- → [提交规范](docs/开发指南/提交规范.md)
- → [构建与测试](docs/开发指南/构建与测试.md)
