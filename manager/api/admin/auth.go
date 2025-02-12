package admin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"

	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	"github.com/whit3rabbit/beehive/manager/models"
// Claims represents the JWT claims for an admin user.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateHashPassword hashes the provided plaintext password.
func GenerateHashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a hashed password with its plaintext version.
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateToken creates a JWT token for the given username.
func GenerateToken(username string, jwtSecret string, tokenExpirationHours int) (string, error) {
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET not configured")
	}

	expirationTime := time.Now().Add(time.Hour * time.Duration(tokenExpirationHours))
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ValidateToken parses and validates the JWT token string.
func ValidateToken(tokenStr string, jwtSecret string) (*Claims, error) {
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET not configured")
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, echo.ErrUnauthorized
	}
	return claims, nil
}

// LoginHandler handles POST /admin/login.
// It verifies admin credentials and returns a JWT token on success.
func LoginHandler(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	// Retrieve the admin record from MongoDB.
	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("admins")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var admin models.Admin
	if err := collection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&admin); err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	// Verify the provided password.
	if err := VerifyPassword(admin.Password, req.Password); err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	// Generate a JWT token with configured expiration
	token, err := GenerateToken(admin.Username, c.Get("jwt_secret").(string), c.Get("token_expiration_hours").(int))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token":    token,
		"username": admin.Username,
	})
}
