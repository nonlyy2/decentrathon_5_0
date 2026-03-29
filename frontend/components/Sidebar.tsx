"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";
import { LayoutDashboard, Users, LogOut } from "lucide-react";

export default function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const { lang, setLang, t } = useI18n();

  const navItems = [
    { href: "/dashboard", label: t("nav.dashboard"), icon: LayoutDashboard },
    { href: "/candidates", label: t("nav.candidates"), icon: Users },
  ];

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

      <div className="p-4 border-t border-slate-800 space-y-3">

        {/* Language toggle */}
        <div>
          <p className="text-xs text-slate-500 mb-1.5 uppercase tracking-wide">Language</p>
          <div className="flex items-center gap-1 bg-slate-800 rounded-lg p-1">
            <button
              onClick={() => setLang("en")}
              className={`flex-1 text-xs py-1.5 rounded-md transition-colors flex items-center justify-center gap-1 ${
                lang === "en" ? "bg-purple-600 text-white" : "text-slate-400 hover:text-white"
              }`}
            >
              🇬🇧 EN
            </button>
            <button
              onClick={() => setLang("ru")}
              className={`flex-1 text-xs py-1.5 rounded-md transition-colors flex items-center justify-center gap-1 ${
                lang === "ru" ? "bg-purple-600 text-white" : "text-slate-400 hover:text-white"
              }`}
            >
              🇷🇺 RU
            </button>
            <button
              onClick={() => setLang("kk")}
              className={`flex-1 text-xs py-1.5 rounded-md transition-colors flex items-center justify-center gap-1 ${
                lang === "kk" ? "bg-purple-600 text-white" : "text-slate-400 hover:text-white"
              }`}
            >
              🇰🇿 KZ
            </button>
          </div>
        </div>

        <div className="text-sm text-slate-400 truncate">{user?.email}</div>
        <div className="text-xs text-slate-500 capitalize">{user?.role}</div>
        <button
          onClick={logout}
          className="flex items-center gap-2 text-sm text-slate-400 hover:text-white transition-colors"
        >
          <LogOut size={16} />
          {t("nav.signout")}
        </button>
      </div>
    </div>
  );
}
