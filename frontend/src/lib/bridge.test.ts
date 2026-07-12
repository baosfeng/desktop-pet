import { describe, expect, it, vi, beforeEach } from "vitest";

// Mock Tauri API
vi.mock("@tauri-apps/api/core", () => ({
  invoke: vi.fn(),
}));

vi.mock("@tauri-apps/api/event", () => ({
  listen: vi.fn(() => Promise.resolve(vi.fn())),
}));

import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";

import {
  type PetEvent,
  getStatus,
  getWindowPosition,
  onPetEvent,
  sendMessage,
  setWindowPosition,
  toggleClickthrough,
} from "../lib/bridge";

describe("bridge", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("sendMessage", () => {
    it("invokes chat with text", async () => {
      await sendMessage("hello");
      expect(invoke).toHaveBeenCalledWith("chat", { text: "hello" });
    });
  });

  describe("toggleClickthrough", () => {
    it("invokes toggle_clickthrough with enabled", async () => {
      await toggleClickthrough(true);
      expect(invoke).toHaveBeenCalledWith("toggle_clickthrough", { enabled: true });
    });

    it("invokes toggle_clickthrough with disabled", async () => {
      await toggleClickthrough(false);
      expect(invoke).toHaveBeenCalledWith("toggle_clickthrough", { enabled: false });
    });
  });

  describe("getWindowPosition", () => {
    it("returns position from invoke", async () => {
      vi.mocked(invoke).mockResolvedValue({ x: 100, y: 200 });
      const pos = await getWindowPosition();
      expect(pos).toEqual({ x: 100, y: 200 });
      expect(invoke).toHaveBeenCalledWith("get_window_position");
    });
  });

  describe("setWindowPosition", () => {
    it("invokes set_window_position with coordinates", async () => {
      await setWindowPosition(300, 400);
      expect(invoke).toHaveBeenCalledWith("set_window_position", { x: 300, y: 400 });
    });
  });

  describe("getStatus", () => {
    it("returns status from invoke", async () => {
      vi.mocked(invoke).mockResolvedValue({ state: "idle" });
      const status = await getStatus();
      expect(status).toEqual({ state: "idle" });
    });
  });

  describe("onPetEvent", () => {
    it("listens to pet:event and calls handler", async () => {
      const handler = vi.fn();
      await onPetEvent(handler);

      expect(listen).toHaveBeenCalledWith("pet:event", expect.any(Function));

      // Get the listener callback that was passed to listen
      const listenCallback = vi.mocked(listen).mock.calls[0]?.[1] as (e: {
        payload: PetEvent;
      }) => void;

      const event = { kind: "agent.reply", data: { text: "hi" } };
      listenCallback({ payload: event });
      expect(handler).toHaveBeenCalledWith(event);
    });

    it("returns an unlisten function", async () => {
      const unlistenFn = vi.fn();
      vi.mocked(listen).mockResolvedValue(unlistenFn);

      const result = await onPetEvent(vi.fn());
      expect(result).toBe(unlistenFn);
    });
  });
});
