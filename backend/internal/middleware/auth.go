package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// иерархия ролей (выше = больше прав)
var roleLevel = map[string]int{
	"committee":  1, // legacy = manager
	"manager":    1,
	"auditor":    2,
	"tech-admin": 3,
	"admin":      4, // legacy = superadmin
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

// superadmin или legacy admin
func AdminOnly() gin.HandlerFunc {
	return RoleAtLeast("admin")
}

// только superadmin (или legacy admin)
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

// tech-admin и выше
func TechAdminOrAbove() gin.HandlerFunc {
	return RoleAtLeast("tech-admin")
}

// auditor и выше
func AuditorOrAbove() gin.HandlerFunc {
	return RoleAtLeast("auditor")
}

// manager и выше
func ManagerOrAbove() gin.HandlerFunc {
	return RoleAtLeast("manager")
}

// auditor — только GET
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

// проверка минимального уровня роли
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

// true если роль пользователя совпадает с одной из переданных
func IsRole(c *gin.Context, roles ...string) bool {
	role := c.GetString("user_role")
	for _, r := range roles {
		if role == r {
			return true
		}
	}
	return false
}

// true если уровень роли >= required
func HasLevel(c *gin.Context, required string) bool {
	role := c.GetString("user_role")
	return roleLevel[role] >= roleLevel[required]
}

// ID пользователя из контекста
func GetUserID(c *gin.Context) int {
	id, _ := c.Get("user_id")
	if v, ok := id.(int); ok {
		return v
	}
	return 0
}

// запрет tech-admin на определённые действия
func TechAdminRestricted() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsRole(c, "tech-admin") {
			c.AbortWithStatusJSON(403, gin.H{"error": "tech admins cannot perform this action"})
			return
		}
		c.Next()
	}
}
