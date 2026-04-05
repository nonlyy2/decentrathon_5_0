"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import api from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";
import { Loader2, RefreshCw, CheckCircle, Clock, PlayCircle, ClipboardList, User } from "lucide-react";

interface ReviewTask {
  id: number;
  candidate_id: number;
  candidate_name: string;
  candidate_email: string;
  assigned_to: number | null;
  assignee_name: string | null;
  assignee_email: string | null;
  status: string;
  review_complexity: number | null;
  created_at: string;
  completed_at: string | null;
}

export default function TasksPage() {
  const { user } = useAuth();
  const [tasks, setTasks] = useState<ReviewTask[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<"all" | "mine" | "unassigned">("mine");
  const [assigning, setAssigning] = useState(false);

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.get(`/tasks?filter=${filter}`);
      setTasks(res.data.tasks || []);
    } catch {
      toast.error("Failed to fetch tasks");
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => { fetchTasks(); }, [fetchTasks]);

  const handleAssign = async () => {
    setAssigning(true);
    try {
      const res = await api.post("/tasks/assign");
      toast.success(`${res.data.message} — ${res.data.assigned} tasks assigned`);
      fetchTasks();
    } catch {
      toast.error("Failed to assign tasks");
    } finally {
      setAssigning(false);
    }
  };

  const handleUpdateStatus = async (taskId: number, status: string) => {
    try {
      await api.patch(`/tasks/${taskId}`, { status });
      toast.success(`Task marked as ${status.replace("_", " ")}`);
      fetchTasks();
    } catch {
      toast.error("Failed to update task");
    }
  };

  const isAdmin = user?.role === "admin" || user?.role === "superadmin";

  const pending = tasks.filter(t => t.status === "pending").length;
  const inProgress = tasks.filter(t => t.status === "in_progress").length;
  const completed = tasks.filter(t => t.status === "completed").length;

  const statusBadge = (s: string) => {
    if (s === "pending") return (
      <Badge className="bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300 text-xs flex items-center gap-1 w-fit">
        <Clock size={11} /> Pending
      </Badge>
    );
    if (s === "in_progress") return (
      <Badge className="bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300 text-xs flex items-center gap-1 w-fit">
        <PlayCircle size={11} /> In Progress
      </Badge>
    );
    return (
      <Badge className="bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300 text-xs flex items-center gap-1 w-fit">
        <CheckCircle size={11} /> Completed
      </Badge>
    );
  };

  const complexityBadge = (c: number | null) => {
    if (c == null) return <span className="text-xs text-muted-foreground">—</span>;
    let cls = "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300";
    if (c > 60) cls = "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300";
    else if (c > 35) cls = "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300";
    return <Badge className={`${cls} text-xs`}>{c.toFixed(0)}</Badge>;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between flex-wrap gap-3">
        <div className="flex items-center gap-3">
          <ClipboardList size={22} className="text-muted-foreground" />
          <div>
            <h1 className="text-2xl font-bold text-foreground">Review Tasks</h1>
            <p className="text-sm text-muted-foreground">Candidates are auto-assigned to managers via round-robin</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={fetchTasks} disabled={loading}>
            <RefreshCw size={14} className={`mr-1 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          {isAdmin && (
            <Button size="sm" onClick={handleAssign} disabled={assigning}
              style={{ backgroundColor: "#c1f11d", color: "#111" }}>
              {assigning
                ? <><Loader2 size={14} className="animate-spin mr-1" /> Assigning…</>
                : <><RefreshCw size={14} className="mr-1" /> Assign Unassigned</>}
            </Button>
          )}
        </div>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: "Pending", count: pending, color: "text-yellow-600", bg: "bg-yellow-50 dark:bg-yellow-900/20", icon: Clock },
          { label: "In Progress", count: inProgress, color: "text-blue-600", bg: "bg-blue-50 dark:bg-blue-900/20", icon: PlayCircle },
          { label: "Completed", count: completed, color: "text-green-600", bg: "bg-green-50 dark:bg-green-900/20", icon: CheckCircle },
        ].map(({ label, count, color, bg, icon: Icon }) => (
          <Card key={label} className={`${bg} border-0`}>
            <CardContent className="p-4 flex items-center gap-3">
              <Icon size={20} className={color} />
              <div>
                <p className="text-2xl font-bold">{count}</p>
                <p className="text-xs text-muted-foreground">{label}</p>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Filter tabs */}
      <div className="flex gap-1 bg-muted rounded-lg p-1 w-fit">
        {(["mine", "all", "unassigned"] as const).map((f) => (
          <button key={f} onClick={() => setFilter(f)}
            className={`px-4 py-1.5 rounded-md text-sm font-medium transition-colors ${
              filter === f ? "bg-background shadow text-foreground" : "text-muted-foreground hover:text-foreground"
            }`}>
            {f === "mine" ? "My Tasks" : f === "all" ? "All Tasks" : "Unassigned"}
          </button>
        ))}
      </div>

      {/* Table */}
      {loading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="animate-spin text-muted-foreground" size={32} />
        </div>
      ) : tasks.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground border rounded-lg">
          <ClipboardList size={36} className="mx-auto mb-3 opacity-30" />
          <p className="font-medium">No tasks found</p>
          {filter === "mine" && (
            <p className="text-sm mt-1">New candidate applications will be automatically assigned to you.</p>
          )}
          {filter === "unassigned" && isAdmin && (
            <Button size="sm" className="mt-4" onClick={handleAssign} disabled={assigning}
              style={{ backgroundColor: "#c1f11d", color: "#111" }}>
              Assign All Now
            </Button>
          )}
        </div>
      ) : (
        <div className="border rounded-lg overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Candidate</TableHead>
                <TableHead>Assigned To</TableHead>
                <TableHead className="text-center">Complexity</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Received</TableHead>
                <TableHead>Completed</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tasks.map((task) => (
                <TableRow key={task.id}>
                  <TableCell>
                    <Link href={`/candidates/${task.candidate_id}`}
                      className="text-primary hover:underline font-medium text-sm">
                      {task.candidate_name}
                    </Link>
                    <p className="text-xs text-muted-foreground">{task.candidate_email}</p>
                  </TableCell>
                  <TableCell>
                    {task.assignee_name || task.assignee_email ? (
                      <div className="flex items-center gap-1.5 text-sm">
                        <User size={13} className="text-muted-foreground shrink-0" />
                        <span>{task.assignee_name || task.assignee_email}</span>
                      </div>
                    ) : (
                      <span className="text-xs text-muted-foreground italic">Unassigned</span>
                    )}
                  </TableCell>
                  <TableCell className="text-center">{complexityBadge(task.review_complexity)}</TableCell>
                  <TableCell>{statusBadge(task.status)}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {new Date(task.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {task.completed_at ? new Date(task.completed_at).toLocaleDateString() : "—"}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      {task.status === "pending" && (
                        <Button size="sm" variant="outline" className="text-xs h-7"
                          onClick={() => handleUpdateStatus(task.id, "in_progress")}>
                          Start
                        </Button>
                      )}
                      {task.status === "in_progress" && (
                        <Button size="sm" variant="outline"
                          className="text-xs h-7 border-green-300 text-green-700 hover:bg-green-50"
                          onClick={() => handleUpdateStatus(task.id, "completed")}>
                          Complete
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
