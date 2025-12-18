package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"

	"stock-reward-api/db"
	"stock-reward-api/logger"
	"stock-reward-api/models"

)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer"))
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			logger.Log.Warn("JWT_SECRET not set; using empty secret")
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"]) }
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			logger.Log.Errorf("failed to parse/validate token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		sub, ok := claims["sub"]
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing subject"})
			return
		}

		var userID int64
		switch v := sub.(type) {
		case float64:
			userID = int64(v)
		case int64:
			userID = v
		case string:
			var parsed int64
			_, err := fmt.Sscanf(v, "%d", &parsed)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token subject"})
				return
			}
			userID = parsed
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token subject type"})
			return
		}

		var user models.User
	err = db.Pool.QueryRow(c.Request.Context(), "SELECT id, name, email, created_at FROM users WHERE id=$1", userID).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)

		if err != nil {
			logger.Log.Errorf("failed to load user %d: %v", userID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*models.User, bool) {
	u, ok := c.Get("user")
	if !ok {
		return nil, false
	}
	usr, ok := u.(*models.User)
	return usr, ok
}
