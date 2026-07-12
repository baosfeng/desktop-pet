import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ChatBubble } from "./ChatBubble";

const mockMessages = [
  { id: "1", role: "assistant" as const, content: "你好！", timestamp: Date.now() },
  { id: "2", role: "user" as const, content: "今天天气如何？", timestamp: Date.now() },
];

describe("ChatBubble", () => {
  it("renders messages with typewriter effect", async () => {
    render(<ChatBubble messages={mockMessages} onSendMessage={vi.fn()} />);
    // The latest assistant message uses typewriter reveal; wait for it
    expect(await screen.findByText("你好！", {}, { timeout: 500 })).toBeDefined();
    // User messages appear immediately
    expect(screen.getByText("今天天气如何？")).toBeDefined();
  });

  it("calls onSendMessage when submitting input", () => {
    const onSend = vi.fn();
    render(<ChatBubble messages={[]} onSendMessage={onSend} />);
    const input = screen.getByPlaceholderText("输入消息...");
    fireEvent.change(input, { target: { value: "test" } });
    fireEvent.submit(screen.getByRole("button", { name: "发送" }));
    expect(onSend).toHaveBeenCalledWith("test");
  });

  it("does not send empty messages", () => {
    const onSend = vi.fn();
    render(<ChatBubble messages={[]} onSendMessage={onSend} />);
    const input = screen.getByPlaceholderText("输入消息...");
    fireEvent.change(input, { target: { value: "   " } });
    fireEvent.submit(screen.getByRole("button", { name: "发送" }));
    expect(onSend).not.toHaveBeenCalled();
  });
});
