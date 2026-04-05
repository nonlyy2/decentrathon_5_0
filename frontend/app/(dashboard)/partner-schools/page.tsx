"use client";

import { useState, useEffect } from "react";
import api from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Loader2, School, Users, TrendingUp } from "lucide-react";
import { toast } from "sonner";

interface PartnerSchool {
  id: number;
  name: string;
  city: string | null;
  contact_email: string | null;
  contact_phone: string | null;
  graduates_per_year: number;
  candidate_count: number;
  avg_ai_score: number;
}

export default function PartnerSchoolsPage() {
  const [schools, setSchools] = useState<PartnerSchool[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.get("/partner-schools")
      .then((res) => setSchools(res.data.schools || []))
      .catch(() => toast.error("Failed to load partner schools"))
      .finally(() => setLoading(false));
  }, []);

  const totalCandidates = schools.reduce((sum, s) => sum + s.candidate_count, 0);
  const avgScore = schools.length > 0
    ? schools.reduce((sum, s) => sum + s.avg_ai_score * s.candidate_count, 0) / Math.max(totalCandidates, 1)
    : 0;

  if (loading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2 className="animate-spin text-muted-foreground" size={32} />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Partner Schools</h1>

      {/* Summary cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="p-4 flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-900/30">
              <School size={18} />
            </div>
            <div>
              <div className="text-2xl font-bold">{schools.length}</div>
              <div className="text-xs text-muted-foreground">Partner Schools</div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 flex items-center gap-3">
            <div className="p-2 rounded-lg bg-green-50 text-green-600 dark:bg-green-900/30">
              <Users size={18} />
            </div>
            <div>
              <div className="text-2xl font-bold">{totalCandidates}</div>
              <div className="text-xs text-muted-foreground">Total Candidates Referred</div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 flex items-center gap-3">
            <div className="p-2 rounded-lg bg-purple-50 text-purple-600 dark:bg-purple-900/30">
              <TrendingUp size={18} />
            </div>
            <div>
              <div className="text-2xl font-bold">{avgScore > 0 ? avgScore.toFixed(1) : "—"}</div>
              <div className="text-xs text-muted-foreground">Average AI Score</div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Schools table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">All Partner Schools</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>School Name</TableHead>
                  <TableHead>City</TableHead>
                  <TableHead>Contact</TableHead>
                  <TableHead className="text-center">Graduates/Year</TableHead>
                  <TableHead className="text-center">Candidates</TableHead>
                  <TableHead className="text-center">Avg AI Score</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {schools.map((school) => (
                  <TableRow key={school.id}>
                    <TableCell className="font-medium">{school.name}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{school.city || "—"}</TableCell>
                    <TableCell>
                      <div className="text-xs">
                        {school.contact_email && <div className="text-muted-foreground">{school.contact_email}</div>}
                        {school.contact_phone && <div className="text-muted-foreground">{school.contact_phone}</div>}
                      </div>
                    </TableCell>
                    <TableCell className="text-center">{school.graduates_per_year}</TableCell>
                    <TableCell className="text-center">
                      <Badge className={school.candidate_count > 0 ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"}>
                        {school.candidate_count}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-center">
                      {school.candidate_count > 0 ? (
                        <span className="font-medium">{school.avg_ai_score.toFixed(1)}</span>
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
