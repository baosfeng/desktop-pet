// ESLint v9 Flat Config
// 文档：https://eslint.org/docs/latest/use/configure/configuration-files
// 前置安装：pnpm add -D eslint @eslint/js typescript-eslint eslint-config-prettier eslint-plugin-react-hooks
//          eslint-plugin-react eslint-plugin-jsx-a11y eslint-plugin-import
import js from "@eslint/js";
import prettier from "eslint-config-prettier";
import importPlugin from "eslint-plugin-import";
import jsxA11y from "eslint-plugin-jsx-a11y";
import reactPlugin from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import tseslint from "typescript-eslint";

// eslint-disable-next-line @typescript-eslint/no-deprecated
export default tseslint.config(
  // 全局忽略
  { ignores: ["dist/", ".vite/", "node_modules/", "public/"] },
  // shadcn-ui 组件使用 CVA 类型链，variant/size 联合类型
  // 在 JSX data-* 属性和 render prop 中触发 no-unsafe-assignment
  {
    name: "shadcn-ui-overrides",
    files: ["src/components/ui/**/*.ts", "src/components/ui/**/*.tsx"],
    rules: {
      "@typescript-eslint/no-unsafe-assignment": "off",
    },
  },

  // 基础 JS 推荐规则
  js.configs.recommended,

  // TypeScript 严格规则 + 风格规则
  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,

  // TypeScript 解析器配置
  {
    languageOptions: {
      parserOptions: {
        projectService: {
          allowDefaultProject: ["eslint.config.js", "vitest.config.ts"],
        },
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
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

  // React 规则
  {
    plugins: {
      react: reactPlugin,
    },
    settings: {
      react: {
        version: "detect",
      },
    },
    rules: {
      ...reactPlugin.configs.recommended.rules,
      // React 18+ 不需要显式 import React
      "react/react-in-jsx-scope": "off",
      // 使用 TypeScript，不需要 prop-types
      "react/prop-types": "off",
      // 自闭合标签
      "react/self-closing-comp": "warn",
      // 简化 boolean 属性表达式
      "react/jsx-boolean-value": ["warn", "never"],
      // 禁止不必要的 Fragment
      "react/jsx-no-useless-fragment": "warn",
      // JSX 中花括号使用规范
      "react/jsx-curly-brace-presence": ["warn", { props: "never", children: "never" }],
      // 禁止在 JSX 中使用未转义字符
      "react/no-unescaped-entities": "error",
    },
  },

  // JSX 无障碍规则
  {
    plugins: {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
      "jsx-a11y": jsxA11y,
    },
    rules: {
      "jsx-a11y/alt-text": "warn",
      "jsx-a11y/aria-props": "warn",
      "jsx-a11y/aria-proptypes": "warn",
      "jsx-a11y/aria-unsupported-elements": "warn",
      "jsx-a11y/role-has-required-aria-props": "warn",
      "jsx-a11y/role-supports-aria-props": "warn",
    },
  },

  // Import 规则
  {
    plugins: {
      import: importPlugin,
    },
    rules: {
      "import/first": "error",
      "import/no-duplicates": "error",
      "import/newline-after-import": "warn",
      "import/no-default-export": "off",
    },
  },

  // 项目自定义规则
  {
    rules: {
      // 禁止 any（除非显式标注）
      "@typescript-eslint/no-explicit-any": "error",

      // 禁止未使用的变量（_ 前缀豁免）
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],

      // 必须显式定义返回值类型（仅导出函数，内部函数由 TS 推断）
      "@typescript-eslint/explicit-function-return-type": ["warn", { allowExpressions: true }],

      // 禁止 require（使用 import）
      "@typescript-eslint/no-require-imports": "error",

      // 禁止空函数
      "@typescript-eslint/no-empty-function": "error",

      // 禁止混淆的 void 表达式
      "@typescript-eslint/no-confusing-void-expression": "error",

      // 强制使用 ===
      eqeqeq: ["error", "always"],

      // 禁止 console.log（使用 logger 工具）
      "no-console": ["warn", { allow: ["warn", "error"] }],

      // 禁止 debugger
      "no-debugger": "error",

      // 禁止 alert
      "no-alert": "warn",

      // 禁止函数参数修改
      "no-param-reassign": "error",

      // 强制对象简写语法
      "object-shorthand": ["error", "always"],
    },
  },

  // 关闭与 Prettier 冲突的规则
  prettier,
);
