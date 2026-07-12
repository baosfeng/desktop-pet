# Desktop Pet — Makefile
# 跨平台统一命令入口（macOS / Linux / Windows via WSL）
# 如果不可用，各语言的底层命令见 docs/开发指南/构建与测试.md

.PHONY: help dev dev-petcore-cli build test lint fmt clean devcontainer-setup

help: ## 显示帮助
	@echo "Desktop Pet 开发命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── 开发 ─────────────────────────────────────

# ─── 开发环境变量 ─────────────────────────────
DEV_ENV := PETCORE_ENV=development PETCORE_DATA_DIR=$(HOME)/.desktop-pet-dev

dev: ## 启动完整开发环境（cargo tauri dev，需在项目根目录执行）
	@echo "🚀 启动 Tauri 开发环境..."
	$(DEV_ENV) cargo tauri dev

dev-frontend: ## 仅启动前端 Vite 开发服务器
	cd frontend && pnpm dev

dev-petcore: ## 仅启动 Go PetCore sidecar 模式（开发配置）
	$(DEV_ENV) cd petcore && go run ./cmd/petcore/

dev-petcore-cli: ## 启动 Go PetCore CLI 模式（交互式对话，开发配置）
	$(DEV_ENV) cd petcore && go run ./cmd/petcore/ --cli

dev-petcore-watch: ## 启动 Go PetCore + 热重载（需安装 air: go install github.com/air-verse/air@latest）
	@command -v air >/dev/null 2>&1 || { \
		echo "❌ 需要安装 air: go install github.com/air-verse/air@latest"; \
		exit 1; \
	}
	$(DEV_ENV) cd petcore && air

# ─── 构建 ─────────────────────────────────────

build: ## 生产构建（Go + 前端 + Tauri）
	./scripts/build.sh

build-petcore: ## 仅编译 Go PetCore（当前平台）
	cd petcore && CGO_ENABLED=0 go build -ldflags="-s -w" -o petcore ./cmd/petcore/

build-petcore-all: ## 编译 Go PetCore（所有平台）
	cd petcore && \
		CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o petcore-darwin-arm64 ./cmd/petcore/ && \
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o petcore-darwin-amd64 ./cmd/petcore/ && \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o petcore-linux-amd64 ./cmd/petcore/ && \
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o petcore-windows-amd64.exe ./cmd/petcore/

build-frontend: ## 仅构建前端
	cd frontend && pnpm build

# ─── 测试 ─────────────────────────────────────

test: test-go test-rust test-frontend ## 运行全部测试

test-go: ## 运行 Go 测试（含竞态检测）
	cd petcore && go test -race -shuffle=on -count=1 ./...

test-rust: ## 运行 Rust 测试
	cd src-tauri && cargo test --all-targets

test-frontend: ## 运行前端测试
	cd frontend && pnpm test -- --run

cover: ## Go 覆盖率报告
	cd petcore && go test -coverprofile=coverage.out -covermode=atomic ./... && go tool cover -html=coverage.out

cover-check: ## 检查 Go 覆盖率 >= 60%
	cd petcore && go test -coverprofile=coverage.out -covermode=atomic ./... && \
		COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//') && \
		echo "Coverage: $${COVERAGE}%" && \
		if (( $$(echo "$$COVERAGE < 60" | bc -l) )); then echo "❌ Below 60%"; exit 1; else echo "✅ OK"; fi

bench: ## 运行 Go 基准测试
	cd petcore && go test -bench=. -benchmem ./...

# ─── Lint ─────────────────────────────────────

lint: lint-go lint-rust lint-frontend ## 运行全部 linter（同 CI lint）

lint-go: ## Go linter（同 CI: golangci-lint run ./...）
	cd petcore && PATH="$$(go env GOPATH)/bin:$$PATH" golangci-lint run ./... --timeout 5m

lint-rust: ## Rust linter + sidecar 构建（同 CI: cargo clippy -- -D warnings）
	@SIDECAR="src-tauri/binaries/petcore-x86_64-unknown-linux-gnu"; \
	NEED_BUILD=0; \
	if [ ! -f "$$SIDECAR" ]; then \
		NEED_BUILD=1; \
	elif [ "petcore/cmd/petcore/main.go" -nt "$$SIDECAR" ] 2>/dev/null; then \
		NEED_BUILD=1; \
	elif [ "petcore/go.mod" -nt "$$SIDECAR" ] 2>/dev/null; then \
		NEED_BUILD=1; \
	fi; \
	if [ "$$NEED_BUILD" = "1" ]; then \
		echo "🔨 构建 sidecar 供 Rust clippy 使用..."; \
		mkdir -p src-tauri/binaries; \
		cd petcore && CGO_ENABLED=0 go build -o ../src-tauri/binaries/petcore-x86_64-unknown-linux-gnu ./cmd/petcore/; \
	else \
		echo "✓ sidecar 已是最新"; \
	fi
	cd src-tauri && cargo clippy -- -D warnings

lint-frontend: ## TypeScript linter（同 CI: pnpm lint）
	cd frontend && pnpm lint

# ─── 推送前检查（拦截 CI 失败）────────────

pre-push: lint build-petcore ## 推送前运行：lint + PetCore 编译（同 CI lint 阶段）
	@echo ""
	@echo "╔══════════════════════════════════════════╗"
	@echo "║  ✅  所有检查通过，可以推送！           ║"
	@echo "╚══════════════════════════════════════════╝"

setup-hooks: ## 安装 Git hooks（lefthook — 优先 pnpm，其次 go install）
	@if [ -f "frontend/node_modules/lefthook/bin/lefthook" ]; then \
		echo "🔗 使用 pnpm lefthook"; \
		cd frontend && pnpm lefthook install; \
	elif [ -f "$$(go env GOPATH)/bin/lefthook" ]; then \
		echo "🔗 使用 Go lefthook"; \
		"$$(go env GOPATH)/bin/lefthook" install; \
	else \
		echo "📦 安装 lefthook..."; \
		cd frontend && pnpm add -D lefthook && pnpm approve-builds lefthook; \
	fi
	@echo "✅ Git hooks 已安装（pre-commit + pre-push）"

# ─── 格式化 ───────────────────────────────────

fmt: fmt-go fmt-rust fmt-frontend ## 格式化所有代码

fmt-go: ## Go 格式化（gofumpt）
	cd petcore && gofumpt -l -w .

fmt-rust: ## Rust 格式化
	cd src-tauri && cargo fmt

fmt-frontend: ## TypeScript 格式化
	cd frontend && pnpm format

# ─── 清理 ─────────────────────────────────────

clean: ## 清理构建产物
	rm -rf petcore/petcore petcore/petcore-* petcore/dist/
	rm -rf src-tauri/target/
	rm -rf frontend/dist/ frontend/.vite/
	rm -rf *.dmg *.app

# ─── Devcontainer ────────────────────────────

devcontainer-setup: ## 手动执行 devcontainer 安装步骤（用于非容器环境）
	@echo "📦 安装前置工具..."
	@command -v rustup >/dev/null 2>&1 || curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
	@command -v go >/dev/null 2>&1 || { echo "请手动安装 Go: https://go.dev/dl/"; exit 1; }
	@command -v pnpm >/dev/null 2>&1 && npm i -g pnpm || npm i -g pnpm
	@cargo install tauri-cli --version "^2" 2>/dev/null || true
	@go install mvdan.cc/gofumpt@latest 2>/dev/null || true
	@go install github.com/air-verse/air@latest 2>/dev/null || true
	@echo "✅ 工具就绪"
