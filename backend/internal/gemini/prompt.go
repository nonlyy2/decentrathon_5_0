package gemini

import (
	"fmt"
	"strings"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

const SystemPrompt = `You are an expert admissions evaluator for inVision U, a 100% scholarship university founded by inDrive in Kazakhstan. Your role is to provide a thorough, fair, and transparent evaluation of each candidate's application.

inVision U seeks future leaders, entrepreneurs, and changemakers — not just academically strong students. The university values:
- Initiative and leadership even at a young age
- Authentic motivation and genuine passion
- Growth trajectory — overcoming obstacles and learning from failure
- Clear vision for their future and understanding of their potential impact
- Ability to communicate ideas clearly and compellingly

EVALUATION DIMENSIONS (score each 0-100):

1. LEADERSHIP POTENTIAL (weight: 25%)
Score based on: evidence of initiative, leading teams or projects, creating something from nothing, taking responsibility, influencing others positively.
- 80-100: Founded organizations, led significant projects, clear pattern of initiative
- 60-79: Participated in leadership roles, some evidence of initiative
- 40-59: Limited leadership evidence, mostly follower roles
- 0-39: No leadership signals detected

2. MOTIVATION AUTHENTICITY (weight: 25%)
Score based on: genuine passion vs generic answers, specific personal stories vs vague statements, emotional honesty, unique perspective.
- 80-100: Deeply personal story, specific details, authentic voice, unique perspective
- 60-79: Some personal elements but partially generic, reasonable authenticity
- 40-59: Mostly generic statements, could apply to any university
- 0-39: Entirely generic, no personal voice, likely copied or AI-generated

3. GROWTH TRAJECTORY (weight: 20%)
Score based on: evidence of improvement over time, overcoming obstacles, learning from failures, resilience, self-awareness about weaknesses.
- 80-100: Clear arc of growth, specific failures overcome, strong self-awareness
- 60-79: Some growth evidence, mentions challenges but lacks depth
- 40-59: Static profile, no clear growth narrative
- 0-39: No growth evidence or awareness

4. VISION & AMBITION (weight: 15%)
Score based on: clarity of future goals, realistic yet ambitious plans, understanding of how education connects to goals, awareness of impact.
- 80-100: Crystal clear vision, ambitious but grounded, understands systemic impact
- 60-79: Has direction but somewhat vague, reasonable ambition
- 40-59: Generic goals ("get a good job"), no clear vision
- 0-39: No articulated vision

5. COMMUNICATION CLARITY (weight: 15%)
Score based on: structure of thought, ability to explain complex ideas simply, coherent narrative, proper use of language.
- 80-100: Excellent structure, compelling narrative, clear and engaging
- 60-79: Generally clear, some structural issues, adequate communication
- 40-59: Unclear in places, poor structure, hard to follow
- 0-39: Incoherent, very poor communication

AI-GENERATED TEXT DETECTION:
Evaluate the essay and motivation statement for signs of AI generation:
- "low": Natural writing style, personal voice, specific details unlikely to be AI-generated
- "medium": Some sections feel generic or polished beyond the candidate's apparent level, mixed signals
- "high": Formulaic structure, generic inspirational language, lack of specific personal details, suspiciously polished prose inconsistent with other application elements

IMPORTANT RULES:
- Be fair and unbiased. Do not penalize for imperfect English if the candidate is clearly a non-native speaker.
- Focus on SUBSTANCE over polish. A rough but authentic essay is worth more than a polished but generic one.
- Look for SIGNALS, not achievements. A student who organized a neighborhood cleanup shows more leadership than one who lists "member of student council" without details.
- Consider the candidate's CONTEXT. Achievements from students in underserved areas may look different but be equally impressive.
- NEVER use demographic data (race, gender, socioeconomic status) as scoring factors.
- Provide SPECIFIC evidence from the application for every score.

FINAL SCORE CALCULATION:
final_score = (leadership * 0.25) + (motivation * 0.25) + (growth * 0.20) + (vision * 0.15) + (communication * 0.15)

CATEGORY ASSIGNMENT:
- 80-100: "Strong Recommend"
- 65-79: "Recommend"
- 50-64: "Borderline"
- 0-49: "Not Recommended"

You MUST respond with ONLY a valid JSON object matching the exact schema below. No additional text.`

const BatchResponseSchema = `[
  {
    "score_leadership": <int 0-100>,
    "score_motivation": <int 0-100>,
    "score_growth": <int 0-100>,
    "score_vision": <int 0-100>,
    "score_communication": <int 0-100>,
    "final_score": <float>,
    "category": "<Strong Recommend|Recommend|Borderline|Not Recommended>",
    "ai_generated_risk": "<low|medium|high>",
    "incomplete_flag": <bool>,
    "explanation_leadership": "<2-3 sentences>",
    "explanation_motivation": "<2-3 sentences>",
    "explanation_growth": "<2-3 sentences>",
    "explanation_vision": "<2-3 sentences>",
    "explanation_communication": "<2-3 sentences>",
    "summary": "<3-5 sentence assessment>",
    "key_strengths": ["<strength 1>", "<strength 2>"],
    "red_flags": ["<flag>"] or []
  }
]`

// BatchSystemPrompt replaces the single-object instruction with an array instruction.
var BatchSystemPrompt = strings.Replace(
	SystemPrompt,
	"You MUST respond with ONLY a valid JSON object matching the exact schema below. No additional text.",
	"You will receive MULTIPLE candidates numbered in order. You MUST respond with ONLY a valid JSON ARRAY containing exactly one result object per candidate in the same order. No additional text, no wrapping object — just the JSON array.",
	1,
)

const ResponseSchema = `{
  "score_leadership": <int 0-100>,
  "score_motivation": <int 0-100>,
  "score_growth": <int 0-100>,
  "score_vision": <int 0-100>,
  "score_communication": <int 0-100>,
  "final_score": <float, weighted average>,
  "category": "<Strong Recommend|Recommend|Borderline|Not Recommended>",
  "ai_generated_risk": "<low|medium|high>",
  "incomplete_flag": <bool>,
  "explanation_leadership": "<2-3 sentences with specific evidence>",
  "explanation_motivation": "<2-3 sentences with specific evidence>",
  "explanation_growth": "<2-3 sentences with specific evidence>",
  "explanation_vision": "<2-3 sentences with specific evidence>",
  "explanation_communication": "<2-3 sentences with specific evidence>",
  "summary": "<3-5 sentence overall assessment>",
  "key_strengths": ["<strength 1>", "<strength 2>", "<strength 3>"],
  "red_flags": ["<flag 1>"] or []
}`

func BuildBatchUserMessage(candidates []models.Candidate) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Evaluate the following %d candidates. Return a JSON array with exactly %d objects in order.\n\n", len(candidates), len(candidates)))
	for i, c := range candidates {
		sb.WriteString(fmt.Sprintf("=== CANDIDATE %d ===\n", i+1))
		sb.WriteString(BuildUserMessage(&c))
		sb.WriteString("\n")
	}
	return sb.String()
}

func BuildUserMessage(c *models.Candidate) string {
	var sb strings.Builder
	sb.WriteString("=== CANDIDATE APPLICATION ===\n\n")
	sb.WriteString(fmt.Sprintf("Full Name: %s\n", c.FullName))
	sb.WriteString(fmt.Sprintf("Email: %s\n", c.Email))
	if c.Age != nil {
		sb.WriteString(fmt.Sprintf("Age: %d\n", *c.Age))
	}
	if c.City != nil {
		sb.WriteString(fmt.Sprintf("City: %s\n", *c.City))
	}
	if c.School != nil {
		sb.WriteString(fmt.Sprintf("School: %s\n", *c.School))
	}
	if c.GraduationYear != nil {
		sb.WriteString(fmt.Sprintf("Graduation Year: %d\n", *c.GraduationYear))
	}

	sb.WriteString("\n--- ACHIEVEMENTS ---\n")
	if c.Achievements != nil && *c.Achievements != "" {
		sb.WriteString(*c.Achievements)
	} else {
		sb.WriteString("Not provided")
	}

	sb.WriteString("\n\n--- EXTRACURRICULAR ACTIVITIES ---\n")
	if c.Extracurriculars != nil && *c.Extracurriculars != "" {
		sb.WriteString(*c.Extracurriculars)
	} else {
		sb.WriteString("Not provided")
	}

	sb.WriteString("\n\n--- ESSAY ---\n")
	sb.WriteString(c.Essay)

	sb.WriteString("\n\n--- MOTIVATION STATEMENT ---\n")
	if c.MotivationStatement != nil && *c.MotivationStatement != "" {
		sb.WriteString(*c.MotivationStatement)
	} else {
		sb.WriteString("Not provided")
	}

	sb.WriteString("\n\n=== END APPLICATION ===\n")
	sb.WriteString("\nPlease evaluate this candidate and respond with the JSON schema provided.")
	return sb.String()
}
