"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { AuthProvider, useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ChevronDown, ChevronUp, Eye, EyeOff, Accessibility } from "lucide-react";
import { getAccessibilityMode, applyAccessibilityMode } from "@/lib/accessibility";

const FAQ_ITEMS = [
  {
    q: "What is inVision U?",
    a: "inVision U is a 100% scholarship university by inDrive, focused on developing the next generation of tech leaders in Central Asia.",
  },
  {
    q: "How does the application process work?",
    a: "Submit your application online (essay, achievements, motivation). Our AI evaluates applications in Stage 1. Top candidates are invited to a Telegram-based interview in Stage 2. A committee makes the final decision.",
  },
  {
    q: "What are the available majors?",
    a: "Creative Engineering, Innovative IT Product Design and Development, Sociology: Leadership and Innovation, Public Policy and Development, and Digital Media and Marketing.",
  },
  {
    q: "Is the interview conducted by a real person?",
    a: "Stage 2 uses an AI-powered interviewer via Telegram. It asks behavioral questions using the STAR method. You can respond with voice or text messages.",
  },
  {
    q: "How is my application scored?",
    a: "AI evaluates you on Leadership (25%), Motivation (25%), Growth (20%), Vision (15%), and Communication (15%). Interview adds Grit and Authenticity scores. Combined: 60% essay + 40% interview.",
  },
  {
    q: "Is my data private?",
    a: "Yes. Your personal information (name, email, age) is excluded from AI analysis. Only your essay and interview responses are evaluated.",
  },
];

function LoginPage() {
  const [showAdmin, setShowAdmin] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [remember, setRemember] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [openFaq, setOpenFaq] = useState<number | null>(null);
  const [a11y, setA11y] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  useEffect(() => { setA11y(getAccessibilityMode()); }, []);

  const toggleA11y = () => {
    const next = !a11y;
    setA11y(next);
    applyAccessibilityMode(next);
  };

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

  useEffect(() => {
    const saved = localStorage.getItem("remember_email");
    if (saved) {
      setEmail(saved);
      setRemember(true);
    }
  }, []);

  return (
    <div className="min-h-screen flex flex-col items-center bg-slate-950 p-4">
      {/* Accessibility button — always visible top-right */}
      <div className="fixed top-3 right-3 z-50">
        <button
          onClick={toggleA11y}
          className={`flex items-center gap-2 px-3 py-2 rounded-full text-sm font-medium border transition-colors ${
            a11y
              ? "bg-white text-black border-white"
              : "bg-slate-800 text-slate-300 border-slate-600 hover:border-white hover:text-white"
          }`}
          aria-pressed={a11y}
          title="Режим для слабовидящих"
        >
          <Accessibility size={15} />
          {a11y ? "Обычный режим" : "Перейти в режим для слабовидящих"}
        </button>
      </div>

      {!showAdmin ? (
        <div className="w-full max-w-2xl space-y-10 py-12">
          {/* Hero */}
          <div className="text-center space-y-3">
            <h1 className="text-5xl font-bold" style={{ color: "#c1f11d" }}>inVision U</h1>
            <p className="text-slate-400 text-lg">100% Scholarship University by <span style={{ color: "#c1f11d" }}>inDrive</span></p>
          </div>

          {/* Apply CTA */}
          <Card className="border-slate-800 bg-slate-900">
            <CardContent className="p-8 space-y-4 text-center">
              <h2 className="text-xl font-semibold text-white">Apply Now</h2>
              <p className="text-slate-400 text-sm">
                Join the next generation of tech leaders in Central Asia.
                Submit your application for the 2026 cohort.
              </p>
              <Link href="/apply">
                <Button
                  className="w-full text-lg py-6 font-bold"
                  style={{ backgroundColor: "#c1f11d", color: "#111" }}
                >
                  Start Application
                </Button>
              </Link>
            </CardContent>
          </Card>

          {/* FAQ Section */}
          <div className="space-y-4">
            <h2 className="text-xl font-semibold text-white text-center">Frequently Asked Questions</h2>
            <div className="space-y-2">
              {FAQ_ITEMS.map((item, i) => (
                <div key={i} className="border border-slate-800 rounded-lg overflow-hidden">
                  <button
                    onClick={() => setOpenFaq(openFaq === i ? null : i)}
                    className="w-full flex items-center justify-between px-5 py-4 text-left text-white hover:bg-slate-800/50 transition-colors"
                  >
                    <span className="text-sm font-medium">{item.q}</span>
                    {openFaq === i ? <ChevronUp size={16} className="text-slate-400 shrink-0" /> : <ChevronDown size={16} className="text-slate-400 shrink-0" />}
                  </button>
                  {openFaq === i && (
                    <div className="px-5 pb-4 text-sm text-slate-400 animate-fade-in">
                      {item.a}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          <div className="text-center">
            <button
              onClick={() => setShowAdmin(true)}
              className="text-slate-600 text-xs hover:text-slate-400 transition-colors"
            >
              Admin Panel
            </button>
          </div>
        </div>
      ) : (
        /* Admin login form */
        <div className="w-full max-w-md flex-1 flex items-center justify-center">
          <Card className="border-slate-800 bg-slate-900 w-full">
            <CardHeader className="text-center">
              <div className="text-3xl font-bold mb-1" style={{ color: "#c1f11d" }}>inVision U</div>
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
                  <div className="relative mt-1">
                    <Input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="bg-slate-800 border-slate-700 text-white pr-10"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200"
                      tabIndex={-1}
                    >
                      {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                    </button>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <input
                    id="remember"
                    type="checkbox"
                    checked={remember}
                    onChange={(e) => setRemember(e.target.checked)}
                    className="rounded border-slate-600 bg-slate-800"
                    style={{ accentColor: "#c1f11d" }}
                  />
                  <Label htmlFor="remember" className="text-slate-400 text-sm cursor-pointer">
                    Remember me
                  </Label>
                </div>
                {error && <p className="text-red-400 text-sm">{error}</p>}
                <Button
                  type="submit"
                  className="w-full font-bold"
                  style={{ backgroundColor: "#c1f11d", color: "#111" }}
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
