<div align="center">
  <img src="docs/assets/logo.png" alt="Desktop Pet" width="120" />

  # 桌面 AI 宠物 🐾

  **一个像 QQ 宠物一样在桌面上陪伴你，接入 AI 大模型、越用越懂你的桌面伴侣。**

  [![CI](https://github.com/user/desktop-pet/actions/workflows/ci.yml/badge.svg)](https://github.com/user/desktop-pet/actions/workflows/ci.yml)
  [![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
  [![Rust](https://img.shields.io/badge/Rust-1.80+-orange)](https://www.rust-lang.org)
  [![Go](https://img.shields.io/badge/Go-1.23+-blue)](https://go.dev)
  [![Tauri](https://img.shields.io/badge/Tauri-v2-purple)](https://tauri.app)

  [English](README.md) | [简体中文](docs/概览/项目简介.md)

</div>

---

## ✨ 特性

| 阶段 | 特性 | 状态 |
|------|------|------|
| Phase 1 | 🖥️ 透明窗口 + Live2D 宠物显示 | 📋 规划中 |
| Phase 1 | 💬 AI 对话（OpenAI / DeepSeek / Ollama） | 📋 规划中 |
| Phase 1 | 🧠 记忆系统（三层架构） | 📋 规划中 |
| Phase 2 | 📖 长期记忆 + 全文检索 | ⏳ 待开始 |
| Phase 3 | ❤️ 情感系统 + 主动行为 | ⏳ 待开始 |
| Phase 4 | 🔌 插件系统（YAML/JS/MCP） | ⏳ 待开始 |
| Phase 5 | 🎤 语音交互 | 🔮 未来 |
| Phase 6 | 🖱️ 桌面控制 | 🔮 未来 |

> **最终目标：通过自然对话控制电脑。**

---

## 🚀 快速开始

### 下载

从 [Releases 页面](https://github.com/user/desktop-pet/releases) 下载最新 `.dmg`（macOS）或 `.msi`（Windows）。

### 从源码构建

```bash
# 1. 前置要求
#    - Rust 1.80+ (rustup)
#    - Go 1.23+
#    - Node.js 22+ (推荐)
#    - pnpm 9+  (npm i -g pnpm)
#    - Tauri CLI v2 (cargo install tauri-cli --version "^2")

# 2. 启动开发环境
git clone https://github.com/user/desktop-pet.git
cd desktop-pet
make dev

# 3. 生产构建
make build
# 产物: src-tauri/target/release/bundle/dmg/DesktopPet-*.dmg (~25MB)
```

详细步骤见 [快速上手](docs/概览/快速上手.md)。

---

## 🏗️ 架构

```
┌──────────────────────────────────────────┐
│          DesktopPet.app (~25MB)           │
│                                           │
│  ┌────────────────────────────────────┐  │
│  │  Tauri Rust 壳                      │  │
│  │  - 透明窗口 / 鼠标穿透 / 置顶       │  │
│  │  - 系统托盘 / 全局快捷键            │  │
│  │  - sidecar 管理 / 自动更新          │  │
│  └──────────┬─────────────────────────┘  │
│             │ Tauri IPC                   │
│  ┌──────────▼─────────────────────────┐  │
│  │  React 前端 (WebView)              │  │
│  │  - PixiJS 8 + Live2D              │  │
│  │  - 聊天气泡 / 设置面板             │  │
│  └────────────────────────────────────┘  │
│                                           │
│  ┌────────────────────────────────────┐  │
│  │  Go PetCore (sidecar, ~10MB)       │  │
│  │  - 状态引擎 / AI Agent             │  │
│  │  - 记忆系统 / 插件系统             │  │
│  │  - CGO_ENABLED=0 静态二进制        │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

设计详情见 [架构总览](docs/概览/架构总览.md)。

---

## 🧰 技术栈

| 层 | 技术 | 用途 |
|----|------|------|
| 桌面容器 | [Tauri v2](https://tauri.app) (Rust) | 系统原生 WebView，透明窗口 |
| 渲染引擎 | [PixiJS 8](https://pixijs.com) + [pixi-live2d-display](https://github.com/guansss/pixi-live2d-display) | Live2D 模型加载和动画 |
| 前端 UI | [React 18](https://react.dev) + [TypeScript 5](https://www.typescriptlang.org) + [Vite 6](https://vitejs.dev) | 组件化 UI |
| 后端内核 | [Go 1.23+](https://go.dev) (CGO_ENABLED=0) | AI Agent、记忆管理、插件 |
| 记忆存储 | SQLite + FTS5 ([modernc.org/sqlite](https://modernc.org/sqlite)) | 纯 Go，零 CGO |
| LLM 协议 | Provider 接口 + 注册表 | 统一接入 OpenAI/DeepSeek/Ollama |
| 工具协议 | [MCP](https://modelcontextprotocol.io) | 行业标准工具扩展 |
| 通信 | Tauri IPC + sidecar stdin/stdout JSON | 零网络开销 |
| 自动更新 | [tauri-plugin-updater](https://github.com/tauri-apps/plugins-workspace/tree/v2/plugins/updater) | GitHub Releases |

---

## 📚 文档

| 文档 | 说明 |
|------|------|
| [设计规范](docs/superpowers/specs/2026-07-09-desktop-pet-design.md) | 完整设计文档 |
| [架构总览](docs/概览/架构总览.md) | 系统架构概览 |
| [快速上手](docs/概览/快速上手.md) | 开发环境搭建 |
| [前端文档](docs/前端/索引.md) | React/PixiJS/Live2D |
| [后端文档](docs/后端/索引.md) | Go PetCore 内核 |
| [代码规范](docs/开发指南/代码规范.md) | 编码风格和 linter 规则 |
| [贡献指南](.github/CONTRIBUTING.md) | 参与贡献 |

## 🤝 贡献

欢迎贡献！请阅读：

1. [贡献指南](.github/CONTRIBUTING.md)
2. [行为准则](.github/CODE_OF_CONDUCT.md)
3. 使用 [Issue 模板](.github/ISSUE_TEMPLATE/) 提交问题

## 📄 许可证

本项目基于 [MIT 许可证](LICENSE) 开源。

---

<div align="center">
  Made with ❤️ and 🐹
</div>
