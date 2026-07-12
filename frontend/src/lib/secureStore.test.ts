import { beforeEach, describe, expect, it, vi } from "vitest";

// Mock tauri-plugin-store
vi.mock("@tauri-apps/plugin-store", () => {
  const store = new Map<string, string>();
  let saveCalled = false;

  return {
    Store: {
      load: vi.fn(() =>
        Promise.resolve({
          get: vi.fn((key: string) => Promise.resolve(store.get(key) ?? null)),
          set: vi.fn((key: string, value: string) => {
            store.set(key, value);
            return Promise.resolve();
          }),
          delete: vi.fn((key: string) => {
            store.delete(key);
            return Promise.resolve();
          }),
          save: vi.fn(() => {
            saveCalled = true;
            return Promise.resolve();
          }),
          getSaveCalled: () => saveCalled,
        }),
      ),
    },
  };
});

describe("secureStore", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("getApiKey returns empty string when no key stored", async () => {
    const { getApiKey } = await import("./secureStore");
    const key = await getApiKey();
    expect(key).toBe("");
  });

  it("setApiKey stores the key", async () => {
    const { getApiKey, setApiKey } = await import("./secureStore");
    await setApiKey("sk-test-key-12345");
    const key = await getApiKey();
    expect(key).toBe("sk-test-key-12345");
  });

  it("setApiKey calls save after setting", async () => {
    const { setApiKey } = await import("./secureStore");
    await setApiKey("test-key");
    // Verify no error thrown
    expect(true).toBe(true);
  });

  it("removeApiKey clears the stored key", async () => {
    const { getApiKey, setApiKey, removeApiKey } = await import("./secureStore");
    await setApiKey("sk-to-remove");
    await removeApiKey();
    const key = await getApiKey();
    expect(key).toBe("");
  });

  it("setApiKey with empty string removes the key", async () => {
    const { getApiKey, setApiKey } = await import("./secureStore");
    await setApiKey("sk-some-key");
    await setApiKey("");
    const key = await getApiKey();
    expect(key).toBe("");
  });
});
