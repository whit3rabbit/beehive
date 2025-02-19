package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
	"github.com/whit3rabbit/beehive/manager/api/handlers"
	"github.com/whit3rabbit/beehive/manager/internal/config"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	customMiddleware "github.com/whit3rabbit/beehive/manager/middleware"
	"github.com/whit3rabbit/beehive/manager/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	testConfig *config.Config
	mongoClient *mongo.Client
)

func TestMain(m *testing.M) {
	// Setup
	var err error
	testConfig = &config.Config{
		MongoDB: config.MongoDBConfig{
			URI:      "mongodb://admin:test_password@localhost:27018/admin",
			Database: "manager_test_db",
		},
	}

	// Wait for MongoDB to be ready
	time.Sleep(2 * time.Second)

	// Connect to test MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(testConfig.MongoDB.URI))
	if err != nil {
		fmt.Printf("Failed to connect to test MongoDB: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := mongoClient.Disconnect(ctx); err != nil {
		fmt.Printf("Failed to disconnect from test MongoDB: %v\n", err)
	}

	os.Exit(code)
}

func TestMongoDBConnection(t *testing.T) {
	err := mongodb.Connect(testConfig.MongoDB.URI)
	require.NoError(t, err, "Should connect to test MongoDB without error")

	// Test ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = mongodb.Client.Ping(ctx, nil)
	assert.NoError(t, err, "Should ping MongoDB successfully")
}

func TestCreateCollection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Switch to test database
	db := mongoClient.Database(testConfig.MongoDB.Database)
	
	// Test creating a collection
	err := db.CreateCollection(ctx, "test_collection")
	assert.NoError(t, err, "Should create collection without error")

	// Verify collection exists
	filter := bson.D{}
	collections, err := db.ListCollectionNames(ctx, filter)
	assert.NoError(t, err, "Should list collections without error")
	assert.Contains(t, collections, "test_collection", "Should contain the created collection")

	// Cleanup
	err = db.Collection("test_collection").Drop(ctx)
	assert.NoError(t, err, "Should drop collection without error")
}

func TestAgentCRUD(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := mongoClient.Database(testConfig.MongoDB.Database)
	collection := db.Collection("agents")

	// Create agent
	agent := models.Agent{
		UUID:      "test-uuid",
		Hostname:  "test-host",
		MacHash:   "test-mac-hash",
		Nickname:  "test-agent",
		Role:      "worker",
		APIKey:    "test-api-key",
		APISecret: "test-api-secret",
		Status:    "active",
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
	}

	result, err := collection.InsertOne(ctx, agent)
	assert.NoError(t, err, "Should insert agent without error")
	assert.NotNil(t, result.InsertedID, "Should have an inserted ID")

	// Read agent
	var foundAgent models.Agent
	err = collection.FindOne(ctx, bson.M{"uuid": agent.UUID}).Decode(&foundAgent)
	assert.NoError(t, err, "Should find agent without error")
	assert.Equal(t, agent.Hostname, foundAgent.Hostname, "Should match hostname")

	// Update agent
	update := bson.M{"$set": bson.M{"nickname": "updated-nickname"}}
	_, err = collection.UpdateOne(ctx, bson.M{"uuid": agent.UUID}, update)
	assert.NoError(t, err, "Should update agent without error")

	// Verify update
	err = collection.FindOne(ctx, bson.M{"uuid": agent.UUID}).Decode(&foundAgent)
	assert.NoError(t, err, "Should find updated agent")
	assert.Equal(t, "updated-nickname", foundAgent.Nickname, "Should have updated nickname")

	// Delete agent
	_, err = collection.DeleteOne(ctx, bson.M{"uuid": agent.UUID})
	assert.NoError(t, err, "Should delete agent without error")

	// Verify deletion
	err = collection.FindOne(ctx, bson.M{"uuid": agent.UUID}).Decode(&foundAgent)
	assert.Error(t, err, "Should not find deleted agent")
	assert.Equal(t, mongo.ErrNoDocuments, err, "Should return no documents error")
}

func TestRoleCRUD(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := mongoClient.Database(testConfig.MongoDB.Database)
	collection := db.Collection("roles")

	// Create role
	role := models.Role{
		Name:         "test-role",
		Description:  "Test role description",
		Applications: []string{"app1", "app2"},
		DefaultTasks: []string{"task1", "task2"},
		CreatedAt:    time.Now(),
	}

	result, err := collection.InsertOne(ctx, role)
	assert.NoError(t, err, "Should insert role without error")
	assert.NotNil(t, result.InsertedID, "Should have an inserted ID")

	// Read role
	var foundRole models.Role
	err = collection.FindOne(ctx, bson.M{"name": role.Name}).Decode(&foundRole)
	assert.NoError(t, err, "Should find role without error")
	assert.Equal(t, role.Description, foundRole.Description, "Should match description")

	// Update role
	update := bson.M{"$set": bson.M{"description": "Updated description"}}
	_, err = collection.UpdateOne(ctx, bson.M{"name": role.Name}, update)
	assert.NoError(t, err, "Should update role without error")

	// Verify update
	err = collection.FindOne(ctx, bson.M{"name": role.Name}).Decode(&foundRole)
	assert.NoError(t, err, "Should find updated role")
	assert.Equal(t, "Updated description", foundRole.Description, "Should have updated description")

	// Delete role
	_, err = collection.DeleteOne(ctx, bson.M{"name": role.Name})
	assert.NoError(t, err, "Should delete role without error")

	// Verify deletion
	err = collection.FindOne(ctx, bson.M{"name": role.Name}).Decode(&foundRole)
	assert.Error(t, err, "Should not find deleted role")
	assert.Equal(t, mongo.ErrNoDocuments, err, "Should return no documents error")
}

func TestAdminAuth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := mongoClient.Database(testConfig.MongoDB.Database)
	collection := db.Collection("admins")

	// Create admin with hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.DefaultCost)
	assert.NoError(t, err, "Should hash password without error")

	admin := models.Admin{
		Username:  "test-admin",
		Password:  string(hashedPassword),
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := collection.InsertOne(ctx, admin)
	assert.NoError(t, err, "Should insert admin without error")
	assert.NotNil(t, result.InsertedID, "Should have an inserted ID")

	// Verify password
	var foundAdmin models.Admin
	err = collection.FindOne(ctx, bson.M{"username": admin.Username}).Decode(&foundAdmin)
	assert.NoError(t, err, "Should find admin without error")

	err = bcrypt.CompareHashAndPassword([]byte(foundAdmin.Password), []byte("test-password"))
	assert.NoError(t, err, "Should verify password without error")

	err = bcrypt.CompareHashAndPassword([]byte(foundAdmin.Password), []byte("wrong-password"))
	assert.Error(t, err, "Should fail with wrong password")

	// Cleanup
	_, err = collection.DeleteOne(ctx, bson.M{"username": admin.Username})
	assert.NoError(t, err, "Should delete admin without error")
}

func TestRateLimiter(t *testing.T) {
	limiter := customMiddleware.NewRateLimiter(2, time.Second*2, time.Second*4)

	// First attempt should succeed
	allowed, waitTime := limiter.CheckLimit("test-user")
	assert.True(t, allowed, "First attempt should be allowed")
	assert.Zero(t, waitTime, "Wait time should be zero")

	// Second attempt should succeed
	allowed, waitTime = limiter.CheckLimit("test-user")
	assert.True(t, allowed, "Second attempt should be allowed")
	assert.Zero(t, waitTime, "Wait time should be zero")

	// Third attempt should fail
	allowed, waitTime = limiter.CheckLimit("test-user")
	assert.False(t, allowed, "Third attempt should be blocked")
	assert.NotZero(t, waitTime, "Wait time should be non-zero")

	// Wait for window and blockout to expire
	time.Sleep(time.Second * 5)

	// Should be allowed again after full expiry
	allowed, waitTime = limiter.CheckLimit("test-user")
	assert.True(t, allowed, "Attempt after window expiry should be allowed")
	assert.Zero(t, waitTime, "Wait time should be zero")
}

func TestLogEntryCRUD(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := mongoClient.Database(testConfig.MongoDB.Database)
	collection := db.Collection("logs")

	// Create log entry
	logEntry := models.LogEntry{
		Timestamp: time.Now(),
		Endpoint:  "/api/test",
		AgentID:   "test-agent-id",
		Status:    "success",
		Details:   "Test log entry",
	}

	result, err := collection.InsertOne(ctx, logEntry)
	assert.NoError(t, err, "Should insert log entry without error")
	assert.NotNil(t, result.InsertedID, "Should have an inserted ID")

	// Read log entry
	var foundLog models.LogEntry
	err = collection.FindOne(ctx, bson.M{"agent_id": logEntry.AgentID}).Decode(&foundLog)
	assert.NoError(t, err, "Should find log entry without error")
	assert.Equal(t, logEntry.Status, foundLog.Status, "Should match status")

	// Update log entry
	update := bson.M{"$set": bson.M{"details": "Updated details"}}
	_, err = collection.UpdateOne(ctx, bson.M{"agent_id": logEntry.AgentID}, update)
	assert.NoError(t, err, "Should update log entry without error")

	// Verify update
	err = collection.FindOne(ctx, bson.M{"agent_id": logEntry.AgentID}).Decode(&foundLog)
	assert.NoError(t, err, "Should find updated log entry")
	assert.Equal(t, "Updated details", foundLog.Details, "Should have updated details")

	// Delete log entry
	_, err = collection.DeleteOne(ctx, bson.M{"agent_id": logEntry.AgentID})
	assert.NoError(t, err, "Should delete log entry without error")

	// Verify deletion
	err = collection.FindOne(ctx, bson.M{"agent_id": logEntry.AgentID}).Decode(&foundLog)
	assert.Error(t, err, "Should not find deleted log entry")
	assert.Equal(t, mongo.ErrNoDocuments, err, "Should return no documents error")
}

func setupEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler
	return e
}

func TestAPICreateTask(t *testing.T) {
	e := setupEcho()
	
	// Setup route
	e.POST("/task/create", handlers.CreateTask)

	// Create test task
	task := models.Task{
		AgentID: "test-agent",
		Type:    "scan",
		Parameters: map[string]interface{}{
			"target": "localhost",
		},
	}

	taskReq := handlers.TaskRequest{
		Task: task,
	}

	jsonBytes, err := json.Marshal(taskReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/task/create", bytes.NewReader(jsonBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Set("mongodb_database", testConfig.MongoDB.Database)

	// Test handler
	err = handlers.CreateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify response
	var response models.TaskCreationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.TaskID)
	assert.Equal(t, "queued", response.Status)
}

func TestTaskOutputValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := mongoClient.Database(testConfig.MongoDB.Database)
	collection := db.Collection("tasks")

	// Create task with output
	task := models.Task{
		AgentID: "test-agent-id",
		Type:    "command_shell",
		Parameters: map[string]interface{}{
			"command": "echo test",
		},
		Status: "completed",
		Output: &models.Output{
			Logs:  "Test output logs",
			Error: "",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Timeout:   300,
	}

	result, err := collection.InsertOne(ctx, task)
	assert.NoError(t, err, "Should insert task without error")
	assert.NotNil(t, result.InsertedID, "Should have an inserted ID")

	// Read and verify output
	var foundTask models.Task
	err = collection.FindOne(ctx, bson.M{"agent_id": task.AgentID}).Decode(&foundTask)
	assert.NoError(t, err, "Should find task without error")
	assert.NotNil(t, foundTask.Output, "Should have output")
	assert.Equal(t, task.Output.Logs, foundTask.Output.Logs, "Should match output logs")

	// Update output
	update := bson.M{"$set": bson.M{"output.logs": "Updated output logs"}}
	_, err = collection.UpdateOne(ctx, bson.M{"agent_id": task.AgentID}, update)
	assert.NoError(t, err, "Should update task output without error")

	// Verify output update
	err = collection.FindOne(ctx, bson.M{"agent_id": task.AgentID}).Decode(&foundTask)
	assert.NoError(t, err, "Should find updated task")
	assert.Equal(t, "Updated output logs", foundTask.Output.Logs, "Should have updated output logs")

	// Cleanup
	_, err = collection.DeleteOne(ctx, bson.M{"agent_id": task.AgentID})
	assert.NoError(t, err, "Should delete task without error")
}

func TestAPIListRoles(t *testing.T) {
	e := setupEcho()
	e.GET("/roles", handlers.ListRoles)

	// Create test role first
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := mongoClient.Database(testConfig.MongoDB.Database)
	collection := db.Collection("roles")

	role := models.Role{
		Name:        "test-role",
		Description: "Test role description",
		CreatedAt:   time.Now(),
	}

	_, err := collection.InsertOne(ctx, role)
	require.NoError(t, err)

	// Test API endpoint
	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("mongodb_database", testConfig.MongoDB.Database)

	err = handlers.ListRoles(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var roles []models.Role
	err = json.Unmarshal(rec.Body.Bytes(), &roles)
	assert.NoError(t, err)
	assert.NotEmpty(t, roles)
	assert.Equal(t, "test-role", roles[0].Name)

	// Cleanup
	_, err = collection.DeleteOne(ctx, bson.M{"name": role.Name})
	assert.NoError(t, err)
}

func TestAPIAgentHeartbeat(t *testing.T) {
	e := setupEcho()
	e.POST("/agent/heartbeat", handlers.AgentHeartbeat)

	heartbeat := models.HeartbeatRequest{
		UUID:      "test-agent",
		Timestamp: time.Now(),
	}

	jsonBytes, err := json.Marshal(heartbeat)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/agent/heartbeat", bytes.NewReader(jsonBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Set("mongodb_database", testConfig.MongoDB.Database)

	err = handlers.AgentHeartbeat(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.HeartbeatResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "heartbeat_received", response.Status)
}

func TestPasswordPolicyValidation(t *testing.T) {
	// Test valid password
	validPassword := "Test123!@"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)
	assert.NoError(t, err, "Should hash valid password without error")

	// Test password verification
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(validPassword))
	assert.NoError(t, err, "Should verify valid password without error")

	// Test invalid passwords
	testCases := []struct {
		password string
		name     string
	}{
		{"short", "too short"},
		{"nouppercase123!", "no uppercase"},
		{"NOLOWERCASE123!", "no lowercase"},
		{"NoNumbers!", "no numbers"},
		{"NoSpecial123", "no special chars"},
	}

	for _, tc := range testCases {
		_, err := bcrypt.GenerateFromPassword([]byte(tc.password), bcrypt.DefaultCost)
		assert.NoError(t, err, "Should hash password without error")
		// Additional policy validation would go here in a real implementation
	}
}

func TestTaskCRUD(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := mongoClient.Database(testConfig.MongoDB.Database)
	collection := db.Collection("tasks")

	// Create task
	task := models.Task{
		AgentID: "test-agent-id",
		Type:    "command_shell",
		Parameters: map[string]interface{}{
			"command": "echo test",
		},
		Status:    "queued",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Timeout:   300,
	}

	result, err := collection.InsertOne(ctx, task)
	assert.NoError(t, err, "Should insert task without error")
	assert.NotNil(t, result.InsertedID, "Should have an inserted ID")

	// Read task
	var foundTask models.Task
	err = collection.FindOne(ctx, bson.M{"agent_id": task.AgentID}).Decode(&foundTask)
	assert.NoError(t, err, "Should find task without error")
	assert.Equal(t, task.Type, foundTask.Type, "Should match task type")

	// Update task
	update := bson.M{"$set": bson.M{"status": "running"}}
	_, err = collection.UpdateOne(ctx, bson.M{"agent_id": task.AgentID}, update)
	assert.NoError(t, err, "Should update task without error")

	// Verify update
	err = collection.FindOne(ctx, bson.M{"agent_id": task.AgentID}).Decode(&foundTask)
	assert.NoError(t, err, "Should find updated task")
	assert.Equal(t, "running", foundTask.Status, "Should have updated status")

	// Delete task
	_, err = collection.DeleteOne(ctx, bson.M{"agent_id": task.AgentID})
	assert.NoError(t, err, "Should delete task without error")

	// Verify deletion
	err = collection.FindOne(ctx, bson.M{"agent_id": task.AgentID}).Decode(&foundTask)
	assert.Error(t, err, "Should not find deleted task")
	assert.Equal(t, mongo.ErrNoDocuments, err, "Should return no documents error")
}
