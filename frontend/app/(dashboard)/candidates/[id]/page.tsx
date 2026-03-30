"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { useFetch } from "@/lib/hooks";
import { useI18n } from "@/lib/i18n";
import { useAIProvider } from "@/lib/aiProvider";
import api from "@/lib/api";
import { CandidateDetail, InterviewStatus, InterviewMessage } from "@/lib/types";
import StatusBadge from "@/components/StatusBadge";
import ScoreRadar from "@/components/ScoreRadar";
import DecisionButtons from "@/components/DecisionButtons";
import KeyStrengthsRedFlags from "@/components/KeyStrengthsRedFlags";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { ArrowLeft, Sparkles, Loader2, AlertTriangle, Trash2, Send, MessageSquare, BotMessageSquare, Copy, Check, Mic, ExternalLink, Pencil, UserX, SkipForward } from "lucide-react";
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

  // Interview (Stage 2)
  const [interviewData, setInterviewData] = useState<InterviewStatus | null>(null);
  const [sendingInvite, setSendingInvite] = useState(false);
  const [inviteLink, setInviteLink] = useState<string | null>(null);
  const [linkCopied, setLinkCopied] = useState(false);
  const [showTranscript, setShowTranscript] = useState(false);
  const [transcript, setTranscript] = useState<InterviewMessage[]>([]);

  // Delete analysis custom modal
  const [showDeleteAnalysis, setShowDeleteAnalysis] = useState(false);
  const [deleteAnalysisText, setDeleteAnalysisText] = useState("");

  // Delete candidate modal
  const [showDeleteCandidate, setShowDeleteCandidate] = useState(false);
  const [deleteCandidateText, setDeleteCandidateText] = useState("");
  const [deletingCandidate, setDeletingCandidate] = useState(false);

  // Edit candidate modal
  const [showEditCandidate, setShowEditCandidate] = useState(false);
  const [editForm, setEditForm] = useState<Record<string, string>>({});
  const [savingEdit, setSavingEdit] = useState(false);

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

  // Fetch interview status
  const fetchInterview = useCallback(() => {
    api.get(`/candidates/${params.id}/interview`).then((res) => setInterviewData(res.data)).catch(() => {});
  }, [params.id]);

  useEffect(() => { fetchInterview(); }, [fetchInterview]);

  const handleSendInvite = async () => {
    setSendingInvite(true);
    try {
      const res = await api.post(`/candidates/${params.id}/telegram-invite`);
      setInviteLink(res.data.deep_link);
      toast.success("Invite created!");
      fetchInterview();
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || "Failed to create invite";
      toast.error(msg);
    } finally {
      setSendingInvite(false);
    }
  };

  const handleCopyLink = () => {
    if (inviteLink) {
      navigator.clipboard.writeText(inviteLink);
      setLinkCopied(true);
      toast.success(t("interview.copied"));
      setTimeout(() => setLinkCopied(false), 2000);
    }
  };

  const handleViewTranscript = async () => {
    try {
      const res = await api.get(`/candidates/${params.id}/interview/messages`);
      setTranscript(res.data);
      setShowTranscript(true);
    } catch {
      toast.error("Failed to load transcript");
    }
  };

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
    setDeletingAnalysis(true);
    try {
      await api.delete(`/candidates/${params.id}/analysis`);
      toast.success("Analysis deleted");
      setAnalysisFailed(false);
      setShowDeleteAnalysis(false);
      setDeleteAnalysisText("");
      refetch();
    } catch {
      toast.error("Failed to delete analysis");
    } finally {
      setDeletingAnalysis(false);
    }
  };

  const handleDeleteCandidate = async () => {
    setDeletingCandidate(true);
    try {
      await api.delete(`/candidates/${params.id}`);
      toast.success("Candidate deleted");
      router.push("/candidates");
    } catch {
      toast.error("Failed to delete candidate");
    } finally {
      setDeletingCandidate(false);
    }
  };

  const openEditCandidate = () => {
    if (!detail) return;
    setEditForm({
      full_name: detail.full_name || "",
      email: detail.email || "",
      phone: detail.phone || "",
      telegram: detail.telegram || "",
      age: detail.age?.toString() || "",
      city: detail.city || "",
      school: detail.school || "",
      graduation_year: detail.graduation_year?.toString() || "",
      achievements: detail.achievements || "",
      extracurriculars: detail.extracurriculars || "",
      essay: detail.essay || "",
      motivation_statement: detail.motivation_statement || "",
    });
    setShowEditCandidate(true);
  };

  const handleSaveEdit = async () => {
    setSavingEdit(true);
    try {
      await api.patch(`/candidates/${params.id}`, {
        full_name: editForm.full_name,
        email: editForm.email,
        phone: editForm.phone || null,
        telegram: editForm.telegram || null,
        age: editForm.age ? parseInt(editForm.age) : null,
        city: editForm.city || null,
        school: editForm.school || null,
        graduation_year: editForm.graduation_year ? parseInt(editForm.graduation_year) : null,
        achievements: editForm.achievements || null,
        extracurriculars: editForm.extracurriculars || null,
        essay: editForm.essay,
        motivation_statement: editForm.motivation_statement || null,
      });
      toast.success("Candidate updated");
      setShowEditCandidate(false);
      refetch();
    } catch {
      toast.error("Failed to update candidate");
    } finally {
      setSavingEdit(false);
    }
  };

  const handleSkipToStage2 = async () => {
    try {
      await api.post(`/candidates/${params.id}/telegram-invite?override=true`);
      toast.success("Candidate forwarded to Stage 2");
      fetchInterview();
      refetch();
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || "Failed to override";
      toast.error(msg);
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
          <Button variant="outline" size="sm" onClick={openEditCandidate}>
            <Pencil size={14} className="mr-1" /> Edit
          </Button>
          <Button variant="outline" size="sm" className="text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200" onClick={() => { setDeleteCandidateText(""); setShowDeleteCandidate(true); }}>
            <UserX size={14} className="mr-1" /> Delete
          </Button>
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
              <Button variant="outline" size="sm" className="text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200" onClick={() => setShowDeleteAnalysis(true)} disabled={deletingAnalysis}>
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
                <div><span className="text-muted-foreground">{t("detail.phone")}:</span> {detail.phone || "—"}</div>
                <div><span className="text-muted-foreground">{t("detail.telegram")}:</span> {detail.telegram || "—"}</div>
                <div><span className="text-muted-foreground">{t("detail.age")}:</span> {detail.age || "—"}</div>
                <div><span className="text-muted-foreground">{t("detail.city")}:</span> {detail.city || "—"}</div>
                <div><span className="text-muted-foreground">{t("detail.school")}:</span> {detail.school || "—"}</div>
                <div><span className="text-muted-foreground">{t("detail.graduation")}:</span> {detail.graduation_year || "—"}</div>
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
                    <div className="flex items-center gap-2">
                      <Badge className={riskColors[a.ai_generated_risk]}>
                        {t("detail.ai_risk")}: {a.ai_generated_score}%
                      </Badge>
                    </div>
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
                    {t("detail.analyzed_by")} {a.model_used} {"\u2014"} {new Date(a.analyzed_at).toLocaleString("en-GB", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit", second: "2-digit" })}
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
                          {d.decided_by_email && <span className="font-medium">{d.decided_by_email} — </span>}
                          {new Date(d.decided_at).toLocaleString("en-GB", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit", second: "2-digit" })}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Interview — Stage 2 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <BotMessageSquare size={16} /> {t("interview.title")}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {!interviewData || interviewData.status === "not_started" ? (
                <div className="space-y-3">
                  {detail.analysis && detail.analysis.final_score >= 65 ? (
                    <>
                      <p className="text-sm text-slate-600">{t("interview.eligible")}</p>
                      {inviteLink ? (
                        <div className="flex items-center gap-2">
                          <code className="text-xs bg-slate-100 px-2 py-1 rounded flex-1 truncate">{inviteLink}</code>
                          <Button size="sm" variant="outline" onClick={handleCopyLink}>
                            {linkCopied ? <Check size={14} className="text-green-500" /> : <Copy size={14} />}
                          </Button>
                        </div>
                      ) : (
                        <Button
                          size="sm"
                          className="bg-blue-600 hover:bg-blue-700"
                          disabled={sendingInvite}
                          onClick={handleSendInvite}
                        >
                          {sendingInvite ? <Loader2 size={14} className="animate-spin mr-1" /> : <ExternalLink size={14} className="mr-1" />}
                          {t("interview.send_invite")}
                        </Button>
                      )}
                      {interviewData?.invite_status && (
                        <p className="text-xs text-muted-foreground">
                          Invite status: <span className="font-medium">{interviewData.invite_status}</span>
                        </p>
                      )}
                    </>
                  ) : (
                    <div className="space-y-2">
                      <p className="text-sm text-slate-500">{t("interview.not_eligible")}</p>
                      <Button size="sm" variant="outline" onClick={handleSkipToStage2}>
                        <SkipForward size={14} className="mr-1" /> Override: Send to Stage 2
                      </Button>
                    </div>
                  )}
                </div>
              ) : (
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">{interviewData.interview?.status || interviewData.status}</Badge>
                    {interviewData.interview && (
                      <span className="text-xs text-muted-foreground">
                        {interviewData.interview.questions_asked} questions asked
                      </span>
                    )}
                  </div>

                  {/* Interview analysis results */}
                  {interviewData.analysis && (
                    <div className="space-y-2 bg-slate-50 rounded-lg p-3">
                      <div className="grid grid-cols-5 gap-2 text-center">
                        {[
                          { label: "Leadership", score: interviewData.analysis.score_leadership },
                          { label: "Grit", score: interviewData.analysis.score_grit },
                          { label: "Authenticity", score: interviewData.analysis.score_authenticity },
                          { label: "Motivation", score: interviewData.analysis.score_motivation },
                          { label: "Vision", score: interviewData.analysis.score_vision },
                        ].map((s) => (
                          <div key={s.label}>
                            <p className="text-xs text-muted-foreground">{s.label}</p>
                            <p className="text-lg font-bold">{s.score}</p>
                          </div>
                        ))}
                      </div>
                      <div className="flex items-center justify-between pt-2 border-t">
                        <div>
                          <span className="text-sm font-medium">Interview Score: </span>
                          <span className="text-lg font-bold text-purple-700">{interviewData.analysis.final_score}</span>
                        </div>
                        <Badge className={categoryColors[interviewData.analysis.category] || ""}>{interviewData.analysis.category}</Badge>
                      </div>
                      {interviewData.analysis.summary && (
                        <p className="text-sm text-slate-600 mt-1">{interviewData.analysis.summary}</p>
                      )}
                      {interviewData.analysis.suspicion_flags && interviewData.analysis.suspicion_flags.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {interviewData.analysis.suspicion_flags.map((f, i) => (
                            <Badge key={i} variant="outline" className="text-xs text-orange-600 border-orange-300">{f}</Badge>
                          ))}
                        </div>
                      )}
                    </div>
                  )}

                  {interviewData.combined_score && (
                    <div className="bg-purple-50 rounded-lg p-3 text-center">
                      <p className="text-xs text-purple-600 uppercase font-medium">{t("interview.combined_score")}</p>
                      <p className="text-2xl font-bold text-purple-800">{Number(interviewData.combined_score).toFixed(1)}</p>
                      <p className="text-xs text-purple-500">60% essay + 40% interview</p>
                    </div>
                  )}

                  {/* Transcript button */}
                  <Button size="sm" variant="outline" onClick={handleViewTranscript}>
                    <Mic size={14} className="mr-1" /> {t("interview.view_transcript")}
                  </Button>

                  {showTranscript && transcript.length > 0 && (
                    <div className="max-h-64 overflow-y-auto space-y-2 border rounded-lg p-3 bg-white">
                      {transcript.map((m, i) => (
                        <div key={i} className={`text-sm ${m.role === "bot" ? "text-blue-700" : "text-slate-800"}`}>
                          <span className="font-medium text-xs uppercase">{m.role === "bot" ? "Interviewer" : "Candidate"}</span>
                          {m.message_type === "voice" && <Mic size={10} className="inline ml-1 text-purple-500" />}
                          <p className="mt-0.5">{m.content}</p>
                        </div>
                      ))}
                    </div>
                  )}
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

      {/* Delete Analysis Modal */}
      <Dialog open={showDeleteAnalysis} onOpenChange={(open) => { setShowDeleteAnalysis(open); if (!open) setDeleteAnalysisText(""); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete AI Analysis</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-slate-600">This will permanently delete the AI analysis for this candidate. Type <span className="font-bold">delete</span> to confirm.</p>
          <Input
            value={deleteAnalysisText}
            onChange={(e) => setDeleteAnalysisText(e.target.value)}
            placeholder='Type "delete" to confirm'
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteAnalysis(false)}>Cancel</Button>
            <Button
              variant="destructive"
              disabled={deleteAnalysisText !== "delete" || deletingAnalysis}
              onClick={handleDeleteAnalysis}
            >
              {deletingAnalysis ? <Loader2 size={14} className="animate-spin mr-1" /> : <Trash2 size={14} className="mr-1" />}
              Delete Analysis
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Candidate Modal */}
      <Dialog open={showDeleteCandidate} onOpenChange={(open) => { setShowDeleteCandidate(open); if (!open) setDeleteCandidateText(""); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Candidate</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-slate-600">This will permanently delete this candidate and all their data (analysis, decisions, comments, interview). Type <span className="font-bold">delete</span> to confirm.</p>
          <Input
            value={deleteCandidateText}
            onChange={(e) => setDeleteCandidateText(e.target.value)}
            placeholder='Type "delete" to confirm'
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteCandidate(false)}>Cancel</Button>
            <Button
              variant="destructive"
              disabled={deleteCandidateText !== "delete" || deletingCandidate}
              onClick={handleDeleteCandidate}
            >
              {deletingCandidate ? <Loader2 size={14} className="animate-spin mr-1" /> : <UserX size={14} className="mr-1" />}
              Delete Candidate
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Candidate Modal */}
      <Dialog open={showEditCandidate} onOpenChange={setShowEditCandidate}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Candidate</DialogTitle>
          </DialogHeader>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label className="text-sm">Full Name *</Label>
              <Input value={editForm.full_name || ""} onChange={(e) => setEditForm({ ...editForm, full_name: e.target.value })} />
            </div>
            <div>
              <Label className="text-sm">Email *</Label>
              <Input value={editForm.email || ""} onChange={(e) => setEditForm({ ...editForm, email: e.target.value })} />
            </div>
            <div>
              <Label className="text-sm">Phone</Label>
              <Input value={editForm.phone || ""} onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })} />
            </div>
            <div>
              <Label className="text-sm">Telegram</Label>
              <Input value={editForm.telegram || ""} onChange={(e) => setEditForm({ ...editForm, telegram: e.target.value })} />
            </div>
            <div>
              <Label className="text-sm">Age</Label>
              <Input type="number" value={editForm.age || ""} onChange={(e) => setEditForm({ ...editForm, age: e.target.value })} />
            </div>
            <div>
              <Label className="text-sm">City</Label>
              <Input value={editForm.city || ""} onChange={(e) => setEditForm({ ...editForm, city: e.target.value })} />
            </div>
            <div>
              <Label className="text-sm">School</Label>
              <Input value={editForm.school || ""} onChange={(e) => setEditForm({ ...editForm, school: e.target.value })} />
            </div>
            <div>
              <Label className="text-sm">Graduation Year</Label>
              <Input type="number" value={editForm.graduation_year || ""} onChange={(e) => setEditForm({ ...editForm, graduation_year: e.target.value })} />
            </div>
          </div>
          <div>
            <Label className="text-sm">Achievements</Label>
            <Textarea rows={2} value={editForm.achievements || ""} onChange={(e) => setEditForm({ ...editForm, achievements: e.target.value })} />
          </div>
          <div>
            <Label className="text-sm">Extracurriculars</Label>
            <Textarea rows={2} value={editForm.extracurriculars || ""} onChange={(e) => setEditForm({ ...editForm, extracurriculars: e.target.value })} />
          </div>
          <div>
            <Label className="text-sm">Essay *</Label>
            <Textarea rows={4} value={editForm.essay || ""} onChange={(e) => setEditForm({ ...editForm, essay: e.target.value })} />
          </div>
          <div>
            <Label className="text-sm">Motivation Statement</Label>
            <Textarea rows={3} value={editForm.motivation_statement || ""} onChange={(e) => setEditForm({ ...editForm, motivation_statement: e.target.value })} />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEditCandidate(false)}>Cancel</Button>
            <Button onClick={handleSaveEdit} disabled={savingEdit || !editForm.full_name || !editForm.email || !editForm.essay}>
              {savingEdit ? <Loader2 size={14} className="animate-spin mr-1" /> : null}
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
