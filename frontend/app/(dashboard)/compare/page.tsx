"use client";

import { useState, useEffect } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import api from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { CandidateDetail } from "@/lib/types";
import StatusBadge from "@/components/StatusBadge";
import ScoreBadge from "@/components/ScoreBadge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";

export default function ComparePage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const { t } = useI18n();
  const ids = (searchParams.get("ids") || "").split(",").filter(Boolean).map(Number);
  const [candidates, setCandidates] = useState<CandidateDetail[]>([]);
  const [loading, setLoading] = useState(true);

  const scoreFields = [
    { key: "score_leadership", label: t("score.leadership") },
    { key: "score_motivation", label: t("score.motivation") },
    { key: "score_growth", label: t("score.growth") },
    { key: "score_vision", label: t("score.vision") },
    { key: "score_communication", label: t("score.communication") },
  ];

  useEffect(() => {
    if (ids.length < 2) return;
    setLoading(true);
    Promise.all(ids.map((id) => api.get(`/candidates/${id}`).then((r) => r.data)))
      .then(setCandidates)
      .catch(() => {})
      .finally(() => setLoading(false));
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  if (ids.length < 2) {
    return (
      <div className="text-center py-20">
        <p className="text-muted-foreground">{t("compare.select_min")}</p>
        <Button variant="outline" className="mt-4" onClick={() => router.push("/candidates")}>
          <ArrowLeft size={16} className="mr-2" /> {t("compare.back")}
        </Button>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="h-8 bg-slate-200 rounded animate-pulse w-48" />
        <div className="grid grid-cols-2 gap-4">
          {[1, 2].map((i) => <Card key={i}><CardContent className="p-6"><div className="h-64 bg-slate-200 rounded animate-pulse" /></CardContent></Card>)}
        </div>
      </div>
    );
  }

  const getHighest = (key: string) => {
    let max = -1;
    let maxId = -1;
    candidates.forEach((c) => {
      const val = c.analysis ? (c.analysis as unknown as Record<string, number>)[key] : 0;
      if (val > max) { max = val; maxId = c.id; }
    });
    return maxId;
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="sm" onClick={() => router.push("/candidates")}>
          <ArrowLeft size={16} />
        </Button>
        <h1 className="text-2xl font-bold">{t("compare.title")}</h1>
      </div>

      {/* Overview comparison */}
      <div className={`grid gap-4`} style={{ gridTemplateColumns: `repeat(${candidates.length}, 1fr)` }}>
        {candidates.map((c) => (
          <Card key={c.id} className="text-center">
            <CardContent className="p-6 space-y-3">
              <h3 className="font-semibold text-lg">{c.full_name}</h3>
              <StatusBadge status={c.status} />
              <div className="flex justify-center gap-2 text-sm text-muted-foreground">
                <span>{c.city || "\u2014"}</span>
                <span>&middot;</span>
                <span>{c.school || "\u2014"}</span>
              </div>
              {c.analysis ? (
                <div className="pt-2">
                  <div className="text-4xl font-bold text-purple-600">{c.analysis.final_score.toFixed(1)}</div>
                  <ScoreBadge score={c.analysis.final_score} category={c.analysis.category} />
                </div>
              ) : (
                <p className="text-muted-foreground text-sm">{t("compare.not_analyzed")}</p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Score breakdown comparison */}
      <Card>
        <CardHeader><CardTitle className="text-base">{t("compare.scores")}</CardTitle></CardHeader>
        <CardContent>
          <div className="space-y-4">
            {scoreFields.map((sf) => {
              const highestId = getHighest(sf.key);
              return (
                <div key={sf.key}>
                  <div className="text-sm font-medium mb-2">{sf.label}</div>
                  <div className="space-y-1">
                    {candidates.map((c) => {
                      const score = c.analysis ? (c.analysis as unknown as Record<string, number>)[sf.key] : 0;
                      const isHighest = c.id === highestId && score > 0;
                      return (
                        <div key={c.id} className="flex items-center gap-3">
                          <div className="w-32 text-xs truncate text-muted-foreground">{c.full_name}</div>
                          <div className="flex-1 bg-slate-100 rounded-full h-4 overflow-hidden">
                            <div
                              className={`h-4 rounded-full transition-all ${isHighest ? "bg-purple-500" : "bg-slate-300"}`}
                              style={{ width: `${score}%` }}
                            />
                          </div>
                          <div className={`w-8 text-sm text-right font-medium ${isHighest ? "text-purple-600" : ""}`}>{score}</div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Summary comparison */}
      <div className={`grid gap-4`} style={{ gridTemplateColumns: `repeat(${candidates.length}, 1fr)` }}>
        {candidates.map((c) => (
          <Card key={c.id}>
            <CardHeader><CardTitle className="text-sm">{c.full_name} \u2014 {t("compare.summary")}</CardTitle></CardHeader>
            <CardContent>
              {c.analysis ? (
                <div className="space-y-3 text-sm">
                  <p className="text-slate-700">{c.analysis.summary}</p>
                  {c.analysis.key_strengths.length > 0 && (
                    <div>
                      <p className="text-xs font-medium text-green-600 mb-1">{t("compare.strengths")}</p>
                      <ul className="list-disc list-inside text-xs text-slate-600">
                        {c.analysis.key_strengths.map((s, i) => <li key={i}>{s}</li>)}
                      </ul>
                    </div>
                  )}
                  {c.analysis.red_flags.length > 0 && (
                    <div>
                      <p className="text-xs font-medium text-red-600 mb-1">{t("compare.red_flags")}</p>
                      <ul className="list-disc list-inside text-xs text-slate-600">
                        {c.analysis.red_flags.map((f, i) => <li key={i}>{f}</li>)}
                      </ul>
                    </div>
                  )}
                </div>
              ) : (
                <p className="text-muted-foreground text-sm">{t("compare.not_analyzed_yet")}</p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
