"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Loader2, AlertTriangle, CheckCircle2, Clock, XCircle, ChevronDown, ChevronUp } from "lucide-react";
import { Suspense } from "react";

interface TMAStatus {
  candidate_id: number;
  full_name: string;
  status: string;
  major: string | null;
  photo_url: string | null;
  final_score: number | null;
  category: string | null;
  interview_status: string | null;
  interview_score: number | null;
  combined_score: number | null;
  key_strengths: string[];
  red_flags: string[];
  summary: string | null;
  created_at: string;
}

const STATUS_CONFIG: Record<string, { icon: React.ReactNode; label: string; color: string; bg: string }> = {
  pending: {
    icon: <Clock size={20} />,
    label: "Application Received",
    color: "text-amber-600",
    bg: "bg-amber-50 border-amber-200",
  },
  analyzed: {
    icon: <CheckCircle2 size={20} />,
    label: "Under Review",
    color: "text-blue-600",
    bg: "bg-blue-50 border-blue-200",
  },
  shortlisted: {
    icon: <CheckCircle2 size={20} />,
    label: "Shortlisted",
    color: "text-green-600",
    bg: "bg-green-50 border-green-200",
  },
  rejected: {
    icon: <XCircle size={20} />,
    label: "Not Selected",
    color: "text-red-600",
    bg: "bg-red-50 border-red-200",
  },
  waitlisted: {
    icon: <Clock size={20} />,
    label: "Waitlisted",
    color: "text-orange-600",
    bg: "bg-orange-50 border-orange-200",
  },
};

const CATEGORY_COLORS: Record<string, string> = {
  "Strong Recommend": "bg-green-100 text-green-800",
  "Recommend": "bg-blue-100 text-blue-800",
  "Borderline": "bg-yellow-100 text-yellow-800",
  "Not Recommended": "bg-red-100 text-red-800",
};

function ScoreBar({ score, color = "bg-lime-400" }: { score: number; color?: string }) {
  return (
    <div className="w-full bg-gray-200 rounded-full h-2 mt-1">
      <div
        className={`h-2 rounded-full transition-all duration-500 ${color}`}
        style={{ width: `${Math.min(100, Math.max(0, score))}%` }}
      />
    </div>
  );
}

function TMAContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const [data, setData] = useState<TMAStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showDetails, setShowDetails] = useState(false);

  const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

  useEffect(() => {
    if (!token) {
      setError("No token provided. Please use the link from your Telegram invite.");
      setLoading(false);
      return;
    }
    fetch(`${apiBase}/tma/status?token=${encodeURIComponent(token)}`)
      .then((r) => {
        if (!r.ok) throw new Error(r.status === 404 ? "Invite not found or expired." : "Failed to load status.");
        return r.json();
      })
      .then((d) => { setData(d); setLoading(false); })
      .catch((e) => { setError(e.message); setLoading(false); });
  }, [token, apiBase]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-3 bg-[#fffee9]">
        <Loader2 className="animate-spin text-lime-500" size={32} />
        <p className="text-sm text-gray-500">Loading your application status...</p>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-3 bg-[#fffee9] px-4">
        <AlertTriangle className="text-amber-500" size={36} />
        <p className="text-center text-gray-700 font-medium">{error || "Something went wrong."}</p>
        <p className="text-xs text-gray-400 text-center">If this persists, contact inVision U admissions.</p>
      </div>
    );
  }

  const statusCfg = STATUS_CONFIG[data.status] || STATUS_CONFIG["pending"];

  return (
    <div className="min-h-screen bg-[#fffee9] pb-8">
      {/* Header */}
      <div className="bg-white border-b px-4 py-4 flex items-center gap-3">
        <div className="w-9 h-9 rounded-lg flex items-center justify-center font-bold text-sm text-gray-900" style={{ backgroundColor: "#c1f11d" }}>
          iU
        </div>
        <div>
          <p className="font-semibold text-sm text-gray-900">inVision University</p>
          <p className="text-xs text-gray-500">Admissions Portal</p>
        </div>
      </div>

      <div className="px-4 pt-5 space-y-4 max-w-md mx-auto">
        {/* Candidate card */}
        <div className="bg-white rounded-2xl shadow-sm border p-4 flex items-center gap-3">
          {data.photo_url ? (
            <img
              src={`${apiBase.replace("/api", "")}${data.photo_url}`}
              alt={data.full_name}
              className="w-14 h-14 rounded-full object-cover border-2 border-lime-300"
            />
          ) : (
            <div className="w-14 h-14 rounded-full flex items-center justify-center text-xl font-bold text-gray-900 border-2 border-lime-300 flex-shrink-0" style={{ backgroundColor: "#c1f11d" }}>
              {data.full_name.charAt(0).toUpperCase()}
            </div>
          )}
          <div className="min-w-0 flex-1">
            <p className="font-bold text-gray-900 truncate">{data.full_name}</p>
            {data.major && (
              <p className="text-xs text-gray-500 mt-0.5">{data.major}</p>
            )}
            <p className="text-xs text-gray-400">Application #{data.candidate_id}</p>
          </div>
        </div>

        {/* Status badge */}
        <div className={`rounded-2xl border p-4 flex items-center gap-3 ${statusCfg.bg}`}>
          <span className={statusCfg.color}>{statusCfg.icon}</span>
          <div>
            <p className={`font-semibold text-sm ${statusCfg.color}`}>{statusCfg.label}</p>
            <p className="text-xs text-gray-500 mt-0.5">
              {data.status === "pending" && "Your application is being reviewed by our admissions team."}
              {data.status === "analyzed" && "AI screening complete — your application is being reviewed."}
              {data.status === "shortlisted" && "Congratulations! You have been shortlisted for interview."}
              {data.status === "rejected" && "Thank you for applying. Unfortunately we cannot proceed at this time."}
              {data.status === "waitlisted" && "You are on our waitlist. We will contact you if a spot opens."}
            </p>
          </div>
        </div>

        {/* Application score card — only if analyzed */}
        {data.final_score != null && (
          <div className="bg-white rounded-2xl shadow-sm border p-4">
            <div className="flex items-center justify-between mb-3">
              <p className="font-semibold text-sm text-gray-900">Application Score</p>
              {data.category && (
                <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${CATEGORY_COLORS[data.category] || "bg-gray-100 text-gray-700"}`}>
                  {data.category}
                </span>
              )}
            </div>
            <div className="flex items-end gap-2 mb-3">
              <span className="text-4xl font-bold text-gray-900">{data.final_score.toFixed(1)}</span>
              <span className="text-sm text-gray-400 mb-1">/ 100</span>
            </div>
            <ScoreBar score={data.final_score} />

            {data.summary && (
              <div className="mt-3 pt-3 border-t">
                <p className="text-xs text-gray-600 leading-relaxed">{data.summary}</p>
              </div>
            )}
          </div>
        )}

        {/* Combined score — if interview done */}
        {data.combined_score != null && (
          <div className="rounded-2xl border p-4" style={{ background: "linear-gradient(135deg, #c1f11d22, #9dd90d33)" }}>
            <p className="text-xs text-gray-600 font-medium uppercase mb-1">Combined Score</p>
            <p className="text-3xl font-bold text-gray-900">{Number(data.combined_score).toFixed(1)}</p>
            <p className="text-xs text-gray-500 mt-1">60% application + 40% interview</p>
          </div>
        )}

        {/* Interview status */}
        {data.interview_status && data.interview_status !== "not_started" && (
          <div className="bg-white rounded-2xl shadow-sm border p-4">
            <p className="font-semibold text-sm text-gray-900 mb-2">Interview Stage</p>
            <div className="flex items-center gap-2">
              <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                data.interview_status === "completed" ? "bg-green-100 text-green-700"
                : data.interview_status === "active" ? "bg-blue-100 text-blue-700"
                : "bg-gray-100 text-gray-600"
              }`}>
                {data.interview_status === "completed" ? "Completed" : data.interview_status === "active" ? "In Progress" : data.interview_status}
              </span>
              {data.interview_score != null && (
                <span className="text-sm font-bold text-gray-900">{data.interview_score.toFixed(1)} pts</span>
              )}
            </div>
          </div>
        )}

        {/* Strengths & flags — collapsible */}
        {((data.key_strengths && data.key_strengths.length > 0) || (data.red_flags && data.red_flags.length > 0)) && (
          <div className="bg-white rounded-2xl shadow-sm border overflow-hidden">
            <button
              className="w-full flex items-center justify-between px-4 py-3 text-sm font-semibold text-gray-900"
              onClick={() => setShowDetails(!showDetails)}
            >
              Feedback Details
              {showDetails ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
            </button>
            {showDetails && (
              <div className="px-4 pb-4 space-y-3 border-t pt-3">
                {data.key_strengths && data.key_strengths.length > 0 && (
                  <div>
                    <p className="text-xs font-semibold text-green-700 mb-1.5">Strengths</p>
                    <ul className="space-y-1">
                      {data.key_strengths.map((s, i) => (
                        <li key={i} className="text-xs text-gray-700 flex items-start gap-1.5">
                          <span className="text-green-500 mt-0.5 flex-shrink-0">✓</span>
                          {s}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
                {data.red_flags && data.red_flags.length > 0 && (
                  <div>
                    <p className="text-xs font-semibold text-amber-700 mb-1.5">Areas to Strengthen</p>
                    <ul className="space-y-1">
                      {data.red_flags.map((f, i) => (
                        <li key={i} className="text-xs text-gray-700 flex items-start gap-1.5">
                          <span className="text-amber-500 mt-0.5 flex-shrink-0">•</span>
                          {f}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {/* Footer */}
        <p className="text-center text-xs text-gray-400 pt-2">
          inVision University · Admissions {new Date().getFullYear()}
        </p>
      </div>
    </div>
  );
}

export default function TMAPage() {
  return (
    <Suspense fallback={
      <div className="flex flex-col items-center justify-center min-h-screen gap-3 bg-[#fffee9]">
        <Loader2 className="animate-spin text-lime-500" size={32} />
      </div>
    }>
      <TMAContent />
    </Suspense>
  );
}
