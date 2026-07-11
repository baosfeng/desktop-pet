#!/usr/bin/env bash
# =============================================================================
# 版本号管理脚本
# 版本格式：MAJOR.MINOR.PATCH
#   MAJOR — 大改动/不兼容变更（breaking changes）
#   MINOR — 小改动/新功能（backward-compatible features）
#   PATCH — Bug 修复/微小调整（backward-compatible fixes）
#
# 用法：
#   ./scripts/bump-version.sh         # 显示当前版本
#   ./scripts/bump-version.sh major   # 升级主版本号（MAJOR +1, MINOR=0, PATCH=0）
#   ./scripts/bump-version.sh minor   # 升级次版本号（MINOR +1, PATCH=0）
#   ./scripts/bump-version.sh patch   # 升级补丁版本号（PATCH +1）
#   ./scripts/bump-version.sh set 1.2.3  # 设置指定版本
# =============================================================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VERSION_FILE="$ROOT_DIR/VERSION"

# 文件列表（需要同步版本号的目标）
declare -a TARGET_FILES=(
    "$ROOT_DIR/src-tauri/Cargo.toml"
    "$ROOT_DIR/src-tauri/tauri.conf.json"
)

# ─── 辅助函数 ─────────────────────────────────

# 读取当前版本
read_version() {
    if [[ ! -f "$VERSION_FILE" ]]; then
        echo "0.0.0" > "$VERSION_FILE"
    fi
    cat "$VERSION_FILE" | tr -d ' \t\n\r'
}

# 校验版本格式
validate_version() {
    local ver="$1"
    if ! echo "$ver" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
        echo "❌ 无效版本号: $ver（需为 MAJOR.MINOR.PATCH 格式，如 1.2.3）" >&2
        exit 1
    fi
}

# 写入 VERSION 文件
write_version_file() {
    local ver="$1"
    echo "$ver" > "$VERSION_FILE"
    echo "✅ VERSION: $ver"
}

# 更新 Cargo.toml 中的 version 字段
update_cargo_toml() {
    local ver="$1"
    local file="$ROOT_DIR/src-tauri/Cargo.toml"
    if [[ -f "$file" ]]; then
        if [[ "$(uname)" == "Darwin" ]]; then
            sed -i '' "s/^version = \".*\"/version = \"$ver\"/" "$file"
        else
            sed -i "s/^version = \".*\"/version = \"$ver\"/" "$file"
        fi
        echo "✅ Cargo.toml: version = \"$ver\""
    fi
}

# 更新 tauri.conf.json 中的 version 字段
update_tauri_conf() {
    local ver="$1"
    local file="$ROOT_DIR/src-tauri/tauri.conf.json"
    if [[ -f "$file" ]]; then
        if [[ "$(uname)" == "Darwin" ]]; then
            sed -i '' "s/\"version\": \".*\"/\"version\": \"$ver\"/" "$file"
        else
            sed -i "s/\"version\": \".*\"/\"version\": \"$ver\"/" "$file"
        fi
        echo "✅ tauri.conf.json: version = \"$ver\""
    fi
}

# 更新所有目标文件
update_all_targets() {
    local ver="$1"
    write_version_file "$ver"
    update_cargo_toml "$ver"
    update_tauri_conf "$ver"
}

# ─── 主逻辑 ─────────────────────────────────

CURRENT_VERSION=$(read_version)
echo "当前版本: $CURRENT_VERSION"

case "${1:-}" in
    major)
        IFS='.' read -r major minor patch <<< "$CURRENT_VERSION"
        NEW_VERSION="$((major + 1)).0.0"
        echo "🔼 升级主版本号: $CURRENT_VERSION → $NEW_VERSION"
        update_all_targets "$NEW_VERSION"
        ;;
    minor)
        IFS='.' read -r major minor patch <<< "$CURRENT_VERSION"
        NEW_VERSION="$major.$((minor + 1)).0"
        echo "🔼 升级次版本号: $CURRENT_VERSION → $NEW_VERSION"
        update_all_targets "$NEW_VERSION"
        ;;
    patch)
        IFS='.' read -r major minor patch <<< "$CURRENT_VERSION"
        NEW_VERSION="$major.$minor.$((patch + 1))"
        echo "🔼 升级补丁版本号: $CURRENT_VERSION → $NEW_VERSION"
        update_all_targets "$NEW_VERSION"
        ;;
    set)
        if [[ -z "${2:-}" ]]; then
            echo "❌ 用法: $0 set <version>" >&2
            exit 1
        fi
        validate_version "$2"
        echo "🔧 设置版本号: $2"
        update_all_targets "$2"
        ;;
    ""|show)
        echo ""
        echo "版本号规则: MAJOR.MINOR.PATCH"
        echo "  major = 大改动/不兼容变更"
        echo "  minor = 小改动/新功能"
        echo "  patch = Bug 修复/微小调整"
        echo ""
        echo "用法:"
        echo "  $0             显示当前版本"
        echo "  $0 major       升级主版本号"
        echo "  $0 minor       升级次版本号"
        echo "  $0 patch       升级补丁版本号"
        echo "  $0 set <ver>   设置指定版本"
        ;;
    *)
        echo "❌ 未知命令: $1" >&2
        echo "用法: $0 {major|minor|patch|set <version>}" >&2
        exit 1
        ;;
esac
