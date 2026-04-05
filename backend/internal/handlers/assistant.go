package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const managerAssistantSystemPrompt = `You are an AI data assistant for inVision U admissions managers.
You have been given a summary of the current admissions data and can help managers:
- Analyze candidate distributions and trends
- Answer questions about specific cohorts (by major, score, status, city, school)
- Suggest which candidates need attention
- Generate textual summaries of data
- Help prioritize workload
- Provide statistical analysis of score distributions

When the manager asks for a chart, graph, table, or diagram, respond with the data in a structured JSON block wrapped in ` + "`" + `~~~chart` + "`" + ` markers. Use this format:
~~~chart
{"type":"bar","title":"Chart Title","data":[{"label":"X","value":123}]}
~~~
Supported chart types: bar, pie, table.
For tables: {"type":"table","title":"Table Title","headers":["Col1","Col2"],"rows":[["val1","val2"]]}

Be concise, data-driven, and actionable. Always base your answers on the provided data context.
IMPORTANT: You have access to city, school, and partner school distribution data. Use it when asked about geographic or school-based questions.`

const userAssistantSystemPrompt = `You are a helpful FAQ assistant for inVision U applicants.
inVision U is a 100% scholarship university founded by inDrive in Kazakhstan.

You can only answer questions based on the following information:
- The application is free and fully funded
- Applicants submit an essay, motivation statement, and optionally a YouTube video
- The process has two stages: AI essay analysis (Stage 1) and Telegram interview (Stage 2)
- Majors: Creative Engineering, Innovative IT Product Design, Sociology: Leadership and Innovation, Public Policy and Development, Digital Media and Marketing
- The university values leadership, authentic motivation, growth mindset, vision, and communication
- Application status can be: pending, analyzed, shortlisted, waitlisted, rejected
- Candidates who are shortlisted proceed to the Telegram interview stage
- All materials must be in English

Do NOT make up information not listed above. If you don't know the answer, say so and suggest contacting the admissions team.
Be friendly, encouraging, and concise.`

// AssistantChat handles AI assistant chat for both managers and regular users.
func AssistantChat(pool *pgxpool.Pool, textGens AITextGenerators, defaultProvider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Message  string `json:"message" binding:"required,min=1,max=2000"`
			Provider string `json:"provider"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		provider := req.Provider
		if provider == "" {
			provider = defaultProvider
		}
		gen, ok := textGens[provider]
		if !ok {
			for k, v := range textGens {
				provider = k
				gen = v
				break
			}
		}
		if gen == nil {
			c.JSON(503, gin.H{"error": "AI service not available"})
			return
		}

		isManager := middleware.HasLevel(c, "manager")

		var systemPrompt, contextData string

		if isManager {
			// Build data context for manager assistant
			contextData = buildManagerDataContext(c.Request.Context(), pool)
			systemPrompt = managerAssistantSystemPrompt + "\n\n=== CURRENT ADMISSIONS DATA ===\n" + contextData
		} else {
			systemPrompt = userAssistantSystemPrompt
		}
		response, err := gen(c.Request.Context(), systemPrompt, req.Message)
		if err != nil {
			c.JSON(500, gin.H{"error": "AI assistant failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"reply":    response,
			"provider": provider,
			"mode":     map[bool]string{true: "manager", false: "user"}[isManager],
		})
	}
}

// buildManagerDataContext returns a text summary of current admissions stats for the AI context.
func buildManagerDataContext(ctx context.Context, pool *pgxpool.Pool) string {
	var sb strings.Builder

	// Status counts
	rows, _ := pool.Query(ctx, `SELECT status, COUNT(*) FROM candidates GROUP BY status ORDER BY COUNT(*) DESC`)
	if rows != nil {
		sb.WriteString("Candidate status breakdown:\n")
		for rows.Next() {
			var status string
			var count int
			if rows.Scan(&status, &count) == nil {
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", status, count))
			}
		}
		rows.Close()
	}

	// Score stats
	var avgScore, minScore, maxScore float64
	var analyzed int
	pool.QueryRow(ctx, `SELECT COALESCE(AVG(final_score),0), COALESCE(MIN(final_score),0), COALESCE(MAX(final_score),0), COUNT(*) FROM analyses`).Scan(&avgScore, &minScore, &maxScore, &analyzed)
	sb.WriteString(fmt.Sprintf("\nAI Analysis scores (n=%d):\n  Average: %.1f | Min: %.1f | Max: %.1f\n", analyzed, avgScore, minScore, maxScore))

	// Category distribution
	catRows, _ := pool.Query(ctx, `SELECT category, COUNT(*) FROM analyses GROUP BY category ORDER BY COUNT(*) DESC`)
	if catRows != nil {
		sb.WriteString("\nCategory distribution:\n")
		for catRows.Next() {
			var cat string
			var cnt int
			if catRows.Scan(&cat, &cnt) == nil {
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", cat, cnt))
			}
		}
		catRows.Close()
	}

	// Major distribution
	majRows, _ := pool.Query(ctx, `SELECT COALESCE(major,'Unknown'), COUNT(*) FROM candidates GROUP BY major ORDER BY COUNT(*) DESC LIMIT 10`)
	if majRows != nil {
		sb.WriteString("\nMajor distribution (candidates):\n")
		for majRows.Next() {
			var major string
			var cnt int
			if majRows.Scan(&major, &cnt) == nil {
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", major, cnt))
			}
		}
		majRows.Close()
	}

	// AI detection summary
	var highRisk, medRisk, lowRisk int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM analyses WHERE ai_generated_risk='high'`).Scan(&highRisk)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM analyses WHERE ai_generated_risk='medium'`).Scan(&medRisk)
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM analyses WHERE ai_generated_risk='low'`).Scan(&lowRisk)
	sb.WriteString(fmt.Sprintf("\nAI-generated text risk:\n  High: %d | Medium: %d | Low: %d\n", highRisk, medRisk, lowRisk))

	// Top 5 candidates by score (name only, no personal data)
	topRows, _ := pool.Query(ctx, `SELECT c.full_name, a.final_score, a.category, COALESCE(c.major,'Unknown') FROM candidates c JOIN analyses a ON c.id=a.candidate_id ORDER BY a.final_score DESC LIMIT 5`)
	if topRows != nil {
		sb.WriteString("\nTop 5 candidates by AI score:\n")
		for topRows.Next() {
			var name, cat, major string
			var score float64
			if topRows.Scan(&name, &score, &cat, &major) == nil {
				sb.WriteString(fmt.Sprintf("  - %s: %.1f (%s) — major: %s\n", name, score, cat, major))
			}
		}
		topRows.Close()
	}

	// City distribution
	cityRows, _ := pool.Query(ctx, `SELECT COALESCE(city,'Unknown'), COUNT(*) FROM candidates GROUP BY city ORDER BY COUNT(*) DESC LIMIT 15`)
	if cityRows != nil {
		sb.WriteString("\nCity distribution:\n")
		for cityRows.Next() {
			var city string
			var cnt int
			if cityRows.Scan(&city, &cnt) == nil {
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", city, cnt))
			}
		}
		cityRows.Close()
	}

	// School distribution
	schoolRows, _ := pool.Query(ctx, `SELECT COALESCE(school,'Unknown'), COUNT(*) FROM candidates GROUP BY school ORDER BY COUNT(*) DESC LIMIT 10`)
	if schoolRows != nil {
		sb.WriteString("\nSchool distribution (top 10):\n")
		for schoolRows.Next() {
			var school string
			var cnt int
			if schoolRows.Scan(&school, &cnt) == nil {
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", school, cnt))
			}
		}
		schoolRows.Close()
	}

	// Partner school distribution
	partnerRows, _ := pool.Query(ctx, `SELECT COALESCE(partner_school,'No partner'), COUNT(*) FROM candidates WHERE partner_school IS NOT NULL GROUP BY partner_school ORDER BY COUNT(*) DESC`)
	if partnerRows != nil {
		sb.WriteString("\nPartner school candidates:\n")
		for partnerRows.Next() {
			var ps string
			var cnt int
			if partnerRows.Scan(&ps, &cnt) == nil {
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", ps, cnt))
			}
		}
		partnerRows.Close()
	}

	// IELTS/TOEFL summary
	var ieltsCount, toeflCount int
	var avgIELTS, avgTOEFL float64
	pool.QueryRow(ctx, `SELECT COUNT(*), COALESCE(AVG(ielts_score),0) FROM candidates WHERE exam_type='IELTS' AND ielts_score IS NOT NULL`).Scan(&ieltsCount, &avgIELTS)
	pool.QueryRow(ctx, `SELECT COUNT(*), COALESCE(AVG(toefl_score),0) FROM candidates WHERE exam_type='TOEFL' AND toefl_score IS NOT NULL`).Scan(&toeflCount, &avgTOEFL)
	sb.WriteString(fmt.Sprintf("\nEnglish proficiency:\n  IELTS: %d candidates, avg: %.1f\n  TOEFL: %d candidates, avg: %.0f\n", ieltsCount, avgIELTS, toeflCount, avgTOEFL))

	return sb.String()
}
