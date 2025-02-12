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
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/whit3rabbit/beehive/manager/api/admin"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	"github.com/whit3rabbit/beehive/manager/models"
)

var validate = validator.New()

type Validatable interface {
	Validate() error
}

// RequestValidationMiddleware validates the request body against the struct tags.
func RequestValidationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

// AdminAuthMiddleware checks for a valid JWT token in the "Authorization" header.
// It expects the header in the format: "Bearer <token>".
func AdminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing or invalid Authorization header"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		jwtSecret := c.Get("jwt_secret").(string)
		claims, err := admin.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
		}
		// Optionally, store admin info in the context for downstream handlers.
		c.Set("admin", claims.Username)
		return next(c)
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
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing API key or signature"})
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
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid API key"})
		}

		// Store agent info in context for downstream handlers
		c.Set("agent_id", agent.ID.Hex())
		c.Set("agent_uuid", agent.UUID)

		// Define a struct to bind the request body to.  We don't actually care
		// about the contents, we just need to read the body for signature validation.
		var body struct{}
		if err := c.Bind(&body); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
		}

		// Get the request body as a byte slice.
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Unable to read request body"})
		}

		// Restore the request body for downstream handlers.
		c.Request().Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		// Use the agent's API secret to compute the HMAC signature.
		mac := hmac.New(sha256.New, []byte(agent.APISecret))
		mac.Write(bodyBytes)
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		// Compare the computed signature with the signature from the header.
		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid signature"})
		}

		return next(c)
	}
}
