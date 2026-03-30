"use client";

import { RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis, Radar, ResponsiveContainer, Tooltip } from "recharts";

interface ScoreRadarProps {
  scores: {
    leadership: number;
    motivation: number;
    growth: number;
    vision: number;
    communication: number;
  };
}

/* eslint-disable @typescript-eslint/no-explicit-any */
function CustomTooltip({ active, payload }: any) {
  if (!active || !payload?.length) return null;
  const { subject, score } = payload[0].payload;
  return (
    <div className="bg-white border border-slate-200 shadow-lg rounded-lg px-3 py-2">
      <p className="text-sm font-medium text-slate-800">{subject}</p>
      <p className="text-lg font-bold text-purple-600">{score}<span className="text-xs text-slate-400 ml-0.5">/100</span></p>
    </div>
  );
}

export default function ScoreRadar({ scores }: ScoreRadarProps) {
  const data = [
    { subject: "Leadership", score: scores.leadership, fullMark: 100 },
    { subject: "Motivation", score: scores.motivation, fullMark: 100 },
    { subject: "Growth", score: scores.growth, fullMark: 100 },
    { subject: "Vision", score: scores.vision, fullMark: 100 },
    { subject: "Communication", score: scores.communication, fullMark: 100 },
  ];

  return (
    <ResponsiveContainer width="100%" height={250}>
      <RadarChart data={data}>
        <PolarGrid strokeDasharray="3 3" />
        <PolarAngleAxis dataKey="subject" tick={{ fontSize: 12 }} />
        <PolarRadiusAxis angle={30} domain={[0, 100]} tick={{ fontSize: 10 }} />
        <Radar dataKey="score" stroke="#7c3aed" fill="#7c3aed" fillOpacity={0.3} strokeWidth={2} activeDot={{ r: 5, fill: "#7c3aed" }} />
        <Tooltip content={<CustomTooltip />} />
      </RadarChart>
    </ResponsiveContainer>
  );
}
