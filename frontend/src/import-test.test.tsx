import { describe, expect, it, vi } from "vitest";
import { SettingsPanel } from "./components/SettingsPanel";

vi.mock("@/stores/petStore", () => ({
  usePetStore: ((selector: (state: Record<string, unknown>) => unknown) => {
    const state: Record<string, unknown> = {
      settings: {
        apiKey: "",
        baseUrl: "https://api.openai.com/v1",
        modelName: "gpt-4o-mini",
        persona: "一只可爱的猫娘",
        opacity: 0.9,
      },
      updateSettings: vi.fn(),
      saveSettings: vi.fn(),
    };
    return selector(state);
  }) as never,
  generateId: (): string => "test-id",
}));

describe("import SettingsPanel", () => {
  it("can import", () => {
    expect(SettingsPanel).toBeDefined();
  });
});
