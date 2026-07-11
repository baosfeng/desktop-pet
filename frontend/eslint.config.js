// ESLint v9 Flat Config
// 文档：https://eslint.org/docs/latest/use/configure/configuration-files
// 前置安装：pnpm add -D eslint @eslint/js typescript-eslint eslint-config-prettier eslint-plugin-react-hooks

import js from "@eslint/js";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
import prettier from "eslint-config-prettier";

export default tseslint.config(
  // 全局忽略
  { ignores: ["dist/", ".vite/", "node_modules/"] },

  // 基础 JS 推荐规则
  js.configs.recommended,

  // TypeScript 严格规则
  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },

  // React Hooks 规则
  {
    plugins: {
      "react-hooks": reactHooks,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
    },
  },

  // 项目自定义规则
  {
    rules: {
      // 禁止 any（除非显式标注）
      "@typescript-eslint/no-explicit-any": "error",

      // 禁止未使用的变量
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],

      // 必须显式定义返回值类型
      "@typescript-eslint/explicit-function-return-type": "warn",

      // 禁止 require（使用 import）
      "@typescript-eslint/no-require-imports": "error",

      // 禁止空函数
      "@typescript-eslint/no-empty-function": "error",

      // 强制使用 ===
      eqeqeq: ["error", "always"],

      // 禁止 console.log（使用 logger 工具）
      "no-console": ["warn", { allow: ["warn", "error"] }],

      // 禁止 debugger
      "no-debugger": "error",
    },
  },

  // 关闭与 Prettier 冲突的规则
  prettier,
);
