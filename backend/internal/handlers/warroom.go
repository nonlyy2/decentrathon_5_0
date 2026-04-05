package handlers

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// mentionRegex ищет @email или @username в тексте поста
var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}|[a-zA-Z0-9_\-]+)`)

type ActivityEntry struct {
	ID          int     `json:"id"`
	UserID      *int    `json:"user_id"`
	UserEmail   *string `json:"user_email"`
	UserName    *string `json:"user_name"`
	CandidateID *int    `json:"candidate_id"`
	CandidateName *string `json:"candidate_name"`
	ActionType  string  `json:"action_type"`
	Content     *string `json:"content"`
	Mentions    []byte  `json:"mentions"`
	CreatedAt   string  `json:"created_at"`
}

// GetActivityFeed возвращает глобальную ленту активности (постранично, сначала новые).
func GetActivityFeed(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if limit > 100 {
			limit = 100
		}

		rows, err := pool.Query(c.Request.Context(), `
			SELECT
				af.id, af.user_id, u.email, u.full_name,
				af.candidate_id, c.full_name,
				af.action_type, af.content, af.mentions, af.created_at::text
			FROM activity_feed af
			LEFT JOIN users u ON u.id = af.user_id
			LEFT JOIN candidates c ON c.id = af.candidate_id
			ORDER BY af.created_at DESC
			LIMIT $1 OFFSET $2
		`, limit, offset)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch activity feed"})
			return
		}
		defer rows.Close()

		var entries []ActivityEntry
		for rows.Next() {
			var e ActivityEntry
			if err := rows.Scan(
				&e.ID, &e.UserID, &e.UserEmail, &e.UserName,
				&e.CandidateID, &e.CandidateName,
				&e.ActionType, &e.Content, &e.Mentions, &e.CreatedAt,
			); err == nil {
				entries = append(entries, e)
			}
		}
		if entries == nil {
			entries = []ActivityEntry{}
		}

		var total int
		pool.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM activity_feed`).Scan(&total)

		c.JSON(http.StatusOK, gin.H{"entries": entries, "total": total})
	}
}

// PostActivityFeed публикует запись в ленту активности с опциональными упоминаниями.
func PostActivityFeed(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CandidateID *int   `json:"candidate_id"`
			ActionType  string `json:"action_type" binding:"required"`
			Content     string `json:"content" binding:"required,min=1,max=2000"`
			Mentions    []int  `json:"mentions"` // явно переданные user ID
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		userID, _ := c.Get("user_id")

		mentionHandles := mentionRegex.FindAllStringSubmatch(req.Content, -1)
		mentionedUserIDs := req.Mentions
		seen := map[int]bool{}
		for _, existing := range mentionedUserIDs {
			seen[existing] = true
		}
		for _, match := range mentionHandles {
			handle := match[1]
			var uid int
			if err2 := pool.QueryRow(c.Request.Context(),
				`SELECT id FROM users WHERE email = $1 OR full_name ILIKE $1 LIMIT 1`, handle,
			).Scan(&uid); err2 == nil && !seen[uid] {
				mentionedUserIDs = append(mentionedUserIDs, uid)
				seen[uid] = true
			}
		}

		var entryID int
		err := pool.QueryRow(c.Request.Context(), `
			INSERT INTO activity_feed (user_id, candidate_id, action_type, content, mentions)
			VALUES ($1, $2, $3, $4, $5::jsonb)
			RETURNING id
		`, userID, req.CandidateID, req.ActionType, req.Content,
			mentionsToJSON(mentionedUserIDs),
		).Scan(&entryID)

		if err != nil {
			c.JSON(500, gin.H{"error": "failed to post to activity feed"})
			return
		}

		for _, mentionedUID := range mentionedUserIDs {
			pool.Exec(c.Request.Context(), `
				INSERT INTO notifications (user_id, from_user_id, candidate_id, activity_id, type)
				VALUES ($1, $2, $3, $4, 'mention')
				ON CONFLICT DO NOTHING
			`, mentionedUID, userID, req.CandidateID, entryID)
		}

		c.JSON(http.StatusCreated, gin.H{"id": entryID, "message": "posted"})
	}
}

// MarkForDiscussion помечает кандидата для обсуждения в War Room.
func MarkForDiscussion(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var req struct {
			Note   string `json:"note"`
			Remove bool   `json:"remove"`
		}
		c.ShouldBindJSON(&req)

		userID, _ := c.Get("user_id")

		if req.Remove {
			_, err = pool.Exec(c.Request.Context(),
				`UPDATE candidates SET needs_discussion=false, discussion_note=NULL, discussion_by=NULL, discussion_at=NULL WHERE id=$1`,
				candidateID)
		} else {
			_, err = pool.Exec(c.Request.Context(),
				`UPDATE candidates SET needs_discussion=true, discussion_note=$1, discussion_by=$2, discussion_at=NOW() WHERE id=$3`,
				req.Note, userID, candidateID)

			if err == nil {
				var candidateName string
				pool.QueryRow(c.Request.Context(), `SELECT full_name FROM candidates WHERE id=$1`, candidateID).Scan(&candidateName)
				content := "Marked for discussion"
				if req.Note != "" {
					content += ": " + req.Note
				}
				pool.Exec(c.Request.Context(), `
					INSERT INTO activity_feed (user_id, candidate_id, action_type, content)
					VALUES ($1, $2, 'needs_discussion', $3)
				`, userID, candidateID, content)
			}
		}

		if err != nil {
			c.JSON(500, gin.H{"error": "failed to update discussion flag"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	}
}

// GetDiscussionCandidates возвращает всех кандидатов, помеченных для обсуждения.
func GetDiscussionCandidates(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(), `
			SELECT c.id, c.full_name, c.email, c.status, c.major,
				c.discussion_note, u.email, c.discussion_at::text,
				a.final_score, a.category
			FROM candidates c
			LEFT JOIN users u ON u.id = c.discussion_by
			LEFT JOIN analyses a ON a.candidate_id = c.id
			WHERE c.needs_discussion = true
			ORDER BY c.discussion_at DESC
		`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch discussion candidates"})
			return
		}
		defer rows.Close()

		type DiscussionCandidate struct {
			ID           int     `json:"id"`
			FullName     string  `json:"full_name"`
			Email        string  `json:"email"`
			Status       string  `json:"status"`
			Major        *string `json:"major"`
			Note         *string `json:"discussion_note"`
			FlaggedBy    *string `json:"flagged_by"`
			FlaggedAt    *string `json:"flagged_at"`
			FinalScore   *float64 `json:"final_score"`
			Category     *string  `json:"category"`
		}

		var list []DiscussionCandidate
		for rows.Next() {
			var d DiscussionCandidate
			if rows.Scan(
				&d.ID, &d.FullName, &d.Email, &d.Status, &d.Major,
				&d.Note, &d.FlaggedBy, &d.FlaggedAt,
				&d.FinalScore, &d.Category,
			) == nil {
				list = append(list, d)
			}
		}
		if list == nil {
			list = []DiscussionCandidate{}
		}
		c.JSON(http.StatusOK, list)
	}
}

// GetNotifications возвращает уведомления текущего пользователя.
func GetNotifications(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		rows, err := pool.Query(c.Request.Context(), `
			SELECT
				n.id, n.type, n.read, n.created_at::text,
				fu.email, fu.full_name,
				c.id, c.full_name,
				af.content, af.action_type
			FROM notifications n
			LEFT JOIN users fu ON fu.id = n.from_user_id
			LEFT JOIN candidates c ON c.id = n.candidate_id
			LEFT JOIN activity_feed af ON af.id = n.activity_id
			WHERE n.user_id = $1
			ORDER BY n.created_at DESC
			LIMIT 50
		`, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch notifications"})
			return
		}
		defer rows.Close()

		type Notification struct {
			ID            int     `json:"id"`
			Type          string  `json:"type"`
			Read          bool    `json:"read"`
			CreatedAt     string  `json:"created_at"`
			FromEmail     *string `json:"from_email"`
			FromName      *string `json:"from_name"`
			CandidateID   *int    `json:"candidate_id"`
			CandidateName *string `json:"candidate_name"`
			Content       *string `json:"content"`
			ActionType    *string `json:"action_type"`
		}

		var notifs []Notification
		for rows.Next() {
			var n Notification
			if rows.Scan(
				&n.ID, &n.Type, &n.Read, &n.CreatedAt,
				&n.FromEmail, &n.FromName,
				&n.CandidateID, &n.CandidateName,
				&n.Content, &n.ActionType,
			) == nil {
				notifs = append(notifs, n)
			}
		}
		if notifs == nil {
			notifs = []Notification{}
		}

		var unread int
		pool.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read=false`, userID).Scan(&unread)

		c.JSON(http.StatusOK, gin.H{"notifications": notifs, "unread": unread})
	}
}

// MarkNotificationsRead отмечает все уведомления пользователя как прочитанные.
func MarkNotificationsRead(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		pool.Exec(c.Request.Context(), `UPDATE notifications SET read=true WHERE user_id=$1`, userID)
		c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
	}
}

func mentionsToJSON(mentions []int) string {
	if len(mentions) == 0 {
		return "[]"
	}
	s := "["
	for i, m := range mentions {
		if i > 0 {
			s += ","
		}
		s += strconv.Itoa(m)
	}
	s += "]"
	return s
}
