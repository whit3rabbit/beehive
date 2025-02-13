package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/whit3rabbit/beehive/manager/internal/logger"
	"github.com/labstack/echo/v4"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/whit3rabbit/beehive/manager/models"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	"go.uber.org/zap"
)

const (
	DefaultContextTimeout = 10 * time.Second
	LongContextTimeout   = 30 * time.Second
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type HeartbeatRequest struct {
    UUID      string    `json:"uuid" validate:"required"`
}

type HeartbeatResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// hashAPIKey hashes the provided API key using SHA256.
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// RegisterAgent handles POST /agent/register.
// @Summary Registers a new agent or updates an existing one
// @Description Registers or updates an agent in the database.
// @Tags agent
// @Accept json
// @Produce json
// @Param agent body models.Agent true "Agent registration info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /agent/register [post]
func RegisterAgent(c echo.Context) error {
	var agent models.Agent
	if err := c.Bind(&agent); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
	}

	// Set created_at if not provided
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = time.Now()
	}

	// Set initial status to "active"
	agent.Status = "active"
	agent.LastSeen = time.Now()

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

	// Update response to use AgentRegistrationResponse
	response := models.AgentRegistrationResponse{
		APIKey:    apiKey,     // unhashed
		Status:    "registered",
		Timestamp: time.Now(),
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

// GetAgentSummary retrieves a summary of a specific agent by UUID.
// @Summary Retrieves a summary of a specific agent
// @Description Gets the summary details of an agent based on its UUID.
// @Tags agent
// @Accept json
// @Produce json
// @Param uuid path string true "Agent UUID"
// @Success 200 {object} models.AgentSummary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /agent/{uuid}/summary [get]
func GetAgentSummary(c echo.Context) error {
	agentUUID := c.Param("uuid")
	if agentUUID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Missing agent UUID"})
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("agents")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var agent models.Agent
	err := collection.FindOne(ctx, bson.M{"uuid": agentUUID}).Decode(&agent)
	if err != nil {
		logger.Error("Agent not found", zap.Error(err), zap.String("agent_uuid", agentUUID))
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Agent not found"})
	}

	summary := agent.ToSummary()
	return c.JSON(http.StatusOK, summary)
}

// AgentHeartbeat handles POST /agent/heartbeat.
// @Summary Updates the heartbeat of an agent
// @Description Updates the last_seen timestamp of an agent.
// @Tags agent
// @Accept json
// @Produce json
// @Param body body struct{UUID string `json:"uuid"`} true "Agent UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /agent/heartbeat [post]
func AgentHeartbeat(c echo.Context) error {
	var req struct {
		UUID string `json:"uuid"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("agents")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":    "active",
			"last_seen": time.Now(),
		},
	}
	_, err := collection.UpdateOne(ctx, bson.M{"uuid": req.UUID}, update)
	if err != nil {
		logger.Error("Failed to update agent heartbeat", zap.Error(err), zap.String("agent_uuid", req.UUID))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update agent heartbeat"})
	}

	response := HeartbeatResponse{
		Status:    "heartbeat_received",
		Timestamp: time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}

// ListAgentTasks handles GET /agent/:agent_id/tasks.
// @Summary Lists tasks for a specific agent
// @Description Retrieves all tasks associated with a given agent ID.
// @Tags agent
// @Accept json
// @Produce json
// @Param agent_id path string true "Agent ID"
// @Success 200 {array} models.Task
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /agent/{agent_id}/tasks [get]
func ListAgentTasks(c echo.Context) error {
	agentID := c.Param("agent_id")
	if agentID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Missing agent ID"})
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"agent_id": agentID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve tasks"})
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to parse tasks"})
	}
	return c.JSON(http.StatusOK, tasks)
}
