package handlers

import (
	"fmt"
	"net/http"
	"time"
	"unicode"

	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// validatePassword — мин. 8 символов, латиница, 1 верхний/нижний, 1 цифра, 1 спецсимвол
func validatePassword(pw string) error {
	if len(pw) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range pw {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			if unicode.IsLetter(ch) && ch > 127 {
				return fmt.Errorf("password must use Latin (English) letters only")
			}
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
				hasSpecial = true
			}
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter (A-Z)")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter (a-z)")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit (0-9)")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character (!@#$%%^&* etc.)")
	}
	return nil
}

func Register(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		if err := validatePassword(req.Password); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to hash password"})
			return
		}

		var user models.User
		err = pool.QueryRow(c.Request.Context(),
			`INSERT INTO users (email, password_hash, role, full_name) VALUES ($1, $2, $3, $4)
			 RETURNING id, email, full_name, role, created_at`,
			req.Email, string(hash), req.Role, req.FullName,
		).Scan(&user.ID, &user.Email, &user.FullName, &user.Role, &user.CreatedAt)

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
