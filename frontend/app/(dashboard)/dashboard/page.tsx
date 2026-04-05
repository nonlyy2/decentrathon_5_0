"use client";

import { useState, useEffect, useCallback } from "react";
import { useFetch } from "@/lib/hooks";
import api from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { DashboardStats } from "@/lib/types";
import { Users, CheckCircle, Star, Clock, TrendingUp, XCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend,
  BarChart, Bar, XAxis, YAxis, CartesianGrid,
} from "recharts";

const STATUS_COLORS: Record<string, string> = {
  Pending: "#94a3b8",
  Analyzed: "#3b82f6",
  Shortlisted: "#22c55e",
  Waitlisted: "#eab308",
  Rejected: "#ef4444",
};

const SCORE_COLORS: Record<string, string> = {
  "0-49": "#ef4444",
  "50-64": "#eab308",
  "65-79": "#3b82f6",
  "80-100": "#22c55e",
};

const CATEGORY_KEYS: Record<string, string> = {
  "Strong Recommend": "cat.strong_recommend",
  Recommend: "cat.recommend",
  Borderline: "cat.borderline",
  "Not Recommended": "cat.not_recommended",
};

const CATEGORY_COLORS: Record<string, string> = {
  "Strong Recommend": "#22c55e",
  Recommend: "#3b82f6",
  Borderline: "#eab308",
  "Not Recommended": "#ef4444",
};

const DIMENSION_OPTIONS = [
  { key: "leadership", label: "Leadership" },
  { key: "motivation", label: "Motivation" },
  { key: "growth", label: "Growth" },
  { key: "vision", label: "Vision" },
  { key: "communication", label: "Communication" },
];

export default function DashboardPage() {
  const { data: stats, loading } = useFetch<DashboardStats>("/stats");
  const { t } = useI18n();
  const [selectedDim, setSelectedDim] = useState("leadership");
  const [dimTopNInput, setDimTopNInput] = useState<string>("");
  const [dimTopN, setDimTopN] = useState<number>(0); // 0 = all
  const [dimData, setDimData] = useState<{ distributions: Record<string, { range: string; count: number }[]>; means: Record<string, number> } | null>(null);

  const fetchDimData = useCallback(async (topN: number) => {
    try {
      const url = topN > 0 ? `/stats?top_n=${topN}` : "/stats";
      const res = await api.get(url);
      setDimData({
        distributions: res.data.dimension_distributions,
        means: res.data.dimension_means,
      });
    } catch {}
  }, []);

  useEffect(() => {
    fetchDimData(dimTopN);
  }, [dimTopN, fetchDimData]);

  const handleTopNApply = () => {
    const n = parseInt(dimTopNInput);
    if (dimTopNInput === "" || dimTopNInput === "0") {
      setDimTopN(0);
    } else if (!isNaN(n) && n > 0) {
      setDimTopN(n);
    }
  };

  if (loading || !stats) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">{t("dash.title")}</h1>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Card key={i}><CardContent className="p-4"><div className="h-16 bg-muted rounded animate-pulse" /></CardContent></Card>
          ))}
        </div>
      </div>
    );
  }

  const pureAnalyzed = stats.analyzed - stats.shortlisted - stats.rejected - stats.waitlisted;

  const statusData = [
    { name: t("dash.pending"), value: stats.pending, key: "Pending" },
    { name: t("dash.analyzed"), value: pureAnalyzed, key: "Analyzed" },
    { name: t("dash.shortlisted"), value: stats.shortlisted, key: "Shortlisted" },
    { name: t("dash.waitlisted"), value: stats.waitlisted, key: "Waitlisted" },
    { name: t("dash.rejected"), value: stats.rejected, key: "Rejected" },
  ].filter((d) => d.value > 0);

  const categoryData = ["Strong Recommend", "Recommend", "Borderline", "Not Recommended"].map((cat) => ({
    name: t(CATEGORY_KEYS[cat]),
    count: stats.category_counts[cat] || 0,
    color: CATEGORY_COLORS[cat],
  }));

  const conversionRate = stats.total_candidates > 0
    ? ((stats.shortlisted / stats.total_candidates) * 100).toFixed(1)
    : "0";

  const analysisRate = stats.total_candidates > 0
    ? ((stats.analyzed / stats.total_candidates) * 100).toFixed(1)
    : "0";

  const statCards = [
    { label: t("dash.total"), value: stats.total_candidates, icon: <Users size={18} />, color: "text-blue-600 bg-blue-50 dark:bg-blue-900/30" },
    { label: t("dash.pending"), value: stats.pending, icon: <Clock size={18} />, color: "text-slate-600 bg-slate-50 dark:bg-slate-800" },
    { label: t("dash.analyzed"), value: stats.analyzed, icon: <CheckCircle size={18} />, color: "text-blue-600 bg-blue-50 dark:bg-blue-900/30" },
    { label: t("dash.shortlisted"), value: stats.shortlisted, icon: <Star size={18} />, color: "text-green-600 bg-green-50 dark:bg-green-900/30" },
    { label: t("dash.waitlisted"), value: stats.waitlisted, icon: <TrendingUp size={18} />, color: "text-yellow-600 bg-yellow-50 dark:bg-yellow-900/30" },
    { label: t("dash.rejected"), value: stats.rejected, icon: <XCircle size={18} />, color: "text-red-600 bg-red-50 dark:bg-red-900/30" },
  ];

  const activeDimDistributions = dimData?.distributions || stats.dimension_distributions;
  const activeDimMeans = dimData?.means || stats.dimension_means;
  const dimDist = activeDimDistributions?.[selectedDim] || [];
  const dimMean = activeDimMeans?.[selectedDim] || 0;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">{t("dash.title")}</h1>

      {/* Stat cards */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
        {statCards.map((s) => (
          <Card key={s.label}>
            <CardContent className="p-4 flex items-center gap-3">
              <div className={`p-2 rounded-lg ${s.color}`}>{s.icon}</div>
              <div>
                <div className="text-2xl font-bold text-foreground">{s.value}</div>
                <div className="text-xs text-muted-foreground">{s.label}</div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* KPI metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-5 text-center">
            <div className="text-4xl font-bold text-purple-600 dark:text-purple-400">
              {stats.avg_score > 0 ? stats.avg_score.toFixed(1) : "\u2014"}
            </div>
            <div className="text-sm text-muted-foreground mt-1">{t("dash.avg_score")}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-5 text-center">
            <div className="text-4xl font-bold text-indigo-600 dark:text-indigo-400">
              {stats.score_median > 0 ? stats.score_median.toFixed(1) : "\u2014"}
            </div>
            <div className="text-sm text-muted-foreground mt-1">Median AI Score</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-5 text-center">
            <div className="text-4xl font-bold text-blue-600 dark:text-blue-400">{analysisRate}%</div>
            <div className="text-sm text-muted-foreground mt-1">{t("dash.analysis_rate")}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-5 text-center">
            <div className="text-4xl font-bold text-green-600 dark:text-green-400">{conversionRate}%</div>
            <div className="text-sm text-muted-foreground mt-1">{t("dash.shortlist_rate")}</div>
          </CardContent>
        </Card>
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Status distribution pie */}
        <Card>
          <CardHeader><CardTitle className="text-base">{t("dash.status_dist")}</CardTitle></CardHeader>
          <CardContent>
            {statusData.length > 0 ? (
              <ResponsiveContainer width="100%" height={260}>
                <PieChart>
                  <Pie
                    data={statusData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    paddingAngle={3}
                    dataKey="value"
                    label={false}
                    labelLine={false}
                  >
                    {statusData.map((entry) => (
                      <Cell key={entry.key} fill={STATUS_COLORS[entry.key] || "#8884d8"} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">{t("dash.no_data")}</div>
            )}
          </CardContent>
        </Card>

        {/* Score distribution bar */}
        <Card>
          <CardHeader><CardTitle className="text-base">{t("dash.score_dist")}</CardTitle></CardHeader>
          <CardContent>
            {stats.score_distribution.some((b) => b.count > 0) ? (
              <ResponsiveContainer width="100%" height={260}>
                <BarChart data={stats.score_distribution} barSize={48}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="range" />
                  <YAxis allowDecimals={false} />
                  <Tooltip />
                  <Bar dataKey="count" radius={[4, 4, 0, 0]}>
                    {stats.score_distribution.map((entry) => (
                      <Cell key={entry.range} fill={SCORE_COLORS[entry.range] || "#8884d8"} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">{t("dash.no_analysis")}</div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Bottom row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Category breakdown */}
        <Card>
          <CardHeader><CardTitle className="text-base">{t("dash.categories")}</CardTitle></CardHeader>
          <CardContent>
            {categoryData.some((c) => c.count > 0) ? (
              <div className="space-y-3 pt-2">
                {categoryData.map((cat) => {
                  const pct = stats.analyzed > 0 ? (cat.count / stats.analyzed) * 100 : 0;
                  return (
                    <div key={cat.name} className="flex items-center gap-3">
                      <div className="w-36 text-sm truncate text-foreground">{cat.name}</div>
                      <div className="flex-1 bg-muted rounded-full h-5 overflow-hidden">
                        <div
                          className="h-5 rounded-full transition-all"
                          style={{ width: `${pct}%`, backgroundColor: cat.color }}
                        />
                      </div>
                      <div className="w-10 text-sm text-right font-medium text-foreground">{cat.count}</div>
                      <div className="w-12 text-xs text-muted-foreground">{pct.toFixed(0)}%</div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">{t("dash.no_analysis")}</div>
            )}
          </CardContent>
        </Card>

        {/* Admissions funnel */}
        <Card>
          <CardHeader><CardTitle className="text-base">{t("dash.funnel")}</CardTitle></CardHeader>
          <CardContent className="pt-2">
            <div className="space-y-3">
              {[
                { label: t("funnel.applied"), count: stats.total_candidates, color: "bg-slate-400" },
                { label: t("funnel.ai_analyzed"), count: stats.analyzed, color: "bg-blue-500" },
                { label: t("funnel.shortlisted"), count: stats.shortlisted, color: "bg-green-500" },
                { label: t("funnel.waitlisted"), count: stats.waitlisted, color: "bg-yellow-500" },
                { label: t("funnel.rejected"), count: stats.rejected, color: "bg-red-500" },
              ].map((step) => {
                const pct = stats.total_candidates > 0 ? (step.count / stats.total_candidates) * 100 : 0;
                return (
                  <div key={step.label} className="flex items-center gap-3">
                    <div className="w-28 text-sm text-foreground">{step.label}</div>
                    <div className="flex-1 bg-muted rounded-full h-5 overflow-hidden">
                      <div
                        className={`h-5 rounded-full transition-all ${step.color}`}
                        style={{ width: `${pct}%` }}
                      />
                    </div>
                    <div className="w-8 text-sm text-right font-medium text-foreground">{step.count}</div>
                    <div className="w-12 text-xs text-muted-foreground">{pct.toFixed(0)}%</div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Dimension Distribution */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Overall Score Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            {stats.score_distribution.some((b) => b.count > 0) ? (
              <>
                <ResponsiveContainer width="100%" height={260}>
                  <BarChart data={stats.score_distribution} barSize={48}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="range" />
                    <YAxis allowDecimals={false} />
                    <Tooltip />
                    <Bar dataKey="count" fill="#7c3aed" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
                <div className="flex items-center justify-center gap-6 mt-2 text-sm">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-blue-500" />
                    <span className="text-muted-foreground">Mean: <span className="font-bold text-blue-600 dark:text-blue-400">{stats.score_mean.toFixed(1)}</span></span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-green-500" />
                    <span className="text-muted-foreground">Median: <span className="font-bold text-green-600 dark:text-green-400">{stats.score_median.toFixed(1)}</span></span>
                  </div>
                </div>
              </>
            ) : (
              <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">No data</div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between flex-wrap gap-2">
              <div className="flex items-center gap-2">
                <CardTitle className="text-base">Dimension Distribution</CardTitle>
                <div className="flex items-center gap-1 text-xs">
                  <span className="text-muted-foreground">Top</span>
                  <input
                    type="number"
                    min={0}
                    placeholder={String(stats.total_candidates)}
                    value={dimTopNInput}
                    onChange={(e) => setDimTopNInput(e.target.value)}
                    onKeyDown={(e) => { if (e.key === "Enter") handleTopNApply(); }}
                    onBlur={handleTopNApply}
                    className="bg-muted border border-border rounded px-2 py-0.5 text-xs w-16 text-center"
                  />
                  <span className="text-muted-foreground">of {stats.total_candidates}</span>
                </div>
              </div>
              <div className="flex gap-1 bg-muted rounded-lg p-0.5">
                {DIMENSION_OPTIONS.map((dim) => (
                  <button
                    key={dim.key}
                    onClick={() => setSelectedDim(dim.key)}
                    className={`text-xs px-2 py-1 rounded-md transition-all duration-200 ${
                      selectedDim === dim.key
                        ? "text-foreground font-semibold shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                    }`}
                    style={selectedDim === dim.key ? { backgroundColor: "#c1f11d", color: "#111" } : undefined}
                  >
                    {dim.label}
                  </button>
                ))}
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {dimDist.length > 0 && dimDist.some((b) => b.count > 0) ? (
              <>
                <ResponsiveContainer width="100%" height={260}>
                  <BarChart data={dimDist} barSize={32}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="range" fontSize={11} />
                    <YAxis allowDecimals={false} />
                    <Tooltip />
                    <Bar dataKey="count" fill="#7c3aed" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
                <div className="flex items-center justify-center gap-6 mt-2 text-sm">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-blue-500" />
                    <span className="text-muted-foreground">Mean: <span className="font-bold text-blue-600 dark:text-blue-400">{dimMean.toFixed(1)}</span></span>
                  </div>
                </div>
              </>
            ) : (
              <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">No data</div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Exam Score Distributions */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* IELTS */}
        <Card>
          <CardHeader><CardTitle className="text-base">IELTS Distribution ({stats.ielts_count})</CardTitle></CardHeader>
          <CardContent>
            {stats.ielts_distribution?.some((b) => b.count > 0) ? (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={stats.ielts_distribution} barSize={24}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="range" fontSize={10} />
                  <YAxis allowDecimals={false} />
                  <Tooltip />
                  <Bar dataKey="count" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">No data</div>
            )}
          </CardContent>
        </Card>

        {/* TOEFL */}
        <Card>
          <CardHeader><CardTitle className="text-base">TOEFL Distribution ({stats.toefl_count})</CardTitle></CardHeader>
          <CardContent>
            {stats.toefl_distribution?.some((b) => b.count > 0) ? (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={stats.toefl_distribution} barSize={24}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="range" fontSize={10} />
                  <YAxis allowDecimals={false} />
                  <Tooltip />
                  <Bar dataKey="count" fill="#22c55e" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">No data</div>
            )}
          </CardContent>
        </Card>

        {/* UNT */}
        <Card>
          <CardHeader><CardTitle className="text-base">UNT Score Distribution ({stats.unt_count})</CardTitle></CardHeader>
          <CardContent>
            {stats.unt_distribution?.some((b) => b.count > 0) ? (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={stats.unt_distribution} barSize={24}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="range" fontSize={10} />
                  <YAxis allowDecimals={false} />
                  <Tooltip />
                  <Bar dataKey="count" fill="#eab308" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">No data</div>
            )}
          </CardContent>
        </Card>

        {/* NIS Grade */}
        <Card>
          <CardHeader><CardTitle className="text-base">NIS 12 Grade Distribution ({stats.nis_count})</CardTitle></CardHeader>
          <CardContent>
            {stats.nis_distribution?.some((b) => b.count > 0) ? (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={stats.nis_distribution} barSize={24}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="range" fontSize={10} />
                  <YAxis allowDecimals={false} />
                  <Tooltip />
                  <Bar dataKey="count" fill="#ef4444" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">No data</div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
