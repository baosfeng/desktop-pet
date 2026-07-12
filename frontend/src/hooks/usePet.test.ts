import { describe, expect, it, vi, beforeEach } from "vitest";

// ---- Mocks (hoisted by Vitest) ----

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

// Mock Live2D animation functions
const mockSetModelMotion = vi.fn();
const mockSetModelExpression = vi.fn();

vi.mock("@/lib/live2d", () => ({
  initLive2D: vi.fn(),
  setModelExpression: (...args: unknown[]): unknown => mockSetModelExpression(...args),
  setModelMotion: (...args: unknown[]): unknown => mockSetModelMotion(...args),
}));

// Mock Tauri API
vi.mock("@tauri-apps/api/core", () => ({
  invoke: vi.fn(),
}));

// Capture the listen handler so we can simulate events
let capturedHandler: ((e: { payload: import("../lib/bridge").PetEvent }) => void) | null = null;

vi.mock("@tauri-apps/api/event", () => ({
  listen: vi.fn((_event: string, handler: (e: { payload: import("../lib/bridge").PetEvent }) => void) => {
    capturedHandler = handler;
    return Promise.resolve(vi.fn());
  }),
}));

/* eslint-disable import/first */
import { act, renderHook } from "@testing-library/react";
import { usePetEvent } from "../hooks/usePet";
import { sendMessage } from "../lib/bridge";
import { generateId, usePetStore } from "../stores/petStore";
/* eslint-enable import/first */

// A dummy app reference (just an object that passes truthiness check)
const dummyApp = { __live2dModel: { internalModel: {} } };

beforeEach(() => {
  capturedHandler = null;
  usePetStore.setState({
    petState: "idle",
    messages: [],
    settings: {
      apiKey: "",
      provider: "deepseek",
      baseUrl: "https://api.deepseek.com/v1",
      modelName: "deepseek-v4-flash",
      persona: "test",
      opacity: 0.9,
    },
    showSettings: false,
    live2dApp: null,
  });
  vi.clearAllMocks();
});

// Helper to fire a pet event through the captured listen handler
function firePetEvent(kind: string, data: Record<string, unknown> = {}): void {
  if (!capturedHandler) throw new Error("usePetEvent not mounted yet");
  capturedHandler({ payload: { kind, data } });
}

describe("live2dApp store field", () => {
  it("starts as null", () => {
    expect(usePetStore.getState().live2dApp).toBeNull();
  });

  it("can be set and cleared", () => {
    usePetStore.getState().setLive2dApp(dummyApp);
    expect(usePetStore.getState().live2dApp).toBe(dummyApp);

    usePetStore.getState().setLive2dApp(null);
    expect(usePetStore.getState().live2dApp).toBeNull();
  });
});

describe("usePetEvent — Live2D animation triggers", () => {
  it("triggers FlickHead + f01 on speaking state", () => {
    usePetStore.getState().setLive2dApp(dummyApp);
    renderHook(() => { usePetEvent(); });

    act(() => {
      firePetEvent("state.changed", { state: "speaking" });
    });

    expect(mockSetModelMotion).toHaveBeenCalledWith(dummyApp, "FlickHead", 0);
    expect(mockSetModelExpression).toHaveBeenCalledWith(dummyApp, "f01");
  });

  it("triggers TapBody on attention state", () => {
    usePetStore.getState().setLive2dApp(dummyApp);
    renderHook(() => { usePetEvent(); });

    act(() => {
      firePetEvent("state.changed", { state: "attention" });
    });

    expect(mockSetModelMotion).toHaveBeenCalledWith(dummyApp, "TapBody", 0);
  });

  it("triggers Pinch on interaction state", () => {
    usePetStore.getState().setLive2dApp(dummyApp);
    renderHook(() => { usePetEvent(); });

    act(() => {
      firePetEvent("state.changed", { state: "interaction" });
    });

    expect(mockSetModelMotion).toHaveBeenCalledWith(dummyApp, "Pinch", 0);
  });

  it("does NOT trigger animations when live2dApp is null", () => {
    renderHook(() => { usePetEvent(); });

    act(() => {
      firePetEvent("state.changed", { state: "speaking" });
    });

    expect(mockSetModelMotion).not.toHaveBeenCalled();
    expect(mockSetModelExpression).not.toHaveBeenCalled();
  });

  it("triggers thinking expression on agent.thinking (status=true)", () => {
    usePetStore.getState().setLive2dApp(dummyApp);
    renderHook(() => { usePetEvent(); });

    act(() => {
      firePetEvent("agent.thinking", { status: true });
    });

    expect(usePetStore.getState().petState).toBe("attention");
    expect(mockSetModelExpression).toHaveBeenCalledWith(dummyApp, "f01");
  });

  it("returns to idle on agent.thinking (status=false)", () => {
    usePetStore.getState().setLive2dApp(dummyApp);
    usePetStore.setState({ petState: "attention" });
    renderHook(() => { usePetEvent(); });

    act(() => {
      firePetEvent("agent.thinking", { status: false });
    });

    expect(usePetStore.getState().petState).toBe("idle");
  });

  it("triggers speaking animation on agent.reply event", () => {
    usePetStore.getState().setLive2dApp(dummyApp);
    renderHook(() => { usePetEvent(); });

    act(() => {
      firePetEvent("agent.reply", { text: "hello", done: false });
    });

    expect(usePetStore.getState().petState).toBe("speaking");
    expect(mockSetModelMotion).toHaveBeenCalledWith(dummyApp, "FlickHead", 0);
    expect(mockSetModelExpression).toHaveBeenCalledWith(dummyApp, "f01");
  });

  it("triggers speaking animation on pet.speak event", () => {
    usePetStore.getState().setLive2dApp(dummyApp);
    renderHook(() => { usePetEvent(); });

    act(() => {
      firePetEvent("pet.speak", { text: "woof!" });
    });

    expect(usePetStore.getState().petState).toBe("speaking");
    expect(mockSetModelMotion).toHaveBeenCalledWith(dummyApp, "FlickHead", 0);
    expect(mockSetModelExpression).toHaveBeenCalledWith(dummyApp, "f01");
  });
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
    expect(settings.baseUrl).toBe("https://api.deepseek.com/v1");
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
