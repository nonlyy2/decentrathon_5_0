"use client";

import { useState, useEffect } from "react";
import api from "@/lib/api";
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line, Legend,
} from "recharts";

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

interface PerformanceData {
  managers: ManagerStats[];
  summary: {
    total_candidates: number;
    shortlisted: number;
    rejected: number;
    pending: number;
  };
  decision_trend: { day: string; count: number }[];
}

export default function AuditorPage() {
  const [data, setData] = useState<PerformanceData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api.get("/auditor/manager-performance")
      .then((r) => setData(r.data))
      .catch(() => setError("Failed to load analytics"))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="p-8 text-center text-muted-foreground">Loading analytics...</div>;
  if (error) return <div className="p-8 text-center text-red-500">{error}</div>;
  if (!data) return null;

  const { managers, summary, decision_trend } = data;

  // Sort managers by total decisions descending
  const sorted = [...managers].sort((a, b) => b.total_decisions - a.total_decisions);

  // Bar chart data: decisions per manager
  const barData = sorted.slice(0, 10).map((m) => ({
    name: m.full_name || m.email.split("@")[0],
    decisions: m.total_decisions,
    upvotes: m.upvotes,
    downvotes: m.downvotes,
    successful: m.successful_cases,
  }));

  return (
    <div className="p-6 space-y-8 max-w-6xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold">Manager Performance Analytics</h1>
        <p className="text-muted-foreground text-sm mt-1">Auditor view — read-only analytics on all manager activity</p>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: "Total Candidates", value: summary.total_candidates, color: "text-blue-600" },
          { label: "Shortlisted", value: summary.shortlisted, color: "text-green-600" },
          { label: "Rejected", value: summary.rejected, color: "text-red-500" },
          { label: "Pending Review", value: summary.pending, color: "text-yellow-600" },
        ].map((card) => (
          <div key={card.label} className="bg-card border border-border rounded-xl p-4">
            <p className="text-xs text-muted-foreground mb-1">{card.label}</p>
            <p className={`text-3xl font-bold ${card.color}`}>{card.value}</p>
          </div>
        ))}
      </div>

      {/* Decision trend */}
      {decision_trend.length > 0 && (
        <div className="bg-card border border-border rounded-xl p-5">
          <h2 className="font-semibold mb-4">Decision Activity (Last 30 Days)</h2>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={decision_trend}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="day" tick={{ fontSize: 11 }} />
              <YAxis allowDecimals={false} tick={{ fontSize: 11 }} />
              <Tooltip />
              <Line type="monotone" dataKey="count" stroke="#c1f11d" strokeWidth={2} dot={false} name="Decisions" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Manager decisions chart */}
      {barData.length > 0 && (
        <div className="bg-card border border-border rounded-xl p-5">
          <h2 className="font-semibold mb-4">Decisions per Manager (Top 10)</h2>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={barData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" tick={{ fontSize: 11 }} />
              <YAxis allowDecimals={false} tick={{ fontSize: 11 }} />
              <Tooltip />
              <Legend />
              <Bar dataKey="upvotes" fill="#22c55e" name="Upvotes" stackId="a" />
              <Bar dataKey="downvotes" fill="#ef4444" name="Downvotes" stackId="a" />
              <Bar dataKey="successful" fill="#c1f11d" name="Led to Shortlist" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Manager table */}
      <div className="bg-card border border-border rounded-xl overflow-hidden">
        <div className="p-5 border-b border-border">
          <h2 className="font-semibold">Manager Performance Table</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-muted/50 text-left">
                <th className="px-4 py-3 font-medium text-muted-foreground">Manager</th>
                <th className="px-4 py-3 font-medium text-muted-foreground text-center">Total</th>
                <th className="px-4 py-3 font-medium text-green-600 text-center">👍 Upvotes</th>
                <th className="px-4 py-3 font-medium text-red-500 text-center">👎 Downvotes</th>
                <th className="px-4 py-3 font-medium text-muted-foreground text-center">Shortlists</th>
                <th className="px-4 py-3 font-medium text-muted-foreground text-center">Rejects</th>
                <th className="px-4 py-3 font-medium text-muted-foreground text-center">Successful</th>
                <th className="px-4 py-3 font-medium text-muted-foreground text-center">Efficiency</th>
                <th className="px-4 py-3 font-medium text-muted-foreground">Last Active</th>
              </tr>
            </thead>
            <tbody>
              {sorted.map((m) => (
                <tr key={m.user_id} className="border-t border-border hover:bg-muted/30">
                  <td className="px-4 py-3">
                    <div className="font-medium">{m.full_name || "—"}</div>
                    <div className="text-xs text-muted-foreground">{m.email}</div>
                    <div className="text-xs text-blue-500 capitalize">{m.role}</div>
                  </td>
                  <td className="px-4 py-3 text-center font-bold">{m.total_decisions}</td>
                  <td className="px-4 py-3 text-center text-green-600">{m.upvotes}</td>
                  <td className="px-4 py-3 text-center text-red-500">{m.downvotes}</td>
                  <td className="px-4 py-3 text-center">{m.shortlists}</td>
                  <td className="px-4 py-3 text-center">{m.rejects}</td>
                  <td className="px-4 py-3 text-center text-green-600 font-medium">{m.successful_cases}</td>
                  <td className="px-4 py-3 text-center">
                    <span className={`font-bold ${m.efficiency_score >= 50 ? "text-green-600" : m.efficiency_score >= 25 ? "text-yellow-600" : "text-red-500"}`}>
                      {m.efficiency_score.toFixed(1)}%
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-muted-foreground">
                    {m.last_active_at ? new Date(m.last_active_at).toLocaleDateString() : "Never"}
                  </td>
                </tr>
              ))}
              {sorted.length === 0 && (
                <tr>
                  <td colSpan={9} className="px-4 py-8 text-center text-muted-foreground">No manager activity yet</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
