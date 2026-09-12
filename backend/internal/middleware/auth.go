package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/inventory/backend/internal/models"
)

var jwtSecretKey = []byte(getSecretKey())

func getSecretKey() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "inventory-management-super-secure-jwt-key-2026"
	}
	return secret
}

type Claims struct {
	UserID   uint        `json:"user_id"`
	Username string      `json:"username"`
	FullName string      `json:"full_name"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}

// RequireAuth checks for valid JWT
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be Bearer <token>"})
			c.Abort()
			return
		}

		claims, err := ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session. Please login again."})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("full_name", claims.FullName)
		c.Next()
	}
}

// RequireAdmin verifies user has admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized context"})
			c.Abort()
			return
		}

		role, ok := roleVal.(models.Role)
		if !ok || role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required for this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}
