"use client";

import { useFetch } from "@/lib/hooks";
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

const CATEGORY_COLORS: Record<string, string> = {
  "Strong Recommend": "#22c55e",
  "Recommend": "#3b82f6",
  "Borderline": "#eab308",
  "Not Recommended": "#ef4444",
};

export default function AnalyticsPage() {
  const { data: stats, loading } = useFetch<DashboardStats>("/stats");

  if (loading || !stats) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Analytics</h1>
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
    { name: "Pending", value: stats.pending },
    { name: "Analyzed", value: pureAnalyzed },
    { name: "Shortlisted", value: stats.shortlisted },
    { name: "Waitlisted", value: stats.waitlisted },
    { name: "Rejected", value: stats.rejected },
  ].filter((d) => d.value > 0);

  const categoryData = ["Strong Recommend", "Recommend", "Borderline", "Not Recommended"].map((cat) => ({
    name: cat,
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
      <h1 className="text-2xl font-bold">Analytics</h1>

      {/* KPI row */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: "Total Applications", value: stats.total_candidates, sub: "all time" },
          { label: "Analysis Rate", value: `${analysisRate}%`, sub: `${stats.analyzed} analyzed` },
          { label: "Shortlist Rate", value: `${conversionRate}%`, sub: `${stats.shortlisted} shortlisted` },
          { label: "Avg AI Score", value: stats.avg_score > 0 ? stats.avg_score.toFixed(1) : "—", sub: "out of 100" },
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
            <CardTitle className="text-base">Status Distribution</CardTitle>
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
                      <Cell key={entry.name} fill={STATUS_COLORS[entry.name] || "#8884d8"} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-48 flex items-center justify-center text-muted-foreground text-sm">
                No data yet
              </div>
            )}
          </CardContent>
        </Card>

        {/* Score distribution bar */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Score Distribution</CardTitle>
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
                No analysis data yet
              </div>
            )}
          </CardContent>
        </Card>

        {/* Category breakdown bar */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Recommendation Categories</CardTitle>
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
                No analysis data yet
              </div>
            )}
          </CardContent>
        </Card>

        {/* Funnel summary */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Admissions Funnel</CardTitle>
          </CardHeader>
          <CardContent className="pt-2">
            <div className="space-y-3">
              {[
                { label: "Applied", count: stats.total_candidates, color: "bg-slate-400" },
                { label: "AI Analyzed", count: stats.analyzed, color: "bg-blue-500" },
                { label: "Shortlisted", count: stats.shortlisted, color: "bg-green-500" },
                { label: "Waitlisted", count: stats.waitlisted, color: "bg-yellow-500" },
                { label: "Rejected", count: stats.rejected, color: "bg-red-500" },
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
    </div>
  );
}
