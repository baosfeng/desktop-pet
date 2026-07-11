# 安全策略

## 报告漏洞

如果发现安全漏洞，**请不要公开提交 Issue**。请通过以下方式私下报告：

1. **GitHub Security Advisory**: 访问仓库 → Security → Report a vulnerability
2. **Email**: 发送至 [security@desktop-pet.app]（如已配置）

我们会在 **48 小时内**确认收到，并尽快修复。

## 支持的版本

| 版本 | 支持状态 |
|------|---------|
| >= 1.0.0 | ✅ 安全更新 |
| < 1.0.0 (预发布) | ⚠️ 仅严重漏洞 |

## 安全措施

本项目采用以下安全实践：

| 措施 | 工具 |
|------|------|
| 依赖漏洞扫描 | Dependabot（每周自动 PR） |
| Rust 依赖审计 | cargo-audit（CI 每次运行） |
| 代码安全分析 | CodeQL（Go/TS/Rust 三语言） |
| 密钥泄露检测 | pre-commit hook（lefthook） |
| Go 代码安全检查 | gosec（golangci-lint 集成） |

## API Key 管理

本项目使用 LLM API Key，必须安全管理：

| 要求 | 说明 |
|------|------|
| **环境变量** | API Key 通过 `LLMConfig.APIKeyEnv` 指定的环境变量读取，绝不硬编码在代码或配置文件中 |
| **最小权限** | 使用的 API Key 仅赋予 LLM 调用权限，不赋予管理/计费权限 |
| **轮换周期** | 建议每 **90 天** 轮换一次 API Key |
| **密钥管理工具** | 推荐使用 1Password CLI、Doppler 或 macOS Keychain 管理环境变量 |
| **泄露响应** | 发现泄露立即在 LLM 服务商控制台吊销密钥，然后更新环境变量 |

## 漏洞处理流程

1. 报告者提交漏洞
2. 维护者确认并评估严重性
3. 开发修复补丁
4. 发布安全更新版本
5. 公开披露（修复后 30 天）
