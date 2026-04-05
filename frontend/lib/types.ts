export interface Candidate {
  id: number;
  full_name: string;
  first_name: string | null;
  last_name: string | null;
  patronymic: string | null;
  email: string;
  phone: string | null;
  telegram: string | null;
  age: number | null;
  date_of_birth: string | null;
  gender: string | null;
  city: string | null;
  home_country: string | null;
  school: string | null;
  graduation_year: number | null;
  nationality: string | null;
  iin: string | null;
  identity_doc_type: string | null;
  instagram: string | null;
  whatsapp: string | null;
  achievements: string | null;
  extracurriculars: string | null;
  essay: string;
  motivation_statement: string | null;
  disability: string | null;
  major: string | null;
  photo_url: string | null;
  photo_ai_flag: boolean;
  photo_ai_note: string | null;
  keywords: string[];
  created_at: string;
  status: string;
  youtube_url: string | null;
  youtube_transcript: string | null;
  youtube_url_valid: boolean | null;
  exam_type: string | null;
  ielts_score: number | null;
  toefl_score: number | null;
  english_cert_url: string | null;
  certificate_type: string | null;
  certificate_url: string | null;
  additional_docs_url: string | null;
  personality_answers: string | null;
  review_complexity: number | null;
  unt_score: number | null;
  nis_grade: string | null;
  partner_school: string | null;
}

export interface CandidateListItem {
  id: number;
  full_name: string;
  email: string;
  city: string | null;
  school: string | null;
  major: string | null;
  status: string;
  created_at: string;
  final_score: number | null;
  category: string | null;
  analyzed_at: string | null;
  model_used: string | null;
  photo_url: string | null;
  photo_ai_flag: boolean;
  age: number | null;
  net_score: number | null;
}

export interface Analysis {
  id: number;
  candidate_id: number;
  score_leadership: number;
  score_motivation: number;
  score_growth: number;
  score_vision: number;
  score_communication: number;
  final_score: number;
  category: string;
  ai_generated_risk: string;
  ai_generated_score: number;
  incomplete_flag: boolean;
  explanation_leadership: string;
  explanation_motivation: string;
  explanation_growth: string;
  explanation_vision: string;
  explanation_communication: string;
  summary: string;
  key_strengths: string[];
  red_flags: string[];
  analyzed_at: string;
  model_used: string;
  recommended_major: string | null;
  major_reason_note: string | null;
}

export interface AnalysisHistoryEntry {
  id: number;
  candidate_id: number;
  score_leadership: number;
  score_motivation: number;
  score_growth: number;
  score_vision: number;
  score_communication: number;
  final_score: number;
  category: string;
  ai_generated_risk: string;
  ai_generated_score: number;
  summary: string | null;
  key_strengths: string[];
  red_flags: string[];
  model_used: string | null;
  analyzed_at: string;
  duration_ms: number;
}

export interface Decision {
  id: number;
  candidate_id: number;
  decision: string;
  notes: string | null;
  decided_by: number;
  decided_by_email: string | null;
  decided_at: string;
}

export interface InterviewAnalysis {
  score_leadership: number;
  score_grit: number;
  score_authenticity: number;
  score_motivation: number;
  score_vision: number;
  explanation_leadership: string;
  explanation_grit: string;
  explanation_authenticity: string;
  explanation_motivation: string;
  explanation_vision: string;
  final_score: number;
  category: string;
  consistency_score: number;
  style_match_score: number;
  suspicion_flags: string[];
  summary: string;
  strengths: string[];
  concerns: string[];
  analyzed_at: string;
  model_used: string;
}

export interface InterviewStatus {
  status: string;
  invite_status?: string;
  invite_token?: string;
  deep_link?: string;
  interview?: {
    id: number;
    status: string;
    language: string;
    questions_asked: number;
    started_at: string;
    completed_at: string | null;
    current_topic: string;
  };
  analysis?: InterviewAnalysis;
  combined_score?: number;
}

export interface InterviewMessage {
  role: string;
  content: string;
  message_type: string;
  voice_duration_sec: number;
  response_time_sec: number;
  created_at: string;
}

export interface CandidateDetail extends Candidate {
  analysis: Analysis | null;
  decisions: Decision[];
}

export interface User {
  id: number;
  email: string;
  full_name: string | null;
  role: string;
  avatar_url: string | null;
  created_at?: string;
}

export interface DashboardStats {
  total_candidates: number;
  analyzed: number;
  pending: number;
  shortlisted: number;
  rejected: number;
  waitlisted: number;
  avg_score: number;
  score_distribution: { range: string; count: number }[];
  category_counts: Record<string, number>;
  score_mean: number;
  score_median: number;
  dimension_means: Record<string, number>;
  dimension_distributions: Record<string, { range: string; count: number }[]>;
  ielts_distribution: { range: string; count: number }[];
  toefl_distribution: { range: string; count: number }[];
  unt_distribution: { range: string; count: number }[];
  nis_distribution: { range: string; count: number }[];
  ielts_count: number;
  toefl_count: number;
  unt_count: number;
  nis_count: number;
}

export interface MajorOption {
  tag: string;
  en: string;
  ru: string;
  kk: string;
}
