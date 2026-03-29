package ollama

import (
	"fmt"
	"strings"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

// Short system prompt specifically designed for local LLMs.
// Keep it minimal — long prompts confuse small models.
const SystemPrompt = `You are an admissions evaluator for inVision U, a scholarship university.
Evaluate the candidate application below and respond with ONLY a valid JSON object.
No explanations, no markdown, no text outside the JSON.

Score each dimension 0-100 based on evidence in the application:
- score_leadership: evidence of initiative, leading projects, taking responsibility
- score_motivation: genuine passion vs generic answers, authentic personal story
- score_growth: improvement over time, overcoming challenges, self-awareness
- score_vision: clarity of future goals, ambition, understanding of impact
- score_communication: clear structured writing, ability to express ideas

ai_generated_score (0-100): probability the essay was AI-written.
Low (0-35) = clearly human: personal details, natural voice, minor imperfections.
High (66-100) = likely AI: generic phrases, formulaic structure, lacks specific personal details.`

// BuildPrompt creates the full user message with candidate data and embedded JSON template.
// Embedding the template in the user message helps small models follow the schema.
func BuildPrompt(c *models.Candidate) string {
	var sb strings.Builder

	sb.WriteString("=== CANDIDATE APPLICATION ===\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", c.FullName))
	if c.Age != nil {
		sb.WriteString(fmt.Sprintf("Age: %d\n", *c.Age))
	}
	if c.City != nil {
		sb.WriteString(fmt.Sprintf("City: %s\n", *c.City))
	}
	if c.School != nil {
		sb.WriteString(fmt.Sprintf("School: %s\n", *c.School))
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

	sb.WriteString(`
=== OUTPUT INSTRUCTIONS ===
Respond with ONLY this JSON object. Fill every field with your evaluation.
{
  "score_leadership": <0-100>,
  "score_motivation": <0-100>,
  "score_growth": <0-100>,
  "score_vision": <0-100>,
  "score_communication": <0-100>,
  "ai_generated_score": <0-100>,
  "summary": "<2-3 sentence assessment of this candidate>",
  "explanation_leadership": "<1-2 sentences with evidence>",
  "explanation_motivation": "<1-2 sentences with evidence>",
  "explanation_growth": "<1-2 sentences with evidence>",
  "explanation_vision": "<1-2 sentences with evidence>",
  "explanation_communication": "<1-2 sentences with evidence>",
  "key_strengths": ["<strength 1>", "<strength 2>"],
  "red_flags": []
}`)

	return sb.String()
}
