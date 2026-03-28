"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { LayoutDashboard, Users, BarChart3, LogOut } from "lucide-react";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/candidates", label: "Candidates", icon: Users },
  { href: "/analytics", label: "Analytics", icon: BarChart3 },
];

export default function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  return (
    <div className="w-64 h-screen bg-slate-900 text-white flex flex-col fixed left-0 top-0">
      <div className="p-6 border-b border-slate-800">
        <h1 className="text-xl font-bold text-purple-400">inVision U</h1>
        <p className="text-xs text-slate-400 mt-1">Admissions AI</p>
      </div>

      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => {
          const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                isActive
                  ? "bg-slate-800 text-purple-400"
                  : "text-slate-300 hover:bg-slate-800/50"
              }`}
            >
              <item.icon size={18} />
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="p-4 border-t border-slate-800">
        <div className="text-sm text-slate-400 mb-2 truncate">{user?.email}</div>
        <div className="text-xs text-slate-500 mb-3 capitalize">{user?.role}</div>
        <button
          onClick={logout}
          className="flex items-center gap-2 text-sm text-slate-400 hover:text-white transition-colors"
        >
          <LogOut size={16} />
          Sign Out
        </button>
      </div>
    </div>
  );
}
