"use client";

import { useState } from "react";
import { useFetch } from "@/lib/hooks";
import { useI18n } from "@/lib/i18n";
import { DashboardStats } from "@/lib/types";
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
  "Recommend": "#3b82f6",
  "Borderline": "#eab308",
  "Not Recommended": "#ef4444",
};

export default function AnalyticsPage() {
  const { data: stats, loading } = useFetch<DashboardStats>("/stats");
  const { t } = useI18n();

  if (loading || !stats) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">{t("analytics.title")}</h1>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="h-48 bg-slate-200 rounded animate-pulse" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  const pureAnalyzed = stats.analyzed - stats.shortlisted - stats.rejected - stats.waitlisted;

  const statusData = [
    { name: t("status.pending"), value: stats.pending, key: "Pending" },
    { name: t("status.analyzed"), value: pureAnalyzed, key: "Analyzed" },
    { name: t("status.shortlisted"), value: stats.shortlisted, key: "Shortlisted" },
    { name: t("status.waitlisted"), value: stats.waitlisted, key: "Waitlisted" },
    { name: t("status.rejected"), value: stats.rejected, key: "Rejected" },
  ].filter((d) => d.value > 0);

  const categoryData = ["Strong Recommend", "Recommend", "Borderline", "Not Recommended"].map((cat) => ({
    name: t(CATEGORY_KEYS[cat]),
    count: stats.category_counts[cat] || 0,
    color: CATEGORY_COLORS[cat],
  }));

  const conversionRate =
    stats.total_candidates > 0
      ? ((stats.shortlisted / stats.total_candidates) * 100).toFixed(1)
      : "0";

  const analysisRate =
    stats.total_candidates > 0
      ? ((stats.analyzed / stats.total_candidates) * 100).toFixed(1)
      : "0";

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t("analytics.title")}</h1>

      {/* KPI row */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: t("analytics.total_apps"), value: stats.total_candidates, sub: t("analytics.all_time") },
          { label: t("dash.analysis_rate"), value: `${analysisRate}%`, sub: `${stats.analyzed} ${t("analytics.analyzed_sub")}` },
          { label: t("dash.shortlist_rate"), value: `${conversionRate}%`, sub: `${stats.shortlisted} ${t("analytics.shortlisted_sub")}` },
          { label: t("dash.avg_score"), value: stats.avg_score > 0 ? stats.avg_score.toFixed(1) : "\u2014", sub: t("analytics.out_of") },
        ].map((item) => (
          <Card key={item.label}>
            <CardContent className="p-4">
              <div className="text-sm text-muted-foreground">{item.label}</div>
              <div className="text-3xl font-bold mt-1 text-purple-600">{item.value}</div>
              <div className="text-xs text-muted-foreground mt-0.5">{item.sub}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Status distribution pie */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("dash.status_dist")}</CardTitle>
          </CardHeader>
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
              <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">
                {t("dash.no_data")}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Score distribution bar */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("dash.score_dist")}</CardTitle>
          </CardHeader>
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
              <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">
                {t("dash.no_analysis")}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Category breakdown bar */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("dash.categories")}</CardTitle>
          </CardHeader>
          <CardContent>
            {categoryData.some((c) => c.count > 0) ? (
              <div className="space-y-3 pt-2">
                {categoryData.map((cat) => {
                  const pct = stats.analyzed > 0 ? (cat.count / stats.analyzed) * 100 : 0;
                  return (
                    <div key={cat.name} className="flex items-center gap-3">
                      <div className="w-36 text-sm truncate">{cat.name}</div>
                      <div className="flex-1 bg-slate-100 rounded-full h-5 overflow-hidden">
                        <div
                          className="h-5 rounded-full transition-all"
                          style={{ width: `${pct}%`, backgroundColor: cat.color }}
                        />
                      </div>
                      <div className="w-10 text-sm text-right font-medium">{cat.count}</div>
                      <div className="w-12 text-xs text-muted-foreground">{pct.toFixed(0)}%</div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">
                {t("dash.no_analysis")}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Funnel summary */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("dash.funnel")}</CardTitle>
          </CardHeader>
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
                    <div className="w-28 text-sm">{step.label}</div>
                    <div className="flex-1 bg-slate-100 rounded-full h-5 overflow-hidden">
                      <div
                        className={`h-5 rounded-full transition-all ${step.color}`}
                        style={{ width: `${pct}%` }}
                      />
                    </div>
                    <div className="w-8 text-sm text-right font-medium">{step.count}</div>
                    <div className="w-12 text-xs text-muted-foreground">{pct.toFixed(0)}%</div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Distribution Graphs */}
      <DistributionGraphs stats={stats} />
    </div>
  );
}

const DIMENSION_OPTIONS = [
  { key: "leadership", label: "Leadership" },
  { key: "motivation", label: "Motivation" },
  { key: "growth", label: "Growth" },
  { key: "vision", label: "Vision" },
  { key: "communication", label: "Communication" },
];

function DistributionGraphs({ stats }: { stats: DashboardStats }) {
  const [selectedDim, setSelectedDim] = useState("leadership");

  const dimDist = stats.dimension_distributions?.[selectedDim] || [];
  const dimMean = stats.dimension_means?.[selectedDim] || 0;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Graph 1: Overall score distribution with mean/median */}
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
                  <span className="text-muted-foreground">Mean: <span className="font-bold text-blue-600">{stats.score_mean.toFixed(1)}</span></span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-green-500" />
                  <span className="text-muted-foreground">Median: <span className="font-bold text-green-600">{stats.score_median.toFixed(1)}</span></span>
                </div>
              </div>
            </>
          ) : (
            <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">No data</div>
          )}
        </CardContent>
      </Card>

      {/* Graph 2: Per-dimension distribution with selector */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">Dimension Distribution</CardTitle>
            <div className="flex gap-1 bg-slate-100 rounded-lg p-0.5">
              {DIMENSION_OPTIONS.map((dim) => (
                <button
                  key={dim.key}
                  onClick={() => setSelectedDim(dim.key)}
                  className={`text-xs px-2 py-1 rounded-md transition-colors ${
                    selectedDim === dim.key
                      ? "bg-purple-600 text-white shadow-sm"
                      : "text-slate-500 hover:text-slate-800"
                  }`}
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
                  <span className="text-muted-foreground">Mean: <span className="font-bold text-blue-600">{dimMean.toFixed(1)}</span></span>
                </div>
              </div>
            </>
          ) : (
            <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">No data</div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
