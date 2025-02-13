package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/whit3rabbit/beehive/manager/api/admin"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	"github.com/whit3rabbit/beehive/manager/models"
)

// RateLimiter defines the interface for rate limiting functionality.
type RateLimiter interface {
	CheckLimit(key string) (bool, time.Duration)
}

// Validatable defines the interface for request validation.
type Validatable interface {
	Validate() error
}

var validate = validator.New()

// RequestValidationMiddleware validates the request body against the struct tags.
func RequestValidationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if validatable, ok := c.Get("body").(Validatable); ok {
			if err := validate.Struct(validatable); err != nil {
				return c.JSON(http.StatusBadRequest, echo.Map{
					"error":   "Validation failed",
					"details": err.Error(),
				})
			}
		}
		return next(c)
	}
}

// RefreshToken handles the token refresh endpoint.
// It validates the current token and issues a new one with extended expiration.
func RefreshToken(c echo.Context) error {
	// Get the token from the request header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error": "Missing or invalid Authorization header",
		})
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate the current token
	jwtSecret := c.Get("jwt_secret").(string)
	claims, err := admin.ValidateToken(tokenStr, jwtSecret)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error": "Invalid token",
		})
	}

	// Generate a new token with an extended expiration
	newToken, err := admin.GenerateToken(claims.Username, jwtSecret, c.Get("token_expiration_hours").(int))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Could not generate token",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token":    newToken,
		"username": claims.Username,
	})
}

// AdminAuthMiddleware checks for a valid JWT token in the "Authorization" header and applies rate limiting.
// It expects the header in the format: "Bearer <token>".
func AdminAuthMiddleware(rateLimiter RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Missing or invalid Authorization header",
				})
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			jwtSecret := c.Get("jwt_secret").(string)
			claims, err := admin.ValidateToken(tokenStr, jwtSecret)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Invalid token",
				})
			}

			// Apply rate limiting
			username := claims.Username
			allowed, waitDuration := rateLimiter.CheckLimit(username)
			if !allowed {
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"error":       "Too many requests",
					"retry_after": waitDuration.Seconds(),
				})
			}

			// Store admin info in the context for downstream handlers
			c.Set("admin", username)
			return next(c)
		}
	}
}

// APIAuthMiddleware validates the X-API-Key and X-Signature headers.
// It checks that the API key exists in the database and that the signature,
// computed as an HMAC-SHA256 of the request body using the API key as secret, is valid.
func APIAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiKey := c.Request().Header.Get("X-API-Key")
		signature := c.Request().Header.Get("X-Signature")

		if apiKey == "" || signature == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Missing API key or signature",
			})
		}

		// Get MongoDB database name from context
		dbName := c.Get("mongodb_database").(string)
		collection := mongodb.Client.Database(dbName).Collection("agents")

		// Find agent by API key
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var agent models.Agent
		err := collection.FindOne(ctx, bson.M{"api_key": apiKey}).Decode(&agent)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Invalid API key",
			})
		}

		// Store agent info in context for downstream handlers
		c.Set("agent_id", agent.ID.Hex())
		c.Set("agent_uuid", agent.UUID)

		// Read and validate request body
		var body struct{}
		if err := c.Bind(&body); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "Invalid request body",
			})
		}

		// Get the request body as a byte slice
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Unable to read request body",
			})
		}
		defer c.Request().Body.Close()

		// Restore the request body for downstream handlers
		c.Request().Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		// Compute and verify HMAC signature
		mac := hmac.New(sha256.New, []byte(agent.APISecret))
		mac.Write(bodyBytes)
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Invalid signature",
			})
		}

		return next(c)
	}
}