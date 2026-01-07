package utils

import (
	"errors"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ================= JWT SECRETS =================
var accessSecret = []byte(os.Getenv("JWT_SECRET"))
var refreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))

// ================= CLAIMS STRUCT =================
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ================= GENERATE TOKENS =================
func GenerateTokens(userID, role string) (string, string, error) {
	accessClaims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	refreshClaims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString(accessSecret)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString(refreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ================= VALIDATE TOKEN =================
func ValidateToken(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ================= GET ENV VARS =================
func GetEnv(key string) string {
	return os.Getenv(key)
}

// ================= CONTEXT HELPERS =================

// GetUserIdFromContext returns the user ID stored in Gin context
func GetUserIdFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("user_id not found in context")
	}

	uid, ok := userID.(string)
	if !ok {
		return "", errors.New("user_id in context has invalid type")
	}

	return uid, nil
}

// GetRoleFromContext returns the role stored in Gin context
func GetRoleFromContext(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", errors.New("role not found in context")
	}

	r, ok := role.(string)
	if !ok {
		return "", errors.New("role in context has invalid type")
	}

	return r, nil
}
