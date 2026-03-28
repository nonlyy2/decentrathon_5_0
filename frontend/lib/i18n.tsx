"use client";

import { createContext, useContext, useState, useEffect, ReactNode } from "react";

type Lang = "en" | "ru";

const translations: Record<Lang, Record<string, string>> = {
  en: {
    // Sidebar
    "nav.dashboard": "Dashboard",
    "nav.candidates": "Candidates",
    "nav.signout": "Sign Out",
    // Dashboard
    "dash.title": "Dashboard",
    "dash.total": "Total",
    "dash.pending": "Pending",
    "dash.analyzed": "Analyzed",
    "dash.shortlisted": "Shortlisted",
    "dash.waitlisted": "Waitlisted",
    "dash.rejected": "Rejected",
    "dash.avg_score": "Average AI Score",
    "dash.analysis_rate": "Analysis Rate",
    "dash.shortlist_rate": "Shortlist Rate",
    "dash.status_dist": "Status Distribution",
    "dash.score_dist": "Score Distribution",
    "dash.categories": "Recommendation Categories",
    "dash.funnel": "Admissions Funnel",
    "dash.no_data": "No data yet",
    "dash.no_analysis": "No analysis data yet",
    // Candidates
    "cand.title": "Candidates",
    "cand.all": "All",
    "cand.search": "Search by name or email...",
    "cand.analyze_all": "Analyze All Pending",
    "cand.running": "Running...",
    "cand.export": "Export CSV",
    "cand.reset": "Reset Analyses",
    "cand.name": "Name",
    "cand.city": "City",
    "cand.score": "Score",
    "cand.status": "Status",
    "cand.created": "Created",
    "cand.analyzed_col": "Analyzed",
    "cand.model": "Model",
    "cand.no_found": "No candidates found",
    "cand.selected": "selected",
    "cand.compare": "Compare",
    "cand.clear": "Clear",
    "cand.previous": "Previous",
    "cand.next": "Next",
    "cand.showing": "Showing",
    "cand.of": "of",
    // Detail
    "detail.back": "Back",
    "detail.reanalyze": "Re-analyze",
    "detail.analyze": "Analyze with AI",
    "detail.analyzing": "Analyzing...",
    "detail.personal": "Personal Information",
    "detail.achievements": "Achievements",
    "detail.extracurriculars": "Extracurriculars",
    "detail.essay": "Essay",
    "detail.motivation": "Motivation Statement",
    "detail.ai_analysis": "AI Analysis",
    "detail.summary": "Summary",
    "detail.score_breakdown": "Score Breakdown",
    "detail.not_analyzed": "Not yet analyzed",
    "detail.committee": "Committee Decision",
    "detail.comments": "Comments",
    "detail.add_comment": "Add a comment...",
    "detail.history": "History",
    // Decisions
    "dec.shortlist": "Shortlist",
    "dec.waitlist": "Waitlist",
    "dec.review": "Review",
    "dec.reject": "Reject",
    // Delete dialog
    "del.title": "Reset All Analyses",
    "del.desc": "This will delete AI analyses for all candidates that have not been shortlisted, rejected, or waitlisted.",
    "del.confirm": "To confirm, type",
    "del.cancel": "Cancel",
    "del.delete": "Delete All",
    "del.deleting": "Deleting...",
    // Apply
    "apply.title": "Apply to inVision U",
    "apply.subtitle": "100% Scholarship University by inDrive",
    "apply.start": "Start Application",
    "apply.admin": "Admin Panel",
    "apply.back": "Back to application",
    "apply.signin": "Sign In",
    "apply.signing": "Signing in...",
    "apply.remember": "Remember me",
    "apply.email": "Email",
    "apply.password": "Password",
    // Compare
    "compare.title": "Compare Candidates",
    "compare.scores": "Score Comparison",
    "compare.not_analyzed": "Not analyzed",
  },
  ru: {
    // Sidebar
    "nav.dashboard": "Панель",
    "nav.candidates": "Кандидаты",
    "nav.signout": "Выйти",
    // Dashboard
    "dash.title": "Панель управления",
    "dash.total": "Всего",
    "dash.pending": "Ожидают",
    "dash.analyzed": "Проанализированы",
    "dash.shortlisted": "Отобраны",
    "dash.waitlisted": "В листе ожидания",
    "dash.rejected": "Отклонены",
    "dash.avg_score": "Средний балл ИИ",
    "dash.analysis_rate": "Доля анализа",
    "dash.shortlist_rate": "Доля отбора",
    "dash.status_dist": "Распределение статусов",
    "dash.score_dist": "Распределение баллов",
    "dash.categories": "Категории рекомендаций",
    "dash.funnel": "Воронка приёма",
    "dash.no_data": "Нет данных",
    "dash.no_analysis": "Нет данных анализа",
    // Candidates
    "cand.title": "Кандидаты",
    "cand.all": "Все",
    "cand.search": "Поиск по имени или email...",
    "cand.analyze_all": "Анализировать всех",
    "cand.running": "Выполняется...",
    "cand.export": "Экспорт CSV",
    "cand.reset": "Сбросить анализы",
    "cand.name": "Имя",
    "cand.city": "Город",
    "cand.score": "Балл",
    "cand.status": "Статус",
    "cand.created": "Создан",
    "cand.analyzed_col": "Проанализирован",
    "cand.model": "Модель",
    "cand.no_found": "Кандидаты не найдены",
    "cand.selected": "выбрано",
    "cand.compare": "Сравнить",
    "cand.clear": "Очистить",
    "cand.previous": "Назад",
    "cand.next": "Вперёд",
    "cand.showing": "Показано",
    "cand.of": "из",
    // Detail
    "detail.back": "Назад",
    "detail.reanalyze": "Повторный анализ",
    "detail.analyze": "Анализ с ИИ",
    "detail.analyzing": "Анализируется...",
    "detail.personal": "Личная информация",
    "detail.achievements": "Достижения",
    "detail.extracurriculars": "Внеклассная деятельность",
    "detail.essay": "Эссе",
    "detail.motivation": "Мотивационное письмо",
    "detail.ai_analysis": "Анализ ИИ",
    "detail.summary": "Резюме",
    "detail.score_breakdown": "Детализация баллов",
    "detail.not_analyzed": "Ещё не проанализирован",
    "detail.committee": "Решение комиссии",
    "detail.comments": "Комментарии",
    "detail.add_comment": "Добавить комментарий...",
    "detail.history": "История",
    // Decisions
    "dec.shortlist": "Отобрать",
    "dec.waitlist": "В ожидание",
    "dec.review": "На ревью",
    "dec.reject": "Отклонить",
    // Delete dialog
    "del.title": "Сбросить все анализы",
    "del.desc": "Это удалит анализы ИИ для всех кандидатов, которые не были отобраны, отклонены или в листе ожидания.",
    "del.confirm": "Для подтверждения введите",
    "del.cancel": "Отмена",
    "del.delete": "Удалить всё",
    "del.deleting": "Удаление...",
    // Apply
    "apply.title": "Подать заявку в inVision U",
    "apply.subtitle": "Университет со 100% стипендией от inDrive",
    "apply.start": "Подать заявку",
    "apply.admin": "Панель администратора",
    "apply.back": "Назад к заявке",
    "apply.signin": "Войти",
    "apply.signing": "Вход...",
    "apply.remember": "Запомнить меня",
    "apply.email": "Email",
    "apply.password": "Пароль",
    // Compare
    "compare.title": "Сравнение кандидатов",
    "compare.scores": "Сравнение баллов",
    "compare.not_analyzed": "Не проанализирован",
  },
};

interface I18nContextType {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: string) => string;
}

const I18nContext = createContext<I18nContextType>({
  lang: "en",
  setLang: () => {},
  t: (key) => key,
});

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>("en");

  useEffect(() => {
    const saved = localStorage.getItem("lang") as Lang;
    if (saved === "ru" || saved === "en") setLangState(saved);
  }, []);

  const setLang = (l: Lang) => {
    setLangState(l);
    localStorage.setItem("lang", l);
  };

  const t = (key: string) => translations[lang][key] || key;

  return (
    <I18nContext.Provider value={{ lang, setLang, t }}>
      {children}
    </I18nContext.Provider>
  );
}

export function useI18n() {
  return useContext(I18nContext);
}
