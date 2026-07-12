import { useCallback, useEffect, useRef, useState } from "react";
import type React from "react";

import { ChatBubble } from "@/components/ChatBubble";
import { OnboardingDialog } from "@/components/OnboardingDialog";
import { SettingsPanel } from "@/components/SettingsPanel";

import { useChat, usePetEvent, useSettings } from "@/hooks/usePet";

import { setWindowOpacity } from "@/lib/bridge";
import { initLive2D, setModelExpression, setModelMotion } from "@/lib/live2d";

import { generateId, usePetStore } from "@/stores/petStore";

import "./app.css";

export default function App(): React.JSX.Element {
  // 初始化事件监听
  usePetEvent();

  // 从 store 读取状态
  const petState = usePetStore((s) => s.petState);
  const [live2dReady, setLive2dReady] = useState(false);
  const [live2dError, setLive2dError] = useState<string | null>(null);
  const { messages, sendMessage } = useChat();
  const { showSettings, toggleSettings } = useSettings();
  const theme = usePetStore((s) => s.settings.theme);
  const opacity = usePetStore((s) => s.settings.opacity);
  const hasCompletedOnboarding = usePetStore((s) => s.settings.hasCompletedOnboarding);
  const updateSettings = usePetStore((s) => s.updateSettings);
  const saveSettings = usePetStore((s) => s.saveSettings);

  const addMessage = usePetStore((s) => s.addMessage);
  const setPetState = usePetStore((s) => s.setPetState);
  const loadApiKey = usePetStore((s) => s.loadApiKey);

  // 启动时从安全存储加载 API Key
  useEffect(() => {
    void loadApiKey();
  }, [loadApiKey]);

  // 互动超时计时器（用于恢复 idle 状态）
  const interactTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Live2D 初始化
  const appRef = useRef<Awaited<ReturnType<typeof initLive2D>>>(null);

  useEffect(() => {
    const canvas = document.getElementById("live2d-canvas") as HTMLCanvasElement | null;
    if (!canvas) return;

    // 模型路径：优先使用用户设置，没有则使用默认 CDN Haru
    const storedPath = usePetStore.getState().settings.modelPath;
    const modelPath =
      storedPath ||
      "https://cdn.jsdelivr.net/gh/Live2D/CubismWebSamples@develop/Samples/Resources/Haru/Haru.model3.json";
    // 模型许可：https://www.live2d.com/eula/live2d-free-material-license-agreement_en.html
    void initLive2D(canvas, {
      modelPath,
      scale: 0.5,
    })
      .then((app) => {
        appRef.current = app;
        usePetStore.getState().setLive2dApp(app);
        setLive2dReady(true);
        document.body.classList.add("live2d-loaded");
      })
      .catch((err: unknown) => {
        const msg = err instanceof Error ? err.message : String(err);
        setLive2dError(msg);
        // 同时输出到终端日志
        import("@tauri-apps/api/core")
          .then(({ invoke }) => {
            invoke("log_from_frontend", { level: "error", message: msg }).catch(() => undefined);
          })
          .catch(() => undefined);
      });

    return (): void => {
      if (interactTimerRef.current !== null) {
        clearTimeout(interactTimerRef.current);
        interactTimerRef.current = null;
      }
      if (appRef.current) {
        appRef.current.destroy(true);
        appRef.current = null;
        usePetStore.getState().setLive2dApp(null);
      }
    };
  }, []);

  // 暗黑模式切换：同步 theme 到 document.documentElement
  useEffect(() => {
    const root = document.documentElement;
    if (theme === "dark") {
      root.classList.add("dark");
      root.setAttribute("data-theme", "dark");
    } else {
      root.classList.remove("dark");
      root.setAttribute("data-theme", "pastoral");
    }
  }, [theme]);

  // 窗口透明度：从 store 恢复并同步到窗口
  useEffect(() => {
    void setWindowOpacity(opacity);
  }, [opacity]);

  // 主题切换回调
  const toggleTheme = useCallback(() => {
    const newTheme = theme === "pastoral" ? "dark" : "pastoral";
    updateSettings({ theme: newTheme });
    saveSettings();
  }, [theme, updateSettings, saveSettings]);

  // ---- 互动逻辑 ----
  const interact = useCallback(() => {
    const app = appRef.current;
    if (!app) return;

    // 随机动作：Haru 支持的 motion group
    const motionGroups = ["TapBody", "Pinch", "Shake", "FlickHead", "Flick"] as const;
    const group = motionGroups[Math.floor(Math.random() * motionGroups.length)] ?? "TapBody";
    const index = Math.floor(Math.random() * 3); // 0-2

    // 随机反馈文字
    const feedbacks = [
      "🥰 好开心~",
      "😊 嘻嘻~",
      "✨ 一起玩吧~",
      "💕 好喜欢你~",
      "🌟 耶~",
      "🐾 摸摸头~",
      "🎵 啦啦啦~",
    ];
    const feedback = feedbacks[Math.floor(Math.random() * feedbacks.length)] ?? "🥰 好开心~";

    // 切换到互动状态
    setPetState("interaction");

    // 触发 Live2D 动作 & 表情
    void setModelMotion(app, group, index);
    void setModelExpression(app, "f01");

    // 显示反馈气泡
    addMessage({
      id: generateId(),
      role: "system",
      content: feedback,
      timestamp: Date.now(),
    });

    // 延迟后恢复到 idle
    if (interactTimerRef.current !== null) {
      clearTimeout(interactTimerRef.current);
    }
    interactTimerRef.current = setTimeout(() => {
      setPetState("idle");
    }, 3000);
  }, [addMessage, setPetState]);

  return (
    <div className="relative w-full h-full flex flex-col bg-transparent overflow-hidden select-none">
      {/* 顶部花草装饰 */}
      <div className="text-center text-lg pt-2 text-primary/60 select-none pointer-events-none shrink-0">
        🌿 🌻 🌱
      </div>

      {/* 中间区域：宠物 + 聊天，自动分配空间 */}
      <div className="flex-1 flex flex-col min-h-0 w-full">
        {/* 宠物区域 */}
        <div className="flex flex-col items-center justify-center shrink-0 py-1">
          <div
            className="relative w-[35%] max-w-[160px] aspect-square flex items-center justify-center cursor-pointer"
            onClick={interact}
            role="button"
            aria-label="点击与宠物互动"
          >
            <canvas id="live2d-canvas" className="w-full h-full block pointer-events-none" />
            {!live2dReady && !live2dError && (
              <div className="absolute inset-0 flex items-center justify-center text-text-brown/50 text-xs">
                🐾 加载中...
              </div>
            )}
            {live2dError && (
              <div className="absolute inset-0 flex items-center justify-center text-text-brown/60 text-xs px-2 text-center">
                🐾 {live2dError}
              </div>
            )}
          </div>
          {/* 草地装饰 */}
          <div className="w-[35%] max-w-[180px] h-3 bg-gradient-to-r from-green-light/40 via-primary/30 to-green-light/40 rounded-full mt-0.5" />
          <div className="text-xs text-text-brown/50 mt-0.5 font-display">
            {petState === "idle"
              ? "晒太阳中~"
              : petState === "attention"
                ? "看着你呢~"
                : petState === "interaction"
                  ? "摸摸我吧~"
                  : "正在说话..."}
          </div>
        </div>

        {/* 聊天区域 */}
        <div className="flex-1 min-h-0 px-3 pb-1.5">
          <ChatBubble messages={messages} onSendMessage={sendMessage} />
        </div>
      </div>

      {/* 底部操作栏 */}
      <div className="flex gap-3 justify-center pb-2 shrink-0">
        <button
          className="btn btn-ghost btn-sm flex items-center gap-1.5 min-h-0 h-auto px-3 py-2 text-sm"
          onClick={toggleTheme}
          type="button"
          aria-label={theme === "pastoral" ? "切换暗黑模式" : "切换亮色模式"}
        >
          {theme === "pastoral" ? "🌙" : "☀️"}
        </button>
        <button
          className="btn btn-primary btn-sm flex items-center gap-1.5 min-h-0 h-auto px-4 py-2 text-sm"
          onClick={toggleSettings}
          type="button"
        >
          ⚙️ 设置
        </button>
        <button
          className="btn btn-accent btn-sm flex items-center gap-1.5 min-h-0 h-auto px-4 py-2 text-sm"
          onClick={interact}
          type="button"
          aria-label="互动"
        >
          🎮 互动
        </button>
      </div>

      {/* 新手引导 */}
      {!hasCompletedOnboarding && <OnboardingDialog />}

      {/* 设置面板 */}
      {showSettings && <SettingsPanel onClose={toggleSettings} />}
    </div>
  );
}
