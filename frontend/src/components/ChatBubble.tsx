import { useCallback, useEffect, useState } from "react";
import type React from "react";

import ReactMarkdown from "react-markdown";

import { Bubble, BubbleContent } from "@/components/ui/bubble";
import { Message, MessageContent, MessageFooter } from "@/components/ui/message";
import {
  MessageScroller,
  MessageScrollerProvider,
  MessageScrollerViewport,
} from "@/components/ui/message-scroller";

import type { Message as MessageType } from "@/stores/petStore";

interface ChatBubbleProps {
  messages: MessageType[];
  onSendMessage: (text: string) => void;
}

const TYPEWRITER_SPEED_MS = 30;

export function ChatBubble({ messages, onSendMessage }: ChatBubbleProps): React.JSX.Element {
  const [inputText, setInputText] = useState("");
  const [typingIndex, setTypingIndex] = useState<number | null>(null);
  const [visibleChars, setVisibleChars] = useState(0);

  const latestAssistantIndex = ((): number | null => {
    for (let i = messages.length - 1; i >= 0; i--) {
      const msg = messages[i];
      if (msg?.role === "assistant") return i;
    }
    return null;
  })();

  // 同步打字机索引到最新助手消息（使用 setTimeout 避免同步 setState）
  useEffect(() => {
    const id = setTimeout(() => {
      setTypingIndex(latestAssistantIndex);
      setVisibleChars(0);
    }, 0);
    return () => {
      clearTimeout(id);
    };
  }, [latestAssistantIndex]);

  useEffect(() => {
    if (typingIndex === null) return;
    const msg = messages[typingIndex];
    if (!msg) return;
    if (visibleChars < msg.content.length) {
      const timer = setTimeout(() => {
        setVisibleChars((p) => p + 1);
      }, TYPEWRITER_SPEED_MS);
      return () => {
        clearTimeout(timer);
      };
    }
  }, [typingIndex, visibleChars, messages]);

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
    <div className="w-full h-full flex flex-col rounded-xl bg-base-100 border border-soft-brown/50 shadow-lg overflow-hidden">
      {/* Messages area */}
      <div className="flex-1 min-h-0">
        <MessageScrollerProvider>
          <MessageScroller>
            <MessageScrollerViewport className="p-3 flex flex-col gap-3">
              {messages.length === 0 && (
                <div className="flex-1 flex items-center justify-center text-muted-foreground/50 text-xs py-8">
                  开始聊天吧 🐾
                </div>
              )}
              {messages.map((msg, idx) => {
                const isTyping = idx === typingIndex && visibleChars < msg.content.length;
                const displayContent =
                  idx === typingIndex ? msg.content.slice(0, visibleChars) : msg.content;
                const isUser = msg.role === "user";

                return (
                  <Message key={msg.id} align={isUser ? "end" : "start"}>
                    <MessageContent>
                      <Bubble
                        variant={isUser ? "default" : "secondary"}
                        align={isUser ? "end" : "start"}
                      >
                        <BubbleContent className={isUser ? "" : "bg-secondary/80"}>
                          {isUser ? (
                            <p className="text-sm">{displayContent}</p>
                          ) : (
                            <div className="text-sm prose prose-sm max-w-none prose-p:my-0.5 prose-headings:my-1">
                              <ReactMarkdown>{displayContent}</ReactMarkdown>
                            </div>
                          )}
                          {isTyping && (
                            <span className="inline-block ml-0.5 font-bold text-foreground/60 animate-blink">
                              |
                            </span>
                          )}
                        </BubbleContent>
                      </Bubble>
                      {isUser && msg.content && !isTyping && (
                        <MessageFooter>
                          <span className="text-[10px] text-muted-foreground/40">已发送</span>
                        </MessageFooter>
                      )}
                    </MessageContent>
                  </Message>
                );
              })}
            </MessageScrollerViewport>
          </MessageScroller>
        </MessageScrollerProvider>
      </div>

      {/* Input form */}
      <form
        onSubmit={handleSubmit}
        className="flex gap-2 px-3 pb-3 pt-2.5 border-t border-soft-brown/30 shrink-0"
      >
        <input
          className="flex-1 h-auto min-h-0 px-4 py-2.5 rounded-lg border border-soft-brown/40 bg-base-100 text-sm text-foreground placeholder:text-muted-foreground/40 outline-none focus:border-primary/60 focus:ring-2 focus:ring-primary/20 transition-all"
          type="text"
          placeholder="输入消息..."
          value={inputText}
          onChange={(e): void => {
            setInputText(e.target.value);
          }}
        />
        <button
          className="shrink-0 h-auto min-h-0 px-5 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium cursor-pointer border-none transition-all duration-150 hover:opacity-90 active:scale-95"
          type="submit"
          aria-label="发送"
        >
          发送
        </button>
      </form>
    </div>
  );
}
