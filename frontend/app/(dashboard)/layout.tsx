"use client";

import { AuthProvider } from "@/lib/auth";
import { I18nProvider } from "@/lib/i18n";
import { AIProviderContextProvider } from "@/lib/aiProvider";
import { ThemeProvider } from "@/lib/theme";
import ProtectedRoute from "@/components/ProtectedRoute";
import Sidebar from "@/components/Sidebar";
import AIAssistant from "@/components/AIAssistant";
import { useEffect, useState } from "react";

function MainContent({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    const check = () => {
      const saved = localStorage.getItem("sidebar_collapsed");
      setCollapsed(saved === "true");
    };
    check();
    // Listen for storage changes from Sidebar
    window.addEventListener("storage", check);
    // Also poll briefly after mount for same-tab changes
    const timer = setInterval(check, 300);
    return () => {
      window.removeEventListener("storage", check);
      clearInterval(timer);
    };
  }, []);

  return (
    <main
      className={`
        flex-1 bg-background p-6 overflow-auto min-h-screen
        transition-[margin] duration-250 ease-[cubic-bezier(0.4,0,0.2,1)]
        hidden md:block
        ${collapsed ? "md:ml-[72px]" : "md:ml-64"}
      `}
    >
      {children}
    </main>
  );
}

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider>
      <AuthProvider>
        <I18nProvider>
          <AIProviderContextProvider>
            <ProtectedRoute>
              <div className="flex min-h-screen bg-background">
                <Sidebar />
                {/* Mobile main — no margin */}
                <main className="flex-1 bg-background p-4 overflow-auto min-h-screen md:hidden">
                  {children}
                </main>
                <MainContent>{children}</MainContent>
                <AIAssistant />
              </div>
            </ProtectedRoute>
          </AIProviderContextProvider>
        </I18nProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}
