import { useCallback, useEffect, useRef, useState } from "react";
import type React from "react";

import ReactMarkdown from "react-markdown";

interface ChatBubbleProps {
  messages: Message[];
  onSendMessage: (text: string) => void;
}

export interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: number;
}

const TYPEWRITER_SPEED_MS = 30;

export function ChatBubble({ messages, onSendMessage }: ChatBubbleProps): React.JSX.Element {
  const [inputText, setInputText] = useState("");
  const [typingIndex, setTypingIndex] = useState<number | null>(null);
  const [visibleChars, setVisibleChars] = useState(0);
  const listRef = useRef<HTMLDivElement>(null);

  // The latest assistant message being typewriter-displayed
  const latestAssistantIndex = ((): number | null => {
    for (let i = messages.length - 1; i >= 0; i--) {
      const msg = messages[i];
      if (msg?.role === "assistant") {
        return i;
      }
    }
    return null;
  })();

  // Start / restart typewriter when a new assistant message arrives
  /* eslint-disable react-hooks/set-state-in-effect -- 打字机效果需要响应式重置 */
  useEffect(() => {
    if (latestAssistantIndex !== typingIndex) {
      setTypingIndex(latestAssistantIndex);
      setVisibleChars(0);
    }
  }, [latestAssistantIndex, typingIndex]);
  /* eslint-enable react-hooks/set-state-in-effect */

  // Typewriter character reveal
  useEffect(() => {
    if (typingIndex === null) return;

    const msg = messages[typingIndex];
    if (!msg) return;

    if (visibleChars < msg.content.length) {
      const timer = setTimeout(() => {
        setVisibleChars((prev) => prev + 1);
      }, TYPEWRITER_SPEED_MS);
      return (): void => {
        clearTimeout(timer);
      };
    }
  }, [typingIndex, visibleChars, messages]);

  // Auto-scroll to bottom
  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight;
    }
  }, [messages, visibleChars]);

  const handleSubmit = useCallback(
    (e: React.SyntheticEvent) => {
      e.preventDefault();
      const trimmed = inputText.trim();
      if (trimmed.length === 0) return;

      onSendMessage(trimmed);
      setInputText("");
    },
    [inputText, onSendMessage],
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter" && !e.shiftKey) {
        handleSubmit(e);
      }
    },
    [handleSubmit],
  );

  return (
    <div className="chat-bubble">
      <div className="chat-messages" ref={listRef}>
        {messages.map((msg, idx) => {
          const isTyping = idx === typingIndex && visibleChars < msg.content.length;
          const displayContent =
            idx === typingIndex ? msg.content.slice(0, visibleChars) : msg.content;

          return (
            <div key={msg.id} className={`chat-message chat-message--${msg.role}`}>
              <div className="chat-message__role">
                {msg.role === "user" ? "你" : msg.role === "assistant" ? "宠物" : "系统"}
              </div>
              <div className="chat-message__content">
                {msg.role === "assistant" || msg.role === "system" ? (
                  <ReactMarkdown>{displayContent}</ReactMarkdown>
                ) : (
                  <p>{displayContent}</p>
                )}
                {isTyping && <span className="cursor-blink">|</span>}
              </div>
            </div>
          );
        })}
      </div>

      <form className="chat-input-form" onSubmit={handleSubmit}>
        <input
          className="chat-input"
          type="text"
          placeholder="输入消息..."
          value={inputText}
          onChange={(e) => {
            setInputText(e.target.value);
          }}
          onKeyDown={handleKeyDown}
        />
        <button className="chat-send-btn" type="submit" aria-label="发送">
          发送
        </button>
      </form>
    </div>
  );
}
