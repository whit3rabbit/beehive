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

	"github.com/whit3rabbit/beehive/manager/internal/config"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
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
