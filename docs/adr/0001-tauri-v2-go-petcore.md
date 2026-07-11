# ADR 0001: 采用 Tauri v2 + Go PetCore 内核-壳分离架构

- **日期**: 2025-07
- **状态**: ✅ 已采纳

## 背景

桌面宠物应用需要同时满足：桌面窗口管理、Live2D 渲染、AI 对话、记忆存储、插件扩展等多个需求。需要一个既能提供原生桌面体验，又能灵活扩展的架构。

## 决策

采用 **内核-壳分离架构**：
- **壳层**: Tauri v2 (Rust) — 窗口管理、系统托盘、自动更新、子进程管理
- **内核**: Go PetCore (CGO_ENABLED=0) — AI Agent、记忆系统、插件系统、状态机
- **前端**: React 18 + TypeScript + PixiJS 8 — Live2D 渲染、聊天气泡、设置面板

通信方式：Tauri IPC（前端↔壳）+ sidecar stdin/stdout JSON（壳↔内核）。

## 备选方案

| 方案 | 优点 | 缺点 |
|------|------|------|
| **Tauri v2 + Go** (选定) | 安装包 ~25MB, 内存 ~40MB, 内核可复用 | 需学 Rust + 两种语言 |
| Electron + Python | 生态成熟 | 安装包 ~210MB, 内存 ~100MB+ |
| Wails v2 (Go) | 单一语言 | 鼠标穿透需自研, 自动更新不完善 |

## 影响

- 开发者需要同时掌握 Rust（壳）、Go（内核）、TypeScript（前端）
- 内核可被 Desktop / CLI / Server 三种壳复用
- 事件驱动设计确保各层松耦合
- 参考项目: Reasonix (Go 多壳), bongo-cat-next (Tauri + Live2D)
