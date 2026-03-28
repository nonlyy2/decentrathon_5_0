"use client";

import { RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis, Radar, ResponsiveContainer } from "recharts";

interface ScoreRadarProps {
  scores: {
    leadership: number;
    motivation: number;
    growth: number;
    vision: number;
    communication: number;
  };
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
        <Radar dataKey="score" stroke="#7c3aed" fill="#7c3aed" fillOpacity={0.3} strokeWidth={2} />
      </RadarChart>
    </ResponsiveContainer>
  );
}
