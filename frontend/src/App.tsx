import { useEffect, useRef } from "react";
import type React from "react";

import { ChatBubble } from "@/components/ChatBubble";
import { SettingsPanel } from "@/components/SettingsPanel";

import { useChat, usePetEvent, useSettings } from "@/hooks/usePet";
import { usePetStore } from "@/stores/petStore";

import { initLive2D } from "@/lib/live2d";

import "./App.css";

export default function App(): React.JSX.Element {
  // 初始化事件监听
  usePetEvent();

  // 从 store 读取状态
  const petState = usePetStore((s) => s.petState);
  const settings = usePetStore((s) => s.settings);
  const { messages, sendMessage } = useChat();
  const { showSettings, toggleSettings } = useSettings();

  // Live2D 初始化
  const appRef = useRef<Awaited<ReturnType<typeof initLive2D>>>(null);

  useEffect(() => {
    const canvas = document.getElementById("live2d-canvas") as HTMLCanvasElement | null;
    if (!canvas) return;

    const modelPath = ""; // 需要在 public/models/ 下放置模型文件
    initLive2D(canvas, {
      modelPath,
      scale: 0.5,
    }).then((app) => {
      appRef.current = app;
    });

    return (): void => {
      if (appRef.current) {
        appRef.current.destroy(true);
        appRef.current = null;
      }
    };
  }, []);

  return (
    <div className="app" data-state={petState}>
      <div className="pet-area">
        <canvas id="live2d-canvas" className="live2d-canvas" />
      </div>

      <ChatBubble messages={messages} onSendMessage={sendMessage} />

      <button
        className="settings-toggle"
        onClick={toggleSettings}
        type="button"
        aria-label="设置"
      >
        ⚙
      </button>

      {showSettings && <SettingsPanel onClose={toggleSettings} />}
    </div>
  );
}
