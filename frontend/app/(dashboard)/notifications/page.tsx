"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import api from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Bell, CheckCheck, RefreshCw } from "lucide-react";

interface Notification {
  id: number;
  type: string;
  read: boolean;
  created_at: string;
  from_email: string | null;
  from_name: string | null;
  candidate_id: number | null;
  candidate_name: string | null;
  content: string | null;
  action_type: string | null;
}

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unread, setUnread] = useState(0);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetchNotifications = useCallback(async () => {
    try {
      const res = await api.get("/notifications");
      // Only show mention notifications
      const mentions = (res.data.notifications || []).filter((n: Notification) => n.type === "mention");
      setNotifications(mentions);
      setUnread(res.data.unread || 0);
    } catch { /* ignore */ } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchNotifications(); }, [fetchNotifications]);

  const markAllRead = async () => {
    await api.post("/notifications/read").catch(() => {});
    setUnread(0);
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    await fetchNotifications();
    setRefreshing(false);
  };

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Bell size={22} style={{ color: "#c1f11d" }} />
            Notifications
          </h1>
          <p className="text-muted-foreground text-sm mt-1">
            Posts and messages where you were @mentioned
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={refreshing}>
            <RefreshCw size={14} className={refreshing ? "animate-spin" : ""} />
          </Button>
          {unread > 0 && (
            <Button variant="outline" size="sm" onClick={markAllRead} className="flex items-center gap-1.5">
              <CheckCheck size={14} /> Mark all read
            </Button>
          )}
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12 text-muted-foreground">Loading...</div>
      ) : notifications.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Bell size={32} className="mx-auto text-muted-foreground/30 mb-3" />
            <p className="text-muted-foreground">No mentions yet.</p>
            <p className="text-xs text-muted-foreground mt-1">
              You will be notified here when someone tags you with @{" "}
              in the War Room.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {notifications.map((n) => (
            <Card
              key={n.id}
              className={`transition-colors ${!n.read ? "border-l-4" : ""}`}
              style={!n.read ? { borderLeftColor: "#c1f11d" } : undefined}
            >
              <CardContent className="p-4 space-y-2">
                <div className="flex items-start justify-between gap-3">
                  <div className="flex items-center gap-2 text-sm">
                    <div
                      className="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
                      style={{ backgroundColor: "#c1f11d", color: "#111" }}
                    >
                      {(n.from_name || n.from_email || "?").charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <span className="font-medium">{n.from_name || n.from_email || "Someone"}</span>
                      <span className="text-muted-foreground"> mentioned you</span>
                      {n.candidate_name && (
                        <span className="text-muted-foreground">
                          {" "}about{" "}
                          <Link href={`/candidates/${n.candidate_id}`} className="text-blue-500 hover:underline font-medium">
                            {n.candidate_name}
                          </Link>
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    {!n.read && (
                      <Badge className="text-[10px] px-1.5 py-0" style={{ backgroundColor: "#c1f11d", color: "#111" }}>New</Badge>
                    )}
                    <span className="text-xs text-muted-foreground">
                      {new Date(n.created_at).toLocaleString()}
                    </span>
                  </div>
                </div>
                {n.content && (
                  <p className="text-sm text-foreground bg-muted/50 rounded-lg px-3 py-2 pl-9 leading-relaxed">
                    {n.content}
                  </p>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
