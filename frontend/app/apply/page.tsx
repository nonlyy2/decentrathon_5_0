"use client";

import { useState, FormEvent, useRef, useEffect } from "react";
import Link from "next/link";
import axios from "axios";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { CheckCircle, GraduationCap, Globe, Rocket, ArrowLeft, Upload, AlertTriangle, X, ChevronDown } from "lucide-react";

const publicApi = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api",
});

interface MajorOption {
  tag: string;
  en: string;
  ru: string;
  kk: string;
}

type Lang = "en" | "ru" | "kk";

export default function ApplyPage() {
  const [lang, setLang] = useState<Lang>("en");
  const [majors, setMajors] = useState<MajorOption[]>([]);
  const [form, setForm] = useState({
    full_name: "", email: "", phone: "", telegram: "", age: "", city: "", school: "",
    graduation_year: "", achievements: "", extracurriculars: "",
    essay: "", motivation_statement: "", disability: "", major: "", youtube_url: "",
  });
  const [photo, setPhoto] = useState<File | null>(null);
  const [photoPreview, setPhotoPreview] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    publicApi.get<MajorOption[]>("/majors").then((r) => setMajors(r.data)).catch(() => {});
    const saved = localStorage.getItem("apply_lang") as Lang;
    if (saved) setLang(saved);
  }, []);

  const switchLang = (l: Lang) => { setLang(l); localStorage.setItem("apply_lang", l); };
  const update = (field: string, value: string) => setForm({ ...form, [field]: value });

  const handlePhotoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > 5 * 1024 * 1024) { setError(t("apply.photo_size_error")); return; }
    setPhoto(file);
    setPhotoPreview(URL.createObjectURL(file));
    setError("");
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");

    if (form.essay.length < 50) {
      setError(t("apply.essay_min"));
      return;
    }
    if (!form.major) {
      setError(t("apply.major_required"));
      return;
    }
    if (!photo) {
      setError(t("apply.photo_required") || "Profile photo is required");
      return;
    }
    if (!form.youtube_url) {
      setError(t("apply.youtube_required"));
      return;
    }
    if (!/(?:youtube\.com\/watch\?(?:.*&)?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})/.test(form.youtube_url)) {
      setError(t("apply.youtube_invalid"));
      return;
    }
    // Validate phone: digits, +, -, spaces, parentheses only
    if (form.phone && !/^\+?[0-9\s\-()]+$/.test(form.phone)) {
      setError(t("apply.phone_invalid"));
      return;
    }
    // Validate telegram: latin letters, digits, underscore only (no cyrillic)
    if (form.telegram) {
      const tg = form.telegram.replace(/^@/, "");
      if (!/^[a-zA-Z0-9_]+$/.test(tg)) {
        setError(t("apply.telegram_invalid"));
        return;
      }
    }
    // Validate city: no digits
    if (form.city && /\d/.test(form.city)) {
      setError(t("apply.city_invalid"));
      return;
    }

    setSubmitting(true);
    try {
      const res = await publicApi.post<{ id: number }>("/apply", {
        full_name: form.full_name,
        email: form.email,
        phone: form.phone || null,
        telegram: form.telegram || null,
        age: form.age ? parseInt(form.age) : null,
        city: form.city || null,
        school: form.school || null,
        graduation_year: form.graduation_year ? parseInt(form.graduation_year) : null,
        achievements: form.achievements || null,
        extracurriculars: form.extracurriculars || null,
        essay: form.essay,
        motivation_statement: form.motivation_statement || null,
        disability: form.disability || null,
        major: form.major || null,
        youtube_url: form.youtube_url,
      });

      // Upload photo if provided
      if (photo && res.data.id) {
        const fd = new FormData();
        fd.append("photo", photo);
        await publicApi.post(`/candidates/${res.data.id}/photo`, fd, {
          headers: { "Content-Type": "multipart/form-data" },
        }).catch(() => {}); // photo upload is best-effort
      }

      setSuccess(true);
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        setError(t("apply.email_taken"));
      } else {
        setError(t("apply.error"));
      }
    } finally {
      setSubmitting(false);
    }
  };

  const t = (key: string) => APPLY_TRANSLATIONS[lang]?.[key] ?? key;
  const majorLabel = (m: MajorOption) => m[lang] || m.en;

  if (success) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <Card className="max-w-md w-full border-green-200 bg-green-50 dark:bg-green-950/30 dark:border-green-800">
          <CardContent className="p-10 text-center">
            <CheckCircle className="mx-auto text-green-500 mb-4" size={52} />
            <h2 className="text-2xl font-bold text-green-800 dark:text-green-200">{t("apply.success_title")}</h2>
            <p className="text-green-700 dark:text-green-300 mt-2 max-w-sm mx-auto text-sm">
              {t("apply.success_desc")}
            </p>
            <Link href="/">
              <Button variant="outline" className="mt-6">
                <ArrowLeft size={16} className="mr-2" /> {t("apply.back_home")}
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Language selector strip */}
      <div className="flex justify-end items-center gap-2 px-6 pt-4">
        {(["en", "ru", "kk"] as const).map((l) => (
          <button
            key={l}
            onClick={() => switchLang(l)}
            aria-pressed={lang === l}
            className={`text-xs px-3 py-1.5 rounded-full border font-medium transition-colors ${
              lang === l
                ? "border-primary text-foreground font-bold"
                : "border-border text-muted-foreground hover:border-primary"
            }`}
            style={lang === l ? { backgroundColor: "#c1f11d", borderColor: "#c1f11d", color: "#111" } : undefined}
          >
            {l.toUpperCase()}
          </button>
        ))}
      </div>

      <div className="max-w-2xl mx-auto px-4 pb-16">
        {/* Hero */}
        <div className="text-center py-8 space-y-3">
          <Link href="/" className="text-primary text-sm hover:opacity-80 inline-flex items-center gap-1">
            <ArrowLeft size={14} /> {t("apply.back")}
          </Link>
          <h1 className="text-4xl font-bold text-foreground">inVision U</h1>
          <p className="text-muted-foreground">{t("apply.subtitle")}</p>
          <div className="flex justify-center gap-6 pt-2">
            {[
              { icon: <GraduationCap size={18} />, label: t("apply.feat_scholarship") },
              { icon: <Globe size={18} />, label: t("apply.feat_global") },
              { icon: <Rocket size={18} />, label: t("apply.feat_tech") },
            ].map((item) => (
              <div key={item.label} className="flex items-center gap-1.5 text-muted-foreground text-xs">
                <span className="text-primary">{item.icon}</span>
                {item.label}
              </div>
            ))}
          </div>
        </div>

        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="text-foreground">{t("apply.form_title")}</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-6" noValidate>

              {/* Personal Info */}
              <section aria-labelledby="personal-section">
                <h3 id="personal-section" className="font-semibold mb-3 text-foreground">{t("apply.personal")}</h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field label={`${t("apply.full_name")} *`} htmlFor="full_name">
                    <Input id="full_name" value={form.full_name} onChange={(e) => update("full_name", e.target.value)} required autoComplete="name" />
                  </Field>
                  <Field label={`${t("apply.email")} *`} htmlFor="email">
                    <Input id="email" type="email" value={form.email} onChange={(e) => update("email", e.target.value)} required autoComplete="email" />
                  </Field>
                  <Field label={`${t("apply.phone")} *`} htmlFor="phone">
                    <Input id="phone" type="tel" value={form.phone} onChange={(e) => update("phone", e.target.value)} placeholder="+7 777 123 4567" required autoComplete="tel" />
                  </Field>
                  <Field label={`${t("apply.telegram")} *`} htmlFor="telegram">
                    <Input id="telegram" value={form.telegram} onChange={(e) => update("telegram", e.target.value)} placeholder="@username" required />
                  </Field>
                  <Field label={`${t("apply.age")} *`} htmlFor="age">
                    <Input id="age" type="number" min="14" max="30" value={form.age} onChange={(e) => update("age", e.target.value)} required />
                  </Field>
                  <Field label={`${t("apply.city")} *`} htmlFor="city">
                    <Input id="city" value={form.city} onChange={(e) => update("city", e.target.value)} required autoComplete="address-level2" />
                  </Field>
                  <Field label={`${t("apply.school")} *`} htmlFor="school">
                    <Input id="school" value={form.school} onChange={(e) => update("school", e.target.value)} required />
                  </Field>
                  <Field label={`${t("apply.graduation_year")} *`} htmlFor="graduation_year">
                    <Input id="graduation_year" type="number" min="2024" max="2030" value={form.graduation_year} onChange={(e) => update("graduation_year", e.target.value)} required />
                  </Field>
                </div>
              </section>

              {/* Photo upload */}
              <section aria-labelledby="photo-section">
                <h3 id="photo-section" className="font-semibold mb-1 text-foreground">{t("apply.photo")} <span className="text-red-500">*</span></h3>
                <p className="text-xs text-muted-foreground mb-3">{t("apply.photo_desc")}</p>
                <div className="flex items-start gap-4">
                  {photoPreview ? (
                    <div className="relative">
                      <img
                        src={photoPreview}
                        alt="Preview"
                        className="w-24 h-24 rounded-xl object-cover border border-border"
                      />
                      <button
                        type="button"
                        onClick={() => { setPhoto(null); setPhotoPreview(null); }}
                        className="absolute -top-2 -right-2 bg-destructive text-white rounded-full p-0.5"
                        aria-label="Remove photo"
                      >
                        <X size={12} />
                      </button>
                    </div>
                  ) : (
                    <button
                      type="button"
                      onClick={() => fileRef.current?.click()}
                      className="w-24 h-24 rounded-xl border-2 border-dashed border-border flex flex-col items-center justify-center gap-1 text-muted-foreground hover:border-primary hover:text-primary transition-colors"
                      aria-label="Upload photo"
                    >
                      <Upload size={20} />
                      <span className="text-[10px]">{t("apply.upload")}</span>
                    </button>
                  )}
                  <div className="flex-1 text-xs text-muted-foreground space-y-1">
                    <p>• {t("apply.photo_tip1")}</p>
                    <p>• {t("apply.photo_tip2")}</p>
                    <p className="text-yellow-600 dark:text-yellow-400 flex items-start gap-1">
                      <AlertTriangle size={11} className="mt-0.5 shrink-0" />
                      {t("apply.photo_ai_note")}
                    </p>
                  </div>
                </div>
                <input
                  ref={fileRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={handlePhotoChange}
                  aria-label="Photo file input"
                />
              </section>

              {/* Major selection */}
              <section aria-labelledby="major-section">
                <h3 id="major-section" className="font-semibold mb-1 text-foreground">{t("apply.major")} *</h3>
                <p className="text-xs text-muted-foreground mb-3">{t("apply.major_desc")}</p>
                <div className="relative">
                  <select
                    value={form.major}
                    onChange={(e) => update("major", e.target.value)}
                    required
                    aria-required="true"
                    aria-label={t("apply.major")}
                    className="w-full appearance-none bg-background border border-input rounded-lg px-3 py-2.5 pr-10 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                  >
                    <option value="">{t("apply.major_placeholder")}</option>
                    {majors.map((m) => (
                      <option key={m.tag} value={m.tag}>{m.tag} — {majorLabel(m)}</option>
                    ))}
                  </select>
                  <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />
                </div>
              </section>

              {/* Background */}
              <section aria-labelledby="background-section">
                <h3 id="background-section" className="font-semibold mb-3 text-foreground">{t("apply.background")}</h3>
                <div className="space-y-3">
                  <Field label={`${t("apply.achievements")} *`} htmlFor="achievements">
                    <Textarea id="achievements" rows={3} value={form.achievements} onChange={(e) => update("achievements", e.target.value)} placeholder={t("apply.achievements_ph")} required />
                  </Field>
                  <Field label={`${t("apply.extracurriculars")} *`} htmlFor="extracurriculars">
                    <Textarea id="extracurriculars" rows={3} value={form.extracurriculars} onChange={(e) => update("extracurriculars", e.target.value)} placeholder={t("apply.extracurriculars_ph")} required />
                  </Field>
                </div>
              </section>

              {/* Essay */}
              <section aria-labelledby="essay-section">
                <Field label={`${t("apply.essay")} *`} htmlFor="essay">
                  <p id="essay-hint" className="text-xs text-muted-foreground mb-2">{t("apply.essay_desc")}</p>
                  <Textarea
                    id="essay"
                    rows={8}
                    value={form.essay}
                    onChange={(e) => update("essay", e.target.value)}
                    placeholder={t("apply.essay_ph")}
                    required
                    aria-describedby="essay-hint essay-count"
                  />
                  <p
                    id="essay-count"
                    className={`text-xs mt-1 ${form.essay.length < 50 && form.essay.length > 0 ? "text-red-500" : "text-muted-foreground"}`}
                    aria-live="polite"
                  >
                    {form.essay.length} {t("apply.chars_min")}
                  </p>
                </Field>
              </section>

              {/* Motivation */}
              <section>
                <Field label={`${t("apply.motivation")} *`} htmlFor="motivation">
                  <Textarea id="motivation" rows={4} value={form.motivation_statement} onChange={(e) => update("motivation_statement", e.target.value)} placeholder={t("apply.motivation_ph")} required />
                </Field>
              </section>

              {/* YouTube Presentation Video */}
              <section>
                <Field label={`${t("apply.youtube_label")} *`} htmlFor="youtube_url">
                  <p className="text-xs text-muted-foreground mb-2">{t("apply.youtube_desc")}</p>
                  <Input
                    id="youtube_url"
                    type="url"
                    value={form.youtube_url}
                    onChange={(e) => update("youtube_url", e.target.value)}
                    placeholder="https://www.youtube.com/watch?v=..."
                    required
                  />
                </Field>
              </section>

              {/* Disability */}
              <section>
                <Field label={t("apply.disability")} htmlFor="disability">
                  <p className="text-xs text-muted-foreground mb-2">{t("apply.disability_desc")}</p>
                  <Textarea id="disability" rows={2} value={form.disability} onChange={(e) => update("disability", e.target.value)} placeholder={t("apply.disability_ph")} />
                </Field>
              </section>

              {error && (
                <p role="alert" className="text-red-500 text-sm flex items-center gap-1.5">
                  <AlertTriangle size={14} /> {error}
                </p>
              )}

              <Button
                type="submit"
                className="w-full py-6 text-base font-semibold"
                style={{ backgroundColor: "#c1f11d", color: "#111827" }}
                disabled={submitting}
              >
                {submitting ? t("apply.submitting") : t("apply.submit")}
              </Button>
            </form>
          </CardContent>
        </Card>

        <p className="text-center text-muted-foreground text-xs py-6">
          inVision U Admissions &mdash; {t("apply.footer")}
        </p>
      </div>
    </div>
  );
}

function Field({ label, htmlFor, children }: { label: string; htmlFor: string; children: React.ReactNode }) {
  return (
    <div>
      <Label htmlFor={htmlFor} className="text-sm font-medium text-foreground mb-1 block">{label}</Label>
      {children}
    </div>
  );
}

// ─── Translations (module-level, used via t() closure inside component) ──────
const APPLY_TRANSLATIONS: Record<Lang, Record<string, string>> = {
  en: {
    "apply.back": "Back",
    "apply.subtitle": "100% Scholarship University by inDrive",
    "apply.feat_scholarship": "Full Scholarship",
    "apply.feat_global": "Global Network",
    "apply.feat_tech": "Tech-Focused",
    "apply.form_title": "Application Form",
    "apply.personal": "Personal Information",
    "apply.full_name": "Full Name",
    "apply.email": "Email",
    "apply.phone": "Phone",
    "apply.telegram": "Telegram",
    "apply.age": "Age",
    "apply.city": "City",
    "apply.school": "School",
    "apply.graduation_year": "Graduation Year",
    "apply.photo": "Profile Photo",
    "apply.photo_desc": "Optional. A real photo helps us verify your identity.",
    "apply.photo_tip1": "Clear face photo, good lighting",
    "apply.photo_tip2": "Max 5 MB, JPEG or PNG",
    "apply.photo_ai_note": "Photos are checked for AI generation. This does NOT affect your application score.",
    "apply.photo_size_error": "Photo must be under 5 MB",
    "apply.upload": "Upload",
    "apply.major": "Preferred Major",
    "apply.major_desc": "Choose the program you are most interested in.",
    "apply.major_placeholder": "Select a major...",
    "apply.major_required": "Please select a major",
    "apply.background": "Background",
    "apply.achievements": "Achievements",
    "apply.achievements_ph": "List your key achievements...",
    "apply.extracurriculars": "Extracurricular Activities",
    "apply.extracurriculars_ph": "Clubs, volunteer work, sports, hobbies...",
    "apply.essay": "Essay",
    "apply.essay_desc": "Tell us about yourself, your experiences, and why you want to join inVision U. Please write in English.",
    "apply.essay_ph": "Write your essay here...",
    "apply.chars_min": "characters (minimum 50)",
    "apply.essay_min": "Essay must be at least 50 characters",
    "apply.motivation": "Motivation Statement",
    "apply.motivation_ph": "What motivates you to apply to inVision U?",
    "apply.disability": "Disability / Accessibility Needs",
    "apply.disability_desc": "Optional. Not used in scoring — only for accommodation during interviews.",
    "apply.disability_ph": "e.g. visual impairment, hearing difficulty...",
    "apply.youtube_label": "YouTube Presentation Video",
    "apply.youtube_desc": "Record a short video presenting your project or yourself and paste the YouTube link here.",
    "apply.youtube_required": "YouTube presentation video link is required",
    "apply.youtube_invalid": "Please enter a valid YouTube URL (youtube.com/watch?v=... or youtu.be/...)",
    "apply.phone_invalid": "Phone must contain only digits and + - ( ) characters",
    "apply.telegram_invalid": "Telegram username must contain only Latin letters, digits, and underscores",
    "apply.city_invalid": "City name must not contain digits",
    "apply.submit": "Submit Application",
    "apply.submitting": "Submitting...",
    "apply.error": "Something went wrong. Please try again.",
    "apply.email_taken": "This email is already registered.",
    "apply.success_title": "Application Submitted!",
    "apply.success_desc": "Thank you for applying to inVision U. We will review your application and contact you via email.",
    "apply.back_home": "Back to Home",
    "apply.footer": "Powered by AI screening technology",
  },
  ru: {
    "apply.back": "Назад",
    "apply.subtitle": "Университет со 100% стипендией от inDrive",
    "apply.feat_scholarship": "Полная стипендия",
    "apply.feat_global": "Глобальная сеть",
    "apply.feat_tech": "Фокус на технологиях",
    "apply.form_title": "Форма заявки",
    "apply.personal": "Личная информация",
    "apply.full_name": "Полное имя",
    "apply.email": "Email",
    "apply.phone": "Телефон",
    "apply.telegram": "Telegram",
    "apply.age": "Возраст",
    "apply.city": "Город",
    "apply.school": "Школа",
    "apply.graduation_year": "Год окончания",
    "apply.photo": "Фото профиля",
    "apply.photo_desc": "Необязательно. Настоящее фото помогает нам верифицировать вашу личность.",
    "apply.photo_tip1": "Чёткое фото лица, хорошее освещение",
    "apply.photo_tip2": "Макс. 5 МБ, JPEG или PNG",
    "apply.photo_ai_note": "Фото проверяются на генерацию ИИ. Это НЕ влияет на оценку заявки.",
    "apply.photo_size_error": "Фото должно быть менее 5 МБ",
    "apply.upload": "Загрузить",
    "apply.major": "Предпочтительное направление",
    "apply.major_desc": "Выберите программу, которая вас наиболее интересует.",
    "apply.major_placeholder": "Выберите направление...",
    "apply.major_required": "Пожалуйста, выберите направление",
    "apply.background": "Опыт и деятельность",
    "apply.achievements": "Достижения",
    "apply.achievements_ph": "Перечислите ваши ключевые достижения...",
    "apply.extracurriculars": "Внеклассная деятельность",
    "apply.extracurriculars_ph": "Кружки, волонтёрство, спорт, хобби...",
    "apply.essay": "Эссе",
    "apply.essay_desc": "Расскажите о себе, своём опыте и почему вы хотите поступить в inVision U. Пишите на английском.",
    "apply.essay_ph": "Напишите ваше эссе здесь...",
    "apply.chars_min": "символов (минимум 50)",
    "apply.essay_min": "Эссе должно содержать не менее 50 символов",
    "apply.motivation": "Мотивационное письмо",
    "apply.motivation_ph": "Что мотивирует вас подать заявку в inVision U?",
    "apply.disability": "Особые потребности / Доступность",
    "apply.disability_desc": "Необязательно. Не влияет на оценку — только для адаптации во время собеседования.",
    "apply.disability_ph": "напр. нарушение зрения, слуха...",
    "apply.youtube_label": "YouTube видео-презентация",
    "apply.youtube_desc": "Запишите короткое видео с презентацией вашего проекта и вставьте ссылку на YouTube.",
    "apply.youtube_required": "Ссылка на YouTube видео обязательна",
    "apply.youtube_invalid": "Введите корректную ссылку YouTube (youtube.com/watch?v=... или youtu.be/...)",
    "apply.phone_invalid": "Телефон должен содержать только цифры и символы + - ( )",
    "apply.telegram_invalid": "Telegram имя пользователя должно содержать только латинские буквы, цифры и подчёркивания",
    "apply.city_invalid": "Название города не должно содержать цифры",
    "apply.submit": "Отправить заявку",
    "apply.submitting": "Отправка...",
    "apply.error": "Что-то пошло не так. Попробуйте ещё раз.",
    "apply.email_taken": "Этот email уже зарегистрирован.",
    "apply.success_title": "Заявка отправлена!",
    "apply.success_desc": "Спасибо за заявку в inVision U. Мы рассмотрим вашу заявку и свяжемся с вами по email.",
    "apply.back_home": "На главную",
    "apply.footer": "Работает на технологии ИИ-скрининга",
  },
  kk: {
    "apply.back": "Артқа",
    "apply.subtitle": "inDrive-тан 100% стипендиялы университет",
    "apply.feat_scholarship": "Толық стипендия",
    "apply.feat_global": "Жаһандық желі",
    "apply.feat_tech": "Технологияларға бағытталған",
    "apply.form_title": "Өтінім нысаны",
    "apply.personal": "Жеке ақпарат",
    "apply.full_name": "Толық аты-жөні",
    "apply.email": "Email",
    "apply.phone": "Телефон",
    "apply.telegram": "Telegram",
    "apply.age": "Жасы",
    "apply.city": "Қала",
    "apply.school": "Мектеп",
    "apply.graduation_year": "Бітіру жылы",
    "apply.photo": "Профиль фотосы",
    "apply.photo_desc": "Міндетті емес. Нақты фото жеке басыңызды растауға көмектеседі.",
    "apply.photo_tip1": "Нақты бет суреті, жақсы жарық",
    "apply.photo_tip2": "Макс. 5 МБ, JPEG немесе PNG",
    "apply.photo_ai_note": "Фотолар ЖИ арқылы жасалғанын тексереді. Бұл өтінімнің бағасына ӘСЕР ЕТПЕЙДІ.",
    "apply.photo_size_error": "Фото 5 МБ-тан аз болуы керек",
    "apply.upload": "Жүктеу",
    "apply.major": "Таңдаулы бағыт",
    "apply.major_desc": "Сізді ең қызықтыратын бағдарламаны таңдаңыз.",
    "apply.major_placeholder": "Бағыт таңдаңыз...",
    "apply.major_required": "Бағытты таңдаңыз",
    "apply.background": "Тәжірибе мен іс-шаралар",
    "apply.achievements": "Жетістіктер",
    "apply.achievements_ph": "Негізгі жетістіктеріңізді тізіп беріңіз...",
    "apply.extracurriculars": "Сыныптан тыс іс-шаралар",
    "apply.extracurriculars_ph": "Үйірмелер, волонтерлік, спорт, хоббилер...",
    "apply.essay": "Эссе",
    "apply.essay_desc": "Өзіңіз, тәжірибеңіз және inVision U-ға неге қосылғыңыз келетіні туралы айтыңыз. Ағылшын тілінде жазыңыз.",
    "apply.essay_ph": "Эссеңізді осында жазыңыз...",
    "apply.chars_min": "таңба (кемінде 50)",
    "apply.essay_min": "Эссе кемінде 50 таңбадан тұруы керек",
    "apply.motivation": "Мотивациялық хат",
    "apply.motivation_ph": "inVision U-ға өтінім беруге не ынталандырады?",
    "apply.disability": "Мүмкіндіктері шектеулі / Қолжетімділік",
    "apply.disability_desc": "Міндетті емес. Бағалауға әсер етпейді — тек сұхбат кезінде бейімдеу үшін.",
    "apply.disability_ph": "мысалы, көру, есту қиындықтары...",
    "apply.youtube_label": "YouTube бейне-презентация",
    "apply.youtube_desc": "Жобаңызды таныстыратын қысқа бейне жазып, YouTube сілтемесін қойыңыз.",
    "apply.youtube_required": "YouTube бейне сілтемесі міндетті",
    "apply.youtube_invalid": "Дұрыс YouTube сілтемесін енгізіңіз (youtube.com/watch?v=... немесе youtu.be/...)",
    "apply.phone_invalid": "Телефон тек сандар мен + - ( ) таңбаларынан тұруы керек",
    "apply.telegram_invalid": "Telegram пайдаланушы аты тек латын әріптерінен, сандардан және астын сызудан тұруы керек",
    "apply.city_invalid": "Қала атауында сандар болмауы керек",
    "apply.submit": "Өтінімді жіберу",
    "apply.submitting": "Жіберілуде...",
    "apply.error": "Бірдеңе дұрыс болмады. Қайталап көріңіз.",
    "apply.email_taken": "Бұл email тіркелген.",
    "apply.success_title": "Өтінім жіберілді!",
    "apply.success_desc": "inVision U-ға өтінім бергеніңіз үшін рахмет. Өтінімді қарастырып, email арқылы хабарласамыз.",
    "apply.back_home": "Басты бетке",
    "apply.footer": "ЖИ скрининг технологиясымен жұмыс жасайды",
  },
};

