"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";
import { useTheme } from "@/lib/theme";
import api from "@/lib/api";
import {
  LayoutDashboard, Users, LogOut, ChevronLeft,
  ChevronRight, UserCog, Moon, Sun, Menu, Flame, Bell, BarChart2,
  StickyNote, ClipboardList,
} from "lucide-react";

// Role level helpers
const ROLE_LEVEL: Record<string, number> = {
  committee: 1, manager: 1, auditor: 2, "tech-admin": 3, admin: 4, superadmin: 4,
};
function hasLevel(role: string, required: string) {
  return (ROLE_LEVEL[role] ?? 0) >= (ROLE_LEVEL[required] ?? 0);
}
function isAuditorOnly(role: string) {
  return role === "auditor";
}

export default function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const { lang, setLang, t } = useI18n();
  const { theme, toggleTheme } = useTheme();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  // Persist collapse state
  useEffect(() => {
    const saved = localStorage.getItem("sidebar_collapsed");
    if (saved === "true") setCollapsed(true);
  }, []);
  const toggleCollapse = () => {
    const next = !collapsed;
    setCollapsed(next);
    localStorage.setItem("sidebar_collapsed", String(next));
  };

  const role = user?.role ?? "manager";
  const auditorOnly = isAuditorOnly(role);

  // Notification badge count
  const [unreadCount, setUnreadCount] = useState(0);
  useEffect(() => {
    const fetchUnread = () => {
      api.get("/notifications").then((r) => setUnreadCount(r.data.unread || 0)).catch(() => {});
    };
    fetchUnread();
    const interval = setInterval(fetchUnread, 30000);
    return () => clearInterval(interval);
  }, []);

  // Navigation items based on role
  const navItems = [
    {
      href: "/dashboard",
      label: t("nav.dashboard"),
      icon: LayoutDashboard,
      show: true,
    },
    {
      href: "/candidates",
      label: t("nav.candidates"),
      icon: Users,
      show: !auditorOnly,
    },
    {
      href: "/war-room",
      label: "War Room",
      icon: Flame,
      show: !auditorOnly,
    },
    {
      href: "/analytics",
      label: "Analytics",
      icon: BarChart2,
      show: true,
    },
    {
      href: "/tasks",
      label: "Tasks",
      icon: ClipboardList,
      show: !auditorOnly && role !== "tech-admin",
    },
    {
      href: "/my-notes",
      label: "My Notes",
      icon: StickyNote,
      show: !auditorOnly && role !== "tech-admin",
    },
    {
      href: "/users",
      label: t("nav.users"),
      icon: UserCog,
      show: hasLevel(role, "tech-admin"),
    },
  ].filter((item) => item.show);

  const sidebarWidth = collapsed ? "w-[72px]" : "w-64";

  const sidebarContent = (
    <div
      className={`
        flex flex-col h-full
        bg-sidebar-bg text-sidebar-text
        border-r border-sidebar-border
        transition-[width] duration-250 ease-[cubic-bezier(0.4,0,0.2,1)]
        ${sidebarWidth}
        overflow-hidden
      `}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-5 border-b border-sidebar-border min-h-[64px]">
        {!collapsed && (
          <div className="animate-fade-in overflow-hidden whitespace-nowrap">
            <h1 className="text-base font-bold" style={{ color: "#c1f11d" }}>
              inVision U
            </h1>
            <p className="text-[10px] text-sidebar-muted mt-0.5">Admissions AI</p>
          </div>
        )}
        <button
          onClick={toggleCollapse}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          className="p-1.5 rounded-lg text-sidebar-muted hover:text-sidebar-text hover:bg-sidebar-hover transition-colors ml-auto"
        >
          {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-2 space-y-0.5 overflow-y-auto overflow-x-hidden" role="navigation" aria-label="Main navigation">
        {navItems.map((item) => {
          const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={isActive ? "page" : undefined}
              title={collapsed ? item.label : undefined}
              className={`
                flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors
                focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-lime-brand focus-visible:ring-offset-1
                ${isActive
                  ? "text-sidebar-bg font-semibold"
                  : "text-sidebar-muted hover:text-sidebar-text hover:bg-sidebar-hover"
                }
              `}
              style={isActive ? { backgroundColor: "#c1f11d", color: "#111827" } : undefined}
            >
              <item.icon size={18} className="shrink-0" />
              {!collapsed && (
                <span className="animate-fade-in overflow-hidden whitespace-nowrap">
                  {item.label}
                </span>
              )}
            </Link>
          );
        })}
      </nav>

      {/* Bottom section */}
      <div className="p-3 border-t border-sidebar-border space-y-2">
        {/* Avatar + Role badge → click to profile */}
        <Link
          href="/profile"
          className={`flex items-center gap-2.5 px-2 py-2 rounded-lg hover:bg-sidebar-hover transition-colors ${collapsed ? "justify-center" : ""}`}
          title={collapsed ? t("nav.profile") : undefined}
        >
          <div
            className="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
            style={{ backgroundColor: "#c1f11d", color: "#111" }}
          >
            {(user?.full_name || user?.email || "U").charAt(0).toUpperCase()}
          </div>
          {!collapsed && (
            <div className="animate-fade-in overflow-hidden min-w-0">
              <p className="text-[11px] text-sidebar-muted truncate">{user?.email}</p>
              <RoleBadge role={role} />
            </div>
          )}
        </Link>

        {/* Language toggle */}
        <div
          className={`flex items-center ${collapsed ? "flex-col gap-1" : "gap-1 bg-sidebar-hover rounded-lg p-1"}`}
          role="group"
          aria-label="Language selection"
        >
          {(["en", "ru", "kk"] as const).map((l) => (
            <button
              key={l}
              onClick={() => setLang(l)}
              aria-label={`Switch to ${l.toUpperCase()}`}
              aria-pressed={lang === l}
              className={`
                flex-1 text-[11px] py-1.5 rounded-md transition-colors
                focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-lime-brand
                ${collapsed ? "px-2" : ""}
                ${lang === l
                  ? "text-sidebar-bg font-semibold"
                  : "text-sidebar-muted hover:text-sidebar-text"
                }
              `}
              style={lang === l ? { backgroundColor: "#c1f11d", color: "#111827" } : undefined}
            >
              {l.toUpperCase()}
            </button>
          ))}
        </div>

        {/* Theme toggle */}
        <button
          onClick={toggleTheme}
          aria-label={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
          title={collapsed ? (theme === "dark" ? "Light mode" : "Dark mode") : undefined}
          className="w-full flex items-center gap-2.5 px-2 py-2 rounded-lg text-sidebar-muted hover:text-sidebar-text hover:bg-sidebar-hover transition-colors text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-lime-brand"
        >
          {theme === "dark" ? <Sun size={15} className="shrink-0" /> : <Moon size={15} className="shrink-0" />}
          {!collapsed && (
            <span className="animate-fade-in whitespace-nowrap text-[12px]">
              {theme === "dark" ? t("theme.light") : t("theme.dark")}
            </span>
          )}
        </button>

        {/* Notifications */}
        <Link
          href="/notifications"
          className="w-full flex items-center gap-2.5 px-2 py-2 rounded-lg text-sidebar-muted hover:text-sidebar-text hover:bg-sidebar-hover transition-colors text-sm relative"
          title={collapsed ? "Notifications" : undefined}
          onClick={() => {
            api.post("/notifications/read").catch(() => {});
            setUnreadCount(0);
          }}
        >
          <div className="relative shrink-0">
            <Bell size={15} />
            {unreadCount > 0 && (
              <span className="absolute -top-1 -right-1 w-3.5 h-3.5 rounded-full text-[9px] flex items-center justify-center font-bold text-black"
                style={{ backgroundColor: "#c1f11d" }}>
                {unreadCount > 9 ? "9+" : unreadCount}
              </span>
            )}
          </div>
          {!collapsed && (
            <span className="animate-fade-in whitespace-nowrap text-[12px] flex-1">
              Notifications {unreadCount > 0 && <span className="ml-1 px-1 rounded text-[10px] font-bold text-black" style={{ backgroundColor: "#c1f11d" }}>{unreadCount}</span>}
            </span>
          )}
        </Link>

        {/* Sign out */}
        <button
          onClick={logout}
          aria-label="Sign out"
          title={collapsed ? t("nav.signout") : undefined}
          className="w-full flex items-center gap-2.5 px-2 py-2 rounded-lg text-sidebar-muted hover:text-red-400 hover:bg-sidebar-hover transition-colors text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-500"
        >
          <LogOut size={15} className="shrink-0" />
          {!collapsed && (
            <span className="animate-fade-in whitespace-nowrap text-[12px]">{t("nav.signout")}</span>
          )}
        </button>
      </div>
    </div>
  );

  return (
    <>
      {/* Desktop sidebar */}
      <aside
        className={`hidden md:flex flex-col fixed left-0 top-0 h-screen z-30 ${sidebarWidth} sidebar-transition`}
        aria-label="Sidebar navigation"
      >
        {sidebarContent}
      </aside>

      {/* Mobile hamburger button */}
      <button
        className="md:hidden fixed top-4 left-4 z-50 p-2 rounded-lg bg-sidebar-bg text-sidebar-text shadow-lg"
        onClick={() => setMobileOpen(true)}
        aria-label="Open navigation menu"
        aria-expanded={mobileOpen}
      >
        <Menu size={20} />
      </button>

      {/* Mobile drawer overlay */}
      {mobileOpen && (
        <div
          className="md:hidden fixed inset-0 z-40 flex"
          role="dialog"
          aria-modal="true"
          aria-label="Navigation menu"
        >
          <div
            className="fixed inset-0 bg-black/50 transition-opacity"
            onClick={() => setMobileOpen(false)}
            aria-hidden="true"
          />
          <aside className="relative flex flex-col w-64 h-full z-50 animate-slide-in">
            {sidebarContent}
          </aside>
        </div>
      )}
    </>
  );
}

function RoleBadge({ role }: { role: string }) {
  const colors: Record<string, string> = {
    superadmin: "bg-red-500/20 text-red-300",
    admin: "bg-red-500/20 text-red-300",
    "tech-admin": "bg-orange-500/20 text-orange-300",
    auditor: "bg-blue-500/20 text-blue-300",
    manager: "bg-green-500/20 text-green-300",
    committee: "bg-green-500/20 text-green-300",
  };
  const labels: Record<string, string> = {
    superadmin: "Super Admin",
    admin: "Super Admin",
    "tech-admin": "Tech Admin",
    auditor: "Auditor",
    manager: "Manager",
    committee: "Manager",
  };
  return (
    <span
      className={`inline-block mt-1 text-[10px] px-1.5 py-0.5 rounded font-medium ${colors[role] ?? "bg-slate-700 text-slate-300"}`}
    >
      {labels[role] ?? role}
    </span>
  );
}
