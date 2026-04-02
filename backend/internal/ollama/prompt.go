package ollama

import (
	"fmt"
	"strings"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

// System prompt for local LLMs — expanded for better evaluation quality
// while keeping it structured enough for small models to follow.
const SystemPrompt = `You are an admissions evaluator for inVision U, a 100% scholarship university.
Evaluate the candidate application and respond with ONLY valid JSON. No markdown, no explanations.

SCORING CRITERIA (0-100 each):
- score_leadership: initiative, leading projects, creating things, taking responsibility. 80+ = founded orgs or led significant projects. 40- = no leadership signals.
- score_motivation: genuine passion vs generic answers, specific personal stories, authentic voice. 80+ = deeply personal, unique perspective. 40- = entirely generic or likely AI-generated.
- score_growth: improvement over time, overcoming obstacles, learning from failure, self-awareness. 80+ = clear growth arc. 40- = no growth evidence.
- score_vision: clarity of future goals, ambitious but realistic plans, understanding of impact. 80+ = crystal clear vision. 40- = no articulated goals.
- score_communication: structured writing, clarity, coherent narrative. 80+ = compelling and engaging. 40- = incoherent.

AI DETECTION (ai_generated_score 0-100):
Low (0-35) = human: personal details, natural voice, minor imperfections.
Medium (36-65) = mixed: some sections feel AI-assisted.
High (66-100) = likely AI: generic phrases, formulaic structure, no personal specifics.

LANGUAGE: Application must be in English. If >20% non-English, penalize communication score.

RULES:
- Focus on substance over polish. Authentic > polished but generic.
- Do not penalize imperfect English grammar from non-native speakers.
- Look for signals (specific examples) not just listed achievements.

final_score = leadership*0.25 + motivation*0.25 + growth*0.20 + vision*0.15 + communication*0.15
category: 80-100="Strong Recommend", 65-79="Recommend", 50-64="Borderline", 0-49="Not Recommended"
ai_generated_risk: 0-35="low", 36-65="medium", 66-100="high"`

// BuildPrompt creates the user message with depersonalized candidate data and JSON template.
// PII (name, email, phone, age, city, telegram) is excluded to prevent bias.
func BuildPrompt(c *models.Candidate) string {
	var sb strings.Builder

	sb.WriteString("=== CANDIDATE APPLICATION ===\n")
	if c.School != nil {
		sb.WriteString(fmt.Sprintf("School: %s\n", *c.School))
	}
	if c.GraduationYear != nil {
		sb.WriteString(fmt.Sprintf("Graduation Year: %d\n", *c.GraduationYear))
	}
	if c.Major != nil {
		sb.WriteString(fmt.Sprintf("Candidate's Chosen Major: %s\n", *c.Major))
	}

	if c.Achievements != nil && *c.Achievements != "" {
		sb.WriteString(fmt.Sprintf("\nAchievements:\n%s\n", *c.Achievements))
	}
	if c.Extracurriculars != nil && *c.Extracurriculars != "" {
		sb.WriteString(fmt.Sprintf("\nExtracurriculars:\n%s\n", *c.Extracurriculars))
	}

	sb.WriteString(fmt.Sprintf("\nEssay:\n%s\n", c.Essay))

	if c.MotivationStatement != nil && *c.MotivationStatement != "" {
		sb.WriteString(fmt.Sprintf("\nMotivation:\n%s\n", *c.MotivationStatement))
	}

	if c.YouTubeTranscript != nil && *c.YouTubeTranscript != "" {
		sb.WriteString(fmt.Sprintf("\nPresentation Video Transcript:\n%s\n", *c.YouTubeTranscript))
	}

	sb.WriteString(`
=== OUTPUT INSTRUCTIONS ===
Respond with ONLY this JSON object. Fill every field.
{
  "score_leadership": <0-100>,
  "score_motivation": <0-100>,
  "score_growth": <0-100>,
  "score_vision": <0-100>,
  "score_communication": <0-100>,
  "ai_generated_score": <0-100>,
  "summary": "<2-3 sentence assessment of this candidate>",
  "explanation_leadership": "<1-2 sentences with evidence. Quote the application in double quotes where possible.>",
  "explanation_motivation": "<1-2 sentences with evidence. Quote the application in double quotes where possible.>",
  "explanation_growth": "<1-2 sentences with evidence. Quote the application in double quotes where possible.>",
  "explanation_vision": "<1-2 sentences with evidence. Quote the application in double quotes where possible.>",
  "explanation_communication": "<1-2 sentences with evidence. Quote the application in double quotes where possible.>",
  "key_strengths": ["<strength 1>", "<strength 2>"],
  "red_flags": [],
  "recommended_major": "<one of: Engineering|Tech|Society|Policy Reform|Art + Media>",
  "major_reason_note": "<1 sentence if differs from chosen major, else empty string>"
}`)

	return sb.String()
}
