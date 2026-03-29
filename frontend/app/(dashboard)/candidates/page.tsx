"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import Link from "next/link";
import { useSearchParams, useRouter } from "next/navigation";
import api from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";
import { useAIProvider } from "@/lib/aiProvider";
import { CandidateListItem, DashboardStats } from "@/lib/types";
import StatusBadge from "@/components/StatusBadge";
import ScoreBadge from "@/components/ScoreBadge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { toast } from "sonner";
import { Loader2, ArrowUpDown, ArrowUp, ArrowDown, Play, Square, Download, Trash2 } from "lucide-react";

export default function CandidatesPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const { t, lang } = useI18n();
  const { provider, setProvider } = useAIProvider();
  const { user } = useAuth();

  const [candidates, setCandidates] = useState<CandidateListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState(searchParams.get("status") || "all");
  const [search, setSearch] = useState(searchParams.get("search") || "");
  const [page, setPage] = useState(Number(searchParams.get("page")) || 0);
  const [counts, setCounts] = useState<Record<string, number>>({});
  const [sortBy, setSortBy] = useState(searchParams.get("sort") || "created_at");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">((searchParams.get("order") as "asc" | "desc") || "desc");
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [bulkAction, setBulkAction] = useState<string | null>(null);

  // Admin controls state
  const [batchRunning, setBatchRunning] = useState(false);
  const [batchProgress, setBatchProgress] = useState<{ done: number; total: number } | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteConfirmText, setDeleteConfirmText] = useState("");
  const [deleting, setDeleting] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const limit = 20;

  const STATUS_TABS = [
    { value: "all", label: t("cand.all") },
    { value: "pending", label: t("status.pending") },
    { value: "analyzed", label: t("status.analyzed") },
    { value: "shortlisted", label: t("status.shortlisted") },
    { value: "waitlisted", label: t("status.waitlisted") },
    { value: "rejected", label: t("status.rejected") },
  ];

  const toggleSelect = (id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === candidates.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(candidates.map((c) => c.id)));
    }
  };

  const handleBulkDecision = async (decision: string) => {
    if (selectedIds.size === 0) return;
    setBulkAction(decision);
    try {
      const res = await api.post("/candidates/bulk-decision", {
        candidate_ids: Array.from(selectedIds),
        decision,
      });
      toast.success(res.data.message);
      setSelectedIds(new Set());
      fetchCandidates();
      fetchCounts();
    } catch {
      toast.error("Failed to apply bulk action");
    } finally {
      setBulkAction(null);
    }
  };

  // Batch analysis polling
  const startBatchPoll = useCallback(() => {
    if (pollRef.current) return;
    pollRef.current = setInterval(async () => {
      try {
        const res = await api.get("/analyze-all/status");
        const { running, processed, total, errors } = res.data;
        setBatchProgress({ done: processed, total });
        if (!running) {
          clearInterval(pollRef.current!);
          pollRef.current = null;
          setBatchRunning(false);
          setBatchProgress(null);
          const errCount = Array.isArray(errors) ? errors.length : 0;
          const ok = processed - errCount;
          if (errCount === 0) toast.success(`Batch complete: ${ok}/${processed} analyzed`);
          else if (ok === 0) toast.error(`Batch failed: all ${errCount} errored`);
          else toast.warning(`Batch done: ${ok} analyzed, ${errCount} failed`);
          fetchCandidates();
          fetchCounts();
        }
      } catch {
        clearInterval(pollRef.current!);
        pollRef.current = null;
        setBatchRunning(false);
      }
    }, 3000);
  }, []);

  useEffect(() => {
    return () => { if (pollRef.current) clearInterval(pollRef.current); };
  }, []);

  // Check if batch is already running on mount
  useEffect(() => {
    api.get("/analyze-all/status").then((res) => {
      if (res.data.running) {
        setBatchRunning(true);
        setBatchProgress({ done: res.data.processed, total: res.data.total });
        startBatchPoll();
      }
    }).catch(() => {});
  }, [startBatchPoll]);

  const handleAnalyzeAll = async () => {
    try {
      const qs = provider ? `?provider=${provider}` : "";
      const res = await api.post(`/analyze-all${qs}`);
      if (!res.data.count) {
        toast.info("No pending candidates to analyze");
        return;
      }
      toast.success(`Analysis started for ${res.data.count} candidates`);
      setBatchRunning(true);
      setBatchProgress({ done: 0, total: res.data.count });
      startBatchPoll();
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
      toast.error(msg ? `Error: ${msg}` : "Failed to start batch analysis");
    }
  };

  const handleStopBatch = async () => {
    try {
      await api.post("/analyze-all/stop");
      toast.info("Stopping batch analysis...");
    } catch {
      toast.error("Failed to stop batch");
    }
  };

  const handleExportCSV = async () => {
    try {
      const res = await api.get(`/candidates/export/csv?lang=${lang}`, { responseType: "blob" });
      const url = window.URL.createObjectURL(new Blob([res.data]));
      const a = document.createElement("a");
      a.href = url;
      a.download = `candidates_${new Date().toISOString().slice(0, 10)}.csv`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch {
      toast.error("Failed to export CSV");
    }
  };

  const handleDeleteAll = async () => {
    setDeleting(true);
    try {
      const res = await api.delete("/analyses");
      toast.success(`Deleted ${res.data.deleted} analyses`);
      setDeleteDialogOpen(false);
      setDeleteConfirmText("");
      fetchCandidates();
      fetchCounts();
    } catch {
      toast.error("Failed to delete analyses");
    } finally {
      setDeleting(false);
    }
  };

  // Sync state to URL params
  useEffect(() => {
    const params = new URLSearchParams();
    if (status !== "all") params.set("status", status);
    if (search) params.set("search", search);
    if (page > 0) params.set("page", String(page));
    if (sortBy !== "created_at") params.set("sort", sortBy);
    if (sortOrder !== "desc") params.set("order", sortOrder);
    const qs = params.toString();
    router.replace(`/candidates${qs ? `?${qs}` : ""}`, { scroll: false });
  }, [status, search, page, sortBy, sortOrder, router]);

  const fetchCounts = useCallback(async () => {
    try {
      const res = await api.get<DashboardStats>("/stats");
      const s = res.data;
      setCounts({
        all: s.total_candidates,
        pending: s.pending,
        analyzed: s.analyzed - s.shortlisted - s.rejected - s.waitlisted,
        shortlisted: s.shortlisted,
        waitlisted: s.waitlisted,
        rejected: s.rejected,
      });
    } catch {
      // counts are non-critical
    }
  }, []);

  useEffect(() => {
    fetchCounts();
  }, [fetchCounts]);

  const fetchCandidates = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (status !== "all") params.set("status", status);
      if (search) params.set("search", search);
      params.set("limit", String(limit));
      params.set("offset", String(page * limit));
      params.set("sort_by", sortBy);
      params.set("order", sortOrder);

      const res = await api.get(`/candidates?${params}`);
      setCandidates(res.data.candidates || []);
      setTotal(res.data.total);
    } catch {
      toast.error("Failed to load candidates");
    } finally {
      setLoading(false);
    }
  }, [status, search, page, sortBy, sortOrder]);

  useEffect(() => {
    fetchCandidates();
  }, [fetchCandidates]);

  useEffect(() => {
    setPage(0);
  }, [status, search]);

  const handleSort = (column: string) => {
    if (sortBy === column) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      setSortBy(column);
      setSortOrder("desc");
    }
    setPage(0);
  };

  const totalPages = Math.ceil(total / limit);
  const isAdmin = user?.role === "admin";

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("cand.title")}</h1>
        {isAdmin && (
          <div className="flex items-center gap-2">
            {/* Provider toggle */}
            <div className="flex items-center gap-1 bg-slate-100 border rounded-lg p-1">
              <button
                onClick={() => setProvider("gemini")}
                className={`text-xs px-2.5 py-1 rounded-md transition-colors ${
                  provider === "gemini" ? "bg-purple-600 text-white shadow-sm" : "text-slate-500 hover:text-slate-800"
                }`}
              >
                ☁ Gemini
              </button>
              <button
                onClick={() => setProvider("ollama")}
                className={`text-xs px-2.5 py-1 rounded-md transition-colors ${
                  provider === "ollama" ? "bg-purple-600 text-white shadow-sm" : "text-slate-500 hover:text-slate-800"
                }`}
              >
                ⚙ Ollama
              </button>
            </div>

            {/* Analyze All / Stop */}
            {batchRunning ? (
              <div className="flex items-center gap-2">
                {batchProgress && (
                  <span className="text-xs text-purple-600 font-medium">
                    {batchProgress.done}/{batchProgress.total}
                  </span>
                )}
                <Button size="sm" variant="destructive" onClick={handleStopBatch}>
                  <Square size={14} className="mr-1" /> {t("cand.stop")}
                </Button>
              </div>
            ) : (
              <Button size="sm" className="bg-purple-600 hover:bg-purple-700" onClick={handleAnalyzeAll}>
                <Play size={14} className="mr-1" /> {t("cand.analyze_all")}
              </Button>
            )}

            {/* Export CSV */}
            <Button size="sm" variant="outline" onClick={handleExportCSV}>
              <Download size={14} className="mr-1" /> {t("cand.export")}
            </Button>

            {/* Reset Analyses */}
            <Button
              size="sm"
              variant="outline"
              className="text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200"
              onClick={() => { setDeleteConfirmText(""); setDeleteDialogOpen(true); }}
              disabled={batchRunning}
            >
              <Trash2 size={14} className="mr-1" /> {t("cand.reset")}
            </Button>
          </div>
        )}
      </div>

      {/* Batch progress bar */}
      {batchRunning && batchProgress && (
        <div className="w-full bg-slate-200 rounded-full h-1.5">
          <div
            className="bg-purple-500 h-1.5 rounded-full transition-all"
            style={{ width: `${batchProgress.total > 0 ? (batchProgress.done / batchProgress.total) * 100 : 0}%` }}
          />
        </div>
      )}

      <div className="flex flex-wrap gap-1 border-b">
        {STATUS_TABS.map((tab) => {
          const count = counts[tab.value];
          const isActive = status === tab.value;
          return (
            <button
              key={tab.value}
              onClick={() => setStatus(tab.value)}
              className={`px-3 py-2 text-sm font-medium rounded-t-md transition-colors flex items-center gap-1.5 ${
                isActive
                  ? "border-b-2 border-purple-600 text-purple-600 bg-purple-50"
                  : "text-slate-500 hover:text-slate-800 hover:bg-slate-100"
              }`}
            >
              {tab.label}
              {count !== undefined && (
                <span className={`text-xs px-1.5 py-0.5 rounded-full ${
                  isActive ? "bg-purple-100 text-purple-700" : "bg-slate-100 text-slate-500"
                }`}>
                  {count}
                </span>
              )}
            </button>
          );
        })}
      </div>

      <div className="flex gap-4">
        <Input
          placeholder={t("cand.search")}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="max-w-sm"
        />
      </div>

      {/* Bulk action bar */}
      {selectedIds.size > 0 && (
        <div className="flex items-center gap-3 bg-purple-50 border border-purple-200 rounded-lg px-4 py-2">
          <span className="text-sm font-medium text-purple-700">{selectedIds.size} {t("cand.selected")}</span>
          <div className="flex gap-2">
            <Button size="sm" className="bg-green-600 hover:bg-green-700 text-white" disabled={!!bulkAction} onClick={() => handleBulkDecision("shortlist")}>
              {bulkAction === "shortlist" ? <Loader2 size={14} className="animate-spin mr-1" /> : null} {t("dec.shortlist")}
            </Button>
            <Button size="sm" className="bg-yellow-500 hover:bg-yellow-600 text-white" disabled={!!bulkAction} onClick={() => handleBulkDecision("waitlist")}>
              {t("dec.waitlist")}
            </Button>
            <Button size="sm" className="bg-red-600 hover:bg-red-700 text-white" disabled={!!bulkAction} onClick={() => handleBulkDecision("reject")}>
              {t("dec.reject")}
            </Button>
          </div>
          {selectedIds.size >= 2 && (
            <Button size="sm" variant="outline" onClick={() => router.push(`/compare?ids=${Array.from(selectedIds).join(",")}`)}>
              {t("cand.compare")} ({selectedIds.size})
            </Button>
          )}
          <Button size="sm" variant="ghost" onClick={() => setSelectedIds(new Set())} className="ml-auto text-slate-500">
            {t("cand.clear")}
          </Button>
        </div>
      )}

      <div className="bg-white rounded-lg border overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-10">
                <input
                  type="checkbox"
                  checked={candidates.length > 0 && selectedIds.size === candidates.length}
                  onChange={toggleSelectAll}
                  className="rounded border-slate-300"
                />
              </TableHead>
              <SortableHead column="full_name" label={t("cand.name")} sortBy={sortBy} sortOrder={sortOrder} onSort={handleSort} />
              <TableHead>{t("cand.city")}</TableHead>
              <SortableHead column="final_score" label={t("cand.score")} sortBy={sortBy} sortOrder={sortOrder} onSort={handleSort} />
              <TableHead>{t("cand.status")}</TableHead>
              <SortableHead column="created_at" label={t("cand.created")} sortBy={sortBy} sortOrder={sortOrder} onSort={handleSort} />
              <SortableHead column="analyzed_at" label={t("cand.analyzed_col")} sortBy={sortBy} sortOrder={sortOrder} onSort={handleSort} />
              <TableHead>{t("cand.model")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && candidates.length === 0 ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  {Array.from({ length: 8 }).map((_, j) => (
                    <TableCell key={j}><div className="h-4 bg-slate-200 rounded animate-pulse w-24" /></TableCell>
                  ))}
                </TableRow>
              ))
            ) : candidates.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                  {t("cand.no_found")}
                </TableCell>
              </TableRow>
            ) : (
              candidates.map((c) => (
                <TableRow
                  key={c.id}
                  className={`hover:bg-slate-50 transition-colors relative ${selectedIds.has(c.id) ? "bg-purple-50" : ""}`}
                >
                  <TableCell className="relative z-20">
                    <input
                      type="checkbox"
                      checked={selectedIds.has(c.id)}
                      onChange={() => toggleSelect(c.id)}
                      className="rounded border-slate-300"
                    />
                  </TableCell>
                  <TableCell className="font-medium">
                    <Link
                      href={`/candidates/${c.id}`}
                      className="absolute inset-0 z-10"
                      aria-label={`Open ${c.full_name}`}
                    />
                    <span className="relative z-20">{c.full_name}</span>
                  </TableCell>
                  <TableCell className="relative z-20">{c.city || "—"}</TableCell>
                  <TableCell className="relative z-20"><ScoreBadge score={c.final_score} category={c.category} /></TableCell>
                  <TableCell className="relative z-20"><StatusBadge status={c.status} /></TableCell>
                  <TableCell className="relative z-20 text-sm text-muted-foreground">
                    <div>{new Date(c.created_at).toLocaleDateString()}</div>
                    <div className="text-xs">{new Date(c.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</div>
                  </TableCell>
                  <TableCell className="relative z-20 text-sm text-muted-foreground">
                    {c.analyzed_at ? (
                      <>
                        <div>{new Date(c.analyzed_at).toLocaleDateString()}</div>
                        <div className="text-xs">{new Date(c.analyzed_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</div>
                      </>
                    ) : "—"}
                  </TableCell>
                  <TableCell className="relative z-20 text-sm text-muted-foreground">
                    {c.model_used ? c.model_used.replace(/^(gemini|ollama)\//, "") : "—"}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {t("cand.showing")} {page * limit + 1}-{Math.min((page + 1) * limit, total)} {t("cand.of")} {total}
          </p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setPage(page - 1)} disabled={page === 0}>
              {t("cand.previous")}
            </Button>
            <Button variant="outline" size="sm" onClick={() => setPage(page + 1)} disabled={page >= totalPages - 1}>
              {t("cand.next")}
            </Button>
          </div>
        </div>
      )}

      {/* Delete confirmation dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-red-600">{t("del.title")}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-slate-600">{t("del.desc")}</p>
          <p className="text-sm text-slate-500 mt-2">
            {t("del.confirm")} <span className="font-mono font-bold text-red-600">{"удалить"}</span>:
          </p>
          <Input
            value={deleteConfirmText}
            onChange={(e) => setDeleteConfirmText(e.target.value)}
            placeholder="удалить"
            className="mt-1"
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>{t("del.cancel")}</Button>
            <Button
              className="bg-red-600 hover:bg-red-700 text-white"
              disabled={deleteConfirmText !== "удалить" || deleting}
              onClick={handleDeleteAll}
            >
              {deleting ? <><Loader2 size={14} className="animate-spin mr-2" />{t("del.deleting")}</> : t("del.delete")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function SortableHead({ column, label, sortBy, sortOrder, onSort }: {
  column: string; label: string; sortBy: string; sortOrder: string; onSort: (col: string) => void;
}) {
  const active = sortBy === column;
  return (
    <TableHead>
      <button
        onClick={() => onSort(column)}
        className="flex items-center gap-1 hover:text-slate-900 transition-colors"
      >
        {label}
        {active ? (
          sortOrder === "asc" ? <ArrowUp size={14} /> : <ArrowDown size={14} />
        ) : (
          <ArrowUpDown size={14} className="text-slate-300" />
        )}
      </button>
    </TableHead>
  );
}
