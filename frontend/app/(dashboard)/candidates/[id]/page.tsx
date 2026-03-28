"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { useFetch } from "@/lib/hooks";
import { useI18n } from "@/lib/i18n";
import { useAIProvider } from "@/lib/aiProvider";
import api from "@/lib/api";
import { CandidateDetail } from "@/lib/types";
import StatusBadge from "@/components/StatusBadge";
import ScoreRadar from "@/components/ScoreRadar";
import DecisionButtons from "@/components/DecisionButtons";
import KeyStrengthsRedFlags from "@/components/KeyStrengthsRedFlags";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ArrowLeft, Sparkles, Loader2, AlertTriangle, Trash2, Send, MessageSquare } from "lucide-react";
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

export default function CandidateDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { t } = useI18n();
  const { data: detail, refetch } = useFetch<CandidateDetail>(`/candidates/${params.id}`);
  const { provider, setProvider } = useAIProvider();
  const [analyzing, setAnalyzing] = useState(false);
  const [analysisFailed, setAnalysisFailed] = useState(false);
  const [analysisError, setAnalysisError] = useState<string | null>(null);
  const [deletingAnalysis, setDeletingAnalysis] = useState(false);
  const [expandedScore, setExpandedScore] = useState<string | null>(null);
  const [comments, setComments] = useState<{ id: number; user_email: string; content: string; created_at: string }[]>([]);
  const [newComment, setNewComment] = useState("");
  const [postingComment, setPostingComment] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const scoreLabels = [
    { key: "score_leadership", label: t("score.leadership"), explKey: "explanation_leadership", weight: "25%" },
    { key: "score_motivation", label: t("score.motivation"), explKey: "explanation_motivation", weight: "25%" },
    { key: "score_growth", label: t("score.growth"), explKey: "explanation_growth", weight: "20%" },
    { key: "score_vision", label: t("score.vision"), explKey: "explanation_vision", weight: "15%" },
    { key: "score_communication", label: t("score.communication"), explKey: "explanation_communication", weight: "15%" },
  ];

  const fetchComments = useCallback(() => {
    api.get(`/candidates/${params.id}/comments`).then((res) => setComments(res.data)).catch(() => {});
  }, [params.id]);

  useEffect(() => { fetchComments(); }, [fetchComments]);

  const handlePostComment = async () => {
    if (!newComment.trim()) return;
    setPostingComment(true);
    try {
      await api.post(`/candidates/${params.id}/comments`, { content: newComment });
      setNewComment("");
      fetchComments();
    } catch {
      toast.error("Failed to post comment");
    } finally {
      setPostingComment(false);
    }
  };

  // Poll analysis status while analyzing
  const startPolling = () => {
    if (pollRef.current) return;
    pollRef.current = setInterval(async () => {
      try {
        const res = await api.get(`/candidates/${params.id}/analysis-status`);
        const { running, failed, error } = res.data;
        if (!running) {
          clearInterval(pollRef.current!);
          pollRef.current = null;
          setAnalyzing(false);
          if (failed) {
            setAnalysisFailed(true);
            setAnalysisError(error || "Analysis failed");
            toast.error("AI analysis failed");
          } else {
            setAnalysisFailed(false);
            setAnalysisError(null);
            toast.success("Analysis complete");
            refetch();
          }
        }
      } catch {
        clearInterval(pollRef.current!);
        pollRef.current = null;
        setAnalyzing(false);
      }
    }, 2000);
  };

  useEffect(() => {
    return () => { if (pollRef.current) clearInterval(pollRef.current); };
  }, []);

  const handleAnalyze = async (force = false) => {
    setAnalysisFailed(false);
    setAnalysisError(null);
    setAnalyzing(true);
    try {
      const params2 = new URLSearchParams();
      if (force) params2.set("force", "true");
      if (provider) params2.set("provider", provider);
      const qs = params2.toString();
      await api.post(`/candidates/${params.id}/analyze${qs ? `?${qs}` : ""}`);
      startPolling();
    } catch {
      setAnalyzing(false);
      toast.error("Failed to start analysis");
    }
  };

  const handleDeleteAnalysis = async () => {
    if (!confirm(t("detail.confirm_delete"))) return;
    setDeletingAnalysis(true);
    try {
      await api.delete(`/candidates/${params.id}/analysis`);
      toast.success("Analysis deleted");
      setAnalysisFailed(false);
      refetch();
    } catch {
      toast.error("Failed to delete analysis");
    } finally {
      setDeletingAnalysis(false);
    }
  };

  if (!detail) {
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
        <div className="flex items-center gap-3">
          {/* Provider toggle */}
          <div className="flex items-center gap-1 bg-slate-100 border rounded-lg p-1">
            <button
              onClick={() => setProvider("gemini")}
              className={`text-xs px-2.5 py-1 rounded-md transition-colors ${
                provider === "gemini" ? "bg-purple-600 text-white shadow-sm" : "text-slate-500 hover:text-slate-800"
              }`}
            >
              ☁ Gemini
            </button>
            <button
              onClick={() => setProvider("ollama")}
              className={`text-xs px-2.5 py-1 rounded-md transition-colors ${
                provider === "ollama" ? "bg-purple-600 text-white shadow-sm" : "text-slate-500 hover:text-slate-800"
              }`}
            >
              ⚙ Ollama
            </button>
          </div>
          {a ? (
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => handleAnalyze(true)} disabled={analyzing || deletingAnalysis}>
                {analyzing ? <><Loader2 className="animate-spin mr-2" size={14} /> {t("detail.re_analyzing")}</> : t("detail.reanalyze")}
              </Button>
              <Button variant="outline" size="sm" className="text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200" onClick={handleDeleteAnalysis} disabled={deletingAnalysis}>
                {deletingAnalysis ? <Loader2 className="animate-spin" size={14} /> : <Trash2 size={14} />}
              </Button>
            </div>
          ) : (
            <Button className="bg-purple-600 hover:bg-purple-700" onClick={() => handleAnalyze()} disabled={analyzing}>
              {analyzing ? <><Loader2 className="animate-spin mr-2" size={14} /> {t("detail.analyzing")}</> : <><Sparkles size={16} className="mr-2" /> {t("detail.analyze")}</>}
            </Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[3fr_2fr] gap-6">
        {/* Left column — Application data */}
        <div className="space-y-4">
          <Card>
            <CardHeader><CardTitle className="text-base">{t("detail.personal")}</CardTitle></CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div><span className="text-muted-foreground">{t("detail.email")}:</span> {detail.email}</div>
                <div><span className="text-muted-foreground">{t("detail.age")}:</span> {detail.age || "\u2014"}</div>
                <div><span className="text-muted-foreground">{t("detail.city")}:</span> {detail.city || "\u2014"}</div>
                <div><span className="text-muted-foreground">{t("detail.school")}:</span> {detail.school || "\u2014"}</div>
                <div><span className="text-muted-foreground">{t("detail.graduation")}:</span> {detail.graduation_year || "\u2014"}</div>
              </div>
            </CardContent>
          </Card>

          {detail.achievements && (
            <Card>
              <CardHeader><CardTitle className="text-base">{t("detail.achievements")}</CardTitle></CardHeader>
              <CardContent><p className="text-sm whitespace-pre-wrap">{detail.achievements}</p></CardContent>
            </Card>
          )}

          {detail.extracurriculars && (
            <Card>
              <CardHeader><CardTitle className="text-base">{t("detail.extracurriculars")}</CardTitle></CardHeader>
              <CardContent><p className="text-sm whitespace-pre-wrap">{detail.extracurriculars}</p></CardContent>
            </Card>
          )}

          <Card>
            <CardHeader><CardTitle className="text-base">{t("detail.essay")}</CardTitle></CardHeader>
            <CardContent>
              <div className="text-sm whitespace-pre-wrap max-h-96 overflow-y-auto leading-relaxed">
                {detail.essay}
              </div>
            </CardContent>
          </Card>

          {detail.motivation_statement && (
            <Card>
              <CardHeader><CardTitle className="text-base">{t("detail.motivation")}</CardTitle></CardHeader>
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
                      <Sparkles size={16} className="text-purple-500" /> {t("detail.ai_analysis")}
                    </CardTitle>
                    <Badge className={riskColors[a.ai_generated_risk]}>
                      {t("detail.ai_risk")}: {a.ai_generated_risk}
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
                <CardHeader><CardTitle className="text-base">{t("detail.summary")}</CardTitle></CardHeader>
                <CardContent>
                  <p className="text-sm leading-relaxed text-slate-700">{a.summary}</p>
                  <p className="text-xs text-muted-foreground mt-3">
                    {t("detail.analyzed_by")} {a.model_used} \u2014 {new Date(a.analyzed_at).toLocaleDateString()}
                  </p>
                </CardContent>
              </Card>

              <KeyStrengthsRedFlags strengths={a.key_strengths} redFlags={a.red_flags} />

              {/* Score breakdown */}
              <Card>
                <CardHeader><CardTitle className="text-base">{t("detail.score_breakdown")}</CardTitle></CardHeader>
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
          ) : analysisFailed ? (
            <Card className="border-red-200">
              <CardContent className="p-8 text-center">
                <AlertTriangle className="mx-auto text-red-400 mb-3" size={32} />
                <p className="text-red-600 font-medium">{t("detail.ai_failed")}</p>
                {analysisError && (
                  <p className="text-xs text-slate-400 mt-2 max-w-xs mx-auto line-clamp-3">{analysisError}</p>
                )}
                <Button
                  className="mt-4 bg-purple-600 hover:bg-purple-700"
                  size="sm"
                  onClick={() => handleAnalyze()}
                  disabled={analyzing}
                >
                  {t("detail.retry")}
                </Button>
              </CardContent>
            </Card>
          ) : analyzing ? (
            <Card>
              <CardContent className="p-8 text-center">
                <Loader2 className="mx-auto text-purple-400 mb-3 animate-spin" size={32} />
                <p className="text-purple-600 font-medium">{t("detail.analyzing")}</p>
                <p className="text-sm text-muted-foreground mt-1">{t("detail.analyzing_wait")}</p>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className="p-8 text-center">
                <Sparkles className="mx-auto text-slate-300 mb-3" size={32} />
                <p className="text-muted-foreground">{t("detail.not_analyzed")}</p>
                <p className="text-sm text-muted-foreground mt-1">{t("detail.click_analyze")}</p>
              </CardContent>
            </Card>
          )}

          {/* Decision panel */}
          <Card>
            <CardHeader><CardTitle className="text-base">{t("detail.committee")}</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <DecisionButtons
                candidateId={detail.id}
                currentStatus={detail.status}
                onDecisionMade={refetch}
              />
              {detail.decisions.length > 0 && (
                <div className="mt-4 space-y-2">
                  <p className="text-xs font-medium text-muted-foreground uppercase">{t("detail.history")}</p>
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

          {/* Comments */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <MessageSquare size={16} /> {t("detail.comments")} ({comments.length})
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex gap-2">
                <Textarea
                  placeholder={t("detail.add_comment")}
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  rows={2}
                  className="text-sm"
                />
                <Button
                  size="sm"
                  className="bg-purple-600 hover:bg-purple-700 self-end"
                  disabled={!newComment.trim() || postingComment}
                  onClick={handlePostComment}
                >
                  {postingComment ? <Loader2 size={14} className="animate-spin" /> : <Send size={14} />}
                </Button>
              </div>
              {comments.length > 0 && (
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {comments.map((cm) => (
                    <div key={cm.id} className="border-l-2 border-purple-200 pl-3 py-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-medium text-purple-600">{cm.user_email}</span>
                        <span className="text-xs text-muted-foreground">
                          {new Date(cm.created_at).toLocaleDateString()}
                        </span>
                      </div>
                      <p className="text-sm text-slate-700 mt-0.5">{cm.content}</p>
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
