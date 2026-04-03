"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import api from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { MessageSquare, Flame, Users, RefreshCw, Send, Eye } from "lucide-react";

interface ActivityEntry {
  id: number;
  user_id: number | null;
  user_email: string | null;
  user_name: string | null;
  candidate_id: number | null;
  candidate_name: string | null;
  action_type: string;
  content: string | null;
  created_at: string;
}

interface DiscussionCandidate {
  id: number;
  full_name: string;
  email: string;
  status: string;
  major: string | null;
  discussion_note: string | null;
  flagged_by: string | null;
  flagged_at: string | null;
  final_score: number | null;
  category: string | null;
}

const ACTION_COLORS: Record<string, string> = {
  comment: "bg-blue-100 text-blue-700",
  decision: "bg-purple-100 text-purple-700",
  needs_discussion: "bg-orange-100 text-orange-700",
  analysis: "bg-green-100 text-green-700",
  status_change: "bg-slate-100 text-slate-700",
};

export default function WarRoomPage() {
  const { user } = useAuth();
  const [tab, setTab] = useState<"feed" | "discussion">("feed");
  const [feed, setFeed] = useState<ActivityEntry[]>([]);
  const [discussion, setDiscussion] = useState<DiscussionCandidate[]>([]);
  const [feedTotal, setFeedTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [posting, setPosting] = useState(false);
  const [newPost, setNewPost] = useState("");
  const [newPostCandidateId, setNewPostCandidateId] = useState("");
  const [refreshing, setRefreshing] = useState(false);

  const isAuditor = user?.role === "auditor";

  const fetchFeed = useCallback(async () => {
    try {
      const r = await api.get("/war-room/feed?limit=50");
      setFeed(r.data.entries || []);
      setFeedTotal(r.data.total || 0);
    } catch { /* ignore */ }
  }, []);

  const fetchDiscussion = useCallback(async () => {
    try {
      const r = await api.get("/war-room/discussion");
      setDiscussion(r.data || []);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => {
    Promise.all([fetchFeed(), fetchDiscussion()]).finally(() => setLoading(false));
  }, [fetchFeed, fetchDiscussion]);

  const handleRefresh = async () => {
    setRefreshing(true);
    await Promise.all([fetchFeed(), fetchDiscussion()]);
    setRefreshing(false);
  };

  const handlePost = async () => {
    const text = newPost.trim();
    if (!text) return;
    setPosting(true);
    try {
      await api.post("/war-room/feed", {
        action_type: "comment",
        content: text,
        candidate_id: newPostCandidateId ? parseInt(newPostCandidateId) : undefined,
      });
      toast.success("Posted to War Room");
      setNewPost("");
      setNewPostCandidateId("");
      await fetchFeed();
    } catch {
      toast.error("Failed to post");
    } finally {
      setPosting(false);
    }
  };

  const handleRemoveDiscussion = async (candidateId: number) => {
    try {
      await api.post(`/candidates/${candidateId}/discuss`, { remove: true });
      toast.success("Removed from discussion list");
      fetchDiscussion();
    } catch {
      toast.error("Failed to update");
    }
  };

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Flame size={22} style={{ color: "#c1f11d" }} />
            Committee War Room
          </h1>
          <p className="text-muted-foreground text-sm mt-1">Global activity feed, discussions, and coordination hub</p>
        </div>
        <Button variant="outline" size="sm" onClick={handleRefresh} disabled={refreshing} className="flex items-center gap-1">
          <RefreshCw size={14} className={refreshing ? "animate-spin" : ""} />
          Refresh
        </Button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-muted rounded-lg p-1 w-fit">
        {[
          { key: "feed", label: "Activity Feed", icon: MessageSquare, count: feedTotal },
          { key: "discussion", label: "Needs Discussion", icon: Users, count: discussion.length },
        ].map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key as "feed" | "discussion")}
            className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              tab === t.key ? "text-black" : "text-muted-foreground hover:text-foreground"
            }`}
            style={tab === t.key ? { backgroundColor: "#c1f11d" } : {}}
          >
            <t.icon size={14} />
            {t.label}
            {t.count > 0 && (
              <span className={`text-xs rounded-full px-1.5 py-0.5 ${tab === t.key ? "bg-black/20 text-black" : "bg-muted-foreground/20"}`}>
                {t.count}
              </span>
            )}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="text-center py-12 text-muted-foreground">Loading...</div>
      ) : (
        <>
          {/* Activity Feed */}
          {tab === "feed" && (
            <div className="space-y-4">
              {/* Post composer (non-auditors) */}
              {!isAuditor && (
                <div className="bg-card border border-border rounded-xl p-4 space-y-3">
                  <p className="text-sm font-medium">Post to War Room</p>
                  <Textarea
                    value={newPost}
                    onChange={(e) => setNewPost(e.target.value)}
                    placeholder="Share a note, observation, or mention a candidate... Use #candidateId to reference a candidate."
                    rows={3}
                    className="resize-none"
                  />
                  <div className="flex items-center gap-2">
                    <input
                      type="number"
                      value={newPostCandidateId}
                      onChange={(e) => setNewPostCandidateId(e.target.value)}
                      placeholder="Candidate ID (optional)"
                      className="text-sm border border-border rounded-lg px-3 py-1.5 w-48 bg-background"
                    />
                    <Button
                      size="sm"
                      onClick={handlePost}
                      disabled={posting || !newPost.trim()}
                      className="flex items-center gap-1.5"
                      style={{ backgroundColor: "#c1f11d", color: "#111" }}
                    >
                      <Send size={14} />
                      {posting ? "Posting..." : "Post"}
                    </Button>
                  </div>
                </div>
              )}

              {/* Feed entries */}
              <div className="space-y-3">
                {feed.length === 0 && (
                  <div className="text-center py-12 text-muted-foreground">
                    No activity yet. Be the first to post!
                  </div>
                )}
                {feed.map((entry) => (
                  <div key={entry.id} className="bg-card border border-border rounded-xl p-4 space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2 text-sm">
                        <div className="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold"
                          style={{ backgroundColor: "#c1f11d", color: "#111" }}>
                          {(entry.user_name || entry.user_email || "?").charAt(0).toUpperCase()}
                        </div>
                        <span className="font-medium">{entry.user_name || entry.user_email || "Unknown"}</span>
                        <Badge variant="outline" className={`text-xs ${ACTION_COLORS[entry.action_type] || "bg-slate-100 text-slate-600"}`}>
                          {entry.action_type.replace(/_/g, " ")}
                        </Badge>
                        {entry.candidate_name && (
                          <span className="text-muted-foreground">
                            about{" "}
                            <Link href={`/candidates/${entry.candidate_id}`} className="text-blue-500 hover:underline font-medium">
                              {entry.candidate_name}
                            </Link>
                            {entry.candidate_id && (
                              <span className="text-xs text-muted-foreground font-mono ml-1">#{entry.candidate_id}</span>
                            )}
                          </span>
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {new Date(entry.created_at).toLocaleString()}
                      </span>
                    </div>
                    {entry.content && (
                      <p className="text-sm text-foreground pl-9 whitespace-pre-wrap">{entry.content}</p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Needs Discussion */}
          {tab === "discussion" && (
            <div className="space-y-3">
              {discussion.length === 0 && (
                <div className="text-center py-12 text-muted-foreground">
                  No candidates flagged for discussion yet.
                  <br />
                  <span className="text-xs">Open a candidate profile and click &quot;Flag for Discussion&quot; to add them here.</span>
                </div>
              )}
              {discussion.map((d) => (
                <div key={d.id} className="bg-card border border-orange-200 rounded-xl p-4 space-y-2">
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex-1 space-y-1">
                      <div className="flex items-center gap-2">
                        <Link href={`/candidates/${d.id}`} className="font-semibold text-blue-500 hover:underline">
                          {d.full_name}
                        </Link>
                        <Badge variant="outline" className="text-xs capitalize">{d.status}</Badge>
                        {d.final_score && (
                          <span className="text-xs text-muted-foreground">Score: {d.final_score.toFixed(1)}</span>
                        )}
                      </div>
                      {d.discussion_note && (
                        <p className="text-sm text-foreground bg-orange-50 rounded-lg px-3 py-1.5 border border-orange-100">
                          {d.discussion_note}
                        </p>
                      )}
                      <p className="text-xs text-muted-foreground">
                        Flagged by {d.flagged_by || "unknown"} · {d.flagged_at ? new Date(d.flagged_at).toLocaleString() : ""}
                      </p>
                    </div>
                    <div className="flex gap-2 shrink-0">
                      <Link href={`/candidates/${d.id}`}>
                        <Button size="sm" variant="outline" className="flex items-center gap-1">
                          <Eye size={13} /> View
                        </Button>
                      </Link>
                      {!isAuditor && (
                        <Button
                          size="sm"
                          variant="outline"
                          className="text-red-500 hover:text-red-600 border-red-200"
                          onClick={() => handleRemoveDiscussion(d.id)}
                        >
                          Remove
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
