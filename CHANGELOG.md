# Changelog

所有显著变更均记录在此文件。
版本号规则见 [docs/开发指南/版本号规范.md](docs/开发指南/版本号规范.md)。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。
版本语义基于 [SemVer](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### ✨ 新功能
- 项目初始规划和文档体系搭建
- Tauri v2 + Go PetCore 架构设计
- 版本号规范文档 + 自动发布流水线
- 版本管理脚本 `scripts/bump-version.sh`

### 🔧 配置/CI
- 三层语言 linter 配置（Go/Rust/TypeScript）
- GitHub Actions CI 流水线（lint → test → build）
- 自动发布工作流 `auto-release.yml`（VERSION 变更 → 自动发版）
- 手动发布工作流 `release.yml`（tag push 触发）
- Pre-commit hooks (lefthook)
- 社区健康文件（CODE_OF_CONDUCT, CONTRIBUTING, Issue/PR 模板）
- 安全扫描（Dependabot, CodeQL, cargo-audit）
- goreleaser 跨平台 Go 编译
- 开发环境脚本（dev.sh / build.sh）
- 代码规范文档

### 📝 文档
- [版本号规范](docs/开发指南/版本号规范.md)：MAJOR.MINOR.PATCH 规则 + 自动发布流程
