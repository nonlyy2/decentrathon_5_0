package handlers

import (
	"net/http"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ListUsers — tech-admin+ can see all users
func ListUsers(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, email, full_name, role, avatar_url, created_at FROM users ORDER BY created_at DESC`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch users"})
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.AvatarURL, &u.CreatedAt); err != nil {
				continue
			}
			users = append(users, u)
		}
		if users == nil {
			users = []models.User{}
		}
		c.JSON(200, users)
	}
}

// GetUser — get single user
func GetUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var u models.User
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, email, full_name, role, avatar_url, created_at FROM users WHERE id = $1`, id,
		).Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.AvatarURL, &u.CreatedAt)
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}
		c.JSON(200, u)
	}
}

// UpdateUser — tech-admin+ can update roles/full_name (cannot edit superadmins)
func UpdateUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		callerRole := c.GetString("user_role")
		callerID := c.GetInt("user_id")

		// Fetch target user's current role
		var targetRole string
		var targetID int
		err := pool.QueryRow(c.Request.Context(), `SELECT id, role FROM users WHERE id = $1`, id).Scan(&targetID, &targetRole)
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}

		// Only superadmin can edit another superadmin
		if (targetRole == "superadmin" || targetRole == "admin") && !middleware.IsRole(c, "superadmin", "admin") {
			c.JSON(403, gin.H{"error": "cannot modify superadmin accounts"})
			return
		}

		// Cannot edit yourself via this endpoint (use /profile instead)
		if targetID == callerID {
			c.JSON(400, gin.H{"error": "use /profile to edit your own account"})
			return
		}

		// tech-admin cannot set superadmin role
		if callerRole == "tech-admin" {
			var req models.UpdateUserRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			if req.Role != nil && (*req.Role == "superadmin" || *req.Role == "admin") {
				c.JSON(403, gin.H{"error": "tech-admin cannot assign superadmin role"})
				return
			}
			_, err = pool.Exec(c.Request.Context(),
				`UPDATE users SET role = COALESCE($1, role), full_name = COALESCE($2, full_name) WHERE id = $3`,
				req.Role, req.FullName, id)
		} else {
			var req models.UpdateUserRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			_, err = pool.Exec(c.Request.Context(),
				`UPDATE users SET role = COALESCE($1, role), full_name = COALESCE($2, full_name) WHERE id = $3`,
				req.Role, req.FullName, id)
		}

		if err != nil {
			c.JSON(500, gin.H{"error": "failed to update user"})
			return
		}
		c.JSON(200, gin.H{"message": "user updated"})
	}
}

// DeleteUser — only superadmin can delete users (except themselves)
func DeleteUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		callerID := c.GetInt("user_id")

		var targetID int
		var targetRole string
		err := pool.QueryRow(c.Request.Context(), `SELECT id, role FROM users WHERE id = $1`, id).Scan(&targetID, &targetRole)
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}

		if targetID == callerID {
			c.JSON(400, gin.H{"error": "cannot delete your own account"})
			return
		}
		if targetRole == "superadmin" || targetRole == "admin" {
			c.JSON(403, gin.H{"error": "cannot delete superadmin accounts"})
			return
		}

		_, err = pool.Exec(c.Request.Context(), `DELETE FROM users WHERE id = $1`, id)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to delete user"})
			return
		}
		c.JSON(200, gin.H{"message": "user deleted"})
	}
}

// GetProfile — returns the current user's profile
func GetProfile(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		var u models.User
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, email, full_name, role, avatar_url, created_at FROM users WHERE id = $1`, userID,
		).Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.AvatarURL, &u.CreatedAt)
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}
		c.JSON(200, u)
	}
}

// UpdateProfile — change own full_name, email (superadmin only), or password
func UpdateProfile(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		userRole := c.GetString("user_role")

		var req models.UpdateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Email change — superadmin only
		if req.Email != nil {
			if userRole != "superadmin" && userRole != "admin" {
				c.JSON(403, gin.H{"error": "only superadmin can change email"})
				return
			}
			_, err := pool.Exec(c.Request.Context(),
				`UPDATE users SET email = $1 WHERE id = $2`, *req.Email, userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to update email (may already be in use)"})
				return
			}
		}

		if req.Password != nil {
			hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 10)
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to hash password"})
				return
			}
			_, err = pool.Exec(c.Request.Context(),
				`UPDATE users SET full_name = COALESCE($1, full_name), password_hash = $2 WHERE id = $3`,
				req.FullName, string(hash), userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to update profile"})
				return
			}
		} else {
			_, err := pool.Exec(c.Request.Context(),
				`UPDATE users SET full_name = COALESCE($1, full_name) WHERE id = $2`,
				req.FullName, userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to update profile"})
				return
			}
		}

		c.JSON(200, gin.H{"message": "profile updated"})
	}
}

// GetMajors — return available majors
func GetMajors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, models.Majors)
	}
}
