package setup

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// SetupMongoDB initializes the MongoDB instance with required users and collections
func SetupMongoDB(config *Config) error {
	ctx := context.Background()

	// Connect to MongoDB without auth first to create user
	mongoURI := fmt.Sprintf("mongodb://%s:%d", config.MongoHost, config.MongoPort)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// Ping MongoDB to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Create the database user
	if err := createDatabaseUser(ctx, client, config); err != nil {
		return fmt.Errorf("failed to create database user: %w", err)
	}

	// Reconnect with the new user credentials
	authURI := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		config.MongoUser,
		config.MongoPass,
		config.MongoHost,
		config.MongoPort,
		config.MongoDatabase,
	)
	authClient, err := mongo.Connect(ctx, options.Client().ApplyURI(authURI))
	if err != nil {
		return fmt.Errorf("failed to connect with authentication: %w", err)
	}
	defer authClient.Disconnect(ctx)

	// Initialize database schema
	if err := initializeSchema(ctx, authClient, config); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// createDatabaseUser creates a new MongoDB user with appropriate permissions
func createDatabaseUser(ctx context.Context, client *mongo.Client, config *Config) error {
	cmd := bson.D{
		{Key: "createUser", Value: config.MongoUser},
		{Key: "pwd", Value: config.MongoPass},
		{Key: "roles", Value: bson.A{
			bson.D{
				{Key: "role", Value: "readWrite"},
				{Key: "db", Value: config.MongoDatabase},
			},
		}},
	}

	err := client.Database(config.MongoDatabase).RunCommand(ctx, cmd).Err()
	if err != nil {
		return fmt.Errorf("failed to create MongoDB user: %w", err)
	}

	return nil
}

// initializeSchema creates all necessary collections and indexes
func initializeSchema(ctx context.Context, client *mongo.Client, config *Config) error {
	db := client.Database(config.MongoDatabase)

	// Create collections
	collections := []string{
		"admins",
		"agents",
		"roles",
		"tasks",
		"logs",
	}

	for _, collName := range collections {
		if err := createCollection(ctx, db, collName); err != nil {
			return fmt.Errorf("failed to create collection %s: %w", collName, err)
		}
	}

	// Create indexes
	if err := createIndexes(ctx, db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Create initial admin user
	if err := createAdminUser(ctx, db, config); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create default role
	if err := createDefaultRole(ctx, db); err != nil {
		return fmt.Errorf("failed to create default role: %w", err)
	}

	return nil
}

// createCollection creates a new collection if it doesn't exist
func createCollection(ctx context.Context, db *mongo.Database, name string) error {
	err := db.CreateCollection(ctx, name)
	if err != nil {
		// Ignore error if collection already exists
		if !mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("failed to create collection %s: %w", name, err)
		}
	}
	return nil
}

// createIndexes creates all required indexes for the collections
func createIndexes(ctx context.Context, db *mongo.Database) error {
	// Admins collection indexes
	adminIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if err := createCollectionIndexes(ctx, db, "admins", adminIndexes); err != nil {
		return err
	}

	// Agents collection indexes
	agentIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "uuid", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "api_key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "last_seen", Value: 1}},
		},
	}
	if err := createCollectionIndexes(ctx, db, "agents", agentIndexes); err != nil {
		return err
	}

	// Tasks collection indexes
	taskIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "agent_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}
	if err := createCollectionIndexes(ctx, db, "tasks", taskIndexes); err != nil {
		return err
	}

	// Logs collection indexes (with TTL)
	logIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "timestamp", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(30 * 24 * 60 * 60), // 30 days TTL
		},
	}
	if err := createCollectionIndexes(ctx, db, "logs", logIndexes); err != nil {
		return err
	}

	return nil
}

// createCollectionIndexes creates indexes for a specific collection
func createCollectionIndexes(ctx context.Context, db *mongo.Database, collection string, indexes []mongo.IndexModel) error {
	_, err := db.Collection(collection).Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes for collection %s: %w", collection, err)
	}
	return nil
}

// createAdminUser creates the initial admin user
func createAdminUser(ctx context.Context, db *mongo.Database, config *Config) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser := bson.D{
		{Key: "username", Value: config.AdminUser},
		{Key: "password", Value: string(hashedPassword)},
		{Key: "created_at", Value: time.Now()},
		{Key: "updated_at", Value: time.Now()},
	}

	_, err = db.Collection("admins").InsertOne(ctx, adminUser)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}

// createDefaultRole creates the default worker role
func createDefaultRole(ctx context.Context, db *mongo.Database) error {
	defaultRole := bson.D{
		{Key: "name", Value: "worker"},
		{Key: "description", Value: "Default worker role"},
		{Key: "created_at", Value: time.Now()},
		{Key: "permissions", Value: bson.A{
			"execute_tasks",
			"report_status",
		}},
	}

	_, err := db.Collection("roles").InsertOne(ctx, defaultRole)
	if err != nil {
		return fmt.Errorf("failed to create default role: %w", err)
	}

	return nil
}