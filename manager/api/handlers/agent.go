package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/whit3rabbit/beehive/manager/models"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
)

// hashAPIKey hashes the provided API key using SHA256.
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// RegisterAgent handles POST /agent/register.
// It registers or updates an agent in the DB.
func RegisterAgent(c echo.Context) error {
	var agent models.Agent
	if err := c.Bind(&agent); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	// Set created_at if not provided
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = time.Now()
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("agents")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"uuid": agent.UUID}
	update := bson.M{"$set": agent}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Error("Failed to register agent", zap.Error(err), zap.String("agent_uuid", agent.UUID))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to register agent"})
	}

	// Generate and store API key
	apiKey, err := generateSecureToken()
	if err != nil {
		logger.Error("Failed to generate API key", zap.Error(err), zap.String("agent_uuid", agent.UUID))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to generate API key"})
	}

	// Generate and store API secret
	apiSecret, err := generateSecureToken()
	if err != nil {
		logger.Error("Failed to generate API secret", zap.Error(err), zap.String("agent_uuid", agent.UUID))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to generate API secret"})
	}

	// Hash API key and secret
	hashedAPIKey := hashAPIKey(apiKey)
	hashedAPISecret := hashAPIKey(apiSecret)

	// Update agent with hashed API key and secret
	agent.APIKey = hashedAPIKey
	agent.APISecret = hashedAPISecret

	// Update document with hashed API key and secret
	update = bson.M{"$set": agent}
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Error("Failed to store API key and secret", zap.Error(err), zap.String("agent_uuid", agent.UUID))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to store API key and secret"})
	}

	response := echo.Map{
		"api_key":     apiKey, // Return the unhashed API key to the agent
		"api_secret":  apiSecret, // Return the unhashed API secret to the agent
		"status":      "registered",
		"timestamp":   time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}

// generateSecureToken generates a secure random token for API key and secret.
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		logger.Error("Failed to generate secure token", zap.Error(err))
		return "", err // Handle error appropriately
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// AgentHeartbeat handles POST /agent/heartbeat.
func AgentHeartbeat(c echo.Context) error {
	var req struct {
		UUID      string    `json:"uuid"`
		Timestamp time.Time `json:"timestamp"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	response := echo.Map{
		"status":   "heartbeat_received",
		"timestamp": time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}

// ListAgentTasks handles GET /agent/:agent_id/tasks.
func ListAgentTasks(c echo.Context) error {
	agentID := c.Param("agent_id")
	if agentID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing agent ID"})
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"agent_id": agentID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve tasks"})
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse tasks"})
	}
	return c.JSON(http.StatusOK, tasks)
}
