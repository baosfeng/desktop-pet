import { useCallback, useEffect } from "react";

import { type PetEvent, onPetEvent, sendMessage, updateConfig } from "@/lib/bridge";
import { generateId, usePetStore } from "@/stores/petStore";
import type { Message, Settings } from "@/stores/petStore";

/**
 * usePetState — 监听 Tauri pet:event 事件，自动更新宠物状态和消息。
 */
export function usePetEvent(): void {
  const setPetState = usePetStore((s) => s.setPetState);
  const addMessage = usePetStore((s) => s.addMessage);
  const appendToLastAssistant = usePetStore((s) => s.appendToLastAssistant);

  useEffect(() => {
    const unlistenPromise = onPetEvent((event: PetEvent) => {
      // 将事件种类映射为宠物可视化状态
      switch (event.kind) {
        case "state.changed":
          // 从 data.state 提取实际 FSM 状态（idle / attention / interaction / speaking）
          if (typeof event.data.state === "string") {
            setPetState(event.data.state as Parameters<typeof setPetState>[0]);
          }
          break;
        case "agent.thinking":
          // 思考中 → 关注状态
          setPetState(event.data.status === true ? "attention" : "idle");
          break;
        case "agent.reply":
        case "pet.speak":
          // 回复/说话 → 说话状态
          setPetState("speaking");
          break;
        case "error":
          // 出错回到待机，并显示错误消息
          setPetState("idle");
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
      void unlistenPromise.then((unlisten) => {
        unlisten();
      });
    };
  }, [setPetState, addMessage, appendToLastAssistant]);
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
