package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func csvHeaders(lang string) []string {
	switch lang {
	case "ru":
		return []string{"ID", "Имя", "Email", "Возраст", "Город", "Школа", "Год выпуска", "Статус", "Дата создания",
			"Итоговый балл", "Категория", "Риск ИИ", "Модель", "Дата анализа"}
	case "kk":
		return []string{"ID", "Аты", "Email", "Жасы", "Қала", "Мектеп", "Бітіру жылы", "Мәртебе", "Жасалған күн",
			"Қорытынды балл", "Санат", "ЖИ тәуекелі", "Модель", "Талдау күні"}
	default:
		return []string{"ID", "Full Name", "Email", "Age", "City", "School", "Graduation Year", "Status", "Created At",
			"Final Score", "Category", "AI Risk", "Model Used", "Analyzed At"}
	}
}

func ExportCandidatesCSV(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rows, err := pool.Query(ctx,
			`SELECT c.id, c.full_name, c.email, COALESCE(c.age::text, ''), COALESCE(c.city, ''),
			        COALESCE(c.school, ''), COALESCE(c.graduation_year::text, ''), c.status, c.created_at,
			        COALESCE(a.final_score::text, ''), COALESCE(a.category, ''),
			        COALESCE(a.ai_generated_risk, ''), COALESCE(a.model_used, ''),
			        COALESCE(a.analyzed_at::text, '')
			 FROM candidates c
			 LEFT JOIN analyses a ON c.id = a.candidate_id
			 ORDER BY c.id`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to query candidates"})
			return
		}
		defer rows.Close()

		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=candidates_%s.csv", time.Now().Format("2006-01-02")))
		// UTF-8 BOM so Excel opens Cyrillic correctly
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

		lang := c.Query("lang")
		headers := csvHeaders(lang)

		w := csv.NewWriter(c.Writer)
		w.Write(headers)

		for rows.Next() {
			var id int
			var fullName, email, age, city, school, gradYear, status string
			var createdAt time.Time
			var finalScore, category, risk, model, analyzedAt string

			if err := rows.Scan(&id, &fullName, &email, &age, &city, &school, &gradYear, &status, &createdAt,
				&finalScore, &category, &risk, &model, &analyzedAt); err != nil {
				continue
			}

			w.Write([]string{
				fmt.Sprintf("%d", id), fullName, email, age, city, school, gradYear, status,
				createdAt.Format("2006-01-02"),
				finalScore, category, risk, model, analyzedAt,
			})
		}

		w.Flush()
		c.Status(http.StatusOK)
	}
}
