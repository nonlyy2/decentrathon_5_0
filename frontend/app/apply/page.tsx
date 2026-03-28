"use client";

import { useState, FormEvent } from "react";
import axios from "axios";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { CheckCircle } from "lucide-react";

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
      <Card className="mt-20 border-green-200 bg-green-50">
        <CardContent className="p-12 text-center">
          <CheckCircle className="mx-auto text-green-500 mb-4" size={48} />
          <h2 className="text-2xl font-bold text-green-800">Application Submitted!</h2>
          <p className="text-green-700 mt-2">Thank you for applying to inVision U. We will review your application.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <div className="text-center py-8">
        <h1 className="text-3xl font-bold text-white">Apply to inVision U</h1>
        <p className="text-slate-400 mt-2">100% Scholarship University by inDrive</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Application Form</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Personal Info */}
            <div>
              <h3 className="font-medium mb-3">Personal Information</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label>Full Name *</Label>
                  <Input value={form.full_name} onChange={(e) => update("full_name", e.target.value)} required />
                </div>
                <div>
                  <Label>Email *</Label>
                  <Input type="email" value={form.email} onChange={(e) => update("email", e.target.value)} required />
                </div>
                <div>
                  <Label>Age</Label>
                  <Input type="number" value={form.age} onChange={(e) => update("age", e.target.value)} />
                </div>
                <div>
                  <Label>City</Label>
                  <Input value={form.city} onChange={(e) => update("city", e.target.value)} />
                </div>
                <div>
                  <Label>School</Label>
                  <Input value={form.school} onChange={(e) => update("school", e.target.value)} />
                </div>
                <div>
                  <Label>Graduation Year</Label>
                  <Input type="number" value={form.graduation_year} onChange={(e) => update("graduation_year", e.target.value)} />
                </div>
              </div>
            </div>

            {/* Background */}
            <div>
              <h3 className="font-medium mb-3">Background</h3>
              <div className="space-y-3">
                <div>
                  <Label>Achievements</Label>
                  <Textarea rows={3} value={form.achievements} onChange={(e) => update("achievements", e.target.value)} placeholder="List your key achievements..." />
                </div>
                <div>
                  <Label>Extracurricular Activities</Label>
                  <Textarea rows={3} value={form.extracurriculars} onChange={(e) => update("extracurriculars", e.target.value)} placeholder="Clubs, volunteer work, sports, hobbies..." />
                </div>
              </div>
            </div>

            {/* Essay */}
            <div>
              <Label>Essay *</Label>
              <Textarea
                rows={8}
                value={form.essay}
                onChange={(e) => update("essay", e.target.value)}
                placeholder="Tell us about yourself, your experiences, and why you want to join inVision U..."
                required
              />
              <p className={`text-xs mt-1 ${form.essay.length < 50 ? "text-red-500" : "text-muted-foreground"}`}>
                {form.essay.length} characters (minimum 50)
              </p>
            </div>

            {/* Motivation */}
            <div>
              <Label>Motivation Statement</Label>
              <Textarea
                rows={4}
                value={form.motivation_statement}
                onChange={(e) => update("motivation_statement", e.target.value)}
                placeholder="What motivates you to apply to inVision U?"
              />
            </div>

            {error && <p className="text-red-500 text-sm">{error}</p>}

            <Button type="submit" className="w-full bg-purple-600 hover:bg-purple-700" disabled={submitting}>
              {submitting ? "Submitting..." : "Submit Application"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </>
  );
}
