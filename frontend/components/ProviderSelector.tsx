"use client";

import { Globe, MonitorCog } from "lucide-react";

interface Props {
  providers: string[];
  value: string;
  onChange: (provider: string) => void;
}

const providerMeta: Record<string, { label: string; icon: typeof Globe; badge: string }> = {
  gemini: { label: "Gemini", icon: Globe, badge: "speed" },
  ollama: { label: "Ollama", icon: MonitorCog, badge: "privacy" },
};

export default function ProviderSelector({ providers, value, onChange }: Props) {
  if (providers.length === 0) return null;

  return (
    <div className="flex items-center gap-1 bg-muted border border-border rounded-lg p-0.5">
      {providers.map((p) => {
        const meta = providerMeta[p] || { label: p, icon: Globe, badge: "" };
        const Icon = meta.icon;
        const isActive = value === p;
        return (
          <button
            key={p}
            onClick={() => onChange(p)}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all duration-200 ${
              isActive
                ? "text-foreground font-semibold shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            }`}
            style={isActive ? { backgroundColor: "#c1f11d", color: "#111" } : undefined}
          >
            <Icon size={14} />
            <span>{meta.label}</span>
            {meta.badge && (
              <span className={`text-xs ${isActive ? "opacity-70" : "text-muted-foreground"}`}>
                {meta.badge}
              </span>
            )}
          </button>
        );
      })}
    </div>
  );
}
