package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func csvHeaders(lang string) []string {
	switch lang {
	case "ru":
		return []string{
			"ID", "Имя", "Email", "Телефон", "Telegram", "Возраст", "Город", "Школа", "Год выпуска",
			"Достижения", "Внеучебная деятельность", "Эссе", "Мотивационное письмо",
			"Особые потребности", "Специальность", "Статус", "Дата создания",
			"Итоговый балл", "Категория", "Лидерство", "Мотивация", "Рост", "Видение", "Коммуникация",
			"Риск ИИ", "ИИ %", "Резюме", "Сильные стороны", "Красные флаги", "Модель", "Дата анализа",
			"Статус интервью", "Комбинированный балл",
		}
	case "kk":
		return []string{
			"ID", "Аты", "Email", "Телефон", "Telegram", "Жасы", "Қала", "Мектеп", "Бітіру жылы",
			"Жетістіктер", "Сабақтан тыс", "Эссе", "Мотивациялық хат",
			"Арнайы қажеттіліктер", "Мамандық", "Мәртебе", "Жасалған күн",
			"Қорытынды балл", "Санат", "Көшбасшылық", "Мотивация", "Өсу", "Көзқарас", "Коммуникация",
			"ЖИ тәуекелі", "ЖИ %", "Түйіндеме", "Күшті жақтар", "Қызыл жалаулар", "Модель", "Талдау күні",
			"Сұхбат мәртебесі", "Біріктірілген балл",
		}
	default:
		return []string{
			"ID", "Full Name", "Email", "Phone", "Telegram", "Age", "City", "School", "Graduation Year",
			"Achievements", "Extracurriculars", "Essay", "Motivation Statement",
			"Disability / Special Needs", "Major", "Status", "Created At",
			"Final Score", "Category", "Leadership", "Motivation", "Growth", "Vision", "Communication",
			"AI Risk", "AI %", "Summary", "Key Strengths", "Red Flags", "Model Used", "Analyzed At",
			"Interview Status", "Combined Score",
		}
	}
}

func ExportCandidatesCSV(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rows, err := pool.Query(ctx,
			`SELECT c.id, c.full_name, c.email, COALESCE(c.phone, ''), COALESCE(c.telegram, ''),
			        COALESCE(c.age::text, ''), COALESCE(c.city, ''),
			        COALESCE(c.school, ''), COALESCE(c.graduation_year::text, ''),
			        COALESCE(c.achievements, ''), COALESCE(c.extracurriculars, ''),
			        COALESCE(c.essay, ''), COALESCE(c.motivation_statement, ''),
			        COALESCE(c.disability, ''), COALESCE(c.major, ''),
			        c.status, c.created_at,
			        COALESCE(a.final_score::text, ''), COALESCE(a.category, ''),
			        COALESCE(a.score_leadership::text, ''), COALESCE(a.score_motivation::text, ''),
			        COALESCE(a.score_growth::text, ''), COALESCE(a.score_vision::text, ''),
			        COALESCE(a.score_communication::text, ''),
			        COALESCE(a.ai_generated_risk, ''), COALESCE(a.ai_generated_score::text, '0'),
			        COALESCE(a.summary, ''),
			        COALESCE(a.key_strengths, ARRAY[]::text[]), COALESCE(a.red_flags, ARRAY[]::text[]),
			        COALESCE(a.model_used, ''), COALESCE(a.analyzed_at::text, ''),
			        COALESCE(c.interview_status, 'not_invited'), COALESCE(c.combined_score::text, '')
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
			var fullName, email, phone, telegram, age, city, school, gradYear string
			var achievements, extracurriculars, essay, motivation, disability, major string
			var status string
			var createdAt time.Time
			var finalScore, category string
			var scoreLeadership, scoreMotivation, scoreGrowth, scoreVision, scoreCommunication string
			var risk, aiScore, summary string
			var strengths, redFlags []string
			var model, analyzedAt string
			var interviewStatus, combinedScore string

			if err := rows.Scan(&id, &fullName, &email, &phone, &telegram, &age, &city, &school, &gradYear,
				&achievements, &extracurriculars, &essay, &motivation,
				&disability, &major, &status, &createdAt,
				&finalScore, &category,
				&scoreLeadership, &scoreMotivation, &scoreGrowth, &scoreVision, &scoreCommunication,
				&risk, &aiScore, &summary, &strengths, &redFlags,
				&model, &analyzedAt,
				&interviewStatus, &combinedScore); err != nil {
				continue
			}

			w.Write([]string{
				fmt.Sprintf("%d", id), fullName, email, phone, telegram, age, city, school, gradYear,
				achievements, extracurriculars, essay, motivation,
				disability, major, status, createdAt.Format("2006-01-02"),
				finalScore, category,
				scoreLeadership, scoreMotivation, scoreGrowth, scoreVision, scoreCommunication,
				risk, aiScore, summary,
				strings.Join(strengths, "; "), strings.Join(redFlags, "; "),
				model, analyzedAt,
				interviewStatus, combinedScore,
			})
		}

		w.Flush()
		c.Status(http.StatusOK)
	}
}

// ImportCandidatesCSV imports candidates from a CSV file upload.
// Expected columns: Full Name, Email, Phone, Telegram, Age, City, School, Graduation Year,
//
//	Achievements, Extracurriculars, Essay, Motivation Statement, Disability, Major
//
// POST /candidates/import/csv
func ImportCandidatesCSV(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "CSV file is required (field: file)"})
			return
		}
		defer file.Close()

		// Read entire file to handle BOM
		raw, err := io.ReadAll(file)
		if err != nil {
			c.JSON(400, gin.H{"error": "failed to read file"})
			return
		}

		// Strip UTF-8 BOM
		content := string(raw)
		content = strings.TrimPrefix(content, "\xEF\xBB\xBF")

		reader := csv.NewReader(strings.NewReader(content))
		reader.LazyQuotes = true
		reader.TrimLeadingSpace = true

		records, err := reader.ReadAll()
		if err != nil {
			c.JSON(400, gin.H{"error": "failed to parse CSV: " + err.Error()})
			return
		}
		if len(records) < 2 {
			c.JSON(400, gin.H{"error": "CSV must have a header row and at least one data row"})
			return
		}

		// Detect column indices by header names
		header := records[0]
		colMap := map[string]int{}
		for i, h := range header {
			key := strings.ToLower(strings.TrimSpace(h))
			colMap[key] = i
		}

		// Column name aliases
		nameAliases := map[string][]string{
			"full_name":            {"full name", "full_name", "имя", "аты", "name"},
			"email":               {"email", "e-mail", "почта"},
			"phone":               {"phone", "телефон"},
			"telegram":            {"telegram", "tg"},
			"age":                 {"age", "возраст", "жасы"},
			"city":                {"city", "город", "қала"},
			"school":              {"school", "школа", "мектеп"},
			"graduation_year":     {"graduation year", "graduation_year", "год выпуска", "бітіру жылы"},
			"achievements":        {"achievements", "достижения", "жетістіктер"},
			"extracurriculars":    {"extracurriculars", "внеучебная деятельность", "сабақтан тыс"},
			"essay":               {"essay", "эссе"},
			"motivation_statement": {"motivation statement", "motivation_statement", "мотивационное письмо", "мотивациялық хат"},
			"disability":          {"disability", "disability / special needs", "особые потребности", "арнайы қажеттіліктер"},
			"major":               {"major", "специальность", "мамандық"},
		}

		colIdx := map[string]int{}
		for field, aliases := range nameAliases {
			for _, alias := range aliases {
				if idx, ok := colMap[alias]; ok {
					colIdx[field] = idx
					break
				}
			}
		}

		// Require at least full_name and email
		if _, ok := colIdx["full_name"]; !ok {
			c.JSON(400, gin.H{"error": "CSV must contain a 'Full Name' column"})
			return
		}
		if _, ok := colIdx["email"]; !ok {
			c.JSON(400, gin.H{"error": "CSV must contain an 'Email' column"})
			return
		}

		getCol := func(row []string, field string) string {
			idx, ok := colIdx[field]
			if !ok || idx >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[idx])
		}

		ctx := c.Request.Context()
		imported := 0
		skipped := 0
		var errors []string

		for i, row := range records[1:] {
			rowNum := i + 2 // 1-indexed, skip header

			fullName := getCol(row, "full_name")
			email := getCol(row, "email")
			if fullName == "" || email == "" {
				skipped++
				continue
			}

			phone := getCol(row, "phone")
			telegram := getCol(row, "telegram")
			ageStr := getCol(row, "age")
			city := getCol(row, "city")
			school := getCol(row, "school")
			gradYearStr := getCol(row, "graduation_year")
			achievements := getCol(row, "achievements")
			extracurriculars := getCol(row, "extracurriculars")
			essay := getCol(row, "essay")
			motivation := getCol(row, "motivation_statement")
			disability := getCol(row, "disability")
			major := getCol(row, "major")

			if essay == "" {
				essay = "N/A" // essay is NOT NULL in DB
			}

			var age *int
			if ageStr != "" {
				if v, err := strconv.Atoi(ageStr); err == nil {
					age = &v
				}
			}
			var gradYear *int
			if gradYearStr != "" {
				if v, err := strconv.Atoi(gradYearStr); err == nil {
					gradYear = &v
				}
			}

			_, err := pool.Exec(ctx,
				`INSERT INTO candidates (full_name, email, phone, telegram, age, city, school, graduation_year,
				 achievements, extracurriculars, essay, motivation_statement, disability, major, status)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,'pending')
				 ON CONFLICT (email) DO NOTHING`,
				fullName, email,
				nilIfEmpty(phone), nilIfEmpty(telegram), age, nilIfEmpty(city), nilIfEmpty(school), gradYear,
				nilIfEmpty(achievements), nilIfEmpty(extracurriculars), essay, nilIfEmpty(motivation),
				nilIfEmpty(disability), nilIfEmpty(major))
			if err != nil {
				log.Printf("CSV import row %d error: %v", rowNum, err)
				errors = append(errors, fmt.Sprintf("row %d (%s): %v", rowNum, email, err))
				skipped++
				continue
			}
			imported++
		}

		result := gin.H{
			"imported": imported,
			"skipped":  skipped,
			"message":  fmt.Sprintf("Imported %d candidates, skipped %d", imported, skipped),
		}
		if len(errors) > 0 {
			result["errors"] = errors
		}
		c.JSON(http.StatusOK, result)
	}
}

