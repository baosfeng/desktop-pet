import { useCallback, useEffect, useRef } from "react";

import { type PetEvent, onPetEvent, sendMessage, updateConfig } from "@/lib/bridge";
import { setModelExpression, setModelMotion } from "@/lib/live2d";
import { generateId, usePetStore } from "@/stores/petStore";
import type { Message, PetState, Settings } from "@/stores/petStore";

/**
 * 根据宠物状态触发对应的 Live2D 表情和动作。
 * app 不存在或模型不支持对应动画时静默跳过。
 */
function triggerLive2DForState(app: unknown, state: PetState): void {
  switch (state) {
    case "speaking":
      void setModelMotion(app as Parameters<typeof setModelMotion>[0], "FlickHead", 0);
      void setModelExpression(app as Parameters<typeof setModelExpression>[0], "f01");
      break;
    case "attention":
      void setModelMotion(app as Parameters<typeof setModelMotion>[0], "TapBody", 0);
      break;
    case "interaction":
      void setModelMotion(app as Parameters<typeof setModelMotion>[0], "Pinch", 0);
      break;
    case "idle":
      // idle 不在此处处理，由 usePetEvent 中的随机间隔管理
      break;
  }
}

/**
 * usePetState — 监听 Tauri pet:event 事件，自动更新宠物状态和消息，
 * 并联动 Live2D 表情/动作。
 */
export function usePetEvent(): void {
  const setPetState = usePetStore((s) => s.setPetState);
  const addMessage = usePetStore((s) => s.addMessage);
  const appendToLastAssistant = usePetStore((s) => s.appendToLastAssistant);

  // 待机随机小动作定时器
  const idleIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const clearIdleInterval = useCallback(() => {
    if (idleIntervalRef.current !== null) {
      clearInterval(idleIntervalRef.current);
      idleIntervalRef.current = null;
    }
  }, []);

  const startIdleInterval = useCallback(() => {
    clearIdleInterval();
    idleIntervalRef.current = setInterval(() => {
      const app = usePetStore.getState().live2dApp;
      if (app) {
        void setModelMotion(app as Parameters<typeof setModelMotion>[0], "Shake", 0);
      }
    }, 5000 + Math.random() * 10000); // 5-15 秒随机间隔
  }, [clearIdleInterval]);

  useEffect(() => {
    const unlistenPromise = onPetEvent((event: PetEvent) => {
      // 通过 store 获取最新的 live2dApp 引用
      const app = usePetStore.getState().live2dApp;

      // 将事件种类映射为宠物可视化状态
      switch (event.kind) {
        case "state.changed":
          // 从 data.state 提取实际 FSM 状态（idle / attention / interaction / speaking）
          if (typeof event.data.state === "string") {
            const newState = event.data.state as Parameters<typeof setPetState>[0];
            setPetState(newState);

            // 联动 Live2D 动画
            if (app) {
              triggerLive2DForState(app, newState);
              // 管理 idle 随机动作定时器
              if (newState === "idle") {
                startIdleInterval();
              } else {
                clearIdleInterval();
              }
            }
          }
          break;
        case "agent.thinking":
          // 思考中 → 关注状态
          {
            const thinking = event.data.status === true;
            setPetState(thinking ? "attention" : "idle");
            if (app && thinking) {
              clearIdleInterval();
              void setModelExpression(app as Parameters<typeof setModelExpression>[0], "f01");
            } else if (app && !thinking) {
              startIdleInterval();
            }
          }
          break;
        case "agent.reply":
        case "pet.speak":
          // 回复/说话 → 说话状态
          setPetState("speaking");
          if (app) {
            clearIdleInterval();
            triggerLive2DForState(app, "speaking");
          }
          break;
        case "error":
          // 出错回到待机，并显示错误消息
          setPetState("idle");
          if (app) startIdleInterval();
          if (typeof event.data.error === "string") {
            addMessage({
              id: generateId(),
              role: "system",
              content: "⚠️ " + event.data.error,
              timestamp: Date.now(),
            });
          }
          break;
        default:
          // 其他事件不修改状态
          break;
      }

      // 处理消息事件
      if (event.kind === "agent.reply" && typeof event.data.text === "string") {
        if (event.data.done === true) {
          return;
        }
        appendToLastAssistant(event.data.text);
      }
      if (event.kind === "pet.speak" && typeof event.data.text === "string") {
        const msg = {
          id: generateId(),
          role: "assistant" as const,
          content: event.data.text,
          timestamp: Date.now(),
        };
        addMessage(msg);
      }
    });

    return (): void => {
      clearIdleInterval();
      void unlistenPromise.then((unlisten) => {
        unlisten();
      });
    };
  }, [setPetState, addMessage, appendToLastAssistant, clearIdleInterval, startIdleInterval]);
}

/**
 * useChat — 发送消息和消息列表。
 */
export function useChat(): { messages: Message[]; sendMessage: (text: string) => void } {
  const messages = usePetStore((s) => s.messages);
  const addMessage = usePetStore((s) => s.addMessage);

  const handleSend = useCallback(
    (text: string) => {
      const userMsg = {
        id: generateId(),
        role: "user" as const,
        content: text,
        timestamp: Date.now(),
      };
      addMessage(userMsg);

      // 为 AI 回复预留一个空消息（打字机效果用）
      const assistantMsg = {
        id: generateId(),
        role: "assistant" as const,
        content: "",
        timestamp: Date.now(),
      };
      addMessage(assistantMsg);

      void sendMessage(text);
    },
    [addMessage],
  );

  return { messages, sendMessage: handleSend };
}

/**
 * useSettings — 设置面板状态和操作。
 */
export function useSettings(): {
  settings: Settings;
  updateSettings: (partial: Partial<Settings>) => void;
  showSettings: boolean;
  toggleSettings: () => void;
  handleSave: () => void;
  handleClose: () => void;
} {
  const settings = usePetStore((s) => s.settings);
  const updateSettings = usePetStore((s) => s.updateSettings);
  const saveSettings = usePetStore((s) => s.saveSettings);
  const loadSettings = usePetStore((s) => s.loadSettings);
  const showSettings = usePetStore((s) => s.showSettings);
  const toggleSettings = usePetStore((s) => s.toggleSettings);

  const handleSave = useCallback(() => {
    // 保存到 localStorage
    saveSettings();

    // 将 LLM 设置同步到 PetCore 后端
    const s = usePetStore.getState().settings;
    void updateConfig({
      apiKey: s.apiKey,
      provider: s.provider,
      baseUrl: s.baseUrl,
      modelName: s.modelName,
      systemPrompt: s.persona,
    });

    toggleSettings();
  }, [saveSettings, toggleSettings]);

  const handleClose = useCallback(() => {
    loadSettings();
    toggleSettings();
  }, [loadSettings, toggleSettings]);

  return {
    settings,
    updateSettings,
    showSettings,
    toggleSettings,
    handleSave,
    handleClose,
  };
}
