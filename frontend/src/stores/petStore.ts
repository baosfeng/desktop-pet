import { create } from "zustand";

export type PetState = "idle" | "attention" | "interaction" | "speaking";

export interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: number;
}

export interface Settings {
  apiKey: string;
  baseUrl: string;
  modelName: string;
  persona: string;
  opacity: number;
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
}

const STORAGE_KEY = "desktop-pet-settings";

const DEFAULT_SETTINGS: Settings = {
  apiKey: "",
  baseUrl: "https://api.openai.com/v1",
  modelName: "gpt-4o-mini",
  persona: "你是一只可爱的桌面宠物，性格活泼友善。",
  opacity: 0.9,
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
  messages: [],
  addMessage: (msg) => {
    set((state) => ({ messages: [...state.messages, msg] }));
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
}));
