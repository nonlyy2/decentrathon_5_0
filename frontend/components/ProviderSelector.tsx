"use client";

import { Globe, MonitorCog } from "lucide-react";

interface Props {
  providers: string[];
  value: string;
  onChange: (provider: string) => void;
}

const providerMeta: Record<string, { label: string; icon: typeof Globe; description: string }> = {
  gemini: { label: "Gemini", icon: Globe, description: "Cloud" },
  ollama: { label: "Ollama", icon: MonitorCog, description: "Local" },
};

export default function ProviderSelector({ providers, value, onChange }: Props) {
  if (providers.length === 0) return null;

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
