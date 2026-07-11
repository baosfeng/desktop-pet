#!/usr/bin/env bash
# devcontainer 初始化脚本（onCreateCommand）
set -euo pipefail

echo "📦 安装系统依赖（Tauri 需要）..."
if command -v apt-get &>/dev/null; then
  sudo apt-get update
  sudo apt-get install -y \
    libwebkit2gtk-4.1-dev \
    build-essential \
    curl \
    wget \
    file \
    libxdo-dev \
    libssl-dev \
    libayatana-appindicator3-dev \
    librsvg2-dev \
    librsvg2-bin
fi

echo "📦 安装 Tauri CLI..."
cargo install tauri-cli --version "^2" 2>/dev/null || true

echo "📦 安装 Go 工具..."
go install mvdan.cc/gofumpt@latest 2>/dev/null || true
go install github.com/air-verse/air@latest 2>/dev/null || true

echo "✅ devcontainer 初始化完成"
