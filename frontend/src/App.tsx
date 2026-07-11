import { useCallback, useEffect, useState } from "react";
import type React from "react";

import { ChatBubble } from "@/components/ChatBubble";
import { SettingsPanel } from "@/components/SettingsPanel";

import { type PetEvent, onPetEvent, sendMessage } from "@/lib/bridge";

import "./App.css";

export interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: number;
}

let nextId = 0;

function generateId(): string {
  nextId += 1;
  return `msg-${String(nextId)}-${String(Date.now())}`;
}

export default function App(): React.JSX.Element {
  const [messages, setMessages] = useState<Message[]>([]);
  const [petState, setPetState] = useState<string>("idle");
  const [showSettings, setShowSettings] = useState(false);

  useEffect(() => {
    const unlistenPromise = onPetEvent((event: PetEvent) => {
      setPetState(event.kind);

      if (event.kind === "message" && typeof event.data.text === "string") {
        const msg: Message = {
          id: generateId(),
          role: "assistant",
          content: event.data.text,
          timestamp: Date.now(),
        };
        setMessages((prev) => [...prev, msg]);
      }
    });

    return (): void => {
      void unlistenPromise.then((unlisten) => {
        unlisten();
      });
    };
  }, []);

  const handleSendMessage = useCallback((text: string) => {
    const userMsg: Message = {
      id: generateId(),
      role: "user",
      content: text,
      timestamp: Date.now(),
    };
    setMessages((prev) => [...prev, userMsg]);
    void sendMessage(text);
  }, []);

  const handleToggleSettings = useCallback(() => {
    setShowSettings((prev) => !prev);
  }, []);

  return (
    <div className="app" data-state={petState}>
      <div className="pet-area">
        <canvas id="live2d-canvas" className="live2d-canvas" />
      </div>

      <ChatBubble messages={messages} onSendMessage={handleSendMessage} />

      <button
        className="settings-toggle"
        onClick={handleToggleSettings}
        type="button"
        aria-label="设置"
      >
        ⚙
      </button>

      {showSettings && <SettingsPanel onClose={handleToggleSettings} />}
    </div>
  );
}
