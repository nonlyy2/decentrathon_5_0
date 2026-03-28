"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useFetch } from "@/lib/hooks";
import api from "@/lib/api";
import { CandidateDetail } from "@/lib/types";
import StatusBadge from "@/components/StatusBadge";
import ScoreRadar from "@/components/ScoreRadar";
import DecisionButtons from "@/components/DecisionButtons";
import KeyStrengthsRedFlags from "@/components/KeyStrengthsRedFlags";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ArrowLeft, Sparkles, Loader2 } from "lucide-react";
import { toast } from "sonner";

const categoryColors: Record<string, string> = {
  "Strong Recommend": "bg-green-100 text-green-800",
  "Recommend": "bg-blue-100 text-blue-800",
  "Borderline": "bg-yellow-100 text-yellow-800",
  "Not Recommended": "bg-red-100 text-red-800",
};

const riskColors: Record<string, string> = {
  low: "bg-green-100 text-green-700",
  medium: "bg-yellow-100 text-yellow-700",
  high: "bg-red-100 text-red-700",
};

const scoreLabels = [
  { key: "score_leadership", label: "Leadership", explKey: "explanation_leadership", weight: "25%" },
  { key: "score_motivation", label: "Motivation", explKey: "explanation_motivation", weight: "25%" },
  { key: "score_growth", label: "Growth", explKey: "explanation_growth", weight: "20%" },
  { key: "score_vision", label: "Vision", explKey: "explanation_vision", weight: "15%" },
  { key: "score_communication", label: "Communication", explKey: "explanation_communication", weight: "15%" },
];

export default function CandidateDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { data: detail, loading, refetch } = useFetch<CandidateDetail>(`/candidates/${params.id}`);
  const [analyzing, setAnalyzing] = useState(false);
  const [expandedScore, setExpandedScore] = useState<string | null>(null);

  const handleAnalyze = async (force = false) => {
    setAnalyzing(true);
    try {
      await api.post(`/candidates/${params.id}/analyze${force ? "?force=true" : ""}`);
      toast.success("Analysis complete");
      refetch();
    } catch {
      toast.error("Analysis failed");
    } finally {
      setAnalyzing(false);
    }
  };

  if (loading || !detail) {
    return (
      <div className="space-y-4">
        <div className="h-8 bg-slate-200 rounded animate-pulse w-48" />
        <div className="grid grid-cols-1 lg:grid-cols-[3fr_2fr] gap-6">
          <div className="space-y-4">{[1, 2, 3].map((i) => <Card key={i}><CardContent className="p-6"><div className="h-24 bg-slate-200 rounded animate-pulse" /></CardContent></Card>)}</div>
          <div className="space-y-4">{[1, 2].map((i) => <Card key={i}><CardContent className="p-6"><div className="h-32 bg-slate-200 rounded animate-pulse" /></CardContent></Card>)}</div>
        </div>
      </div>
    );
  }

  const a = detail.analysis;

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={() => router.push("/candidates")}>
            <ArrowLeft size={16} />
          </Button>
          <h1 className="text-2xl font-bold">{detail.full_name}</h1>
          <StatusBadge status={detail.status} />
        </div>
        {a ? (
          <Button variant="outline" size="sm" onClick={() => handleAnalyze(true)} disabled={analyzing}>
            {analyzing ? <><Loader2 className="animate-spin mr-2" size={14} /> Re-analyzing...</> : "Re-analyze"}
          </Button>
        ) : (
          <Button className="bg-purple-600 hover:bg-purple-700" onClick={() => handleAnalyze()} disabled={analyzing}>
            {analyzing ? <><Loader2 className="animate-spin mr-2" size={14} /> Analyzing...</> : <><Sparkles size={16} className="mr-2" /> Analyze with AI</>}
          </Button>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[3fr_2fr] gap-6">
        {/* Left column — Application data */}
        <div className="space-y-4">
          <Card>
            <CardHeader><CardTitle className="text-base">Personal Information</CardTitle></CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div><span className="text-muted-foreground">Email:</span> {detail.email}</div>
                <div><span className="text-muted-foreground">Age:</span> {detail.age || "—"}</div>
                <div><span className="text-muted-foreground">City:</span> {detail.city || "—"}</div>
                <div><span className="text-muted-foreground">School:</span> {detail.school || "—"}</div>
                <div><span className="text-muted-foreground">Graduation:</span> {detail.graduation_year || "—"}</div>
              </div>
            </CardContent>
          </Card>

          {detail.achievements && (
            <Card>
              <CardHeader><CardTitle className="text-base">Achievements</CardTitle></CardHeader>
              <CardContent><p className="text-sm whitespace-pre-wrap">{detail.achievements}</p></CardContent>
            </Card>
          )}

          {detail.extracurriculars && (
            <Card>
              <CardHeader><CardTitle className="text-base">Extracurriculars</CardTitle></CardHeader>
              <CardContent><p className="text-sm whitespace-pre-wrap">{detail.extracurriculars}</p></CardContent>
            </Card>
          )}

          <Card>
            <CardHeader><CardTitle className="text-base">Essay</CardTitle></CardHeader>
            <CardContent>
              <div className="text-sm whitespace-pre-wrap max-h-96 overflow-y-auto leading-relaxed">
                {detail.essay}
              </div>
            </CardContent>
          </Card>

          {detail.motivation_statement && (
            <Card>
              <CardHeader><CardTitle className="text-base">Motivation Statement</CardTitle></CardHeader>
              <CardContent><p className="text-sm whitespace-pre-wrap">{detail.motivation_statement}</p></CardContent>
            </Card>
          )}
        </div>

        {/* Right column — Analysis + Decisions */}
        <div className="space-y-4">
          {a ? (
            <>
              {/* Score overview */}
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-base flex items-center gap-2">
                      <Sparkles size={16} className="text-purple-500" /> AI Analysis
                    </CardTitle>
                    <Badge className={riskColors[a.ai_generated_risk]}>
                      AI Risk: {a.ai_generated_risk}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center gap-4">
                    <div className="text-4xl font-bold text-purple-600">{a.final_score.toFixed(1)}</div>
                    <Badge className={`${categoryColors[a.category]} text-sm`}>{a.category}</Badge>
                  </div>
                  <ScoreRadar scores={{
                    leadership: a.score_leadership,
                    motivation: a.score_motivation,
                    growth: a.score_growth,
                    vision: a.score_vision,
                    communication: a.score_communication,
                  }} />
                </CardContent>
              </Card>

              {/* Summary */}
              <Card>
                <CardHeader><CardTitle className="text-base">Summary</CardTitle></CardHeader>
                <CardContent>
                  <p className="text-sm leading-relaxed text-slate-700">{a.summary}</p>
                  <p className="text-xs text-muted-foreground mt-3">
                    Analyzed by {a.model_used} on {new Date(a.analyzed_at).toLocaleDateString()}
                  </p>
                </CardContent>
              </Card>

              <KeyStrengthsRedFlags strengths={a.key_strengths} redFlags={a.red_flags} />

              {/* Score breakdown */}
              <Card>
                <CardHeader><CardTitle className="text-base">Score Breakdown</CardTitle></CardHeader>
                <CardContent className="space-y-3">
                  {scoreLabels.map((s) => {
                    const score = a[s.key as keyof typeof a] as number;
                    const explanation = a[s.explKey as keyof typeof a] as string;
                    const isExpanded = expandedScore === s.key;
                    return (
                      <div key={s.key}>
                        <button
                          className="w-full text-left"
                          onClick={() => setExpandedScore(isExpanded ? null : s.key)}
                        >
                          <div className="flex items-center justify-between mb-1">
                            <span className="text-sm font-medium">{s.label} <span className="text-muted-foreground text-xs">({s.weight})</span></span>
                            <span className="text-sm font-bold">{score}</span>
                          </div>
                          <div className="w-full bg-slate-100 rounded-full h-2">
                            <div
                              className="h-2 rounded-full bg-purple-500 transition-all"
                              style={{ width: `${score}%` }}
                            />
                          </div>
                        </button>
                        {isExpanded && (
                          <p className="text-xs text-slate-600 mt-1 pl-2 border-l-2 border-purple-200">
                            {explanation}
                          </p>
                        )}
                      </div>
                    );
                  })}
                </CardContent>
              </Card>
            </>
          ) : (
            <Card>
              <CardContent className="p-8 text-center">
                <Sparkles className="mx-auto text-slate-300 mb-3" size={32} />
                <p className="text-muted-foreground">Not yet analyzed</p>
                <p className="text-sm text-muted-foreground mt-1">Click &quot;Analyze with AI&quot; to generate scores</p>
              </CardContent>
            </Card>
          )}

          {/* Decision panel */}
          <Card>
            <CardHeader><CardTitle className="text-base">Committee Decision</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <DecisionButtons
                candidateId={detail.id}
                currentStatus={detail.status}
                onDecisionMade={refetch}
              />
              {detail.decisions.length > 0 && (
                <div className="mt-4 space-y-2">
                  <p className="text-xs font-medium text-muted-foreground uppercase">History</p>
                  {detail.decisions.map((d) => (
                    <div key={d.id} className="flex items-start gap-2 text-sm border-l-2 border-slate-200 pl-3">
                      <div>
                        <StatusBadge status={d.decision === "shortlist" ? "shortlisted" : d.decision === "reject" ? "rejected" : d.decision === "waitlist" ? "waitlisted" : "analyzed"} />
                        {d.notes && <p className="text-xs text-slate-500 italic mt-1">{d.notes}</p>}
                        <p className="text-xs text-muted-foreground mt-1">
                          {new Date(d.decided_at).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
