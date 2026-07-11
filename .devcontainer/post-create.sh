#!/usr/bin/env bash
# devcontainer 启动后脚本（postCreateCommand）
set -euo pipefail

echo "📦 安装前端依赖..."
cd frontend && pnpm install

echo "📦 初始化 Go 模块..."
cd ../petcore && go mod tidy

echo "✅ 项目依赖就绪"
echo ""
echo "🚀 启动开发环境: cd src-tauri && cargo tauri dev"
echo "   或分步调试: 见 docs/概览/快速上手.md"
