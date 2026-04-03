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
import { ThumbsUp, ThumbsDown } from "lucide-react";

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
  upvotes: number;
  downvotes: number;
  net_score: number;
  upvote_threshold: number;
}

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
        if (data.winning_decision === "upvote") {
          toast.success(`Auto-shortlisted! Net score: +${data.net_score}`);
        } else if (data.winning_decision === "downvote") {
          toast.error(`Auto-rejected! Net score: ${data.net_score}`);
        } else {
          toast.success(`Consensus reached: ${data.winning_decision}`);
        }
      } else {
        const netStr = data.net_score >= 0 ? `+${data.net_score}` : `${data.net_score}`;
        toast.success(`Vote recorded. Net score: ${netStr} (need +${data.upvote_threshold} to auto-shortlist)`);
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

  const upvotes = reviewInfo?.upvotes ?? 0;
  const downvotes = reviewInfo?.downvotes ?? 0;
  const netScore = reviewInfo?.net_score ?? 0;
  const threshold = reviewInfo?.upvote_threshold ?? 3;
  const hasVotes = upvotes > 0 || downvotes > 0;

  return (
    <>
      <div className="flex gap-2 flex-wrap">
        <Button
          size="sm"
          className="bg-green-600 hover:bg-green-700 text-white flex items-center gap-1.5"
          onClick={() => handleClick("upvote")}
        >
          <ThumbsUp size={14} />
          Upvote
        </Button>
        <Button
          size="sm"
          className="bg-red-600 hover:bg-red-700 text-white flex items-center gap-1.5"
          onClick={() => handleClick("downvote")}
        >
          <ThumbsDown size={14} />
          Downvote
        </Button>
      </div>

      {/* Score display */}
      {hasVotes && (
        <div className="mt-3 space-y-2">
          <div className="flex items-center gap-3 text-sm">
            <span className="flex items-center gap-1 text-green-600">
              <ThumbsUp size={13} /> {upvotes}
            </span>
            <span className="flex items-center gap-1 text-red-500">
              <ThumbsDown size={13} /> {downvotes}
            </span>
            <span className={`font-bold ${netScore >= 0 ? "text-green-600" : "text-red-500"}`}>
              Net: {netScore >= 0 ? `+${netScore}` : netScore}
            </span>
            {netScore >= threshold ? (
              <Badge className="bg-green-100 text-green-700 text-xs">Auto-shortlist at +{threshold}</Badge>
            ) : (
              <span className="text-xs text-muted-foreground">
                Need +{threshold - netScore} more for shortlist
              </span>
            )}
          </div>
          {/* Progress bar toward threshold */}
          <div className="w-full bg-slate-100 rounded-full h-2">
            <div
              className={`h-2 rounded-full transition-all ${netScore >= threshold ? "bg-green-500" : netScore >= 0 ? "bg-lime-400" : "bg-red-400"}`}
              style={{ width: `${Math.min(Math.max((netScore / threshold) * 100, 0), 100)}%` }}
            />
          </div>
        </div>
      )}

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              {selected === "upvote" ? (
                <><ThumbsUp size={18} className="text-green-600" /> Upvote Candidate</>
              ) : (
                <><ThumbsDown size={18} className="text-red-500" /> Downvote Candidate</>
              )}
            </DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {selected === "upvote"
              ? `Your upvote will raise this candidate's score. At +${threshold}, they are automatically shortlisted.`
              : `Your downvote will lower this candidate's score. At -${threshold}, they are automatically rejected.`}
            {hasVotes && ` Current net score: ${netScore >= 0 ? `+${netScore}` : netScore}.`}
          </p>
          <Textarea
            placeholder={t("dec.notes_placeholder")}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={3}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsOpen(false)}>{t("dec.cancel")}</Button>
            <Button
              onClick={handleConfirm}
              disabled={submitting}
              className={selected === "upvote" ? "bg-green-600 hover:bg-green-700 text-white" : "bg-red-600 hover:bg-red-700 text-white"}
            >
              {submitting ? t("dec.saving") : `Confirm ${selected === "upvote" ? "Upvote" : "Downvote"}`}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
