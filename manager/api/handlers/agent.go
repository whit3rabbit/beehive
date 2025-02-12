package handlers

import (
	"context"
	"net/http"
	"time"

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

	collection := mongodb.Client.Database(os.Getenv("MONGODB_DATABASE")).Collection("agents")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"uuid": agent.UUID}
	update := bson.M{"$set": agent}
	opts := options.Update().SetUpsert(true)
	if _, err := collection.UpdateOne(ctx, filter, update, opts); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to register agent"})
	}

	// Normally you'd generate an API key; here we return a dummy key.
	response := echo.Map{
		"api_key":  "dummy_api_key",
		"status":   "registered",
		"timestamp": time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}

// AgentHeartbeat handles POST /agent/heartbeat.
// It receives heartbeat signals from agents.
func AgentHeartbeat(c echo.Context) error {
	var req struct {
		UUID      string    `json:"uuid"`
		Timestamp time.Time `json:"timestamp"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	// Optionally update the agent's last seen timestamp in the DB.

	response := echo.Map{
		"status":   "heartbeat_received",
		"timestamp": time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}

// ListAgentTasks handles GET /agent/:agent_id/tasks.
// It returns all tasks assigned to a specific agent.
func ListAgentTasks(c echo.Context) error {
	agentID := c.Param("agent_id")
	if agentID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing agent ID"})
	}

	collection := mongodb.Client.Database("manager_db").Collection("tasks")
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
