# 贡献指南

感谢你对 Desktop Pet 的关注！请花几分钟阅读以下指南。

## 开发环境

参考 [快速上手](../docs/概览/快速上手.md) 搭建开发环境。最低要求：

- Rust 1.80+
- Go 1.22+
- Node.js 20+ (推荐 22+)
- pnpm 9+
- Tauri CLI v2

## 代码规范

每层语言有其独立的 linter 配置，提交前必须全部通过：

```bash
# Go
cd petcore && golangci-lint run ./...

# Rust
cd src-tauri && cargo clippy -- -D warnings && cargo fmt --check

# TypeScript
cd frontend && pnpm lint && pnpm format:check
```

详细规范见 [代码规范](../docs/开发指南/代码规范.md)。

## 测试要求

```bash
# 全部测试
cd petcore && go test ./...
cd src-tauri && cargo test
cd frontend && pnpm test
```

- 新增代码必须配测试
- 现有测试失败必须修复
- 使用 `t.Parallel()` 并行执行

## 提交信息格式

每条提交必须使用中文描述，推荐使用 Conventional Commits 前缀：

```
feat: 添加 LLM Provider 注册表
fix: 修复 sidecar 崩溃后无法重启
docs: 更新构建文档
refactor: 重构状态机事件处理
test: 添加 FSM 模糊测试
```

## 分支策略

| 分支 | 用途 |
|------|------|
| `main` | 稳定发布分支，只接受 PR |
| `develop` | 开发集成分支 |
| `feat/*` | 功能分支，从 develop 切出 |
| `fix/*` | 修复分支 |

## PR 流程

1. 从 `develop` 切出功能分支
2. 实现代码 + 测试
3. 确保所有 linter 和测试通过
4. 提交 PR 到 `develop`
5. 等待 Code Review

## 提问

Issues 请使用相应模板提交：
- 🐛 Bug 报告 → 使用 `bug_report.yml` 模板
- 💡 功能建议 → 使用 `feature_request.yml` 模板
