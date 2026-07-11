#!/usr/bin/env bash
# ============================================
# build.sh — 生产构建（Go PetCore + 前端 + Tauri）
# 支持多平台打包 + macOS 公证
# ============================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PETCORE_DIR="$ROOT_DIR/petcore"
FRONTEND_DIR="$ROOT_DIR/frontend"
SRC_TAURI_DIR="$ROOT_DIR/src-tauri"

# 默认构建目标平台（当前平台）
TARGET_OS="${TARGET_OS:-$(go env GOOS)}"
TARGET_ARCH="${TARGET_ARCH:-$(go env GOARCH)}"

# 多平台构建（all 表示构建所有支持的平台）
BUILD_ALL="${BUILD_ALL:-false}"

# 包格式（默认 macOS dmg，支持 dmg / msi / deb / appimage）
BUNDLES="${BUNDLES:-dmg}"

# 版本号（从 git tag 获取，默认 0.1.0）
VERSION="$(git describe --tags --abbrev=0 2>/dev/null || echo '0.1.0')"

echo ""
echo "=============================="
echo "📦 Desktop Pet — Build v$VERSION"
echo "平台: $TARGET_OS/$TARGET_ARCH"
echo "包格式: $BUNDLES"
echo "全平台: $BUILD_ALL"
echo "=============================="
echo ""

# ---- Sidecar 命名映射 ----
sidecar_name() {
  local os="$1" arch="$2"
  case "${os}-${arch}" in
    darwin-arm64)  echo "petcore-aarch64-apple-darwin" ;;
    darwin-amd64)  echo "petcore-x86_64-apple-darwin" ;;
    windows-amd64) echo "petcore-x86_64-pc-windows-msvc.exe" ;;
    linux-amd64)   echo "petcore-x86_64-unknown-linux-gnu" ;;
    *) echo "" ;;
  esac
}

# ---- Step 1: 编译 Go PetCore ----
build_petcore() {
  local os="$1" arch="$2"
  local name
  name="$(sidecar_name "$os" "$arch")"
  [ -z "$name" ] && { echo "   ⏭️  跳过 $os/$arch（不支持的组合）"; return; }

  echo "🔨 编译 PetCore ($os/$arch)..."
  cd "$PETCORE_DIR"
  [ ! -f go.sum ] && go mod tidy

  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -ldflags="-s -w -X main.version=$VERSION" \
             -o "$SRC_TAURI_DIR/binaries/$name" \
             ./cmd/petcore/
  echo "   ✅ $name"
}

if [ "$BUILD_ALL" = "true" ]; then
  echo "🔨 [1/3] 编译 Go PetCore（全平台）..."
  build_petcore darwin arm64
  build_petcore darwin amd64
  build_petcore linux amd64
  build_petcore windows amd64
else
  echo "🔨 [1/3] 编译 Go PetCore..."
  build_petcore "$TARGET_OS" "$TARGET_ARCH"
fi
echo ""

# ---- Step 2: 构建前端 ----
echo "🔨 [2/3] 构建前端..."
cd "$FRONTEND_DIR"
if [ ! -d node_modules ]; then
  pnpm install --frozen-lockfile
fi
pnpm build
echo "   ✅ 前端构建完成"
echo ""

# ---- Step 3: 打包 Tauri ----
echo "🔨 [3/3] 打包 Tauri..."
cd "$SRC_TAURI_DIR"

# 如果启用全平台构建，为各平台分别打包
if [ "$BUILD_ALL" = "true" ]; then
  echo "   ⚠️  全平台 Tauri 打包请在对应平台 CI 中执行"
  echo "   ⏭️  当前仅验证本地平台构建"
fi

cargo tauri build --bundles "$BUNDLES"

echo ""
echo "=============================="
echo "✅ 构建完成!"
echo "产物: $SRC_TAURI_DIR/target/release/bundle/"
echo "=============================="
echo ""
echo "📋 macOS 公证提示（如果需要分发）:"
echo "   设置以下环境变量后重新运行:"
echo "     export APPLE_ID=\"your@apple.id\""
echo "     export APPLE_PASSWORD=\"@keychain:AC_PASSWORD\""
echo "     export APPLE_TEAM_ID=\"TEAM123456\""
echo "     export APPLE_SIGNING_IDENTITY=\"Developer ID Application\""
echo ""
echo "   或用 xcrun 手动公证:"
echo "     xcrun notarytool submit DesktopPet.dmg --apple-id \$APPLE_ID --password \$APPLE_PASSWORD --team-id \$APPLE_TEAM_ID --wait"
echo "     xcrun stapler staple DesktopPet.dmg"
