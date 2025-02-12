package admin

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Claims represents the JWT claims structure
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// LoginHandler processes admin login requests
func LoginHandler(username, password string) (string, error) {
	// Instead of separate error messages for username and password,
	// use a generic error message for both.
	if username != os.Getenv("ADMIN_DEFAULT_USERNAME") {
		return "", fmt.Errorf("invalid credentials")
	}

	// Compare password hash
	if err := bcrypt.CompareHashAndPassword(
		[]byte(os.Getenv("ADMIN_DEFAULT_PASSWORD")),
		[]byte(password),
	); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// Create token
	expirationHours := getEnvAsInt("TOKEN_EXPIRATION_HOURS", 24)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// ValidateToken validates the JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// Helper function to get environment variables as integers
func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}
