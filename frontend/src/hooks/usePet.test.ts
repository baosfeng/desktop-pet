import { describe, expect, it, vi, beforeEach } from "vitest";

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

// Mock Tauri API
vi.mock("@tauri-apps/api/core", () => ({
  invoke: vi.fn(),
}));

vi.mock("@tauri-apps/api/event", () => ({
  listen: vi.fn(() => Promise.resolve(vi.fn())),
}));

/* eslint-disable import/first */
import { generateId, usePetStore } from "../stores/petStore";
import { sendMessage } from "../lib/bridge";
/* eslint-enable import/first */

beforeEach(() => {
  usePetStore.setState({
    messages: [],
    settings: {
      apiKey: "",
      provider: "deepseek",
      baseUrl: "https://api.deepseek.com/v1",
      modelName: "deepseek-chat",
      persona: "test",
      opacity: 0.9,
    },
    showSettings: false,
  });
  vi.clearAllMocks();
});

describe("useChat (store-level)", () => {
  it("has empty messages initially", () => {
    expect(usePetStore.getState().messages).toEqual([]);
  });

  it("adds user message and assistant placeholder on send", () => {
    const msg = {
      id: generateId(),
      role: "user" as const,
      content: "hello",
      timestamp: Date.now(),
    };
    usePetStore.getState().addMessage(msg);

    // Add assistant placeholder (mimics useChat.sendMessage)
    const asstMsg = {
      id: generateId(),
      role: "assistant" as const,
      content: "",
      timestamp: Date.now(),
    };
    usePetStore.getState().addMessage(asstMsg);

    const msgs = usePetStore.getState().messages;
    expect(msgs).toHaveLength(2);
    expect(msgs[0]?.role).toBe("user");
    expect(msgs[0]?.content).toBe("hello");
    expect(msgs[1]?.role).toBe("assistant");
    expect(msgs[1]?.content).toBe("");
  });

  it("sends message via bridge", async () => {
    await sendMessage("hello");
    const { invoke } = await import("@tauri-apps/api/core");
    expect(invoke).toHaveBeenCalledWith("chat", { text: "hello" });
  });
});

describe("useSettings (store-level)", () => {
  it("returns default settings", () => {
    const { settings } = usePetStore.getState();
    expect(settings.baseUrl).toBe("https://api.openai.com/v1");
  });

  it("toggles settings via store", () => {
    expect(usePetStore.getState().showSettings).toBe(false);
    usePetStore.getState().toggleSettings();
    expect(usePetStore.getState().showSettings).toBe(true);
  });

  it("saves and reloads settings", () => {
    usePetStore.getState().updateSettings({ modelName: "custom-model" });
    usePetStore.getState().saveSettings();
    usePetStore.getState().loadSettings();
    expect(usePetStore.getState().settings.modelName).toBe("custom-model");
  });
});
