/**
 * 安全存储模块
 *
 * 使用 Tauri 的 tauri-plugin-store 插件将敏感数据（API Key）存储在
 * 操作系统的应用数据目录中，而非浏览器 localStorage。
 *
 * macOS: ~/Library/Application Support/com.desktop-pet.app/
 * Windows: %APPDATA%/com.desktop-pet.app/
 * Linux: ~/.local/share/com.desktop-pet.app/
 */

import { Store } from "@tauri-apps/plugin-store";

const STORE_FILE = "secrets.json";
const API_KEY_KEY = "apiKey";

let storePromise: Promise<Store> | null = null;

function getStore(): Promise<Store> {
  storePromise ??= Store.load(STORE_FILE);
  return storePromise;
}

/**
 * 从安全存储中获取 API Key。
 * 如果尚未存储或读取失败，返回空字符串。
 */
export async function getApiKey(): Promise<string> {
  try {
    const store = await getStore();
    const key = await store.get<string>(API_KEY_KEY);
    return key ?? "";
  } catch {
    return "";
  }
}

/**
 * 将 API Key 保存到安全存储。
 */
export async function setApiKey(key: string): Promise<void> {
  const store = await getStore();
  await store.set(API_KEY_KEY, key);
  await store.save();
}

/**
 * 从安全存储中删除 API Key。
 */
export async function removeApiKey(): Promise<void> {
  const store = await getStore();
  await store.delete(API_KEY_KEY);
  await store.save();
}
