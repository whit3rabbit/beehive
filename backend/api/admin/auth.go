package admin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
	"unicode"

	"github.com/whit3rabbit/beehive/manager/internal/logger"
	"go.uber.org/zap"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"

	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	"github.com/whit3rabbit/beehive/manager/models"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var loginMutex sync.RWMutex
var loginAttempts = make(map[string]struct {
	count       int
	lastAttempt time.Time
})

// validatePassword checks if the password meets the given policy.
func validatePassword(password string, policy models.PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return fmt.Errorf("password must be at least %d characters", policy.MinLength)
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if policy.RequireLowercase && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if policy.RequireNumbers && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if policy.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// GenerateHashPassword hashes the provided plaintext password.
func GenerateHashPassword(password string, policy models.PasswordPolicy) (string, error) {
	if err := validatePassword(password, policy); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a hashed password with its plaintext version.
func VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		logger.Warn("Invalid username or password", zap.Error(err))
		return err
	}
	return nil
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

	// Get password policy from context
	passwordPolicy, ok := c.Get("password_policy").(models.PasswordPolicy)
	if !ok {
		logger.Error("Password policy not properly configured")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Internal server error"})
	}

	// Validate password
	if err := validatePassword(req.Password, passwordPolicy); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	username := req.Username

	loginMutex.Lock()
	if attempts, exists := loginAttempts[username]; exists {
		if attempts.count >= 5 && time.Since(attempts.lastAttempt) < 15*time.Minute {
			loginMutex.Unlock()
			return c.JSON(http.StatusTooManyRequests, echo.Map{"error": "Too many login attempts. Please wait 15 minutes."})
		}
	}
	loginMutex.Unlock()

	// Retrieve the admin record from MongoDB.
	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("admins")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var admin models.Admin
	if err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&admin); err != nil {
		logger.Warn("Invalid username or password", zap.Error(err), zap.String("username", username))
		updateLoginAttempts(username, false)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	// Verify the provided password.
	if err := VerifyPassword(admin.Password, req.Password); err != nil {
		updateLoginAttempts(username, false)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	// Retrieve jwt_secret from context
	jwtSecret, ok := c.Get("jwt_secret").(string)
	if !ok || jwtSecret == "" {
		logger.Error("JWT secret not properly configured")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Internal server error"})
	}

	tokenExpiration, ok := c.Get("token_expiration_hours").(int)
	if !ok || tokenExpiration <= 0 {
		logger.Error("Token expiration not properly configured")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Internal server error"})
	}

	// Generate a JWT token with configured expiration
	token, err := GenerateToken(admin.Username, jwtSecret, tokenExpiration)
	if err != nil {
		logger.Error("Could not generate token", zap.Error(err), zap.String("username", admin.Username))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	updateLoginAttempts(username, true)

	return c.JSON(http.StatusOK, echo.Map{
		"token":    token,
		"username": admin.Username,
	})
}

// CleanupLoginAttempts periodically cleans up expired login attempts.
func CleanupLoginAttempts(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("Stopping login attempts cleanup routine")
				return
			case <-ticker.C:
				loginMutex.Lock()
				now := time.Now()
				cleanupCount := 0

				for username, attempt := range loginAttempts {
					if now.Sub(attempt.lastAttempt) > 15*time.Minute {
						delete(loginAttempts, username)
						cleanupCount++
					}
				}

				loginMutex.Unlock()

				if cleanupCount > 0 {
					logger.Info("Cleaned up login attempts",
						zap.Int("count", cleanupCount))
				}
			}
		}
	}()
}

// updateLoginAttempts updates the login attempts count for a given username.
func updateLoginAttempts(username string, success bool) {
	loginMutex.Lock()
	defer loginMutex.Unlock()

	if success {
		delete(loginAttempts, username)
		return
	}

	attempts, exists := loginAttempts[username]
	if !exists {
		loginAttempts[username] = struct {
			count       int
			lastAttempt time.Time
		}{count: 1, lastAttempt: time.Now()}
	} else {
		attempts.count++
		attempts.lastAttempt = time.Now()
		loginAttempts[username] = attempts
	}
}