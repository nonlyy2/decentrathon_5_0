export interface Candidate {
  id: number;
  full_name: string;
  email: string;
  phone: string | null;
  telegram: string | null;
  age: number | null;
  city: string | null;
  school: string | null;
  graduation_year: number | null;
  achievements: string | null;
  extracurriculars: string | null;
  essay: string;
  motivation_statement: string | null;
  created_at: string;
  status: string;
}

export interface CandidateListItem {
  id: number;
  full_name: string;
  email: string;
  city: string | null;
  school: string | null;
  status: string;
  created_at: string;
  final_score: number | null;
  category: string | null;
  analyzed_at: string | null;
  model_used: string | null;
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
}

export interface Decision {
  id: number;
  candidate_id: number;
  decision: string;
  notes: string | null;
  decided_by: number;
  decided_at: string;
}

export interface InterviewAnalysis {
  score_leadership: number;
  score_grit: number;
  score_authenticity: number;
  score_motivation: number;
  score_vision: number;
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
  interview?: {
    id: number;
    status: string;
    language: string;
    questions_asked: number;
    started_at: string;
    completed_at: string | null;
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
  role: string;
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
}
