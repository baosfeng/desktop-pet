import { AnimatePresence, motion } from "motion/react";
import { useCallback, useEffect, useRef, useState } from "react";
import ReactMarkdown from "react-markdown";
import type React from "react";

import type { Message } from "@/stores/petStore";

interface ChatBubbleProps {
  messages: Message[];
  onSendMessage: (text: string) => void;
}

const TYPEWRITER_SPEED_MS = 30;

export function ChatBubble({ messages, onSendMessage }: ChatBubbleProps): React.JSX.Element {
  const [inputText, setInputText] = useState("");
  const [typingIndex, setTypingIndex] = useState<number | null>(null);
  const [visibleChars, setVisibleChars] = useState(0);
  const listRef = useRef<HTMLDivElement>(null);

  const latestAssistantIndex = ((): number | null => {
    for (let i = messages.length - 1; i >= 0; i--) {
      const msg = messages[i];
      if (msg?.role === "assistant") return i;
    }
    return null;
  })();

  useEffect(() => {
    if (latestAssistantIndex !== typingIndex) {
      setTypingIndex(latestAssistantIndex); // eslint-disable-line react-hooks/set-state-in-effect
      setVisibleChars(0);
    }
  }, [latestAssistantIndex, typingIndex]);

  useEffect(() => {
    if (typingIndex === null) return;
    const msg = messages[typingIndex];
    if (!msg) return;
    if (visibleChars < msg.content.length) {
      const timer = setTimeout((): void => { setVisibleChars((p) => p + 1); }, TYPEWRITER_SPEED_MS);
      return (): void => { clearTimeout(timer); };
    }
  }, [typingIndex, visibleChars, messages]);

  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight;
    }
  }, [messages, visibleChars]);

  const handleSubmit = useCallback(
    (e: React.SyntheticEvent) => {
      e.preventDefault();
      const trimmed = inputText.trim();
      if (!trimmed) return;
      onSendMessage(trimmed);
      setInputText("");
    },
    [inputText, onSendMessage],
  );

  return (
    <motion.div
      className="absolute bottom-4 left-4 right-4 z-50 flex flex-col max-h-[40vh] rounded-[16px] bg-cream border border-soft-brown/50 shadow-lg overflow-hidden"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Messages list */}
      <div
        className="flex-1 overflow-y-auto p-3 flex flex-col gap-2 scrollbar-thin"
        ref={listRef}
        style={{ scrollbarWidth: "thin", scrollbarColor: "#D4E0C8 transparent" }}
      >
        <AnimatePresence initial={false}>
          {messages.map((msg, idx) => {
            const isTyping = idx === typingIndex && visibleChars < msg.content.length;
            const displayContent =
              idx === typingIndex ? msg.content.slice(0, visibleChars) : msg.content;
            const isUser = msg.role === "user";

            return (
              <motion.div
                key={msg.id}
                className={`max-w-[85%] ${isUser ? "self-end" : "self-start"}`}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.25 }}
              >
                <div
                  className={`text-[11px] font-semibold mb-[2px] uppercase tracking-[0.5px] ${
                    isUser ? "text-right text-peach/70" : "text-secondary/70"
                  }`}
                >
                  {isUser ? "你" : "宠物"}
                </div>
                <div
                  className={`rounded-[12px] px-3 py-2 text-[13px] leading-[1.55] break-words ${
                    isUser
                      ? "bg-peach text-text-brown rounded-br-[4px]"
                      : "bg-secondary text-text-brown rounded-bl-[4px]"
                  }`}
                >
                  {isUser ? (
                    <p>{displayContent}</p>
                  ) : (
                    <ReactMarkdown>{displayContent}</ReactMarkdown>
                  )}
                  {isTyping && <span className="inline-block ml-[1px] font-bold text-text-brown/60 animate-blink">|</span>}
                </div>
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>

      {/* Input form */}
      <form
        onSubmit={handleSubmit}
        className="flex gap-2 px-3 pb-3 pt-2 border-t border-soft-brown/30"
      >
        <input
          className="flex-1 px-3 py-2 border border-soft-brown/40 rounded-[8px] bg-cream text-text-brown text-[13px] outline-none focus:border-primary/60 transition-colors placeholder:text-text-brown/30"
          type="text"
          placeholder="输入消息..."
          value={inputText}
          onChange={(e): void => { setInputText(e.target.value); }}
        />
        <motion.button
          className="px-4 py-2 border-none rounded-[8px] bg-primary text-primary-content text-[13px] font-medium cursor-pointer"
          type="submit"
          aria-label="发送"
          whileHover={{ scale: 1.04 }}
          whileTap={{ scale: 0.97 }}
        >
          发送
        </motion.button>
      </form>
    </motion.div>
  );
}
