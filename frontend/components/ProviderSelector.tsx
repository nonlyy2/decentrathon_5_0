"use client";

import { useState, useEffect } from "react";
import api from "@/lib/api";
import { Globe, MonitorCog } from "lucide-react";

interface Props {
  value: string;
  onChange: (provider: string) => void;
}

const providerMeta: Record<string, { label: string; icon: typeof Globe; description: string }> = {
  gemini: { label: "Gemini", icon: Globe, description: "Cloud API" },
  ollama: { label: "Ollama", icon: MonitorCog, description: "Local" },
};

export function useAIProvider() {
  const [providers, setProviders] = useState<string[]>([]);
  const [defaultProvider, setDefaultProvider] = useState<string>("");
  const [selected, setSelected] = useState<string>("");

  useEffect(() => {
    api
      .get("/ai-providers")
      .then((res) => {
        const { providers: p, default_provider: d } = res.data;
        setProviders(p);
        setDefaultProvider(d);
        const saved = localStorage.getItem("ai_provider");
        if (saved && p.includes(saved)) {
          setSelected(saved);
        } else {
          setSelected(d);
        }
      })
      .catch(() => {
        // fallback
        setProviders(["gemini"]);
        setDefaultProvider("gemini");
        setSelected("gemini");
      });
  }, []);

  const selectProvider = (p: string) => {
    setSelected(p);
    localStorage.setItem("ai_provider", p);
  };

  return { providers, defaultProvider, selected, selectProvider };
}

export default function ProviderSelector({ value, onChange }: Props) {
  const { providers } = useAIProvider();

  if (providers.length <= 1) return null;

  return (
    <div className="flex items-center gap-1 bg-slate-100 rounded-lg p-0.5">
      {providers.map((p) => {
        const meta = providerMeta[p] || { label: p, icon: Globe, description: "" };
        const Icon = meta.icon;
        const isActive = value === p;
        return (
          <button
            key={p}
            onClick={() => onChange(p)}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all ${
              isActive
                ? "bg-white text-purple-700 shadow-sm"
                : "text-slate-500 hover:text-slate-700"
            }`}
          >
            <Icon size={14} />
            <span>{meta.label}</span>
            {meta.description && (
              <span className={`text-xs ${isActive ? "text-purple-400" : "text-slate-400"}`}>
                {meta.description}
              </span>
            )}
          </button>
        );
      })}
    </div>
  );
}
