package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"manager/admin"
)

// AdminAuthMiddleware checks for a valid JWT token in the "Authorization" header.
// It expects the header in the format: "Bearer <token>".
func AdminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing or invalid Authorization header"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := admin.ValidateToken(tokenStr)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
		}
		// Optionally, store admin info in the context for downstream handlers.
		c.Set("admin", claims.Username)
		return next(c)
	}
}

// APIAuthMiddleware validates the X-API-Key and X-Signature headers.
// It checks that the API key matches an expected value and that the signature,
// computed as an HMAC-SHA256 of the request body using a secret key, is valid.
func APIAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiKey := c.Request().Header.Get("X-API-Key")
		signature := c.Request().Header.Get("X-Signature")

		if apiKey == "" || signature == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing API key or signature"})
		}

		validAPIKey := os.Getenv("API_KEY")
		apiSecret := []byte(os.Getenv("API_SECRET"))

		// Verify the API key.
		if apiKey != validAPIKey {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid API key"})
		}

		// Read the request body to compute the HMAC signature.
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Unable to read request body"})
		}
		// Restore the request body for downstream handlers.
		c.Request().Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		mac := hmac.New(sha256.New, apiSecret)
		mac.Write(bodyBytes)
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid signature"})
		}

		return next(c)
	}
}
