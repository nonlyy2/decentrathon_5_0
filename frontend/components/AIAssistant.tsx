"use client";

import { useState, useRef, useEffect } from "react";
import api from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { MessageCircle, X, Send, Bot, User, Loader2 } from "lucide-react";

interface ChatMessage {
  role: "user" | "assistant";
  content: string;
}

export default function AIAssistant() {
  const { user } = useAuth();
  const [open, setOpen] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  const isManager = user && ["manager", "committee", "tech-admin", "superadmin", "admin", "auditor"].includes(user.role);

  useEffect(() => {
    if (open && messages.length === 0) {
      const intro: ChatMessage = {
        role: "assistant",
        content: isManager
          ? "👋 Hi! I'm your AI data assistant. I can help you analyze candidate data, identify trends, and answer questions about the current admissions cycle. What would you like to know?"
          : "👋 Hi! I'm your inVision U FAQ assistant. I can answer questions about the application process, university, and what to expect. How can I help you?",
      };
      setMessages([intro]);
    }
  }, [open, isManager, messages.length]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = async () => {
    const text = input.trim();
    if (!text || loading) return;

    setInput("");
    const userMsg: ChatMessage = { role: "user", content: text };
    setMessages((prev) => [...prev, userMsg]);
    setLoading(true);

    try {
      const res = await api.post("/ai/assistant", { message: text });
      setMessages((prev) => [...prev, { role: "assistant", content: res.data.reply }]);
    } catch {
      setMessages((prev) => [...prev, {
        role: "assistant",
        content: "Sorry, I couldn't process your request. Please try again.",
      }]);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <>
      {/* Floating button */}
      <button
        onClick={() => setOpen((o) => !o)}
        className="fixed bottom-6 right-6 z-50 w-14 h-14 rounded-full shadow-lg flex items-center justify-center transition-all hover:scale-105"
        style={{ backgroundColor: "#c1f11d" }}
        aria-label="Open AI Assistant"
      >
        {open ? <X size={22} color="#111" /> : <MessageCircle size={22} color="#111" />}
      </button>

      {/* Chat panel */}
      {open && (
        <div className="fixed bottom-24 right-6 z-50 w-96 max-w-[calc(100vw-2rem)] bg-card border border-border rounded-2xl shadow-2xl flex flex-col overflow-hidden"
          style={{ height: "480px" }}>
          {/* Header */}
          <div className="flex items-center gap-2 px-4 py-3 border-b border-border" style={{ backgroundColor: "#c1f11d22" }}>
            <Bot size={18} style={{ color: "#c1f11d" }} />
            <div className="flex-1">
              <p className="text-sm font-semibold">
                {isManager ? "AI Data Assistant" : "inVision U Assistant"}
              </p>
              <p className="text-[11px] text-muted-foreground">
                {isManager ? "Powered by Gemini · Admissions analytics" : "Powered by Gemini · FAQ & Help"}
              </p>
            </div>
            <button onClick={() => setOpen(false)} className="text-muted-foreground hover:text-foreground">
              <X size={16} />
            </button>
          </div>

          {/* Messages */}
          <div className="flex-1 overflow-y-auto p-4 space-y-3">
            {messages.map((msg, i) => (
              <div key={i} className={`flex gap-2 ${msg.role === "user" ? "flex-row-reverse" : ""}`}>
                <div className={`w-7 h-7 rounded-full flex items-center justify-center shrink-0 ${msg.role === "assistant" ? "bg-lime-100" : "bg-blue-100"}`}>
                  {msg.role === "assistant" ? <Bot size={14} style={{ color: "#c1f11d" }} /> : <User size={14} className="text-blue-600" />}
                </div>
                <div className={`max-w-[80%] text-sm rounded-2xl px-3 py-2 whitespace-pre-wrap ${
                  msg.role === "assistant"
                    ? "bg-muted text-foreground rounded-tl-sm"
                    : "text-white rounded-tr-sm"
                }`}
                  style={msg.role === "user" ? { backgroundColor: "#c1f11d", color: "#111" } : {}}>
                  {msg.content}
                </div>
              </div>
            ))}
            {loading && (
              <div className="flex gap-2">
                <div className="w-7 h-7 rounded-full bg-lime-100 flex items-center justify-center">
                  <Bot size={14} style={{ color: "#c1f11d" }} />
                </div>
                <div className="bg-muted rounded-2xl rounded-tl-sm px-3 py-2">
                  <Loader2 size={16} className="animate-spin text-muted-foreground" />
                </div>
              </div>
            )}
            <div ref={bottomRef} />
          </div>

          {/* Input */}
          <div className="p-3 border-t border-border flex gap-2 items-end">
            <Textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={isManager ? "Ask about candidates, trends, scores..." : "Ask about the application process..."}
              rows={2}
              className="resize-none text-sm"
            />
            <Button
              size="sm"
              onClick={handleSend}
              disabled={loading || !input.trim()}
              className="h-10 w-10 p-0 shrink-0"
              style={{ backgroundColor: "#c1f11d", color: "#111" }}
            >
              <Send size={16} />
            </Button>
          </div>
        </div>
      )}
    </>
  );
}
