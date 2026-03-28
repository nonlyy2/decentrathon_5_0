"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { AuthProvider, useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

function LoginPage() {
  const [showAdmin, setShowAdmin] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [remember, setRemember] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(email, password);
      if (remember) {
        localStorage.setItem("remember_email", email);
      } else {
        localStorage.removeItem("remember_email");
      }
      router.push("/dashboard");
    } catch {
      setError("Invalid email or password");
    } finally {
      setLoading(false);
    }
  };

  // Load remembered email on mount
  useEffect(() => {
    const saved = localStorage.getItem("remember_email");
    if (saved) {
      setEmail(saved);
      setRemember(true);
    }
  }, []);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-slate-950 p-4">
      {!showAdmin ? (
        /* Applicant-first landing */
        <div className="w-full max-w-lg text-center space-y-8">
          <div>
            <h1 className="text-4xl font-bold text-purple-400 mb-2">inVision U</h1>
            <p className="text-slate-400 text-lg">100% Scholarship University by inDrive</p>
          </div>

          <Card className="border-slate-800 bg-slate-900">
            <CardContent className="p-8 space-y-4">
              <h2 className="text-xl font-semibold text-white">Apply Now</h2>
              <p className="text-slate-400 text-sm">
                Join the next generation of tech leaders in Central Asia.
                Submit your application for the 2026 cohort.
              </p>
              <Link href="/apply">
                <Button className="w-full bg-purple-600 hover:bg-purple-700 text-lg py-6">
                  Start Application
                </Button>
              </Link>
            </CardContent>
          </Card>

          <button
            onClick={() => setShowAdmin(true)}
            className="text-slate-600 text-xs hover:text-slate-400 transition-colors"
          >
            Admin Panel
          </button>
        </div>
      ) : (
        /* Admin login form */
        <div className="w-full max-w-md">
          <Card className="border-slate-800 bg-slate-900">
            <CardHeader className="text-center">
              <div className="text-3xl font-bold text-purple-400 mb-1">inVision U</div>
              <CardTitle className="text-slate-300 text-sm font-normal">
                Admissions Screening Platform
              </CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <Label htmlFor="email" className="text-slate-300">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="bg-slate-800 border-slate-700 text-white"
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="password" className="text-slate-300">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="bg-slate-800 border-slate-700 text-white"
                    required
                  />
                </div>
                <div className="flex items-center gap-2">
                  <input
                    id="remember"
                    type="checkbox"
                    checked={remember}
                    onChange={(e) => setRemember(e.target.checked)}
                    className="rounded border-slate-600 bg-slate-800 text-purple-600 focus:ring-purple-500"
                  />
                  <Label htmlFor="remember" className="text-slate-400 text-sm cursor-pointer">
                    Remember me
                  </Label>
                </div>
                {error && <p className="text-red-400 text-sm">{error}</p>}
                <Button
                  type="submit"
                  className="w-full bg-purple-600 hover:bg-purple-700"
                  disabled={loading}
                >
                  {loading ? "Signing in..." : "Sign In"}
                </Button>
              </form>
              <button
                onClick={() => setShowAdmin(false)}
                className="w-full text-center text-slate-500 text-xs mt-4 hover:text-slate-300 transition-colors"
              >
                Back to application
              </button>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

export default function Home() {
  return (
    <AuthProvider>
      <LoginPage />
    </AuthProvider>
  );
}
