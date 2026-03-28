"use client";

import { useState, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import api from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { useAIProvider } from "@/lib/aiProvider";
import { useAuth } from "@/lib/auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { toast } from "sonner";
import { Loader2, Play, Download, Trash2, ShieldAlert } from "lucide-react";

export default function AdminPage() {
  const router = useRouter();
  const { t, lang } = useI18n();
  const { provider, setProvider } = useAIProvider();
  const { user } = useAuth();

  const [batchRunning, setBatchRunning] = useState(false);
  const [batchProgress, setBatchProgress] = useState<{ done: number; total: number } | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteConfirmText, setDeleteConfirmText] = useState("");
  const [deleting, setDeleting] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

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
        }
      } catch {
        clearInterval(pollRef.current!);
        pollRef.current = null;
        setBatchRunning(false);
      }
    }, 3000);
  }, []);

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
    } catch {
      toast.error("Failed to delete analyses");
    } finally {
      setDeleting(false);
    }
  };

  if (user?.role !== "admin") {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <ShieldAlert size={48} className="text-red-400" />
        <p className="text-lg font-medium text-slate-600">Access denied</p>
        <Button variant="outline" onClick={() => router.push("/dashboard")}>Go to Dashboard</Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t("admin.title")}</h1>

      {/* Provider toggle */}
      <div className="flex items-center gap-3">
        <span className="text-sm text-slate-500">AI Provider:</span>
        <div className="flex items-center gap-1 bg-slate-100 border rounded-lg p-1">
          <button
            onClick={() => setProvider("gemini")}
            className={`text-xs px-3 py-1.5 rounded-md transition-colors ${
              provider === "gemini" ? "bg-purple-600 text-white shadow-sm" : "text-slate-500 hover:text-slate-800"
            }`}
          >
            ☁ Gemini
          </button>
          <button
            onClick={() => setProvider("ollama")}
            className={`text-xs px-3 py-1.5 rounded-md transition-colors ${
              provider === "ollama" ? "bg-purple-600 text-white shadow-sm" : "text-slate-500 hover:text-slate-800"
            }`}
          >
            ⚙ Ollama
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Batch Analysis */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Play size={16} className="text-purple-500" /> {t("admin.analyze_section")}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm text-slate-500">{t("admin.analyze_desc")}</p>
            {batchRunning && batchProgress && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm text-purple-600">
                  <Loader2 size={12} className="animate-spin" />
                  <span>{batchProgress.done}/{batchProgress.total}</span>
                </div>
                <div className="w-full bg-slate-200 rounded-full h-1.5">
                  <div
                    className="bg-purple-500 h-1.5 rounded-full transition-all"
                    style={{ width: `${batchProgress.total > 0 ? (batchProgress.done / batchProgress.total) * 100 : 0}%` }}
                  />
                </div>
              </div>
            )}
            <Button
              className="w-full bg-purple-600 hover:bg-purple-700"
              onClick={handleAnalyzeAll}
              disabled={batchRunning}
            >
              {batchRunning
                ? <><Loader2 size={14} className="animate-spin mr-2" />{t("cand.running")}</>
                : t("cand.analyze_all")}
            </Button>
          </CardContent>
        </Card>

        {/* Export */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Download size={16} className="text-blue-500" /> {t("admin.export_section")}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm text-slate-500">{t("admin.export_desc")}</p>
            <Button variant="outline" className="w-full" onClick={handleExportCSV}>
              <Download size={14} className="mr-2" /> {t("cand.export")}
            </Button>
          </CardContent>
        </Card>

        {/* Reset */}
        <Card className="border-red-100">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2 text-red-600">
              <Trash2 size={16} /> {t("admin.reset_section")}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm text-slate-500">{t("admin.reset_desc")}</p>
            <Button
              variant="outline"
              className="w-full text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200"
              onClick={() => { setDeleteConfirmText(""); setDeleteDialogOpen(true); }}
              disabled={batchRunning}
            >
              <Trash2 size={14} className="mr-2" /> {t("cand.reset")}
            </Button>
          </CardContent>
        </Card>
      </div>

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
