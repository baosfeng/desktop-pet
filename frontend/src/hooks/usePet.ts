import { useCallback, useEffect } from "react";

import { type PetEvent, onPetEvent, sendMessage } from "@/lib/bridge";
import { generateId, usePetStore } from "@/stores/petStore";

/**
 * usePetState — 监听 Tauri pet:event 事件，自动更新宠物状态和消息。
 */
export function usePetEvent(): void {
  const setPetState = usePetStore((s) => s.setPetState);
  const addMessage = usePetStore((s) => s.addMessage);
  const appendToLastAssistant = usePetStore((s) => s.appendToLastAssistant);

  useEffect(() => {
    const unlistenPromise = onPetEvent((event: PetEvent) => {
      // 更新状态
      setPetState(event.kind as Parameters<typeof setPetState>[0]);

      // 处理消息事件
      if (event.kind === "agent.reply" && typeof event.data.text === "string") {
        if ((event.data as Record<string, unknown>).done === true) {
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
export function useChat() {
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
export function useSettings() {
  const settings = usePetStore((s) => s.settings);
  const updateSettings = usePetStore((s) => s.updateSettings);
  const saveSettings = usePetStore((s) => s.saveSettings);
  const loadSettings = usePetStore((s) => s.loadSettings);
  const showSettings = usePetStore((s) => s.showSettings);
  const toggleSettings = usePetStore((s) => s.toggleSettings);

  const handleSave = useCallback(() => {
    saveSettings();
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
