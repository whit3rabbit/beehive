package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
	"os"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/whit3rabbit/beehive/manager/models"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
)

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

	agent.ID = primitive.NewObjectID()

	filter := bson.M{"uuid": agent.UUID}
	update := bson.M{"$set": agent}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to register agent"})
	}

	apiKey := generateSecureToken() // Implement secure token generation

	response := echo.Map{
		"api_key":   apiKey,
		"status":    "registered",
		"timestamp": time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}

// generateSecureToken generates a secure random token for API key.
func generateSecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "" // Handle error appropriately
	}
	return base64.StdEncoding.EncodeToString(b)
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
