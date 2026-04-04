"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import api from "@/lib/api";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { Loader2, Search, StickyNote } from "lucide-react";

interface CandidateNote {
  id: number;
  full_name: string;
  email: string;
  note_content: string;
  note_updated_at: string;
}

export default function MyNotesPage() {
  const [candidates, setCandidates] = useState<CandidateNote[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");

  useEffect(() => {
    const fetchCandidatesWithNotes = async () => {
      setLoading(true);
      try {
        // Fetch all candidates to allow browsing notes
        const res = await api.get("/candidates?limit=100&offset=0");
        const items = res.data.candidates || [];

        // Fetch notes for each candidate in parallel
        const withNotes = await Promise.all(
          items.map(async (c: { id: number; full_name: string; email: string }) => {
            try {
              const noteRes = await api.get(`/candidates/${c.id}/notes`);
              return {
                id: c.id,
                full_name: c.full_name,
                email: c.email,
                note_content: noteRes.data.content || "",
                note_updated_at: noteRes.data.updated_at || "",
              };
            } catch {
              return {
                id: c.id,
                full_name: c.full_name,
                email: c.email,
                note_content: "",
                note_updated_at: "",
              };
            }
          })
        );

        // Show only candidates that have notes, sorted by most recent
        const filtered = withNotes
          .filter((c: CandidateNote) => c.note_content.length > 0)
          .sort((a: CandidateNote, b: CandidateNote) =>
            new Date(b.note_updated_at).getTime() - new Date(a.note_updated_at).getTime()
          );
        setCandidates(filtered);
      } catch {
        toast.error("Failed to load notes");
      } finally {
        setLoading(false);
      }
    };
    fetchCandidatesWithNotes();
  }, []);

  const filteredCandidates = candidates.filter(
    (c) =>
      c.full_name.toLowerCase().includes(search.toLowerCase()) ||
      c.email.toLowerCase().includes(search.toLowerCase()) ||
      c.note_content.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <h1 className="text-2xl font-bold text-foreground flex items-center gap-2">
          <StickyNote size={24} /> My Notes
        </h1>
        <p className="text-sm text-muted-foreground">
          Private notes you&apos;ve written about candidates. Only visible to you.
        </p>
      </div>

      <div className="relative max-w-sm">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input value={search} onChange={(e) => setSearch(e.target.value)}
          placeholder="Search notes..." className="pl-9" />
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="animate-spin text-muted-foreground" size={32} />
        </div>
      ) : filteredCandidates.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <StickyNote size={40} className="mx-auto mb-3 opacity-30" />
          <p>No notes yet.</p>
          <p className="text-sm mt-1">
            Go to a candidate&apos;s profile to write private notes about them.
          </p>
        </div>
      ) : (
        <div className="grid gap-4">
          {filteredCandidates.map((c) => (
            <Link key={c.id} href={`/candidates/${c.id}`}
              className="block border rounded-lg p-4 bg-card hover:border-primary transition-colors">
              <div className="flex items-start justify-between">
                <div>
                  <p className="font-medium text-foreground">{c.full_name}</p>
                  <p className="text-xs text-muted-foreground">{c.email}</p>
                </div>
                <p className="text-xs text-muted-foreground">
                  {c.note_updated_at ? new Date(c.note_updated_at).toLocaleDateString() : ""}
                </p>
              </div>
              <p className="text-sm text-muted-foreground mt-2 line-clamp-3 whitespace-pre-wrap">
                {c.note_content}
              </p>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
