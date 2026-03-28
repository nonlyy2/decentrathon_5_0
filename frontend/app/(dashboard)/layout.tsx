"use client";

import { AuthProvider } from "@/lib/auth";
import { I18nProvider } from "@/lib/i18n";
import ProtectedRoute from "@/components/ProtectedRoute";
import Sidebar from "@/components/Sidebar";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <I18nProvider>
        <ProtectedRoute>
          <div className="flex min-h-screen">
            <Sidebar />
            <main className="flex-1 ml-64 bg-slate-50 p-6 overflow-auto">
              {children}
            </main>
          </div>
        </ProtectedRoute>
      </I18nProvider>
    </AuthProvider>
  );
}
