package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartnerSchool struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	City             *string   `json:"city"`
	ContactEmail     *string   `json:"contact_email"`
	ContactPhone     *string   `json:"contact_phone"`
	GraduatesPerYear int       `json:"graduates_per_year"`
	CreatedAt        time.Time `json:"created_at"`
	CandidateCount   int       `json:"candidate_count"`
	AvgAIScore       float64   `json:"avg_ai_score"`
}

func GetPartnerSchools(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		rows, err := pool.Query(ctx,
			`SELECT ps.id, ps.name, ps.city, ps.contact_email, ps.contact_phone, ps.graduates_per_year, ps.created_at,
				COALESCE((SELECT COUNT(*) FROM candidates c WHERE c.partner_school = ps.name), 0) AS candidate_count,
				COALESCE((SELECT AVG(a.final_score) FROM analyses a JOIN candidates c ON c.id = a.candidate_id WHERE c.partner_school = ps.name), 0) AS avg_score
			 FROM partner_schools ps
			 ORDER BY candidate_count DESC, ps.name ASC`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch partner schools"})
			return
		}
		defer rows.Close()

		schools := []PartnerSchool{}
		for rows.Next() {
			var s PartnerSchool
			if err := rows.Scan(&s.ID, &s.Name, &s.City, &s.ContactEmail, &s.ContactPhone,
				&s.GraduatesPerYear, &s.CreatedAt, &s.CandidateCount, &s.AvgAIScore); err != nil {
				continue
			}
			schools = append(schools, s)
		}

		c.JSON(http.StatusOK, gin.H{"schools": schools})
	}
}
