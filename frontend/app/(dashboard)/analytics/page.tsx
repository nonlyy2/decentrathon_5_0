"use client";

import { useState, useEffect, useCallback } from "react";
import api from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { DashboardStats } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { BarChart3, TrendingUp, Users, Target, Award, Activity } from "lucide-react";

interface CandidateScoreVariance {
  candidate_id: number;
  full_name: string;
  analysis_count: number;
  mean_score: number;
  std_dev: number;
  min_score: number;
  max_score: number;
  score_range: number;
}

interface VarianceSummary {
  candidates: CandidateScoreVariance[];
  overall_mean_stdev: number;
  high_variance_count: number;
  total_multi_analyzed: number;
}

interface ManagerStats {
  user_id: number;
  email: string;
  full_name: string | null;
  role: string;
  total_decisions: number;
  upvotes: number;
  downvotes: number;
  shortlists: number;
  rejects: number;
  waitlists: number;
  reviews: number;
  successful_cases: number;
  efficiency_score: number;
  last_active_at: string | null;
}

interface DayCount {
  day: string;
  count: number;
}

function BarRow({ label, value, max, color }: { label: string; value: number; max: number; color: string }) {
  const pct = max > 0 ? (value / max) * 100 : 0;
  return (
    <div className="flex items-center gap-3">
      <div className="w-28 text-xs text-muted-foreground truncate text-right">{label}</div>
      <div className="flex-1 bg-muted rounded-full h-2">
        <div className={`h-2 rounded-full transition-all ${color}`} style={{ width: `${pct}%` }} />
      </div>
      <div className="w-10 text-xs font-medium text-right">{value}</div>
    </div>
  );
}

export default function AnalyticsPage() {
  const { user } = useAuth();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [managerData, setManagerData] = useState<{ managers: ManagerStats[]; decision_trend: DayCount[] } | null>(null);
  const [varianceData, setVarianceData] = useState<VarianceSummary | null>(null);
  const [loading, setLoading] = useState(true);

  const isAuditorOrAbove = ["auditor", "tech-admin", "superadmin", "admin"].includes(user?.role ?? "");

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [statsRes, managerRes, varianceRes] = await Promise.allSettled([
        api.get<DashboardStats>("/stats"),
        isAuditorOrAbove ? api.get("/auditor/manager-performance") : Promise.reject(null),
        isAuditorOrAbove ? api.get<VarianceSummary>("/auditor/analysis-variance") : Promise.reject(null),
      ]);
      if (statsRes.status === "fulfilled") setStats(statsRes.value.data);
      if (managerRes.status === "fulfilled") setManagerData(managerRes.value.data);
      if (varianceRes.status === "fulfilled") setVarianceData(varianceRes.value.data);
    } finally {
      setLoading(false);
    }
  }, [isAuditorOrAbove]);

  useEffect(() => { fetchData(); }, [fetchData]);

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        {[1,2,3,4].map(i => (
          <Card key={i}><CardContent className="p-6"><div className="h-32 bg-muted rounded animate-pulse" /></CardContent></Card>
        ))}
      </div>
    );
  }

  if (!stats) {
    return <div className="p-6 text-muted-foreground">Failed to load analytics data.</div>;
  }

  const total = stats.total_candidates;
  const statusData = [
    { label: "Pending", value: stats.pending, color: "bg-slate-400" },
    { label: "Shortlisted", value: stats.shortlisted, color: "bg-green-500" },
    { label: "Waitlisted", value: stats.waitlisted, color: "bg-yellow-500" },
    { label: "Rejected", value: stats.rejected, color: "bg-red-500" },
  ];

  const maxCategoryCount = Math.max(...Object.values(stats.category_counts || {}), 1);
  const maxScore = Math.max(...(stats.score_distribution || []).map(d => d.count), 1);

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <BarChart3 size={22} style={{ color: "#c1f11d" }} />
          Analytics
        </h1>
        <p className="text-muted-foreground text-sm mt-1">Candidate pipeline overview and performance metrics</p>
      </div>

      {/* Top stat cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          { label: "Total Candidates", value: total, icon: Users, color: "text-blue-500" },
          { label: "Shortlisted", value: stats.shortlisted, icon: Award, color: "text-green-500" },
          { label: "Avg Score", value: stats.avg_score ? stats.avg_score.toFixed(1) : "—", icon: Target, color: "text-purple-500" },
          { label: "Analyzed", value: stats.analyzed, icon: Activity, color: "text-orange-500" },
        ].map((item) => (
          <Card key={item.label}>
            <CardContent className="p-4 flex items-center gap-3">
              <item.icon size={24} className={item.color} />
              <div>
                <p className="text-xs text-muted-foreground">{item.label}</p>
                <p className="text-2xl font-bold text-foreground">{item.value}</p>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Status funnel */}
        <Card>
          <CardHeader><CardTitle className="text-base flex items-center gap-2"><TrendingUp size={16} /> Status Distribution</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            {statusData.map((s) => (
              <BarRow key={s.label} label={s.label} value={s.value} max={total} color={s.color} />
            ))}
            <div className="pt-2 border-t text-xs text-muted-foreground">
              Conversion rate: <span className="font-semibold text-green-600">{total > 0 ? ((stats.shortlisted / total) * 100).toFixed(1) : 0}%</span>
            </div>
          </CardContent>
        </Card>

        {/* AI Category breakdown */}
        <Card>
          <CardHeader><CardTitle className="text-base flex items-center gap-2"><Award size={16} /> AI Category Breakdown</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            {Object.entries(stats.category_counts || {}).sort((a,b) => b[1]-a[1]).map(([cat, cnt]) => (
              <BarRow key={cat} label={cat} value={cnt as number} max={maxCategoryCount} color="bg-purple-500" />
            ))}
            {Object.keys(stats.category_counts || {}).length === 0 && (
              <p className="text-sm text-muted-foreground">No analyzed candidates yet.</p>
            )}
          </CardContent>
        </Card>

        {/* Score distribution */}
        <Card>
          <CardHeader><CardTitle className="text-base flex items-center gap-2"><BarChart3 size={16} /> Score Distribution</CardTitle></CardHeader>
          <CardContent className="space-y-2">
            {(stats.score_distribution || []).map((d) => (
              <BarRow key={d.range} label={d.range} value={d.count} max={maxScore} color="bg-blue-500" />
            ))}
            {(stats.score_distribution || []).length === 0 && (
              <p className="text-sm text-muted-foreground">No score data yet.</p>
            )}
            {(stats.score_mean !== undefined || stats.score_median !== undefined) && (
              <div className="pt-2 border-t flex gap-4 text-xs text-muted-foreground">
                {stats.score_mean !== undefined && <span>Mean: <span className="font-semibold text-foreground">{stats.score_mean.toFixed(1)}</span></span>}
                {stats.score_median !== undefined && <span>Median: <span className="font-semibold text-foreground">{stats.score_median.toFixed(1)}</span></span>}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Dimension averages */}
        <Card>
          <CardHeader><CardTitle className="text-base flex items-center gap-2"><Target size={16} /> Dimension Averages</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            {Object.entries(stats.dimension_means || {}).map(([dim, mean]) => (
              <BarRow key={dim} label={dim.charAt(0).toUpperCase() + dim.slice(1)} value={Math.round(mean as number)} max={100} color="bg-lime-500" />
            ))}
            {Object.keys(stats.dimension_means || {}).length === 0 && (
              <p className="text-sm text-muted-foreground">No dimension data yet.</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Analysis Variance / Bias Detection — auditor+ only */}
      {isAuditorOrAbove && varianceData && varianceData.total_multi_analyzed > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Activity size={16} /> AI Scoring Consistency (Repeated Analyses)
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Summary stats */}
            <div className="grid grid-cols-3 gap-4">
              <div className="bg-muted/50 rounded-lg p-3 text-center">
                <p className="text-2xl font-bold text-foreground">{varianceData.total_multi_analyzed}</p>
                <p className="text-xs text-muted-foreground">Candidates with 2+ analyses</p>
              </div>
              <div className="bg-muted/50 rounded-lg p-3 text-center">
                <p className="text-2xl font-bold text-foreground">{varianceData.overall_mean_stdev.toFixed(2)}</p>
                <p className="text-xs text-muted-foreground">Mean Std Dev</p>
              </div>
              <div className={`rounded-lg p-3 text-center ${varianceData.high_variance_count > 0 ? "bg-red-50 dark:bg-red-950/30" : "bg-green-50 dark:bg-green-950/30"}`}>
                <p className={`text-2xl font-bold ${varianceData.high_variance_count > 0 ? "text-red-600" : "text-green-600"}`}>{varianceData.high_variance_count}</p>
                <p className="text-xs text-muted-foreground">High Variance (σ {">"} 5)</p>
              </div>
            </div>

            {/* Variance table */}
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-xs uppercase text-muted-foreground">
                    <th className="text-left py-2 px-3">Candidate</th>
                    <th className="text-right py-2 px-3"># Analyses</th>
                    <th className="text-right py-2 px-3">Mean</th>
                    <th className="text-right py-2 px-3">Std Dev</th>
                    <th className="text-right py-2 px-3">Min</th>
                    <th className="text-right py-2 px-3">Max</th>
                    <th className="text-right py-2 px-3">Range</th>
                  </tr>
                </thead>
                <tbody>
                  {varianceData.candidates.slice(0, 20).map((cv) => (
                    <tr key={cv.candidate_id} className="border-b border-border last:border-0 hover:bg-muted/30">
                      <td className="py-2 px-3">
                        <a href={`/candidates/${cv.candidate_id}`} className="text-blue-600 hover:underline dark:text-blue-400 font-medium">
                          {cv.full_name}
                        </a>
                      </td>
                      <td className="py-2 px-3 text-right font-mono">{cv.analysis_count}</td>
                      <td className="py-2 px-3 text-right font-mono">{cv.mean_score.toFixed(1)}</td>
                      <td className="py-2 px-3 text-right">
                        <span className={`font-mono font-semibold ${cv.std_dev > 5 ? "text-red-600" : cv.std_dev > 2 ? "text-amber-600" : "text-green-600"}`}>
                          {cv.std_dev.toFixed(2)}
                        </span>
                      </td>
                      <td className="py-2 px-3 text-right font-mono">{cv.min_score.toFixed(1)}</td>
                      <td className="py-2 px-3 text-right font-mono">{cv.max_score.toFixed(1)}</td>
                      <td className="py-2 px-3 text-right font-mono">{cv.score_range.toFixed(1)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {varianceData.candidates.length > 20 && (
              <p className="text-xs text-muted-foreground text-center">Showing top 20 by highest variance. {varianceData.candidates.length - 20} more candidates with repeated analyses.</p>
            )}
          </CardContent>
        </Card>
      )}

      {/* Manager performance table — auditor+ only */}
      {isAuditorOrAbove && managerData && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Users size={16} /> Manager Performance
            </CardTitle>
          </CardHeader>
          <CardContent className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-xs uppercase text-muted-foreground">
                  <th className="text-left py-2 px-3">Manager</th>
                  <th className="text-right py-2 px-3">Decisions</th>
                  <th className="text-right py-2 px-3">👍 Up</th>
                  <th className="text-right py-2 px-3">👎 Down</th>
                  <th className="text-right py-2 px-3">Shortlisted</th>
                  <th className="text-right py-2 px-3">Efficiency</th>
                  <th className="text-left py-2 px-3">Last Active</th>
                </tr>
              </thead>
              <tbody>
                {(managerData.managers || []).map((m) => (
                  <tr key={m.user_id} className="border-b border-border last:border-0 hover:bg-muted/30">
                    <td className="py-2 px-3">
                      <div className="font-medium">{m.full_name || m.email}</div>
                      {m.full_name && <div className="text-xs text-muted-foreground">{m.email}</div>}
                      <Badge variant="outline" className="text-[10px] mt-0.5">{m.role}</Badge>
                    </td>
                    <td className="py-2 px-3 text-right font-mono">{m.total_decisions}</td>
                    <td className="py-2 px-3 text-right text-green-600 font-mono">{m.upvotes}</td>
                    <td className="py-2 px-3 text-right text-red-600 font-mono">{m.downvotes}</td>
                    <td className="py-2 px-3 text-right text-green-700 font-mono">{m.successful_cases}</td>
                    <td className="py-2 px-3 text-right">
                      <span className={`font-semibold ${m.efficiency_score >= 50 ? "text-green-600" : "text-amber-600"}`}>
                        {m.efficiency_score.toFixed(0)}%
                      </span>
                    </td>
                    <td className="py-2 px-3 text-xs text-muted-foreground">
                      {m.last_active_at ? new Date(m.last_active_at).toLocaleDateString() : "—"}
                    </td>
                  </tr>
                ))}
                {(managerData.managers || []).length === 0 && (
                  <tr><td colSpan={7} className="py-6 text-center text-muted-foreground">No manager activity yet.</td></tr>
                )}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}

      {/* Decision trend — auditor+ only */}
      {isAuditorOrAbove && managerData?.decision_trend && managerData.decision_trend.length > 0 && (
        <Card>
          <CardHeader><CardTitle className="text-base flex items-center gap-2"><Activity size={16} /> Decision Activity (Last 30 Days)</CardTitle></CardHeader>
          <CardContent>
            <div className="flex items-end gap-1 h-24">
              {(() => {
                const maxVal = Math.max(...managerData.decision_trend.map(d => d.count), 1);
                return managerData.decision_trend.map((d) => (
                  <div key={d.day} className="flex-1 flex flex-col items-center gap-1 group relative">
                    <div
                      className="w-full rounded-t"
                      style={{ height: `${(d.count / maxVal) * 80}px`, backgroundColor: "#c1f11d" }}
                    />
                    <div className="absolute bottom-full mb-1 hidden group-hover:block bg-card border border-border rounded px-2 py-1 text-xs whitespace-nowrap shadow-sm z-10">
                      {d.day}: {d.count} decisions
                    </div>
                  </div>
                ));
              })()}
            </div>
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>{managerData.decision_trend[0]?.day}</span>
              <span>{managerData.decision_trend[managerData.decision_trend.length - 1]?.day}</span>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
