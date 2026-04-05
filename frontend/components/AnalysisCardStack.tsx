"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { AnalysisHistoryEntry } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import { X, Layers } from "lucide-react";

const categoryColors: Record<string, string> = {
  "Strong Recommend": "bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300",
  "Recommend": "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300",
  "Borderline": "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/40 dark:text-yellow-300",
  "Not Recommended": "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300",
};

interface Props {
  current: {
    final_score: number;
    category: string;
    model_used: string;
    analyzed_at: string;
    summary: string;
  };
  history: AnalysisHistoryEntry[];
}

export default function AnalysisCardStack({ current, history }: Props) {
  const [expanded, setExpanded] = useState(false);

  const allCards = [
    {
      id: "current",
      final_score: current.final_score,
      category: current.category,
      model_used: current.model_used,
      analyzed_at: current.analyzed_at,
      summary: current.summary,
      isCurrent: true,
    },
    ...history.map((h) => ({
      id: String(h.id),
      final_score: h.final_score,
      category: h.category,
      model_used: h.model_used || "unknown",
      analyzed_at: h.analyzed_at,
      summary: h.summary || "",
      isCurrent: false,
    })),
  ];

  if (history.length === 0) return null;

  const totalCount = allCards.length;

  // Stacked card offsets (collapsed)
  const stackOffsets = history.slice(0, 3).map((_, i) => ({
    y: (i + 1) * 6,
    x: (i + 1) * 3,
    rotate: (i + 1) * -1.5,
    scale: 1 - (i + 1) * 0.03,
  }));

  // Fan-out positions for expanded view
  const getFanPosition = (index: number, total: number) => {
    const spreadAngle = Math.min(total * 8, 40);
    const startAngle = -spreadAngle / 2;
    const step = total > 1 ? spreadAngle / (total - 1) : 0;
    const angle = startAngle + index * step;
    const yOffset = index * 80;
    return { rotate: angle, y: yOffset, x: 0 };
  };

  return (
    <>
      {/* Collapsed stack */}
      {!expanded && (
        <div className="relative cursor-pointer" onClick={() => setExpanded(true)}>
          {/* Background cards (stack effect) */}
          {stackOffsets.map((offset, i) => (
            <div
              key={`bg-${i}`}
              className="absolute inset-0 rounded-lg border bg-card shadow-sm"
              style={{
                transform: `translateY(${offset.y}px) translateX(${offset.x}px) rotate(${offset.rotate}deg) scale(${offset.scale})`,
                zIndex: 10 - i,
                opacity: 1 - i * 0.15,
              }}
            />
          ))}
          {/* Top card (current) */}
          <div className="relative z-20 rounded-lg border bg-card p-4 shadow-md hover:shadow-lg transition-shadow">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="text-2xl font-bold text-purple-600">{current.final_score.toFixed(1)}</div>
                <Badge className={categoryColors[current.category] || ""}>{current.category}</Badge>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="text-[10px] gap-1">
                  <Layers size={10} />
                  {totalCount} analyses
                </Badge>
              </div>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              {current.model_used} — {new Date(current.analyzed_at).toLocaleString("en-GB", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" })}
            </p>
            <p className="text-xs text-muted-foreground mt-1 italic">Click to view all analyses</p>
          </div>
          {/* Extra padding for the stack */}
          <div style={{ height: Math.min(history.length, 3) * 6 + 4 }} />
        </div>
      )}

      {/* Expanded fan-out overlay */}
      <AnimatePresence>
        {expanded && (
          <>
            {/* Backdrop */}
            <motion.div
              className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setExpanded(false)}
            />
            {/* Cards container */}
            <motion.div
              className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto py-12 px-4"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setExpanded(false)}
            >
              <div className="relative w-full max-w-lg" onClick={(e) => e.stopPropagation()}>
                {/* Close button */}
                <motion.button
                  className="absolute -top-2 -right-2 z-[60] w-8 h-8 rounded-full bg-card border shadow-lg flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
                  initial={{ scale: 0 }}
                  animate={{ scale: 1, transition: { delay: 0.2 } }}
                  onClick={() => setExpanded(false)}
                >
                  <X size={16} />
                </motion.button>

                {/* Fan-out cards */}
                <div className="space-y-3">
                  {allCards.map((card, i) => {
                    const fan = getFanPosition(i, allCards.length);
                    return (
                      <motion.div
                        key={card.id}
                        className={`relative rounded-lg border bg-card p-4 shadow-lg ${card.isCurrent ? "ring-2 ring-purple-500" : ""}`}
                        initial={{
                          opacity: 0,
                          y: -50,
                          rotate: fan.rotate * 2,
                          scale: 0.8,
                        }}
                        animate={{
                          opacity: 1,
                          y: 0,
                          rotate: 0,
                          scale: 1,
                          transition: {
                            delay: i * 0.08,
                            type: "spring",
                            stiffness: 300,
                            damping: 25,
                          },
                        }}
                        exit={{
                          opacity: 0,
                          y: -30,
                          scale: 0.9,
                          transition: { delay: (allCards.length - i) * 0.04 },
                        }}
                      >
                        {card.isCurrent && (
                          <Badge className="absolute -top-2 -left-2 bg-purple-600 text-white text-[10px]">
                            Current
                          </Badge>
                        )}
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <div className={`text-2xl font-bold ${card.isCurrent ? "text-purple-600" : "text-muted-foreground"}`}>
                              {card.final_score.toFixed(1)}
                            </div>
                            <Badge className={categoryColors[card.category] || ""}>{card.category}</Badge>
                          </div>
                          <span className="text-xs text-muted-foreground">
                            #{i + 1}
                          </span>
                        </div>
                        {card.summary && (
                          <p className="text-sm text-muted-foreground mt-2 line-clamp-2">
                            {card.summary}
                          </p>
                        )}
                        <p className="text-xs text-muted-foreground mt-2">
                          {card.model_used} — {new Date(card.analyzed_at).toLocaleString("en-GB", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" })}
                        </p>
                      </motion.div>
                    );
                  })}
                </div>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
}
