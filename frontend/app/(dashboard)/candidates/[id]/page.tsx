"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { useFetch } from "@/lib/hooks";
import { useI18n } from "@/lib/i18n";
import { useAIProvider } from "@/lib/aiProvider";
import api from "@/lib/api";
import { CandidateDetail, InterviewStatus, InterviewMessage, MajorOption } from "@/lib/types";
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
import { ArrowLeft, Sparkles, Loader2, AlertTriangle, Trash2, Send, MessageSquare, BotMessageSquare, Copy, Check, Mic, ExternalLink, Pencil, UserX, SkipForward, GitCompare, Brain } from "lucide-react";
import { toast } from "sonner";

const categoryColors: Record<string, string> = {
  "Strong Recommend": "bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300",
  "Recommend": "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300",
  "Borderline": "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/40 dark:text-yellow-300",
  "Not Recommended": "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300",
};

const riskColors: Record<string, string> = {
  low: "bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300",
  medium: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-300",
  high: "bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300",
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

  // Analysis timer
  const [analysisElapsed, setAnalysisElapsed] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Copy state for contact fields
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const copyField = (value: string, field: string) => {
    navigator.clipboard.writeText(value);
    setCopiedField(field);
    toast.success("Copied!");
    setTimeout(() => setCopiedField(null), 2000);
  };

  // Interview (Stage 2)
  const [interviewData, setInterviewData] = useState<InterviewStatus | null>(null);
  const [sendingInvite, setSendingInvite] = useState(false);
  const [inviteLink, setInviteLink] = useState<string | null>(null);
  const [linkCopied, setLinkCopied] = useState(false);
  // effectiveInviteLink: prefer local state (just-created), fall back to status endpoint deep_link
  const effectiveInviteLink = inviteLink || (interviewData?.deep_link as string | undefined) || null;
  const [showTranscript, setShowTranscript] = useState(false);
  const [transcript, setTranscript] = useState<InterviewMessage[]>([]);

  // Majors
  const [majors, setMajors] = useState<MajorOption[]>([]);
  useEffect(() => {
    api.get("/majors").then((res) => setMajors(res.data || [])).catch(() => {});
  }, []);

  // Similar candidates
  const [similarCandidates, setSimilarCandidates] = useState<{ id: number; full_name: string; major: string | null; final_score: number; category: string; status: string }[]>([]);
  useEffect(() => {
    if (detail?.analysis) {
      api.get(`/candidates/${params.id}/similar`).then((res) => setSimilarCandidates(res.data || [])).catch(() => {});
    }
  }, [detail?.analysis, params.id]);

  // Stage 2 expanded criteria
  const [expandedStage2Score, setExpandedStage2Score] = useState<string | null>(null);

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

  const hasRussianContent = (() => {
    const text = (detail?.essay || "") + " " + (detail?.motivation_statement || "") + " " + (detail?.achievements || "");
    if (text.trim().length < 50) return false;
    const cyrillicCount = (text.match(/[а-яёА-ЯЁ]/g) || []).length;
    return cyrillicCount / text.length > 0.25;
  });

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
    const link = inviteLink || (interviewData?.deep_link as string | undefined) || null;
    if (link) {
      navigator.clipboard.writeText(link);
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
    // Start elapsed timer
    setAnalysisElapsed(0);
    timerRef.current = setInterval(() => {
      setAnalysisElapsed((s) => s + 1);
    }, 1000);
    pollRef.current = setInterval(async () => {
      try {
        const res = await api.get(`/candidates/${params.id}/analysis-status`);
        const { running, failed, error } = res.data;
        if (!running) {
          clearInterval(pollRef.current!);
          pollRef.current = null;
          if (timerRef.current) { clearInterval(timerRef.current); timerRef.current = null; }
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
        if (timerRef.current) { clearInterval(timerRef.current); timerRef.current = null; }
        setAnalyzing(false);
      }
    }, 2000);
  };

  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  const handleAnalyze = async (force = false) => {
    setAnalysisFailed(false);
    setAnalysisError(null);
    setAnalyzing(true);
    setAnalysisElapsed(0);
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
      major: detail.major || "",
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
        major: editForm.major || null,
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
      const res = await api.post(`/candidates/${params.id}/telegram-invite?override=true`);
      setInviteLink(res.data.deep_link);
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
          <Button variant="outline" size="sm" className="text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200 dark:hover:bg-red-950" onClick={() => { setDeleteCandidateText(""); setShowDeleteCandidate(true); }}>
            <UserX size={14} className="mr-1" /> Delete Candidate
          </Button>
          {/* Provider toggle */}
          <div className="flex items-center gap-1 bg-muted border border-border rounded-lg p-1">
            <button
              onClick={() => setProvider("gemini")}
              className={`text-xs px-2.5 py-1 rounded-md transition-all duration-200 ${
                provider === "gemini" ? "text-foreground font-semibold shadow-sm" : "text-muted-foreground hover:text-foreground"
              }`}
              style={provider === "gemini" ? { backgroundColor: "#c1f11d", color: "#111" } : undefined}
            >
              ☁ Gemini <span className="text-[10px] opacity-70">speed</span>
            </button>
            <button
              onClick={() => setProvider("ollama")}
              className={`text-xs px-2.5 py-1 rounded-md transition-all duration-200 ${
                provider === "ollama" ? "text-foreground font-semibold shadow-sm" : "text-muted-foreground hover:text-foreground"
              }`}
              style={provider === "ollama" ? { backgroundColor: "#c1f11d", color: "#111" } : undefined}
            >
              ⚙ Ollama <span className="text-[10px] opacity-70">privacy</span>
            </button>
          </div>
          {a ? (
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => handleAnalyze(true)} disabled={analyzing || deletingAnalysis}>
                {analyzing ? <><Loader2 className="animate-spin mr-2" size={14} /> {t("detail.re_analyzing")}</> : t("detail.reanalyze")}
              </Button>
              <Button variant="outline" size="sm" className="text-orange-600 hover:text-orange-700 hover:bg-orange-50 border-orange-200 dark:hover:bg-orange-950" onClick={() => setShowDeleteAnalysis(true)} disabled={deletingAnalysis}>
                {deletingAnalysis ? <Loader2 className="animate-spin mr-1" size={14} /> : <Trash2 size={14} className="mr-1" />}
                <span className="text-xs">Del Analysis</span>
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
              <div className="flex gap-4 mb-4">
                {detail.photo_url ? (
                  <div className="flex-shrink-0">
                    <img
                      src={`${process.env.NEXT_PUBLIC_API_URL?.replace("/api", "") || "http://localhost:8080"}${detail.photo_url}`}
                      alt={detail.full_name}
                      className="w-20 h-20 rounded-lg object-cover border"
                    />
                    {detail.photo_ai_flag && (
                      <div className="flex items-center gap-1 mt-1.5 text-amber-600 bg-amber-50 border border-amber-200 rounded px-2 py-1">
                        <AlertTriangle size={11} />
                        <span className="text-xs font-medium">AI-generated photo</span>
                      </div>
                    )}
                    {detail.photo_ai_note && (
                      <p className="text-xs text-muted-foreground mt-1 max-w-[80px] truncate" title={detail.photo_ai_note}>
                        {detail.photo_ai_note}
                      </p>
                    )}
                  </div>
                ) : (
                  <div className="w-20 h-20 rounded-lg bg-slate-100 border flex items-center justify-center flex-shrink-0">
                    <span className="text-2xl font-bold text-slate-400">{detail.full_name.charAt(0).toUpperCase()}</span>
                  </div>
                )}
                <div className="flex-1 min-w-0">
                  {detail.major && (
                    <div className="mb-2">
                      <span className="inline-block bg-lime-100 text-lime-800 text-xs font-semibold px-2 py-0.5 rounded">
                        {majors.find((m) => m.tag === detail.major)?.en || detail.major}
                      </span>
                    </div>
                  )}
                  <div className="text-sm text-muted-foreground">{detail.email}</div>
                  {detail.phone && <div className="text-sm text-muted-foreground">{detail.phone}</div>}
                  {detail.telegram && <div className="text-sm text-muted-foreground">{detail.telegram}</div>}
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div className="flex items-center gap-1.5">
                  <span className="text-muted-foreground">{t("detail.email")}:</span>
                  <span className="truncate">{detail.email}</span>
                  <button onClick={() => copyField(detail.email, "email")} className="text-muted-foreground hover:text-foreground shrink-0" title="Copy email">
                    {copiedField === "email" ? <Check size={12} className="text-green-500" /> : <Copy size={12} />}
                  </button>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="text-muted-foreground">{t("detail.phone")}:</span>
                  <span>{detail.phone || "—"}</span>
                  {detail.phone && (
                    <button onClick={() => copyField(detail.phone!, "phone")} className="text-muted-foreground hover:text-foreground shrink-0" title="Copy phone">
                      {copiedField === "phone" ? <Check size={12} className="text-green-500" /> : <Copy size={12} />}
                    </button>
                  )}
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="text-muted-foreground">{t("detail.telegram")}:</span>
                  <span>{detail.telegram || "—"}</span>
                  {detail.telegram && (
                    <button onClick={() => copyField(detail.telegram!, "telegram")} className="text-muted-foreground hover:text-foreground shrink-0" title="Copy Telegram">
                      {copiedField === "telegram" ? <Check size={12} className="text-green-500" /> : <Copy size={12} />}
                    </button>
                  )}
                </div>
                <div><span className="text-muted-foreground">{t("detail.age")}:</span> {detail.age || "—"}</div>
                <div><span className="text-muted-foreground">{t("detail.city")}:</span> {detail.city || "—"}</div>
                <div><span className="text-muted-foreground">{t("detail.school")}:</span> {detail.school || "—"}</div>
                <div><span className="text-muted-foreground">{t("detail.graduation")}:</span> {detail.graduation_year || "—"}</div>
                {detail.major && (
                  <div><span className="text-muted-foreground">Major:</span> {majors.find((m) => m.tag === detail.major)?.en || detail.major}</div>
                )}
              </div>
              {/* Recommended major from analysis */}
              {detail.analysis && detail.major && (
                <div className="mt-3 pt-3 border-t flex items-start gap-2 text-sm bg-lime-50 dark:bg-lime-950/30 rounded-lg px-3 py-2">
                  <Brain size={14} className="text-lime-600 dark:text-lime-400 mt-0.5 shrink-0" />
                  <div>
                    <span className="font-medium text-lime-800 dark:text-lime-300">Recommended Major: </span>
                    <span className="text-lime-700 dark:text-lime-400">{majors.find((m) => m.tag === detail.major)?.en || detail.major}</span>
                    <p className="text-xs text-lime-600 dark:text-lime-500 mt-0.5">Based on essay themes and identified strengths</p>
                  </div>
                </div>
              )}
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
        <div className="space-y-4 min-w-0">
          {a ? (
            <>
              {/* Score overview */}
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-base flex items-center gap-2">
                      <Sparkles size={16} className="text-purple-500" /> {t("detail.ai_analysis")}
                    </CardTitle>
                    <div className="flex items-center gap-2 flex-wrap">
                      <Badge className={riskColors[a.ai_generated_risk]}>
                        {t("detail.ai_risk")}: {a.ai_generated_score}%
                      </Badge>
                      {hasRussianContent() && (
                        <Badge className="bg-orange-100 text-orange-700 border border-orange-300 dark:bg-orange-900/40 dark:text-orange-300 dark:border-orange-700">
                          ⚠ Answers not in English
                        </Badge>
                      )}
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
                  <p className="text-sm leading-relaxed text-foreground">{a.summary}</p>
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
                          <p className="text-xs text-muted-foreground mt-1 pl-2 border-l-2 border-purple-200 dark:border-purple-700">
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
                <p className="text-2xl font-mono font-bold text-purple-500 mt-3">
                  {String(Math.floor(analysisElapsed / 60)).padStart(2, "0")}:{String(analysisElapsed % 60).padStart(2, "0")}
                </p>
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

          {/* Interview — Stage 2 */}
          <Card className="overflow-hidden">
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
                      <p className="text-sm text-muted-foreground">{t("interview.eligible")}</p>
                      
                      {effectiveInviteLink ? (
                        <div className="flex items-center gap-2 min-w-0">
                          <code className="text-xs bg-muted px-2 py-1 rounded flex-1 min-w-0 overflow-x-auto whitespace-nowrap text-foreground">
                            {effectiveInviteLink}
                          </code>
                          <Button size="sm" variant="outline" className="shrink-0" onClick={handleCopyLink}>
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
                      {effectiveInviteLink ? (
                        <div className="flex items-center gap-2 w-full">
                          <div className="flex-1 min-w-0 overflow-hidden rounded bg-muted">
                            <code className="text-xs px-2 py-1 block w-full overflow-x-auto whitespace-nowrap text-foreground">
                              {effectiveInviteLink}
                            </code>
                          </div>
                          <Button size="sm" variant="outline" className="shrink-0" onClick={handleCopyLink}>
                            {linkCopied ? <Check size={14} className="text-green-500" /> : <Copy size={14} />}
                          </Button>
                        </div>
                      ) : (
                        <Button size="sm" variant="outline" onClick={handleSkipToStage2}>
                          <SkipForward size={14} className="mr-1" /> Override: Send to Stage 2
                        </Button>
                      )}
                    </div>
                  )}
                </div>
              ) : (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{interviewData.interview?.status || interviewData.status}</Badge>
                      {interviewData.interview && (
                        <span className="text-xs text-muted-foreground">
                          {interviewData.interview.questions_asked} questions asked
                        </span>
                      )}
                    </div>
                    <Button size="sm" variant="ghost" className="text-xs text-muted-foreground h-7 px-2" onClick={() => { fetchInterview(); if (showTranscript) handleViewTranscript(); }}>
                      ↻ Refresh
                    </Button>
                  </div>

                  {/* Evaluating / needs evaluation state */}
                  {!interviewData.analysis && interviewData.interview && (() => {
                    const isTrulyActive = interviewData.interview!.status === "active"
                      && interviewData.interview!.current_topic !== "evaluating"
                      && interviewData.interview!.current_topic !== "closing";
                    return (
                    <div className="flex items-center justify-between gap-2 text-sm text-purple-600 dark:text-purple-400 bg-purple-50 dark:bg-purple-900/30 rounded-lg p-3">
                      <div className="flex items-center gap-2">
                        {isTrulyActive ? (
                          <>
                            <Loader2 size={14} className="animate-spin" />
                            <span>Interview in progress — AI analysis will appear here once complete.</span>
                          </>
                        ) : (
                          <span>Interview finished but not yet evaluated.</span>
                        )}
                      </div>
                      {!isTrulyActive && (
                        <Button
                          size="sm"
                          variant="outline"
                          className="shrink-0 border-purple-300 text-purple-700 hover:bg-purple-100 dark:border-purple-600 dark:text-purple-300 dark:hover:bg-purple-900/50"
                          onClick={async () => {
                            try {
                              await api.post(`/candidates/${params.id}/interview/re-evaluate`);
                              toast.success("Evaluation started — this may take up to a minute");
                              // Poll for result every 5s, up to 2 min
                              let attempts = 0;
                              const poll = setInterval(async () => {
                                attempts++;
                                try {
                                  const res = await api.get(`/candidates/${params.id}/interview`);
                                  if (res.data.analysis) {
                                    clearInterval(poll);
                                    setInterviewData(res.data);
                                    toast.success("Interview evaluation complete!");
                                  } else if (attempts >= 24) {
                                    clearInterval(poll);
                                    toast.error("Evaluation timed out — try refreshing");
                                  }
                                } catch { /* ignore poll errors */ }
                              }, 5000);
                            } catch {
                              toast.error("Failed to start evaluation");
                            }
                          }}
                        >
                          Run AI Evaluation
                        </Button>
                      )}
                    </div>
                    );
                  })()}

                  {/* Interview AI analysis */}
                  {interviewData.analysis && (
                    <div className="space-y-3 bg-muted/50 rounded-lg p-3">
                      {/* Score grid with collapsible explanations */}
                      <div className="space-y-2">
                        {[
                          { key: "leadership", label: "Leadership", score: interviewData.analysis.score_leadership, explanation: interviewData.analysis.explanation_leadership },
                          { key: "grit", label: "Grit", score: interviewData.analysis.score_grit, explanation: interviewData.analysis.explanation_grit },
                          { key: "authenticity", label: "Authenticity", score: interviewData.analysis.score_authenticity, explanation: interviewData.analysis.explanation_authenticity },
                          { key: "motivation", label: "Motivation", score: interviewData.analysis.score_motivation, explanation: interviewData.analysis.explanation_motivation },
                          { key: "vision", label: "Vision", score: interviewData.analysis.score_vision, explanation: interviewData.analysis.explanation_vision },
                        ].map((s) => {
                          const isExp = expandedStage2Score === s.key;
                          return (
                          <div key={s.key} className="bg-card rounded p-2 border border-border">
                            <button
                              className="w-full text-left"
                              onClick={() => setExpandedStage2Score(isExp ? null : s.key)}
                            >
                              <div className="flex items-center justify-between">
                                <span className="text-xs font-medium text-muted-foreground">{s.label}</span>
                                <span className="text-sm font-bold">{s.score}<span className="text-xs text-muted-foreground font-normal">/100</span></span>
                              </div>
                              <div className="w-full bg-muted rounded-full h-1.5 mt-1">
                                <div className="h-1.5 rounded-full bg-purple-500 transition-all" style={{ width: `${s.score}%` }} />
                              </div>
                            </button>
                            {isExp && s.explanation && (
                              <p className="text-xs text-muted-foreground mt-1.5 leading-relaxed pl-2 border-l-2 border-purple-200 dark:border-purple-700">{s.explanation}</p>
                            )}
                          </div>
                          );
                        })}
                      </div>

                      {/* Final score + category */}
                      <div className="flex items-center justify-between pt-2 border-t">
                        <div>
                          <span className="text-sm font-medium">Interview Score: </span>
                          <span className="text-lg font-bold text-purple-700">{interviewData.analysis.final_score}</span>
                        </div>
                        <Badge className={categoryColors[interviewData.analysis.category] || ""}>{interviewData.analysis.category}</Badge>
                      </div>

                      {/* Summary */}
                      {interviewData.analysis.summary && (
                        <p className="text-sm text-foreground">{interviewData.analysis.summary}</p>
                      )}

                      {/* Consistency + style match */}
                      <div className="grid grid-cols-2 gap-2 pt-1">
                        <div className="bg-card rounded p-2 text-center border border-border">
                          <p className="text-xs text-muted-foreground">Essay Consistency</p>
                          <p className="text-base font-bold text-foreground">{interviewData.analysis.consistency_score}</p>
                          <div className="w-full bg-muted rounded-full h-1 mt-1">
                            <div className="h-1 rounded-full bg-blue-400" style={{ width: `${interviewData.analysis.consistency_score}%` }} />
                          </div>
                        </div>
                        <div className="bg-card rounded p-2 text-center border border-border">
                          <p className="text-xs text-muted-foreground">Style Match</p>
                          <p className="text-base font-bold text-foreground">{interviewData.analysis.style_match_score}</p>
                          <div className="w-full bg-muted rounded-full h-1 mt-1">
                            <div className="h-1 rounded-full bg-green-400" style={{ width: `${interviewData.analysis.style_match_score}%` }} />
                          </div>
                        </div>
                      </div>

                      {/* Strengths */}
                      {interviewData.analysis.strengths && interviewData.analysis.strengths.length > 0 && (
                        <div>
                          <p className="text-xs font-semibold text-green-700 mb-1">✓ Strengths</p>
                          <ul className="space-y-1">
                            {interviewData.analysis.strengths.map((s, i) => (
                              <li key={i} className="text-xs text-foreground flex items-start gap-1">
                                <span className="text-green-500 mt-0.5">•</span>{s}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}

                      {/* Concerns */}
                      {interviewData.analysis.concerns && interviewData.analysis.concerns.length > 0 && (
                        <div>
                          <p className="text-xs font-semibold text-amber-700 mb-1">⚠ Concerns</p>
                          <ul className="space-y-1">
                            {interviewData.analysis.concerns.map((c, i) => (
                              <li key={i} className="text-xs text-foreground flex items-start gap-1">
                                <span className="text-amber-500 mt-0.5">•</span>{c}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}

                      {/* Anti-cheat flags */}
                      {interviewData.analysis.suspicion_flags && interviewData.analysis.suspicion_flags.length > 0 && (
                        <div className="pt-1 border-t space-y-1.5">
                          <p className="text-xs text-orange-600 dark:text-orange-400 font-semibold">⚑ Anti-cheat flags detected</p>
                          {interviewData.analysis.suspicion_flags.map((f, i) => {
                            const flagInfo: Record<string, { label: string; reason: string }> = {
                              multiple_fast_responses: {
                                label: "Multiple fast responses",
                                reason: "Candidate answered 50+ character messages in under 5 seconds — indicates likely copy-paste from prepared text.",
                              },
                              style_shift_detected: {
                                label: "Writing style shift",
                                reason: "Significant change in average word length between first and second half of answers — suggests different authorship or assistance mid-interview.",
                              },
                              all_voice: {
                                label: "All voice messages",
                                reason: "Candidate used only voice messages, which reduces verifiability of authentic typed responses.",
                              },
                              no_responses: {
                                label: "No responses recorded",
                                reason: "No candidate messages were logged during the interview session.",
                              },
                            };
                            const info = flagInfo[f] || { label: f.replace(/_/g, " "), reason: "Unusual behavioral pattern detected by the anti-cheat system." };
                            return (
                              <div key={i} className="bg-orange-50 dark:bg-orange-950/30 border border-orange-200 dark:border-orange-800 rounded p-2">
                                <p className="text-xs font-medium text-orange-700 dark:text-orange-300">{info.label}</p>
                                <p className="text-xs text-orange-600 dark:text-orange-400 mt-0.5">{info.reason}</p>
                              </div>
                            );
                          })}
                        </div>
                      )}

                      <div className="flex items-center justify-between pt-1 border-t">
                        <p className="text-xs text-muted-foreground">
                          Analyzed by {interviewData.analysis.model_used} — {new Date(interviewData.analysis.analyzed_at).toLocaleString("en-GB", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" })}
                        </p>
                        <Button
                          size="sm"
                          variant="ghost"
                          className="text-xs h-6 px-2 text-muted-foreground hover:text-foreground"
                          onClick={async () => {
                            try {
                              await api.post(`/candidates/${params.id}/interview/re-evaluate`);
                              toast.success("Re-evaluation started");
                              const poll = setInterval(async () => {
                                const res = await api.get(`/candidates/${params.id}/interview`);
                                if (res.data.analysis?.analyzed_at !== interviewData?.analysis?.analyzed_at) {
                                  clearInterval(poll);
                                  fetchInterview();
                                  toast.success("Re-evaluation complete");
                                }
                              }, 5000);
                              setTimeout(() => clearInterval(poll), 120000);
                            } catch {
                              toast.error("Failed to re-evaluate");
                            }
                          }}
                        >
                          Re-evaluate
                        </Button>
                      </div>
                    </div>
                  )}


                  {/* Transcript button */}
                  <Button size="sm" variant="outline" onClick={handleViewTranscript}>
                    <Mic size={14} className="mr-1" /> {t("interview.view_transcript")}
                  </Button>

                  {showTranscript && transcript.length > 0 && (
                    <div className="max-h-64 overflow-y-auto space-y-2 border border-border rounded-lg p-3 bg-card">
                      {transcript.map((m, i) => (
                        <div key={i} className={`text-sm ${m.role === "bot" ? "text-blue-600 dark:text-blue-400" : "text-foreground"}`}>
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
          {/* Decision panel */}
          <Card>
            <CardHeader><CardTitle className="text-base">{t("detail.committee")}</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <DecisionButtons
                candidateId={detail.id}
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

      {/* Similar candidates */}
      {similarCandidates.length > 0 && (
        <Card className="mt-4">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-base">Similar Candidates (±3 pts)</CardTitle>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  const ids = [detail.id, ...similarCandidates.map((sc) => sc.id)].join(",");
                  router.push(`/compare?ids=${ids}`);
                }}
              >
                <GitCompare size={14} className="mr-1" /> Compare All
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {similarCandidates.map((sc) => (
                <button
                  key={sc.id}
                  onClick={() => router.push(`/candidates/${sc.id}`)}
                  className="flex items-center justify-between p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors text-left"
                >
                  <div>
                    <p className="text-sm font-medium text-foreground">{sc.full_name}</p>
                    {sc.major && (
                      <p className="text-xs text-muted-foreground">{majors.find((m) => m.tag === sc.major)?.en || sc.major}</p>
                    )}
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-bold text-purple-600 dark:text-purple-400">{sc.final_score.toFixed(1)}</p>
                    <Badge className={`text-[10px] ${categoryColors[sc.category] || ""}`}>{sc.category}</Badge>
                  </div>
                </button>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

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
          <div className="col-span-2">
            <Label className="text-sm">Major / Program</Label>
            <select
              className="w-full mt-1 h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              value={editForm.major || ""}
              onChange={(e) => setEditForm({ ...editForm, major: e.target.value })}
            >
              <option value="">— Not specified —</option>
              {majors.map((m) => (
                <option key={m.tag} value={m.tag}>{m.tag} — {m.en}</option>
              ))}
            </select>
          </div>
          <div className="col-span-2">
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
