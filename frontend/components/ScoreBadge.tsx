const categoryColors: Record<string, string> = {
  "Strong Recommend": "bg-green-100 text-green-800",
  "Recommend": "bg-blue-100 text-blue-800",
  "Borderline": "bg-yellow-100 text-yellow-800",
  "Not Recommended": "bg-red-100 text-red-800",
};

export default function ScoreBadge({ score, category }: { score: number | null; category: string | null }) {
  if (score === null || score === undefined) {
    return <span className="text-slate-400">—</span>;
  }

  const colorClass = categoryColors[category || ""] || "bg-slate-100 text-slate-800";

  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-sm font-medium ${colorClass}`}>
      {Math.round(score)}
    </span>
  );
}
