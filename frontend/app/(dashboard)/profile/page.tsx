"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";
import api from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { User, KeyRound, Shield } from "lucide-react";

interface Profile {
  id: number;
  email: string;
  full_name: string | null;
  role: string;
  avatar_url: string | null;
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
  superadmin: "bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300",
  admin: "bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300",
  "tech-admin": "bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300",
  auditor: "bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300",
  manager: "bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300",
  committee: "bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300",
};

export default function ProfilePage() {
  useAuth();
  const { t } = useI18n();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPw, setConfirmPw] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    api.get<Profile>("/profile").then((r) => {
      setProfile(r.data);
      setFullName(r.data.full_name ?? "");
      setEmail(r.data.email);
    }).catch(() => {});
  }, []);

  const handleSave = async () => {
    if (password && password !== confirmPw) {
      toast.error("Passwords do not match");
      return;
    }
    if (password && password.length < 6) {
      toast.error("Password must be at least 6 characters");
      return;
    }

    setSaving(true);
    try {
      const isSuperAdmin = profile?.role === "superadmin" || profile?.role === "admin";
      await api.patch("/profile", {
        full_name: fullName || null,
        email: isSuperAdmin && email !== profile?.email ? email : undefined,
        password: password || undefined,
      });
      toast.success(t("profile.saved"));
      setPassword("");
      setConfirmPw("");
    } catch {
      toast.error(t("profile.save_error"));
    } finally {
      setSaving(false);
    }
  };

  if (!profile) {
    return (
      <div className="space-y-4">
        {[1, 2].map((i) => (
          <Card key={i}>
            <CardContent className="p-6">
              <div className="h-24 bg-muted rounded animate-pulse" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  const role = profile.role;

  return (
    <div className="max-w-xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold text-foreground">{t("nav.profile")}</h1>

      {/* Account info */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-base flex items-center gap-2">
            <User size={16} /> {t("profile.account_info")}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Avatar placeholder + role badge */}
          <div className="flex items-center gap-4">
            <div
              className="w-16 h-16 rounded-full flex items-center justify-center text-xl font-bold shrink-0"
              style={{ backgroundColor: "#c1f11d", color: "#111" }}
              aria-label="User avatar"
            >
              {(profile.full_name || profile.email).charAt(0).toUpperCase()}
            </div>
            <div>
              <p className="font-semibold text-foreground">{profile.full_name || profile.email}</p>
              <p className="text-sm text-muted-foreground">{profile.email}</p>
              <span className={`text-xs px-2 py-0.5 rounded-full font-medium mt-1 inline-block ${ROLE_COLORS[role] ?? "bg-muted text-muted-foreground"}`}>
                <Shield size={10} className="inline mr-1" />
                {ROLE_LABELS[role] ?? role}
              </span>
            </div>
          </div>

          <div className="text-xs text-muted-foreground pt-1">
            {t("profile.member_since")}: {new Date(profile.created_at).toLocaleDateString()}
          </div>

          {/* Full name field */}
          <div>
            <Label htmlFor="full_name" className="text-sm">{t("profile.full_name")}</Label>
            <Input
              id="full_name"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              placeholder={t("profile.full_name_ph")}
              className="mt-1"
            />
          </div>

          {/* Email — editable for superadmin */}
          <div>
            <Label className="text-sm">{t("profile.email")}</Label>
            {role === "superadmin" || role === "admin" ? (
              <Input
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1"
                type="email"
              />
            ) : (
              <Input value={profile.email} readOnly className="mt-1 opacity-60 cursor-not-allowed" aria-readonly />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Change password */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-base flex items-center gap-2">
            <KeyRound size={16} /> {t("profile.change_password")}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <Label htmlFor="new_password" className="text-sm">{t("profile.new_password")}</Label>
            <Input
              id="new_password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              autoComplete="new-password"
              className="mt-1"
            />
          </div>
          <div>
            <Label htmlFor="confirm_password" className="text-sm">{t("profile.confirm_password")}</Label>
            <Input
              id="confirm_password"
              type="password"
              value={confirmPw}
              onChange={(e) => setConfirmPw(e.target.value)}
              placeholder="••••••••"
              autoComplete="new-password"
              className="mt-1"
            />
            {password && confirmPw && password !== confirmPw && (
              <p className="text-xs text-red-500 mt-1" role="alert">{t("profile.pw_mismatch")}</p>
            )}
          </div>
        </CardContent>
      </Card>

      <Button
        onClick={handleSave}
        disabled={saving}
        className="w-full font-semibold"
        style={{ backgroundColor: "#c1f11d", color: "#111" }}
      >
        {saving ? t("profile.saving") : t("profile.save")}
      </Button>
    </div>
  );
}
