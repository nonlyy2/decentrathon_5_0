package handlers

import (
	"net/http"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Register(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to hash password"})
			return
		}

		var user models.User
		err = pool.QueryRow(c.Request.Context(),
			`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
			 RETURNING id, email, role, created_at`,
			req.Email, string(hash), req.Role,
		).Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt)

		if err != nil {
			if isDuplicateKey(err) {
				c.JSON(409, gin.H{"error": "email already registered"})
				return
			}
			c.JSON(500, gin.H{"error": "failed to create user"})
			return
		}

		c.JSON(201, user)
	}
}

func Login(pool *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		var user models.User
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`,
			req.Email,
		).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"role":    user.Role,
			"exp":     time.Now().Add(72 * time.Hour).Unix(),
		})

		tokenStr, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, models.LoginResponse{
			Token: tokenStr,
			User:  user,
		})
	}
}

func isDuplicateKey(err error) bool {
	return err != nil && contains(err.Error(), "duplicate key")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
