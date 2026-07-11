import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],

  // 防止 Vite 将源码中的敏感信息泄露到产物中
  esbuild: {
    pure: ["console.log"],
  },

  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },

  // 开发服务器配置（Tauri 开发模式需要）
  server: {
    port: 1420,
    strictPort: true,
  },

  // 构建配置
  build: {
    target: "ES2022",
    outDir: "dist",
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          pixi: ["pixi.js"],
          react: ["react", "react-dom"],
        },
      },
    },
    // Tauri v2 需要禁用 assets 内联，否则路径会出问题
    assetsInlineLimit: 0,
  },

  // 环境变量前缀
  envPrefix: ["VITE_", "TAURI_"],
});
