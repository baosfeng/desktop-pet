import { create } from "zustand";
import { subscribeWithSelector } from "zustand/middleware";
import { getApiKey, setApiKey, removeApiKey } from "@/lib/secureStore";

export type PetState = "idle" | "attention" | "interaction" | "speaking";

export interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: number;
}

export type Theme = "pastoral" | "dark";

export interface Settings {
  apiKey: string;
  provider: string;
  baseUrl: string;
  modelName: string;
  persona: string;
  opacity: number;
  theme: Theme;
  modelPath: string;
  hasCompletedOnboarding: boolean;
}

interface PetStore {
  // 宠物状态
  petState: PetState;
  setPetState: (state: PetState) => void;

  // 消息
  messages: Message[];
  addMessage: (msg: Message) => void;
  appendToLastAssistant: (text: string) => void;
  clearMessages: () => void;

  // 设置
  settings: Settings;
  updateSettings: (partial: Partial<Settings>) => void;
  loadSettings: () => void;
  saveSettings: () => void;
  loadApiKey: () => Promise<void>;
  saveApiKey: () => Promise<void>;
  apiKeyLoaded: boolean;

  // 设置面板可见性
  showSettings: boolean;
  toggleSettings: () => void;

  // Live2D Application 引用（用于 hook 触发动画）
  live2dApp: unknown;
  setLive2dApp: (app: unknown) => void;
}

const STORAGE_KEY = "desktop-pet-settings";
const MSG_STORAGE_KEY = "desktop-pet-messages";
const MAX_MESSAGES = 200;

function loadMessagesFromStorage(): Message[] {
  try {
    const raw = localStorage.getItem(MSG_STORAGE_KEY);
    if (raw !== null) {
      const parsed: unknown = JSON.parse(raw);
      if (Array.isArray(parsed)) {
        return (parsed as Message[]).slice(-MAX_MESSAGES);
      }
    }
  } catch {
    // ignore
  }
  return [];
}

const DEFAULT_SETTINGS: Settings = {
  apiKey: "",
  provider: "deepseek",
  baseUrl: "https://api.deepseek.com/v1",
  modelName: "deepseek-v4-flash",
  persona: "你是一只可爱的桌面宠物，性格活泼友善。",
  opacity: 0.9,
  theme: "pastoral",
  modelPath: "",
  hasCompletedOnboarding: false,
};

let msgCounter = 0;

export function generateId(): string {
  msgCounter += 1;
  return `msg-${String(msgCounter)}-${String(Date.now())}`;
}

/**
 * 从 localStorage 加载非敏感设置项（排除 apiKey）。
 */
function loadSettingsFromStorage(): Settings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw !== null) {
      const parsed = JSON.parse(raw) as Partial<Settings>;
      // 确保从 localStorage 读取时 apiKey 始终为空字符串
      // apiKey 通过安全存储（tauri-plugin-store）管理
      return { ...DEFAULT_SETTINGS, ...parsed, apiKey: "" };
    }
  } catch {
    // ignore
  }
  return { ...DEFAULT_SETTINGS };
}

export const usePetStore = create<PetStore>()(subscribeWithSelector((set, get) => ({
  // Pet state
  petState: "idle",
  setPetState: (state) => {
    set({ petState: state });
  },

  // Messages
  messages: loadMessagesFromStorage(),
  addMessage: (msg) => {
    set((state) => ({ messages: [...state.messages, msg].slice(-MAX_MESSAGES) }));
  },

  appendToLastAssistant: (text) => {
    set((state) => {
      const msgs = [...state.messages];
      for (let i = msgs.length - 1; i >= 0; i--) {
        const m = msgs[i];
        if (m?.role === "assistant") {
          msgs[i] = { ...m, content: m.content + text };
          break;
        }
      }
      return { messages: msgs };
    });
  },

  clearMessages: () => {
    set({ messages: [] });
  },

  // Settings
  settings: loadSettingsFromStorage(),
  apiKeyLoaded: false,
  updateSettings: (partial) => {
    set((state) => ({ settings: { ...state.settings, ...partial } }));
  },
  loadSettings: () => {
    set({ settings: loadSettingsFromStorage() });
  },
  saveSettings: () => {
    const { settings } = get();
    // 保存到 localStorage 时排除 apiKey（敏感字段走安全存储）
    const { apiKey: _, ...safeSettings } = settings;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(safeSettings));
  },

  /**
   * 从安全存储（tauri-plugin-store）加载 API Key。
   * 应用启动时调用。
   */
  loadApiKey: async () => {
    const key = await getApiKey();
    set((state) => ({
      settings: { ...state.settings, apiKey: key },
      apiKeyLoaded: true,
    }));
  },

  /**
   * 将 API Key 保存到安全存储（tauri-plugin-store）。
   */
  saveApiKey: async () => {
    const { settings } = get();
    if (settings.apiKey) {
      await setApiKey(settings.apiKey);
    } else {
      await removeApiKey();
    }
  },

  // Settings panel
  showSettings: false,
  toggleSettings: () => {
    set((state) => ({ showSettings: !state.showSettings }));
  },

  // Live2D
  live2dApp: null,
  setLive2dApp: (app) => {
    set({ live2dApp: app });
  },
})));

// --- 消息持久化：debounce 500ms 写入 localStorage ---
let saveTimer: ReturnType<typeof setTimeout> | null = null;

usePetStore.subscribe(
  (state: PetStore) => state.messages,
  (messages: Message[]) => {
    if (saveTimer !== null) clearTimeout(saveTimer);
    saveTimer = setTimeout(() => {
      try {
        const trimmed = messages.length > MAX_MESSAGES ? messages.slice(-MAX_MESSAGES) : messages;
        localStorage.setItem(MSG_STORAGE_KEY, JSON.stringify(trimmed));
      } catch {
        // ignore
      }
    }, 500);
  },
);
