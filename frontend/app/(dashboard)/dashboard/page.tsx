"use client";

import { useFetch } from "@/lib/hooks";
import { DashboardStats } from "@/lib/types";
import StatCard from "@/components/StatCard";
import { Users, CheckCircle, Star, Clock } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from "recharts";

const SCORE_COLORS: Record<string, string> = {
  "0-49": "#ef4444",
  "50-64": "#eab308",
  "65-79": "#3b82f6",
  "80-100": "#22c55e",
};

export default function DashboardPage() {
  const { data: stats, loading } = useFetch<DashboardStats>("/stats");

  if (loading || !stats) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <div className="grid grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i}><CardContent className="p-6"><div className="h-16 bg-slate-200 rounded animate-pulse" /></CardContent></Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard title="Total Candidates" value={stats.total_candidates} icon={<Users className="text-blue-600" size={24} />} color="bg-blue-100" />
        <StatCard title="Analyzed" value={stats.analyzed} icon={<CheckCircle className="text-green-600" size={24} />} color="bg-green-100" />
        <StatCard title="Shortlisted" value={stats.shortlisted} icon={<Star className="text-purple-600" size={24} />} color="bg-purple-100" />
        <StatCard title="Pending" value={stats.pending} icon={<Clock className="text-gray-600" size={24} />} color="bg-gray-100" />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-lg">Score Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            {stats.score_distribution.some((b) => b.count > 0) ? (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={stats.score_distribution}>
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
              <div className="h-48 flex items-center justify-center text-muted-foreground">
                No analysis data yet. Analyze candidates to see distribution.
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Average Score</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center justify-center h-48">
            <div className="text-5xl font-bold text-purple-600">
              {stats.avg_score > 0 ? stats.avg_score.toFixed(1) : "—"}
            </div>
            <div className="text-sm text-muted-foreground mt-2">out of 100</div>
          </CardContent>
        </Card>
      </div>

      {Object.keys(stats.category_counts).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Category Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {["Strong Recommend", "Recommend", "Borderline", "Not Recommended"].map((cat) => {
                const count = stats.category_counts[cat] || 0;
                const total = stats.analyzed || 1;
                const colors: Record<string, string> = {
                  "Strong Recommend": "bg-green-500",
                  "Recommend": "bg-blue-500",
                  "Borderline": "bg-yellow-500",
                  "Not Recommended": "bg-red-500",
                };
                return (
                  <div key={cat} className="flex items-center gap-3">
                    <div className="w-40 text-sm">{cat}</div>
                    <div className="flex-1 bg-slate-100 rounded-full h-4">
                      <div
                        className={`h-4 rounded-full ${colors[cat]}`}
                        style={{ width: `${(count / total) * 100}%` }}
                      />
                    </div>
                    <div className="w-8 text-sm text-right">{count}</div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
