"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";
import api from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { toast } from "sonner";
import { UserCog, Plus, Pencil, Trash2, Shield, Eye, EyeOff, KeyRound, Copy, Check } from "lucide-react";

interface UserItem {
  id: number;
  email: string;
  full_name: string | null;
  role: string;
  created_at: string;
}

const ROLE_LABELS: Record<string, string> = {
  superadmin: "Super Admin",
  admin: "Super Admin",
  "tech-admin": "Tech Admin",
  auditor: "Auditor",
  manager: "Manager",
  committee: "Manager",
};
const ROLE_COLORS: Record<string, string> = {
  superadmin: "bg-red-100 text-red-700",
  admin: "bg-red-100 text-red-700",
  "tech-admin": "bg-orange-100 text-orange-700",
  auditor: "bg-blue-100 text-blue-700",
  manager: "bg-green-100 text-green-700",
  committee: "bg-green-100 text-green-700",
};
const SELECTABLE_ROLES = ["manager", "auditor", "tech-admin", "superadmin"];
const ROLE_LEVEL: Record<string, number> = { committee: 1, manager: 1, auditor: 2, "tech-admin": 3, admin: 4, superadmin: 4 };

export default function UsersPage() {
  const { user } = useAuth();
  const { t } = useI18n();
  const [users, setUsers] = useState<UserItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [editUser, setEditUser] = useState<UserItem | null>(null);
  const [editRole, setEditRole] = useState("");
  const [editName, setEditName] = useState("");
  const [editEmail, setEditEmail] = useState("");
  const [saving, setSaving] = useState(false);
  const [deleteUser, setDeleteUser] = useState<UserItem | null>(null);
  const [resetPwUser, setResetPwUser] = useState<UserItem | null>(null);
  const [resetting, setResetting] = useState(false);
  const [resetResult, setResetResult] = useState<{ password?: string; warning?: string } | null>(null);
  const [copiedPw, setCopiedPw] = useState(false);
  const [registerOpen, setRegisterOpen] = useState(false);
  const [regForm, setRegForm] = useState({ email: "", password: "", role: "manager", full_name: "" });
  const [registering, setRegistering] = useState(false);
  const [showRegPw, setShowRegPw] = useState(false);

  const PW_RULES = [
    { test: (p: string) => p.length >= 8, text: "At least 8 characters" },
    { test: (p: string) => /[A-Z]/.test(p), text: "One uppercase letter (A-Z)" },
    { test: (p: string) => /[a-z]/.test(p), text: "One lowercase letter (a-z)" },
    { test: (p: string) => /[0-9]/.test(p), text: "One digit (0-9)" },
    { test: (p: string) => /[^A-Za-z0-9]/.test(p), text: "One special character (!@#$…)" },
  ];

  const myRole = user?.role ?? "manager";
  const myLevel = ROLE_LEVEL[myRole] ?? 0;

  const fetch = async () => {
    try {
      const r = await api.get<UserItem[]>("/users");
      setUsers(r.data);
    } catch {
      toast.error("Failed to load users");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetch(); }, []);

  const openEdit = (u: UserItem) => {
    setEditUser(u);
    setEditRole(u.role);
    setEditName(u.full_name ?? "");
    setEditEmail(u.email);
  };

  const handleSaveEdit = async () => {
    if (!editUser) return;
    setSaving(true);
    try {
      const body: Record<string, unknown> = { role: editRole, full_name: editName || null };
      // Only send email if superadmin and it changed
      if (myLevel >= 4 && editEmail && editEmail !== editUser.email) {
        body.email = editEmail;
      }
      await api.patch(`/users/${editUser.id}`, body);
      toast.success("User updated");
      setEditUser(null);
      fetch();
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
      toast.error(msg ?? "Failed to update");
    } finally {
      setSaving(false);
    }
  };

  const handleResetPassword = async () => {
    if (!resetPwUser) return;
    setResetting(true);
    setResetResult(null);
    try {
      const res = await api.post(`/users/${resetPwUser.id}/reset-password`);
      setResetResult({ password: res.data.new_password, warning: res.data.warning });
      if (!res.data.new_password) {
        toast.success("Password reset — new password emailed to user");
      }
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
      toast.error(msg ?? "Failed to reset password");
    } finally {
      setResetting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteUser) return;
    try {
      await api.delete(`/users/${deleteUser.id}`);
      toast.success("User deleted");
      setDeleteUser(null);
      fetch();
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
      toast.error(msg ?? "Failed to delete");
    }
  };

  const handleRegister = async () => {
    if (!regForm.email || !regForm.password) return;
    setRegistering(true);
    try {
      await api.post("/auth/register", regForm);
      toast.success("User created");
      setRegisterOpen(false);
      setRegForm({ email: "", password: "", role: "manager", full_name: "" });
      fetch();
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
      toast.error(msg ?? "Failed to create user");
    } finally {
      setRegistering(false);
    }
  };

  const canEdit = (target: UserItem) => {
    // Cannot edit superadmin unless you are superadmin
    if ((target.role === "superadmin" || target.role === "admin") && myLevel < 4) return false;
    // Cannot edit yourself here (use profile)
    if (target.id === user?.id) return false;
    return myLevel >= 3; // tech-admin+
  };

  const canDelete = (target: UserItem) => {
    if (target.role === "superadmin" || target.role === "admin") return false;
    if (target.id === user?.id) return false;
    return myLevel >= 4; // superadmin only
  };

  if (loading) {
    return (
      <div className="space-y-4">
        {[1,2,3].map(i => (
          <Card key={i}><CardContent className="p-4"><div className="h-12 bg-muted rounded animate-pulse"/></CardContent></Card>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground flex items-center gap-2">
          <UserCog size={22} /> {t("users.title")}
        </h1>
        {myLevel >= 3 && (
          <Button
            onClick={() => setRegisterOpen(true)}
            className="flex items-center gap-2 font-semibold"
            style={{ backgroundColor: "#c1f11d", color: "#111" }}
          >
            <Plus size={16} /> {t("users.add")}
          </Button>
        )}
      </div>

      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full text-sm" role="table">
              <thead>
                <tr className="border-b border-border text-muted-foreground text-xs uppercase">
                  <th className="text-left px-4 py-3 font-medium">{t("users.name")}</th>
                  <th className="text-left px-4 py-3 font-medium">{t("users.email")}</th>
                  <th className="text-left px-4 py-3 font-medium">{t("users.role")}</th>
                  <th className="text-left px-4 py-3 font-medium">{t("users.joined")}</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id} className="border-b border-border last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <div
                          className="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
                          style={{ backgroundColor: "#c1f11d", color: "#111" }}
                          aria-hidden="true"
                        >
                          {(u.full_name || u.email).charAt(0).toUpperCase()}
                        </div>
                        <span className="font-medium text-foreground">{u.full_name || "—"}</span>
                        {u.id === user?.id && (
                          <span className="text-[10px] bg-muted text-muted-foreground px-1.5 rounded">{t("users.you")}</span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{u.email}</td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${ROLE_COLORS[u.role] ?? "bg-muted text-muted-foreground"}`}>
                        <Shield size={10} className="inline mr-1" />
                        {ROLE_LABELS[u.role] ?? u.role}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground text-xs">
                      {new Date(u.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        {canEdit(u) && (
                          <button
                            onClick={() => openEdit(u)}
                            className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                            aria-label={`Edit ${u.email}`}
                            title="Edit user"
                          >
                            <Pencil size={14} />
                          </button>
                        )}
                        {myLevel >= 4 && u.id !== user?.id && u.role !== "superadmin" && u.role !== "admin" && (
                          <button
                            onClick={() => { setResetPwUser(u); setResetResult(null); }}
                            className="p-1.5 rounded-lg text-muted-foreground hover:text-amber-600 hover:bg-amber-50 transition-colors"
                            aria-label={`Reset password for ${u.email}`}
                            title="Reset password"
                          >
                            <KeyRound size={14} />
                          </button>
                        )}
                        {canDelete(u) && (
                          <button
                            onClick={() => setDeleteUser(u)}
                            className="p-1.5 rounded-lg text-muted-foreground hover:text-red-500 hover:bg-red-50 transition-colors"
                            aria-label={`Delete ${u.email}`}
                            title="Delete user"
                          >
                            <Trash2 size={14} />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Edit dialog */}
      <Dialog open={!!editUser} onOpenChange={(o) => !o && setEditUser(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("users.edit_title")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div>
              <Label>{t("users.name")}</Label>
              <Input value={editName} onChange={(e) => setEditName(e.target.value)} className="mt-1" />
            </div>
            {myLevel >= 4 && (
              <div>
                <Label>Email {<span className="text-xs text-muted-foreground ml-1">(superadmin only)</span>}</Label>
                <Input type="email" value={editEmail} onChange={(e) => setEditEmail(e.target.value)} className="mt-1" />
              </div>
            )}
            <div>
              <Label>{t("users.role")}</Label>
              <select
                value={editRole}
                onChange={(e) => setEditRole(e.target.value)}
                className="w-full mt-1 bg-background border border-input rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
              >
                {SELECTABLE_ROLES
                  .filter((r) => myLevel >= 4 || r !== "superadmin")
                  .map((r) => (
                    <option key={r} value={r}>{ROLE_LABELS[r] ?? r}</option>
                  ))}
              </select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditUser(null)}>{t("dec.cancel")}</Button>
            <Button onClick={handleSaveEdit} disabled={saving} style={{ backgroundColor: "#c1f11d", color: "#111" }}>
              {saving ? t("profile.saving") : t("profile.save")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={!!deleteUser} onOpenChange={(o) => !o && setDeleteUser(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-destructive">{t("users.delete_title")}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {t("users.delete_confirm")} <strong>{deleteUser?.email}</strong>?
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteUser(null)}>{t("dec.cancel")}</Button>
            <Button variant="destructive" onClick={handleDelete}>{t("users.delete_btn")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Reset Password dialog */}
      <Dialog open={!!resetPwUser} onOpenChange={(o) => { if (!o) { setResetPwUser(null); setResetResult(null); } }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <KeyRound size={16} /> Reset Password
            </DialogTitle>
          </DialogHeader>
          {!resetResult ? (
            <>
              <p className="text-sm text-muted-foreground">
                Generate a new random password for <strong>{resetPwUser?.email}</strong>.
                {" "}If SMTP is configured, it will be emailed automatically.
              </p>
              <DialogFooter>
                <Button variant="outline" onClick={() => setResetPwUser(null)}>Cancel</Button>
                <Button
                  className="bg-amber-500 hover:bg-amber-600 text-white"
                  onClick={handleResetPassword}
                  disabled={resetting}
                >
                  {resetting ? "Generating..." : "Generate & Send Password"}
                </Button>
              </DialogFooter>
            </>
          ) : (
            <div className="space-y-4">
              {resetResult.warning && (
                <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 text-sm text-amber-700">
                  <p className="font-medium mb-1">SMTP not configured</p>
                  <p className="text-xs">{resetResult.warning}</p>
                </div>
              )}
              {resetResult.password && (
                <div className="space-y-2">
                  <Label>New temporary password — copy and share manually:</Label>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 bg-muted rounded px-3 py-2 text-sm font-mono">{resetResult.password}</code>
                    <button
                      onClick={() => {
                        navigator.clipboard.writeText(resetResult!.password!);
                        setCopiedPw(true);
                        setTimeout(() => setCopiedPw(false), 2000);
                      }}
                      className="p-2 rounded hover:bg-muted transition-colors"
                      title="Copy password"
                    >
                      {copiedPw ? <Check size={16} className="text-green-500" /> : <Copy size={16} />}
                    </button>
                  </div>
                </div>
              )}
              <DialogFooter>
                <Button onClick={() => { setResetPwUser(null); setResetResult(null); }}>Done</Button>
              </DialogFooter>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Register new user */}
      <Dialog open={registerOpen} onOpenChange={setRegisterOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("users.add_title")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <div>
              <Label>{t("users.name")}</Label>
              <Input value={regForm.full_name} onChange={(e) => setRegForm({ ...regForm, full_name: e.target.value })} className="mt-1" />
            </div>
            <div>
              <Label>{t("users.email")}</Label>
              <Input type="email" value={regForm.email} onChange={(e) => setRegForm({ ...regForm, email: e.target.value })} className="mt-1" required />
            </div>
            <div>
              <Label>{t("users.password")}</Label>
              <div className="relative mt-1">
                <Input
                  type={showRegPw ? "text" : "password"}
                  value={regForm.password}
                  onChange={(e) => setRegForm({ ...regForm, password: e.target.value })}
                  className="pr-10"
                  autoComplete="new-password"
                  required
                />
                <button type="button" onClick={() => setShowRegPw(!showRegPw)} tabIndex={-1}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                  {showRegPw ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
              {regForm.password && (
                <ul className="mt-1.5 space-y-0.5">
                  {PW_RULES.map((r) => (
                    <li key={r.text} className={`text-xs flex items-center gap-1 ${r.test(regForm.password) ? "text-green-600 dark:text-green-400" : "text-muted-foreground"}`}>
                      <span>{r.test(regForm.password) ? "✓" : "○"}</span>{r.text}
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <div>
              <Label>{t("users.role")}</Label>
              <select
                value={regForm.role}
                onChange={(e) => setRegForm({ ...regForm, role: e.target.value })}
                className="w-full mt-1 bg-background border border-input rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              >
                {SELECTABLE_ROLES
                  .filter((r) => myLevel >= 4 || r !== "superadmin")
                  .map((r) => (
                    <option key={r} value={r}>{ROLE_LABELS[r] ?? r}</option>
                  ))}
              </select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRegisterOpen(false)}>{t("dec.cancel")}</Button>
            <Button onClick={handleRegister} disabled={registering} style={{ backgroundColor: "#c1f11d", color: "#111" }}>
              {registering ? t("profile.saving") : t("users.create")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
