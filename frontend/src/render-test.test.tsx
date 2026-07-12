import { render, screen } from "@testing-library/react";
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

describe("render SettingsPanel", () => {
  it("renders", () => {
    render(<SettingsPanel onClose={vi.fn()} />);
    expect(screen.getByText("⚙️ 设置")).toBeDefined();
  });
});
