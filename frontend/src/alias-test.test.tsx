import { describe, expect, it, vi } from "vitest";
import { usePetStore } from "@/stores/petStore";

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

describe("alias mock", () => {
  it("works", () => {
    const settings = usePetStore((s) => s);
    expect(settings.settings.baseUrl).toBe("https://api.openai.com/v1");
  });
});
