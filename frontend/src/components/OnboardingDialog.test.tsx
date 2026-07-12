import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { usePetStore } from "@/stores/petStore";

// Mock the dialog component from shadcn
vi.mock("@/components/ui/dialog", () => ({
  Dialog: ({ children, open }: { children: React.ReactNode; open: boolean }) =>
    open ? <div data-testid="dialog">{children}</div> : null,
  DialogContent: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="dialog-content">{children}</div>
  ),
  DialogHeader: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="dialog-header">{children}</div>
  ),
  DialogTitle: ({ children }: { children: React.ReactNode }) => (
    <h2 data-testid="dialog-title">{children}</h2>
  ),
}));

describe("OnboardingDialog", () => {
  beforeEach(() => {
    // Reset store to unboarded state
    usePetStore.setState({
      settings: {
        ...usePetStore.getState().settings,
        hasCompletedOnboarding: false,
      },
    });
  });

  it("renders the first step when not completed onboarding", async () => {
    const { OnboardingDialog } = await import("./OnboardingDialog");
    render(<OnboardingDialog />);

    expect(screen.getByText(/欢迎来到桌面 AI 宠物/)).toBeTruthy();
    expect(screen.getByText("下一步 →")).toBeTruthy();
  });

  it("shows '开始使用' button on last step", async () => {
    const { OnboardingDialog } = await import("./OnboardingDialog");
    render(<OnboardingDialog />);

    const user = userEvent.setup();

    // Advance through 4 steps (click next 3 times)
    for (let i = 0; i < 3; i++) {
      const nextBtn = screen.getByText("下一步 →");
      await user.click(nextBtn);
    }

    // Should now see the last step with "开始使用" button
    expect(screen.getByText("开始使用 🚀")).toBeTruthy();
  });

  it("completes onboarding when clicking '开始使用'", async () => {
    const { OnboardingDialog } = await import("./OnboardingDialog");
    render(<OnboardingDialog />);

    const user = userEvent.setup();

    // Advance to last step
    for (let i = 0; i < 3; i++) {
      const nextBtn = screen.getByText("下一步 →");
      await user.click(nextBtn);
    }

    // Click "开始使用"
    const startBtn = screen.getByText("开始使用 🚀");
    await user.click(startBtn);

    // Verify onboarding is completed
    const state = usePetStore.getState();
    expect(state.settings.hasCompletedOnboarding).toBe(true);
  });

  it("has step indicator dots", async () => {
    const { OnboardingDialog } = await import("./OnboardingDialog");
    const { container } = render(<OnboardingDialog />);

    // Should have 4 step dots
    const dots = container.querySelectorAll(".w-2.h-2.rounded-full");
    expect(dots.length).toBe(4);
  });
});
