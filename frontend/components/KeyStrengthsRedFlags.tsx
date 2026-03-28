import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CheckCircle, AlertTriangle } from "lucide-react";

interface Props {
  strengths: string[];
  redFlags: string[];
}

export default function KeyStrengthsRedFlags({ strengths, redFlags }: Props) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <CardHeader className="p-0 pb-2">
              <CardTitle className="text-sm flex items-center gap-2 text-green-600">
                <CheckCircle size={16} /> Key Strengths
              </CardTitle>
            </CardHeader>
            {strengths.length > 0 ? (
              <ul className="space-y-1">
                {strengths.map((s, i) => (
                  <li key={i} className="text-sm flex items-start gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-green-500 mt-1.5 shrink-0" />
                    {s}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-slate-400 italic">No key strengths identified</p>
            )}
          </div>

          <div>
            <CardHeader className="p-0 pb-2">
              <CardTitle className="text-sm flex items-center gap-2 text-red-600">
                <AlertTriangle size={16} /> Red Flags
              </CardTitle>
            </CardHeader>
            {redFlags.length > 0 ? (
              <ul className="space-y-1">
                {redFlags.map((f, i) => (
                  <li key={i} className="text-sm flex items-start gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-red-500 mt-1.5 shrink-0" />
                    {f}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-green-600 flex items-center gap-1">
                <CheckCircle size={14} /> No red flags detected
              </p>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
