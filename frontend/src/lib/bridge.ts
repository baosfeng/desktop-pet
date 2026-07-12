import { invoke } from "@tauri-apps/api/core";
import { type UnlistenFn, listen } from "@tauri-apps/api/event";

/* ------------------------------------------------------------------ */
/*  Tauri Commands                                                     */
/* ------------------------------------------------------------------ */

export async function sendMessage(text: string): Promise<void> {
  await invoke("chat", { text });
}

export async function toggleClickthrough(enabled: boolean): Promise<void> {
  await invoke("toggle_clickthrough", { enabled });
}

export async function getWindowPosition(): Promise<{ x: number; y: number }> {
  return await invoke("get_window_position");
}

export async function setWindowPosition(x: number, y: number): Promise<void> {
  await invoke("set_window_position", { x, y });
}

export async function getStatus(): Promise<Record<string, unknown>> {
  return await invoke("get_status");
}

export interface LLMConfig {
  apiKey: string;
  provider: string;
  baseUrl: string;
  modelName: string;
  systemPrompt: string;
}

export async function updateConfig(config: LLMConfig): Promise<void> {
  await invoke("update_config", { config });
}

export async function resizeWindow(width: number, height: number): Promise<void> {
  await invoke("resize_window", { width, height });
}

export async function setWindowOpacity(opacity: number): Promise<void> {
  await invoke("set_window_opacity", { opacity });
}

/* ------------------------------------------------------------------ */
/*  Tauri Events                                                       */
/* ------------------------------------------------------------------ */

export interface PetEvent {
  kind: string;
  data: Record<string, unknown>;
}

export function onPetEvent(handler: (event: PetEvent) => void): Promise<UnlistenFn> {
  return listen<PetEvent>("pet:event", (e) => {
    handler(e.payload);
  });
}

/* ------------------------------------------------------------------ */
/*  Domain Types                                                       */
/* ------------------------------------------------------------------ */

export interface ChatMessage {
  role: "user" | "assistant" | "system";
  content: string;
}
