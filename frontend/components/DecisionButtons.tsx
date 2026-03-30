"use client";

import { useState, useEffect, useCallback } from "react";
import api from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { Decision } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { toast } from "sonner";

interface DecisionButtonsProps {
  candidateId: number;
  onDecisionMade: () => void;
}

interface VoteSummary {
  decision: string;
  count: number;
}

interface DecisionsResponse {
  decisions: Decision[];
  vote_summary: VoteSummary[] | null;
  total_reviews: number;
  required_reviews: number;
}

const decisionColors: Record<string, string> = {
  shortlist: "bg-green-100 text-green-700",
  reject: "bg-red-100 text-red-700",
  waitlist: "bg-yellow-100 text-yellow-700",
  review: "bg-blue-100 text-blue-700",
};

export default function DecisionButtons({ candidateId, onDecisionMade }: DecisionButtonsProps) {
  const { t } = useI18n();
  const [selected, setSelected] = useState<string | null>(null);
  const [notes, setNotes] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [reviewInfo, setReviewInfo] = useState<DecisionsResponse | null>(null);

  const fetchReviewInfo = useCallback(() => {
    api.get(`/candidates/${candidateId}/decisions`).then((res) => setReviewInfo(res.data)).catch(() => {});
  }, [candidateId]);

  useEffect(() => { fetchReviewInfo(); }, [fetchReviewInfo]);

  const decisions = [
    { value: "shortlist", label: t("dec.shortlist"), className: "bg-green-600 hover:bg-green-700 text-white" },
    { value: "waitlist", label: t("dec.waitlist"), className: "bg-yellow-500 hover:bg-yellow-600 text-white" },
    { value: "review", label: t("dec.review"), className: "bg-blue-600 hover:bg-blue-700 text-white" },
    { value: "reject", label: t("dec.reject"), className: "bg-red-600 hover:bg-red-700 text-white" },
  ];

  const handleClick = (decision: string) => {
    setSelected(decision);
    setNotes("");
    setIsOpen(true);
  };

  const handleConfirm = async () => {
    if (!selected) return;
    setSubmitting(true);
    try {
      const res = await api.post(`/candidates/${candidateId}/decision`, {
        decision: selected,
        notes: notes || null,
      });
      const data = res.data;
      if (data.consensus_reached) {
        toast.success(`Consensus reached: ${data.winning_decision} (${data.total_reviews}/${data.required_reviews} reviews)`);
      } else {
        toast.success(`Vote recorded (${data.total_reviews}/${data.required_reviews} reviews)`);
      }
      setIsOpen(false);
      fetchReviewInfo();
      onDecisionMade();
    } catch {
      toast.error("Failed to save decision");
    } finally {
      setSubmitting(false);
    }
  };

  const totalVotes = reviewInfo?.total_reviews || 0;
  const requiredVotes = reviewInfo?.required_reviews || 4;
  const progress = Math.min((totalVotes / requiredVotes) * 100, 100);

  return (
    <>
      <div className="flex gap-2 flex-wrap">
        {decisions.map((d) => (
          <Button
            key={d.value}
            size="sm"
            className={d.className}
            onClick={() => handleClick(d.value)}
          >
            {d.label}
          </Button>
        ))}
      </div>

      {/* Review progress */}
      {totalVotes > 0 && (
        <div className="mt-3 space-y-2">
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>Reviews: {totalVotes}/{requiredVotes}</span>
            {totalVotes < requiredVotes && (
              <span className="text-orange-500 font-medium">{requiredVotes - totalVotes} more needed</span>
            )}
            {totalVotes >= requiredVotes && (
              <span className="text-green-600 font-medium">Consensus reached</span>
            )}
          </div>
          <div className="w-full bg-slate-100 rounded-full h-2">
            <div
              className={`h-2 rounded-full transition-all ${totalVotes >= requiredVotes ? "bg-green-500" : "bg-purple-500"}`}
              style={{ width: `${progress}%` }}
            />
          </div>
          {reviewInfo?.vote_summary && reviewInfo.vote_summary.length > 0 && (
            <div className="flex gap-1.5 flex-wrap">
              {reviewInfo.vote_summary.map((v) => (
                <Badge key={v.decision} variant="outline" className={`text-xs ${decisionColors[v.decision] || ""}`}>
                  {v.decision}: {v.count}
                </Badge>
              ))}
            </div>
          )}
        </div>
      )}

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("dec.confirm_title")}: {selected?.charAt(0).toUpperCase()}{selected?.slice(1)}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Your vote will be recorded. The final decision requires {requiredVotes} reviews.
            {totalVotes > 0 && ` Currently ${totalVotes} vote(s) recorded.`}
          </p>
          <Textarea
            placeholder={t("dec.notes_placeholder")}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={3}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsOpen(false)}>{t("dec.cancel")}</Button>
            <Button onClick={handleConfirm} disabled={submitting}>
              {submitting ? t("dec.saving") : t("dec.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
