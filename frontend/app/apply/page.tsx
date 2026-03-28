"use client";

import { useState, FormEvent } from "react";
import Link from "next/link";
import axios from "axios";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { CheckCircle, GraduationCap, Globe, Rocket, ArrowLeft } from "lucide-react";

const publicApi = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api",
});

export default function ApplyPage() {
  const [form, setForm] = useState({
    full_name: "", email: "", age: "", city: "", school: "",
    graduation_year: "", achievements: "", extracurriculars: "",
    essay: "", motivation_statement: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");

  const update = (field: string, value: string) => setForm({ ...form, [field]: value });

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");

    if (form.essay.length < 50) {
      setError("Essay must be at least 50 characters");
      return;
    }

    setSubmitting(true);
    try {
      await publicApi.post("/apply", {
        full_name: form.full_name,
        email: form.email,
        age: form.age ? parseInt(form.age) : null,
        city: form.city || null,
        school: form.school || null,
        graduation_year: form.graduation_year ? parseInt(form.graduation_year) : null,
        achievements: form.achievements || null,
        extracurriculars: form.extracurriculars || null,
        essay: form.essay,
        motivation_statement: form.motivation_statement || null,
      });
      setSuccess(true);
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        setError("This email is already registered.");
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (success) {
    return (
      <div className="pt-20">
        <Card className="border-green-200 bg-green-50">
          <CardContent className="p-12 text-center">
            <CheckCircle className="mx-auto text-green-500 mb-4" size={48} />
            <h2 className="text-2xl font-bold text-green-800">Application Submitted!</h2>
            <p className="text-green-700 mt-2 max-w-md mx-auto">
              Thank you for applying to inVision U. Our AI system will review your application, and our admissions committee will make a decision.
            </p>
            <Link href="/">
              <Button variant="outline" className="mt-6">
                <ArrowLeft size={16} className="mr-2" /> Back to Home
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <>
      {/* Hero section */}
      <div className="text-center py-10 space-y-4">
        <Link href="/" className="text-purple-400 text-sm hover:text-purple-300 inline-flex items-center gap-1">
          <ArrowLeft size={14} /> Back
        </Link>
        <h1 className="text-4xl font-bold text-white">
          inVision U
        </h1>
        <p className="text-lg text-slate-300">100% Scholarship University by inDrive</p>
        <p className="text-slate-400 max-w-lg mx-auto">
          A new kind of university for the next generation of tech leaders in Central Asia.
          Apply for the 2026 cohort below.
        </p>
        <div className="flex justify-center gap-8 pt-4">
          {[
            { icon: <GraduationCap size={20} />, label: "Full Scholarship" },
            { icon: <Globe size={20} />, label: "Global Network" },
            { icon: <Rocket size={20} />, label: "Tech-Focused" },
          ].map((item) => (
            <div key={item.label} className="flex items-center gap-2 text-slate-300 text-sm">
              <span className="text-purple-400">{item.icon}</span>
              {item.label}
            </div>
          ))}
        </div>
      </div>

      <Card className="border-slate-700 bg-slate-800/50 backdrop-blur">
        <CardHeader>
          <CardTitle className="text-white">Application Form</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Personal Info */}
            <div>
              <h3 className="font-medium mb-3 text-slate-200">Personal Information</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <Label className="text-slate-300">Full Name *</Label>
                  <Input value={form.full_name} onChange={(e) => update("full_name", e.target.value)} required className="bg-slate-700/50 border-slate-600 text-white" />
                </div>
                <div>
                  <Label className="text-slate-300">Email *</Label>
                  <Input type="email" value={form.email} onChange={(e) => update("email", e.target.value)} required className="bg-slate-700/50 border-slate-600 text-white" />
                </div>
                <div>
                  <Label className="text-slate-300">Age</Label>
                  <Input type="number" value={form.age} onChange={(e) => update("age", e.target.value)} className="bg-slate-700/50 border-slate-600 text-white" />
                </div>
                <div>
                  <Label className="text-slate-300">City</Label>
                  <Input value={form.city} onChange={(e) => update("city", e.target.value)} className="bg-slate-700/50 border-slate-600 text-white" />
                </div>
                <div>
                  <Label className="text-slate-300">School</Label>
                  <Input value={form.school} onChange={(e) => update("school", e.target.value)} className="bg-slate-700/50 border-slate-600 text-white" />
                </div>
                <div>
                  <Label className="text-slate-300">Graduation Year</Label>
                  <Input type="number" value={form.graduation_year} onChange={(e) => update("graduation_year", e.target.value)} className="bg-slate-700/50 border-slate-600 text-white" />
                </div>
              </div>
            </div>

            {/* Background */}
            <div>
              <h3 className="font-medium mb-3 text-slate-200">Background</h3>
              <div className="space-y-3">
                <div>
                  <Label className="text-slate-300">Achievements</Label>
                  <Textarea rows={3} value={form.achievements} onChange={(e) => update("achievements", e.target.value)} placeholder="List your key achievements..." className="bg-slate-700/50 border-slate-600 text-white placeholder:text-slate-500" />
                </div>
                <div>
                  <Label className="text-slate-300">Extracurricular Activities</Label>
                  <Textarea rows={3} value={form.extracurriculars} onChange={(e) => update("extracurriculars", e.target.value)} placeholder="Clubs, volunteer work, sports, hobbies..." className="bg-slate-700/50 border-slate-600 text-white placeholder:text-slate-500" />
                </div>
              </div>
            </div>

            {/* Essay */}
            <div>
              <Label className="text-slate-300">Essay *</Label>
              <p className="text-xs text-slate-400 mb-2">Tell us about yourself, your experiences, and why you want to join inVision U.</p>
              <Textarea
                rows={8}
                value={form.essay}
                onChange={(e) => update("essay", e.target.value)}
                placeholder="Write your essay here..."
                required
                className="bg-slate-700/50 border-slate-600 text-white placeholder:text-slate-500"
              />
              <p className={`text-xs mt-1 ${form.essay.length < 50 && form.essay.length > 0 ? "text-red-400" : "text-slate-500"}`}>
                {form.essay.length} characters (minimum 50)
              </p>
            </div>

            {/* Motivation */}
            <div>
              <Label className="text-slate-300">Motivation Statement</Label>
              <Textarea
                rows={4}
                value={form.motivation_statement}
                onChange={(e) => update("motivation_statement", e.target.value)}
                placeholder="What motivates you to apply to inVision U?"
                className="bg-slate-700/50 border-slate-600 text-white placeholder:text-slate-500"
              />
            </div>

            {error && <p className="text-red-400 text-sm">{error}</p>}

            <Button type="submit" className="w-full bg-purple-600 hover:bg-purple-700 py-6 text-lg" disabled={submitting}>
              {submitting ? "Submitting..." : "Submit Application"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <p className="text-center text-slate-500 text-xs py-6">
        inVision U Admissions &mdash; Powered by AI screening technology
      </p>
    </>
  );
}
