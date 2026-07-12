import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

/* eslint-disable import/first */
import { SettingsPanel } from "./components/SettingsPanel";

const { mockStore } = vi.hoisted(() => {
  const mockSettings = {
    apiKey: "",
    baseUrl: "https://api.openai.com/v1",
    modelName: "gpt-4o-mini",
    persona: "一只可爱的猫娘",
    opacity: 0.9,
  };

  return {
    mockStore: {
      usePetStore: (selector?: (state: Record<string, unknown>) => unknown) => {
        const state = {
          settings: { ...mockSettings },
          updateSettings: vi.fn(),
          saveSettings: vi.fn(),
        };
        return selector ? selector(state) : state;
      },
      generateId: (): string => "test-id",
    }
  };
});

vi.mock("@/stores/petStore", () => mockStore);

describe("SettingsPanel", () => {
  it("renders settings form", () => {
    render(<SettingsPanel onClose={vi.fn()} />);
    expect(screen.getByText("⚙️ 设置")).toBeDefined();
  });
});
