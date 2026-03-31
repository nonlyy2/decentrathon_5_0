package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Role hierarchy (higher index = more permissions)
var roleLevel = map[string]int{
	"committee":  1, // legacy alias for manager
	"manager":    1,
	"auditor":    2,
	"tech-admin": 3,
	"admin":      4, // legacy alias for superadmin
	"superadmin": 4,
}

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing or invalid token"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token claims"})
			return
		}

		c.Set("user_id", int(claims["user_id"].(float64)))
		c.Set("user_email", claims["email"].(string))
		c.Set("user_role", claims["role"].(string))
		c.Next()
	}
}

// AdminOnly — superadmin OR admin (legacy)
func AdminOnly() gin.HandlerFunc {
	return RoleAtLeast("admin")
}

// SuperAdminOnly — only superadmin (or legacy admin)
func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_role")
		if role != "superadmin" && role != "admin" {
			c.AbortWithStatusJSON(403, gin.H{"error": "superadmin access required"})
			return
		}
		c.Next()
	}
}

// TechAdminOrAbove — tech-admin, superadmin, admin
func TechAdminOrAbove() gin.HandlerFunc {
	return RoleAtLeast("tech-admin")
}

// AuditorOrAbove — auditor and above can read everything
func AuditorOrAbove() gin.HandlerFunc {
	return RoleAtLeast("auditor")
}

// ManagerOrAbove — basic dashboard access
func ManagerOrAbove() gin.HandlerFunc {
	return RoleAtLeast("manager")
}

// ReadOnly — auditors can only use GET
func AuditorReadOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_role")
		level := roleLevel[role]
		if level == roleLevel["auditor"] && c.Request.Method != "GET" {
			c.AbortWithStatusJSON(403, gin.H{"error": "auditors have read-only access"})
			return
		}
		c.Next()
	}
}

// RoleAtLeast checks that the user's role level is >= the required role's level
func RoleAtLeast(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_role")
		if roleLevel[role] < roleLevel[required] {
			c.AbortWithStatusJSON(403, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

// IsRole returns true if the user has exactly one of the given roles
func IsRole(c *gin.Context, roles ...string) bool {
	role := c.GetString("user_role")
	for _, r := range roles {
		if role == r {
			return true
		}
	}
	return false
}

// HasLevel returns true if the user's role level >= given level
func HasLevel(c *gin.Context, required string) bool {
	role := c.GetString("user_role")
	return roleLevel[role] >= roleLevel[required]
}
