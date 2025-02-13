package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			URI:      "mongodb://admin:test_password@localhost:27018",
			Database: "manager_test_db",
		},
	}

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

	db := mongoClient.Database(testConfig.MongoDB.Database)
	
	// Test creating a collection
	err := db.CreateCollection(ctx, "test_collection")
	assert.NoError(t, err, "Should create collection without error")

	// Verify collection exists
	collections, err := db.ListCollectionNames(ctx, nil)
	assert.NoError(t, err, "Should list collections without error")
	assert.Contains(t, collections, "test_collection", "Should contain the created collection")

	// Cleanup
	err = db.Collection("test_collection").Drop(ctx)
	assert.NoError(t, err, "Should drop collection without error")
}
