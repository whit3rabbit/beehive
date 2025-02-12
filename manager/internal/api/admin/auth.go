package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"

	"manager/internal/mongodb"
	"manager/models"
)

// jwtKey is used for signing tokens. Replace it with a secure key from your configuration.
var jwtKey = []byte("your_secret_key")

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

// GenerateToken creates a JWT token for the given username with the specified duration.
func GenerateToken(username string, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateToken parses and validates the JWT token string.
func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
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
	collection := mongodb.Client.Database("manager_db").Collection("admins")
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

	// Generate a JWT token (e.g., valid for 24 hours).
	token, err := GenerateToken(admin.Username, 24*time.Hour)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token":    token,
		"username": admin.Username,
	})
}
