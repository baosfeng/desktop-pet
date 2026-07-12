import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Button } from "./Button";

describe("Button", () => {
  it("renders children text", () => {
    render(<Button>发送</Button>);
    expect(screen.getByText("发送")).toBeDefined();
  });

  it("applies variant class", () => {
    render(<Button variant="accent">保存</Button>);
    const btn = screen.getByText("保存");
    expect(btn.className).toContain("bg-accent");
  });

  it("fires onClick handler", async () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>点击</Button>);
    await userEvent.click(screen.getByText("点击"));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("defaults to primary variant", () => {
    render(<Button>默认</Button>);
    const btn = screen.getByText("默认");
    expect(btn.className).toContain("bg-primary");
  });
});
