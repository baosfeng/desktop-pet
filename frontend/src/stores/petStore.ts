import { create } from "zustand";

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
};

let msgCounter = 0;

export function generateId(): string {
  msgCounter += 1;
  return `msg-${String(msgCounter)}-${String(Date.now())}`;
}

function loadSettingsFromStorage(): Settings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw !== null) {
      const parsed = JSON.parse(raw) as Partial<Settings>;
      return { ...DEFAULT_SETTINGS, ...parsed };
    }
  } catch {
    // ignore
  }
  return { ...DEFAULT_SETTINGS };
}

export const usePetStore = create<PetStore>((set, get) => ({
  // Pet state
  petState: "idle",
  setPetState: (state) => { set({ petState: state }); },

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

  clearMessages: () => { set({ messages: [] }); },

  // Settings
  settings: loadSettingsFromStorage(),
  updateSettings: (partial) => {
    set((state) => ({ settings: { ...state.settings, ...partial } }));
  },
  loadSettings: () => { set({ settings: loadSettingsFromStorage() }); },
  saveSettings: () => {
    const { settings } = get();
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  },

  // Settings panel
  showSettings: false,
  toggleSettings: () => { set((state) => ({ showSettings: !state.showSettings })); },

  // Live2D
  live2dApp: null,
  setLive2dApp: (app) => { set({ live2dApp: app }); },
}));

// --- 消息持久化：debounce 500ms 写入 localStorage ---
let saveTimer: ReturnType<typeof setTimeout> | null = null;

usePetStore.subscribe(
  (state: PetStore) => state.messages,
  (messages: Message[]) => {
    if (saveTimer !== null) clearTimeout(saveTimer);
    saveTimer = setTimeout(() => {
      try {
        const trimmed = messages.length > MAX_MESSAGES
          ? messages.slice(-MAX_MESSAGES)
          : messages;
        localStorage.setItem(MSG_STORAGE_KEY, JSON.stringify(trimmed));
      } catch {
        // ignore
      }
    }, 500);
  },
);
