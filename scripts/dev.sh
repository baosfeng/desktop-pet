#!/usr/bin/env bash
# ============================================
# dev.sh — 开发环境一键启动
# ============================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PETCORE_DIR="$ROOT_DIR/petcore"
FRONTEND_DIR="$ROOT_DIR/frontend"
SRC_TAURI_DIR="$ROOT_DIR/src-tauri"

echo "🚀 启动开发环境..."

# 检查前置工具
command -v go >/dev/null 2>&1 || { echo "❌ 需要安装 Go"; exit 1; }
command -v pnpm >/dev/null 2>&1 || { echo "❌ 需要安装 pnpm (npm i -g pnpm)"; exit 1; }
command -v cargo >/dev/null 2>&1 || { echo "❌ 需要安装 Rust"; exit 1; }

# 安装前端依赖
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
  echo "📦 安装前端依赖..."
  cd "$FRONTEND_DIR" && pnpm install
fi

# 检查 tauri-cli
if ! cargo tauri --version >/dev/null 2>&1; then
  echo "📦 安装 Tauri CLI..."
  cargo install tauri-cli --version "^2"
fi

echo ""
echo "▶️  启动方式选择:"
echo "  1) cargo tauri dev         — 完整开发（推荐）"
echo "  2) 分步启动               — 后端/前端/壳分别调试"
echo ""
echo "正在启动完整开发模式..."
cd "$SRC_TAURI_DIR" && cargo tauri dev
