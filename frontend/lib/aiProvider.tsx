"use client";

import { createContext, useContext, useState, useEffect, ReactNode } from "react";

type AIProvider = "gemini" | "ollama";

interface AIProviderContextType {
  provider: AIProvider;
  setProvider: (p: AIProvider) => void;
}

const AIProviderContext = createContext<AIProviderContextType>({
  provider: "gemini",
  setProvider: () => {},
});

export function AIProviderContextProvider({ children }: { children: ReactNode }) {
  const [provider, setProviderState] = useState<AIProvider>("gemini");

  useEffect(() => {
    const saved = localStorage.getItem("ai_provider") as AIProvider;
    if (saved === "gemini" || saved === "ollama") {
      setProviderState(saved);
    } else {
      // default to gemini, clear any stale value
      localStorage.setItem("ai_provider", "gemini");
    }
  }, []);

  const setProvider = (p: AIProvider) => {
    setProviderState(p);
    localStorage.setItem("ai_provider", p);
  };

  return (
    <AIProviderContext.Provider value={{ provider, setProvider }}>
      {children}
    </AIProviderContext.Provider>
  );
}

export function useAIProvider() {
  return useContext(AIProviderContext);
}
