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
    <div className="relative w-full h-full flex flex-col bg-transparent overflow-hidden select-none">
      {/* 顶部花草装饰 */}
      <div className="text-center text-lg pt-2 text-primary/60 select-none pointer-events-none shrink-0">
        🌿 🌻 🌱
      </div>

      {/* 中间区域：宠物 + 聊天，自动分配空间 */}
      <div className="flex-1 flex flex-col min-h-0 w-full">
        {/* 宠物区域 */}
        <div className="flex flex-col items-center justify-center shrink-0 py-1">
          <div className="relative w-[35%] max-w-[160px] aspect-square flex items-center justify-center">
            <canvas id="live2d-canvas" className="w-full h-full block" />
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
            {petState === "idle" ? "晒太阳中~" : petState === "attention" ? "看着你呢~" : petState === "interaction" ? "摸摸我吧~" : petState === "speaking" ? "正在说话..." : "在想什么呢？"}
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
          className="btn btn-primary btn-sm flex items-center gap-1.5 min-h-0 h-auto px-4 py-2 text-sm"
          onClick={toggleSettings}
          type="button"
        >
          ⚙️ 设置
        </button>
        <button
          className="btn btn-accent btn-sm flex items-center gap-1.5 min-h-0 h-auto px-4 py-2 text-sm"
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
