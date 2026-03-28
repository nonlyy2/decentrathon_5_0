"use client";

import { useState } from "react";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { toast } from "sonner";

interface DecisionButtonsProps {
  candidateId: number;
  currentStatus: string;
  onDecisionMade: () => void;
}

const decisions = [
  { value: "shortlist", label: "Shortlist", className: "bg-green-600 hover:bg-green-700 text-white" },
  { value: "waitlist", label: "Waitlist", className: "bg-yellow-500 hover:bg-yellow-600 text-white" },
  { value: "review", label: "Review", className: "bg-blue-600 hover:bg-blue-700 text-white" },
  { value: "reject", label: "Reject", className: "bg-red-600 hover:bg-red-700 text-white" },
];

const statusMap: Record<string, string> = {
  shortlist: "shortlisted",
  reject: "rejected",
  waitlist: "waitlisted",
  review: "analyzed",
};

export default function DecisionButtons({ candidateId, currentStatus, onDecisionMade }: DecisionButtonsProps) {
  const [selected, setSelected] = useState<string | null>(null);
  const [notes, setNotes] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const handleClick = (decision: string) => {
    setSelected(decision);
    setNotes("");
    setIsOpen(true);
  };

  const handleConfirm = async () => {
    if (!selected) return;
    setSubmitting(true);
    try {
      await api.post(`/candidates/${candidateId}/decision`, {
        decision: selected,
        notes: notes || null,
      });
      toast.success(`Candidate ${selected}ed`);
      setIsOpen(false);
      onDecisionMade();
    } catch {
      toast.error("Failed to save decision");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <div className="flex gap-2 flex-wrap">
        {decisions.map((d) => (
          <Button
            key={d.value}
            size="sm"
            className={d.className}
            onClick={() => handleClick(d.value)}
            disabled={currentStatus === statusMap[d.value]}
          >
            {d.label}
          </Button>
        ))}
      </div>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm: {selected?.charAt(0).toUpperCase()}{selected?.slice(1)}</DialogTitle>
          </DialogHeader>
          <Textarea
            placeholder="Add notes (optional)..."
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={3}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsOpen(false)}>Cancel</Button>
            <Button onClick={handleConfirm} disabled={submitting}>
              {submitting ? "Saving..." : "Confirm"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
