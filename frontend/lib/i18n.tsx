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
    "detail.email": "Email",
    "detail.age": "Age",
    "detail.city": "City",
    "detail.school": "School",
    "detail.graduation": "Graduation",
    "detail.ai_risk": "AI Risk",
    "detail.analyzed_by": "Analyzed by",
    "detail.retry": "Retry",
    "detail.analyzing_wait": "This may take 10\u201330 seconds",
    "detail.click_analyze": "Click \"Analyze with AI\" to generate scores",
    "detail.re_analyzing": "Re-analyzing...",
    "detail.ai_failed": "AI Analysis Failed",
    "detail.confirm_delete": "Delete this analysis? The candidate will return to Pending status.",
    // Decisions
    "dec.shortlist": "Shortlist",
    "dec.waitlist": "Waitlist",
    "dec.review": "Review",
    "dec.reject": "Reject",
    "dec.confirm_title": "Confirm",
    "dec.notes_placeholder": "Add notes (optional)...",
    "dec.cancel": "Cancel",
    "dec.saving": "Saving...",
    "dec.confirm": "Confirm",
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
    "compare.back": "Back to Candidates",
    "compare.select_min": "Select at least 2 candidates to compare.",
    "compare.summary": "Summary",
    "compare.strengths": "Strengths",
    "compare.red_flags": "Red Flags",
    "compare.not_analyzed_yet": "Not analyzed yet",
    // Status labels
    "status.pending": "Pending",
    "status.analyzed": "Analyzed",
    "status.shortlisted": "Shortlisted",
    "status.waitlisted": "Waitlisted",
    "status.rejected": "Rejected",
    // Score dimensions
    "score.leadership": "Leadership",
    "score.motivation": "Motivation",
    "score.growth": "Growth",
    "score.vision": "Vision",
    "score.communication": "Communication",
    // Analysis card
    "analysis.key_strengths": "Key Strengths",
    "analysis.red_flags": "Red Flags",
    "analysis.no_strengths": "No key strengths identified",
    "analysis.no_red_flags": "No red flags detected",
    // Analytics page
    "analytics.title": "Analytics",
    "analytics.total_apps": "Total Applications",
    "analytics.all_time": "all time",
    "analytics.analyzed_sub": "analyzed",
    "analytics.shortlisted_sub": "shortlisted",
    "analytics.out_of": "out of 100",
    // Funnel labels
    "funnel.applied": "Applied",
    "funnel.ai_analyzed": "AI Analyzed",
    "funnel.shortlisted": "Shortlisted",
    "funnel.waitlisted": "Waitlisted",
    "funnel.rejected": "Rejected",
    // Categories
    "cat.strong_recommend": "Strong Recommend",
    "cat.recommend": "Recommend",
    "cat.borderline": "Borderline",
    "cat.not_recommended": "Not Recommended",
  },
  ru: {
    // Sidebar
    "nav.dashboard": "\u041f\u0430\u043d\u0435\u043b\u044c",
    "nav.candidates": "\u041a\u0430\u043d\u0434\u0438\u0434\u0430\u0442\u044b",
    "nav.signout": "\u0412\u044b\u0439\u0442\u0438",
    // Dashboard
    "dash.title": "\u041f\u0430\u043d\u0435\u043b\u044c \u0443\u043f\u0440\u0430\u0432\u043b\u0435\u043d\u0438\u044f",
    "dash.total": "\u0412\u0441\u0435\u0433\u043e",
    "dash.pending": "\u041e\u0436\u0438\u0434\u0430\u044e\u0442",
    "dash.analyzed": "\u041f\u0440\u043e\u0430\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d\u044b",
    "dash.shortlisted": "\u041e\u0442\u043e\u0431\u0440\u0430\u043d\u044b",
    "dash.waitlisted": "\u0412 \u043b\u0438\u0441\u0442\u0435 \u043e\u0436\u0438\u0434\u0430\u043d\u0438\u044f",
    "dash.rejected": "\u041e\u0442\u043a\u043b\u043e\u043d\u0435\u043d\u044b",
    "dash.avg_score": "\u0421\u0440\u0435\u0434\u043d\u0438\u0439 \u0431\u0430\u043b\u043b \u0418\u0418",
    "dash.analysis_rate": "\u0414\u043e\u043b\u044f \u0430\u043d\u0430\u043b\u0438\u0437\u0430",
    "dash.shortlist_rate": "\u0414\u043e\u043b\u044f \u043e\u0442\u0431\u043e\u0440\u0430",
    "dash.status_dist": "\u0420\u0430\u0441\u043f\u0440\u0435\u0434\u0435\u043b\u0435\u043d\u0438\u0435 \u0441\u0442\u0430\u0442\u0443\u0441\u043e\u0432",
    "dash.score_dist": "\u0420\u0430\u0441\u043f\u0440\u0435\u0434\u0435\u043b\u0435\u043d\u0438\u0435 \u0431\u0430\u043b\u043b\u043e\u0432",
    "dash.categories": "\u041a\u0430\u0442\u0435\u0433\u043e\u0440\u0438\u0438 \u0440\u0435\u043a\u043e\u043c\u0435\u043d\u0434\u0430\u0446\u0438\u0439",
    "dash.funnel": "\u0412\u043e\u0440\u043e\u043d\u043a\u0430 \u043f\u0440\u0438\u0451\u043c\u0430",
    "dash.no_data": "\u041d\u0435\u0442 \u0434\u0430\u043d\u043d\u044b\u0445",
    "dash.no_analysis": "\u041d\u0435\u0442 \u0434\u0430\u043d\u043d\u044b\u0445 \u0430\u043d\u0430\u043b\u0438\u0437\u0430",
    // Candidates
    "cand.title": "\u041a\u0430\u043d\u0434\u0438\u0434\u0430\u0442\u044b",
    "cand.all": "\u0412\u0441\u0435",
    "cand.search": "\u041f\u043e\u0438\u0441\u043a \u043f\u043e \u0438\u043c\u0435\u043d\u0438 \u0438\u043b\u0438 email...",
    "cand.analyze_all": "\u0410\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u0442\u044c \u0432\u0441\u0435\u0445",
    "cand.running": "\u0412\u044b\u043f\u043e\u043b\u043d\u044f\u0435\u0442\u0441\u044f...",
    "cand.export": "\u042d\u043a\u0441\u043f\u043e\u0440\u0442 CSV",
    "cand.reset": "\u0421\u0431\u0440\u043e\u0441\u0438\u0442\u044c \u0430\u043d\u0430\u043b\u0438\u0437\u044b",
    "cand.name": "\u0418\u043c\u044f",
    "cand.city": "\u0413\u043e\u0440\u043e\u0434",
    "cand.score": "\u0411\u0430\u043b\u043b",
    "cand.status": "\u0421\u0442\u0430\u0442\u0443\u0441",
    "cand.created": "\u0421\u043e\u0437\u0434\u0430\u043d",
    "cand.analyzed_col": "\u041f\u0440\u043e\u0430\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d",
    "cand.model": "\u041c\u043e\u0434\u0435\u043b\u044c",
    "cand.no_found": "\u041a\u0430\u043d\u0434\u0438\u0434\u0430\u0442\u044b \u043d\u0435 \u043d\u0430\u0439\u0434\u0435\u043d\u044b",
    "cand.selected": "\u0432\u044b\u0431\u0440\u0430\u043d\u043e",
    "cand.compare": "\u0421\u0440\u0430\u0432\u043d\u0438\u0442\u044c",
    "cand.clear": "\u041e\u0447\u0438\u0441\u0442\u0438\u0442\u044c",
    "cand.previous": "\u041d\u0430\u0437\u0430\u0434",
    "cand.next": "\u0412\u043f\u0435\u0440\u0451\u0434",
    "cand.showing": "\u041f\u043e\u043a\u0430\u0437\u0430\u043d\u043e",
    "cand.of": "\u0438\u0437",
    // Detail
    "detail.back": "\u041d\u0430\u0437\u0430\u0434",
    "detail.reanalyze": "\u041f\u043e\u0432\u0442\u043e\u0440\u043d\u044b\u0439 \u0430\u043d\u0430\u043b\u0438\u0437",
    "detail.analyze": "\u0410\u043d\u0430\u043b\u0438\u0437 \u0441 \u0418\u0418",
    "detail.analyzing": "\u0410\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u0443\u0435\u0442\u0441\u044f...",
    "detail.personal": "\u041b\u0438\u0447\u043d\u0430\u044f \u0438\u043d\u0444\u043e\u0440\u043c\u0430\u0446\u0438\u044f",
    "detail.achievements": "\u0414\u043e\u0441\u0442\u0438\u0436\u0435\u043d\u0438\u044f",
    "detail.extracurriculars": "\u0412\u043d\u0435\u043a\u043b\u0430\u0441\u0441\u043d\u0430\u044f \u0434\u0435\u044f\u0442\u0435\u043b\u044c\u043d\u043e\u0441\u0442\u044c",
    "detail.essay": "\u042d\u0441\u0441\u0435",
    "detail.motivation": "\u041c\u043e\u0442\u0438\u0432\u0430\u0446\u0438\u043e\u043d\u043d\u043e\u0435 \u043f\u0438\u0441\u044c\u043c\u043e",
    "detail.ai_analysis": "\u0410\u043d\u0430\u043b\u0438\u0437 \u0418\u0418",
    "detail.summary": "\u0420\u0435\u0437\u044e\u043c\u0435",
    "detail.score_breakdown": "\u0414\u0435\u0442\u0430\u043b\u0438\u0437\u0430\u0446\u0438\u044f \u0431\u0430\u043b\u043b\u043e\u0432",
    "detail.not_analyzed": "\u0415\u0449\u0451 \u043d\u0435 \u043f\u0440\u043e\u0430\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d",
    "detail.committee": "\u0420\u0435\u0448\u0435\u043d\u0438\u0435 \u043a\u043e\u043c\u0438\u0441\u0441\u0438\u0438",
    "detail.comments": "\u041a\u043e\u043c\u043c\u0435\u043d\u0442\u0430\u0440\u0438\u0438",
    "detail.add_comment": "\u0414\u043e\u0431\u0430\u0432\u0438\u0442\u044c \u043a\u043e\u043c\u043c\u0435\u043d\u0442\u0430\u0440\u0438\u0439...",
    "detail.history": "\u0418\u0441\u0442\u043e\u0440\u0438\u044f",
    "detail.email": "\u042d\u043b. \u043f\u043e\u0447\u0442\u0430",
    "detail.age": "\u0412\u043e\u0437\u0440\u0430\u0441\u0442",
    "detail.city": "\u0413\u043e\u0440\u043e\u0434",
    "detail.school": "\u0428\u043a\u043e\u043b\u0430",
    "detail.graduation": "\u0412\u044b\u043f\u0443\u0441\u043a",
    "detail.ai_risk": "\u0420\u0438\u0441\u043a \u0418\u0418",
    "detail.analyzed_by": "\u041c\u043e\u0434\u0435\u043b\u044c",
    "detail.retry": "\u041f\u043e\u0432\u0442\u043e\u0440\u0438\u0442\u044c",
    "detail.analyzing_wait": "\u042d\u0442\u043e \u043c\u043e\u0436\u0435\u0442 \u0437\u0430\u043d\u044f\u0442\u044c 10\u201330 \u0441\u0435\u043a\u0443\u043d\u0434",
    "detail.click_analyze": "\u041d\u0430\u0436\u043c\u0438\u0442\u0435 \"\u0410\u043d\u0430\u043b\u0438\u0437 \u0441 \u0418\u0418\" \u0434\u043b\u044f \u0433\u0435\u043d\u0435\u0440\u0430\u0446\u0438\u0438 \u0431\u0430\u043b\u043b\u043e\u0432",
    "detail.re_analyzing": "\u041f\u043e\u0432\u0442\u043e\u0440\u043d\u044b\u0439 \u0430\u043d\u0430\u043b\u0438\u0437...",
    "detail.ai_failed": "\u0410\u043d\u0430\u043b\u0438\u0437 \u0418\u0418 \u043d\u0435 \u0443\u0434\u0430\u043b\u0441\u044f",
    "detail.confirm_delete": "\u0423\u0434\u0430\u043b\u0438\u0442\u044c \u0430\u043d\u0430\u043b\u0438\u0437? \u041a\u0430\u043d\u0434\u0438\u0434\u0430\u0442 \u0432\u0435\u0440\u043d\u0451\u0442\u0441\u044f \u0432 \u0441\u0442\u0430\u0442\u0443\u0441 \u041e\u0436\u0438\u0434\u0430\u0435\u0442.",
    // Decisions
    "dec.shortlist": "\u041e\u0442\u043e\u0431\u0440\u0430\u0442\u044c",
    "dec.waitlist": "\u0412 \u043e\u0436\u0438\u0434\u0430\u043d\u0438\u0435",
    "dec.review": "\u041d\u0430 \u0440\u0435\u0432\u044c\u044e",
    "dec.reject": "\u041e\u0442\u043a\u043b\u043e\u043d\u0438\u0442\u044c",
    "dec.confirm_title": "\u041f\u043e\u0434\u0442\u0432\u0435\u0440\u0436\u0434\u0435\u043d\u0438\u0435",
    "dec.notes_placeholder": "\u0414\u043e\u0431\u0430\u0432\u0438\u0442\u044c \u0437\u0430\u043c\u0435\u0442\u043a\u0438 (\u043d\u0435\u043e\u0431\u044f\u0437\u0430\u0442\u0435\u043b\u044c\u043d\u043e)...",
    "dec.cancel": "\u041e\u0442\u043c\u0435\u043d\u0430",
    "dec.saving": "\u0421\u043e\u0445\u0440\u0430\u043d\u0435\u043d\u0438\u0435...",
    "dec.confirm": "\u041f\u043e\u0434\u0442\u0432\u0435\u0440\u0434\u0438\u0442\u044c",
    // Delete dialog
    "del.title": "\u0421\u0431\u0440\u043e\u0441\u0438\u0442\u044c \u0432\u0441\u0435 \u0430\u043d\u0430\u043b\u0438\u0437\u044b",
    "del.desc": "\u042d\u0442\u043e \u0443\u0434\u0430\u043b\u0438\u0442 \u0430\u043d\u0430\u043b\u0438\u0437\u044b \u0418\u0418 \u0434\u043b\u044f \u0432\u0441\u0435\u0445 \u043a\u0430\u043d\u0434\u0438\u0434\u0430\u0442\u043e\u0432, \u043a\u043e\u0442\u043e\u0440\u044b\u0435 \u043d\u0435 \u0431\u044b\u043b\u0438 \u043e\u0442\u043e\u0431\u0440\u0430\u043d\u044b, \u043e\u0442\u043a\u043b\u043e\u043d\u0435\u043d\u044b \u0438\u043b\u0438 \u0432 \u043b\u0438\u0441\u0442\u0435 \u043e\u0436\u0438\u0434\u0430\u043d\u0438\u044f.",
    "del.confirm": "\u0414\u043b\u044f \u043f\u043e\u0434\u0442\u0432\u0435\u0440\u0436\u0434\u0435\u043d\u0438\u044f \u0432\u0432\u0435\u0434\u0438\u0442\u0435",
    "del.cancel": "\u041e\u0442\u043c\u0435\u043d\u0430",
    "del.delete": "\u0423\u0434\u0430\u043b\u0438\u0442\u044c \u0432\u0441\u0451",
    "del.deleting": "\u0423\u0434\u0430\u043b\u0435\u043d\u0438\u0435...",
    // Apply
    "apply.title": "\u041f\u043e\u0434\u0430\u0442\u044c \u0437\u0430\u044f\u0432\u043a\u0443 \u0432 inVision U",
    "apply.subtitle": "\u0423\u043d\u0438\u0432\u0435\u0440\u0441\u0438\u0442\u0435\u0442 \u0441\u043e 100% \u0441\u0442\u0438\u043f\u0435\u043d\u0434\u0438\u0435\u0439 \u043e\u0442 inDrive",
    "apply.start": "\u041f\u043e\u0434\u0430\u0442\u044c \u0437\u0430\u044f\u0432\u043a\u0443",
    "apply.admin": "\u041f\u0430\u043d\u0435\u043b\u044c \u0430\u0434\u043c\u0438\u043d\u0438\u0441\u0442\u0440\u0430\u0442\u043e\u0440\u0430",
    "apply.back": "\u041d\u0430\u0437\u0430\u0434 \u043a \u0437\u0430\u044f\u0432\u043a\u0435",
    "apply.signin": "\u0412\u043e\u0439\u0442\u0438",
    "apply.signing": "\u0412\u0445\u043e\u0434...",
    "apply.remember": "\u0417\u0430\u043f\u043e\u043c\u043d\u0438\u0442\u044c \u043c\u0435\u043d\u044f",
    "apply.email": "Email",
    "apply.password": "\u041f\u0430\u0440\u043e\u043b\u044c",
    // Compare
    "compare.title": "\u0421\u0440\u0430\u0432\u043d\u0435\u043d\u0438\u0435 \u043a\u0430\u043d\u0434\u0438\u0434\u0430\u0442\u043e\u0432",
    "compare.scores": "\u0421\u0440\u0430\u0432\u043d\u0435\u043d\u0438\u0435 \u0431\u0430\u043b\u043b\u043e\u0432",
    "compare.not_analyzed": "\u041d\u0435 \u043f\u0440\u043e\u0430\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d",
    "compare.back": "\u041d\u0430\u0437\u0430\u0434 \u043a \u043a\u0430\u043d\u0434\u0438\u0434\u0430\u0442\u0430\u043c",
    "compare.select_min": "\u0412\u044b\u0431\u0435\u0440\u0438\u0442\u0435 \u043c\u0438\u043d\u0438\u043c\u0443\u043c 2 \u043a\u0430\u043d\u0434\u0438\u0434\u0430\u0442\u043e\u0432 \u0434\u043b\u044f \u0441\u0440\u0430\u0432\u043d\u0435\u043d\u0438\u044f.",
    "compare.summary": "\u0420\u0435\u0437\u044e\u043c\u0435",
    "compare.strengths": "\u0421\u0438\u043b\u044c\u043d\u044b\u0435 \u0441\u0442\u043e\u0440\u043e\u043d\u044b",
    "compare.red_flags": "\u041a\u0440\u0430\u0441\u043d\u044b\u0435 \u0444\u043b\u0430\u0433\u0438",
    "compare.not_analyzed_yet": "\u0415\u0449\u0451 \u043d\u0435 \u043f\u0440\u043e\u0430\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d",
    // Status labels
    "status.pending": "\u041e\u0436\u0438\u0434\u0430\u0435\u0442",
    "status.analyzed": "\u041f\u0440\u043e\u0430\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d",
    "status.shortlisted": "\u041e\u0442\u043e\u0431\u0440\u0430\u043d",
    "status.waitlisted": "\u0412 \u043e\u0436\u0438\u0434\u0430\u043d\u0438\u0438",
    "status.rejected": "\u041e\u0442\u043a\u043b\u043e\u043d\u0451\u043d",
    // Score dimensions
    "score.leadership": "\u041b\u0438\u0434\u0435\u0440\u0441\u0442\u0432\u043e",
    "score.motivation": "\u041c\u043e\u0442\u0438\u0432\u0430\u0446\u0438\u044f",
    "score.growth": "\u0420\u043e\u0441\u0442",
    "score.vision": "\u0412\u0438\u0434\u0435\u043d\u0438\u0435",
    "score.communication": "\u041a\u043e\u043c\u043c\u0443\u043d\u0438\u043a\u0430\u0446\u0438\u044f",
    // Analysis card
    "analysis.key_strengths": "\u041a\u043b\u044e\u0447\u0435\u0432\u044b\u0435 \u0441\u0438\u043b\u044c\u043d\u044b\u0435 \u0441\u0442\u043e\u0440\u043e\u043d\u044b",
    "analysis.red_flags": "\u041a\u0440\u0430\u0441\u043d\u044b\u0435 \u0444\u043b\u0430\u0433\u0438",
    "analysis.no_strengths": "\u041a\u043b\u044e\u0447\u0435\u0432\u044b\u0435 \u0441\u0438\u043b\u044c\u043d\u044b\u0435 \u0441\u0442\u043e\u0440\u043e\u043d\u044b \u043d\u0435 \u0432\u044b\u044f\u0432\u043b\u0435\u043d\u044b",
    "analysis.no_red_flags": "\u041a\u0440\u0430\u0441\u043d\u044b\u0435 \u0444\u043b\u0430\u0433\u0438 \u043d\u0435 \u043e\u0431\u043d\u0430\u0440\u0443\u0436\u0435\u043d\u044b",
    // Analytics page
    "analytics.title": "\u0410\u043d\u0430\u043b\u0438\u0442\u0438\u043a\u0430",
    "analytics.total_apps": "\u0412\u0441\u0435\u0433\u043e \u0437\u0430\u044f\u0432\u043e\u043a",
    "analytics.all_time": "\u0437\u0430 \u0432\u0441\u0451 \u0432\u0440\u0435\u043c\u044f",
    "analytics.analyzed_sub": "\u043f\u0440\u043e\u0430\u043d\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d\u043e",
    "analytics.shortlisted_sub": "\u043e\u0442\u043e\u0431\u0440\u0430\u043d\u043e",
    "analytics.out_of": "\u0438\u0437 100",
    // Funnel labels
    "funnel.applied": "\u041f\u043e\u0434\u0430\u043d\u043e",
    "funnel.ai_analyzed": "\u0418\u0418 \u0430\u043d\u0430\u043b\u0438\u0437",
    "funnel.shortlisted": "\u041e\u0442\u043e\u0431\u0440\u0430\u043d\u044b",
    "funnel.waitlisted": "\u0412 \u043e\u0436\u0438\u0434\u0430\u043d\u0438\u0438",
    "funnel.rejected": "\u041e\u0442\u043a\u043b\u043e\u043d\u0435\u043d\u044b",
    // Categories
    "cat.strong_recommend": "\u0421\u0438\u043b\u044c\u043d\u043e \u0440\u0435\u043a\u043e\u043c\u0435\u043d\u0434\u043e\u0432\u0430\u043d",
    "cat.recommend": "\u0420\u0435\u043a\u043e\u043c\u0435\u043d\u0434\u043e\u0432\u0430\u043d",
    "cat.borderline": "\u041f\u043e\u0433\u0440\u0430\u043d\u0438\u0447\u043d\u044b\u0439",
    "cat.not_recommended": "\u041d\u0435 \u0440\u0435\u043a\u043e\u043c\u0435\u043d\u0434\u043e\u0432\u0430\u043d",
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
