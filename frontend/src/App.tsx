import { useEffect, useRef, useState } from "react";
import type React from "react";

import { ChatBubble } from "@/components/ChatBubble";
import { SettingsPanel } from "@/components/SettingsPanel";

import { useChat, usePetEvent, useSettings } from "@/hooks/usePet";
import { usePetStore } from "@/stores/petStore";

import { initLive2D } from "@/lib/live2d";

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

  // Live2D 初始化
  const appRef = useRef<Awaited<ReturnType<typeof initLive2D>>>(null);

  useEffect(() => {
    const canvas = document.getElementById("live2d-canvas") as HTMLCanvasElement | null;
    if (!canvas) return;

    // 使用 Live2D 官方示例模型 Haru（CDN 加载）
    // 模型许可：https://www.live2d.com/eula/live2d-free-material-license-agreement_en.html
    const modelPath = "https://cdn.jsdelivr.net/gh/Live2D/CubismWebSamples@develop/Samples/Resources/Haru/Haru.model3.json";
    void initLive2D(canvas, {
      modelPath,
      scale: 0.5,
    })
      .then((app) => {
        appRef.current = app;
        setLive2dReady(true);
        document.body.classList.add("live2d-loaded");
      })
      .catch((err: unknown) => {
        const msg = err instanceof Error ? err.message : String(err);
        setLive2dError(msg);
        // 同时输出到终端日志
        import("@tauri-apps/api/core").then(({ invoke }) => {
          invoke("log_from_frontend", { level: "error", message: msg }).catch(() => {});
        }).catch(() => {});
      });

    return (): void => {
      if (appRef.current) {
        appRef.current.destroy(true);
        appRef.current = null;
      }
    };
  }, []);

  return (
    <div
      className="relative w-full h-full flex flex-col items-center bg-transparent overflow-hidden select-none"
      data-state={petState}
    >
      {/* 顶部花草装饰 */}
      <div className="text-center text-lg pt-3 text-primary/60 select-none pointer-events-none">
        🌿 🌻 🌱
      </div>

      {/* 宠物区域 */}
      <div className="flex-1 flex flex-col items-center justify-center w-full">
        <div className="relative w-[120px] h-[120px] flex items-center justify-center">
          <canvas id="live2d-canvas" className="w-full h-full block" />
          {!live2dReady && !live2dError && (
            <div className="absolute inset-0 flex items-center justify-center text-text-brown/50 text-[12px]">
              🐾 加载中...
            </div>
          )}
          {live2dError && (
            <div className="absolute inset-0 flex items-center justify-center text-text-brown/60 text-[11px] px-2 text-center">
              🐾 {live2dError}
            </div>
          )}
        </div>
        {/* 草地装饰 */}
        <div className="w-[140px] h-[16px] bg-gradient-to-r from-green-light/40 via-primary/30 to-green-light/40 rounded-full mt-1" />
        <div className="text-[11px] text-text-brown/50 mt-1 font-display">
          {petState === "idle" ? "晒太阳中~" : petState === "attention" ? "看着你呢~" : petState === "interaction" ? "摸摸我吧~" : petState === "speaking" ? "正在说话..." : "在想什么呢？"}
        </div>
      </div>

      {/* 聊天气泡 */}
      <div className="w-full px-3 pb-2">
        <ChatBubble messages={messages} onSendMessage={sendMessage} />
      </div>

      {/* 底部操作栏 */}
      <div className="flex gap-2 justify-center pb-3">
        <button
          className="flex items-center gap-1 px-3 py-1.5 rounded-[8px] bg-primary text-primary-content text-[12px] font-medium cursor-pointer border-none transition-all duration-150 hover:scale-105 active:scale-95"
          onClick={toggleSettings}
          type="button"
        >
          ⚙️ 设置
        </button>
        <button
          className="flex items-center gap-1 px-3 py-1.5 rounded-[8px] bg-accent text-accent-content text-[12px] font-medium cursor-pointer border-none transition-all duration-150 hover:scale-105 active:scale-95"
          onClick={(): void => {
            // TODO: 互动功能
          }}
          type="button"
          aria-label="互动"
        >
          🎮 互动
        </button>
      </div>

      {/* 设置面板 */}
      {showSettings && <SettingsPanel onClose={toggleSettings} />}
    </div>
  );
}
