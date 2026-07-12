import { describe, expect, it, vi, beforeEach } from "vitest";

import { usePetStore, generateId } from "../stores/petStore";

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    clear: () => {
      store = {};
    },
  };
})();

vi.stubGlobal("localStorage", localStorageMock);

// 重置 store 状态
beforeEach(() => {
  const { clearMessages } = usePetStore.getState();
  clearMessages();
  usePetStore.setState({
    petState: "idle",
    messages: [],
    showSettings: false,
  });
  localStorage.clear();
});

describe("petStore", () => {
  describe("petState", () => {
    it("starts with idle state", () => {
      const { petState } = usePetStore.getState();
      expect(petState).toBe("idle");
    });

    it("updates pet state", () => {
      const { setPetState } = usePetStore.getState();
      setPetState("speaking");
      expect(usePetStore.getState().petState).toBe("speaking");
    });
  });

  describe("messages", () => {
    it("adds a message", () => {
      const msg = {
        id: generateId(),
        role: "user" as const,
        content: "hello",
        timestamp: Date.now(),
      };
      usePetStore.getState().addMessage(msg);
      expect(usePetStore.getState().messages).toHaveLength(1);
      expect(usePetStore.getState().messages[0]?.content).toBe("hello");
    });

    it("appends text to the last assistant message", () => {
      const msg = {
        id: generateId(),
        role: "assistant" as const,
        content: "Hi",
        timestamp: Date.now(),
      };
      usePetStore.getState().addMessage(msg);
      usePetStore.getState().appendToLastAssistant(" there");
      const msgs = usePetStore.getState().messages;
      expect(msgs).toHaveLength(1);
      expect(msgs[0]?.content).toBe("Hi there");
    });

    it("appends text only to assistant messages", () => {
      const userMsg = {
        id: generateId(),
        role: "user" as const,
        content: "hello",
        timestamp: Date.now(),
      };
      const asstMsg = {
        id: generateId(),
        role: "assistant" as const,
        content: "Hi",
        timestamp: Date.now(),
      };
      usePetStore.getState().addMessage(userMsg);
      usePetStore.getState().addMessage(asstMsg);
      usePetStore.getState().appendToLastAssistant(" there");
      const msgs = usePetStore.getState().messages;
      expect(msgs[msgs.length - 1]?.content).toBe("Hi there");
    });

    it("clears messages", () => {
      usePetStore.getState().addMessage({
        id: generateId(),
        role: "user",
        content: "test",
        timestamp: Date.now(),
      });
      usePetStore.getState().clearMessages();
      expect(usePetStore.getState().messages).toHaveLength(0);
    });
  });

  describe("settings", () => {
    it("has default settings", () => {
      const { settings } = usePetStore.getState();
      expect(settings.baseUrl).toBe("https://api.openai.com/v1");
      expect(settings.modelName).toBe("gpt-4o-mini");
    });

    it("updates settings partially", () => {
      usePetStore.getState().updateSettings({ modelName: "gpt-4" });
      expect(usePetStore.getState().settings.modelName).toBe("gpt-4");
    });

    it("persists settings to localStorage", () => {
      usePetStore.getState().updateSettings({ opacity: 0.5 });
      usePetStore.getState().saveSettings();
      const saved = JSON.parse(localStorage.getItem("desktop-pet-settings") ?? "{}") as Record<string, unknown>;
      expect(saved.opacity).toBe(0.5);
    });

    it("loads settings from localStorage", () => {
      localStorage.setItem(
        "desktop-pet-settings",
        JSON.stringify({ modelName: "gpt-4-turbo" }),
      );
      usePetStore.getState().loadSettings();
      expect(usePetStore.getState().settings.modelName).toBe("gpt-4-turbo");
    });
  });

  describe("showSettings", () => {
    it("toggles settings panel", () => {
      expect(usePetStore.getState().showSettings).toBe(false);
      usePetStore.getState().toggleSettings();
      expect(usePetStore.getState().showSettings).toBe(true);
      usePetStore.getState().toggleSettings();
      expect(usePetStore.getState().showSettings).toBe(false);
    });
  });
});

describe("generateId", () => {
  it("generates unique IDs", () => {
    const id1 = generateId();
    const id2 = generateId();
    expect(id1).not.toBe(id2);
  });

  it("includes msg- prefix", () => {
    expect(generateId()).toMatch(/^msg-/);
  });
});
