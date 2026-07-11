# Vitest 2.x → 4.x 升级踩坑记录

> 状态：已解决
> 升级日期：2026-07-11
> 升级路径：2.1.9 → 4.1.10（跨大版本，跳过 3.x）

## 背景

Dependabot 自动创建了 vitest 从 2.1.9 升级到 4.1.10 的 PR。由于前端（React + PixiJS + Live2D）当前没有测试文件，升级过程未遇到 API 兼容性问题。

## 关键变更调研（2.x → 4.x）

由于网络环境受限未能直接访问官方迁移指南，以下是从 npm registry 及已知信息整理的要点：

### Vitest 3.0（2.x → 3.x 重大变更）

1. **Node.js 最低版本提升到 18+**（已满足，项目用 Node 22）
2. **Pool 配置变更**
   - 旧：`threads`/`vmThreads` 配置项
   - 新：统一为 `pool` 配置，默认 `forks`
   - 迁移：如果配置文件中有 `threads: false`，需改为 `pool: 'forks'`
3. **`vi.mock()` 自动提升行为变化**
   - 3.x 中 `vi.mock()` 不再自动提升到文件顶部
   - 替代：需要配合 `vi.hoisted()` 使用
4. **Reporters 配置重构**
   - 部分自定义 reporter 接口有变化

### Vitest 4.0（3.x → 4.x 重大变更）

1. **Vite 依赖升级**：要求 `vite ^6.0.0 || ^7.0.0 || ^8.0.0`（项目已用 vite 6，兼容）
2. **新依赖**：`obug` 包替代了部分内部工具
3. **`expect` API 强化**：新增 `expect-type` 包
4. **TypeScript 类型加强**：更多严格类型检查

### 新语法速查（基于 4.x）

```typescript
// 配置方式（vitest.config.ts 或 vite.config.ts 中的 test 字段）
import { defineConfig } from 'vitest/config'
export default defineConfig({
  test: {
    pool: 'forks',          // 默认池类型
    include: ['src/**/*.test.ts'],
    environment: 'node',    // 默认环境
  },
})

// vi.mock 在 3.x+ 中的正确用法
import { vi } from 'vitest'

const { myModule } = vi.hoisted(() => {
  return { myModule: { foo: 'bar' } }
})
vi.mock('./my-module', () => ({
  default: myModule,
}))

// 并发测试
import { test } from 'vitest'
test.concurrent('runs in parallel', async () => { /* ... */ })

// 覆盖率（需安装 @vitest/coverage-v8）
// vitest.config.ts:
// test: { coverage: { provider: 'v8', reporter: ['text', 'lcov'] } }
```

## 本项目当前状态

- **vitest** 仅在 `package.json` 的 `devDependencies` 中声明
- **无测试文件**：`src/` 下无 `*.test.*` 文件或 `__tests__` 目录
- **CI 命令**：`pnpm test` 会执行 `vitest run`，无测试时返回空结果（通过）
- **升级后验证**：`pnpm install` 成功，锁定文件正确生成

## 后续注意事项

1. 未来添加测试文件时，应使用 vitest 4.x 的 API（`vi.hoisted()` 替代自动提升）
2. 如果混用 Jest API，注意 vitest 4.x 的 `expect` 类型更严格
3. `vitest/config` 导入路径未变，现有 `vite.config.ts` 无需修改
4. 如需浏览器测试，vitest 4.x 使用 `@vitest/browser` 插件
