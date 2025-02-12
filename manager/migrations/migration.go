package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/whit3rabbit/beehive/manager/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Migration defines a single migration step.
type Migration struct {
	Version     int
	Description string
	Up          func(db *mongo.Database) error
	Down        func(db *mongo.Database) error
}

// RunMigrations applies the migrations to the database.
func RunMigrations(db *mongo.Database, migrations []Migration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a collection to track migrations
	migrationCollectionName := "migrations"
	migrationCollection := db.Collection(migrationCollectionName)

	// Ensure the migration collection exists
	names, err := db.ListCollectionNames(ctx, mongo.ListCollectionsOptions{})
	if err != nil {
		logger.Error("failed to list collections", zap.Error(err))
		return fmt.Errorf("failed to list collections: %w", err)
	}
	collectionExists := false
	for _, name := range names {
		if name == migrationCollectionName {
			collectionExists = true
			break
		}
	}
	if !collectionExists {
		err = db.CreateCollection(ctx, migrationCollectionName)
		if err != nil {
			logger.Error("failed to create migration collection", zap.Error(err))
			return fmt.Errorf("failed to create migration collection: %w", err)
		}
	}

	// Get the last applied migration version
	var lastAppliedMigration struct {
		Version int `bson:"version"`
	}
	err = migrationCollection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})).Decode(&lastAppliedMigration)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Error("failed to get last applied migration", zap.Error(err))
		return fmt.Errorf("failed to get last applied migration: %w", err)
	}

	lastVersion := lastAppliedMigration.Version

	// Apply migrations
	for _, migration := range migrations {
		if migration.Version > lastVersion {
			logger.Info("Applying migration", zap.Int("version", migration.Version), zap.String("description", migration.Description))
			if err := migration.Up(db); err != nil {
				logger.Error("failed to apply migration", zap.Error(err), zap.Int("migration_version", migration.Version))
				return fmt.Errorf("failed to apply migration version %d: %w", migration.Version, err)
			}

			// Record the migration
			_, err = migrationCollection.InsertOne(ctx, bson.M{"version": migration.Version, "description": migration.Description, "applied_at": time.Now()})
			if err != nil {
				logger.Error("failed to record migration", zap.Error(err), zap.Int("migration_version", migration.Version))
				return fmt.Errorf("failed to record migration version %d: %w", migration.Version, err)
			}
			logger.Info("Migration applied successfully", zap.Int("version", migration.Version))
		}
	}

	logger.Info("All migrations applied successfully")
	return nil
}

// Define a helper function to create a collection with options
func createCollection(db *mongo.Database, collectionName string, options *options.CreateCollectionOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.CreateCollection(ctx, collectionName, options)
	if err != nil {
		logger.Error("failed to create collection", zap.Error(err), zap.String("collection_name", collectionName))
		return fmt.Errorf("failed to create collection %s: %w", collectionName, err)
	}

	logger.Info("Collection created successfully", zap.String("collection_name", collectionName))
	return nil
}

// Define a helper function to create indexes
func createIndex(db *mongo.Database, collectionName string, keys interface{}, options *options.IndexOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.Collection(collectionName)
	model := mongo.IndexModel{
		Keys:    keys,
		Options: options,
	}

	_, err := collection.Indexes().CreateOne(ctx, model)
	if err != nil {
		logger.Error("failed to create index", zap.Error(err), zap.String("collection_name", collectionName))
		return fmt.Errorf("failed to create index on collection %s: %w", collectionName, err)
	}

	logger.Info("Index created successfully", zap.String("collection_name", collectionName))
	return nil
}

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Migration defines a single migration step.
type Migration struct {
	Version     int
	Description string
	Up          func(db *mongo.Database) error
	Down        func(db *mongo.Database) error
}

// RunMigrations applies the migrations to the database.
func RunMigrations(db *mongo.Database, migrations []Migration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a collection to track migrations
	migrationCollectionName := "migrations"
	migrationCollection := db.Collection(migrationCollectionName)

	// Ensure the migration collection exists
	names, err := db.ListCollectionNames(ctx, mongo.ListCollectionsOptions{})
	if err != nil {
		logger.Error("failed to list collections", zap.Error(err))
		return fmt.Errorf("failed to list collections: %w", err)
	}
	collectionExists := false
	for _, name := range names {
		if name == migrationCollectionName {
			collectionExists = true
			break
		}
	}
	if !collectionExists {
		err = db.CreateCollection(ctx, migrationCollectionName)
		if err != nil {
			logger.Error("failed to create migration collection", zap.Error(err))
			return fmt.Errorf("failed to create migration collection: %w", err)
		}
	}

	// Get the last applied migration version
	var lastAppliedMigration struct {
		Version int `bson:"version"`
	}
	err = migrationCollection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})).Decode(&lastAppliedMigration)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Error("failed to get last applied migration", zap.Error(err))
		return fmt.Errorf("failed to get last applied migration: %w", err)
	}

	lastVersion := lastAppliedMigration.Version

	// Apply migrations
	for _, migration := range migrations {
		if migration.Version > lastVersion {
			log.Printf("Applying migration version %d: %s", migration.Version, migration.Description)
			if err := migration.Up(db); err != nil {
				logger.Error("failed to apply migration", zap.Error(err), zap.Int("migration_version", migration.Version))
				return fmt.Errorf("failed to apply migration version %d: %w", migration.Version, err)
			}

			// Record the migration
			_, err = migrationCollection.InsertOne(ctx, bson.M{"version": migration.Version, "description": migration.Description, "applied_at": time.Now()})
			if err != nil {
				logger.Error("failed to record migration", zap.Error(err), zap.Int("migration_version", migration.Version))
				return fmt.Errorf("failed to record migration version %d: %w", migration.Version, err)
			}
			log.Printf("Migration version %d applied successfully", migration.Version)
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}

// Define a helper function to create a collection with options
func createCollection(db *mongo.Database, collectionName string, options *options.CreateCollectionOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.CreateCollection(ctx, collectionName, options)
	if err != nil {
		logger.Error("failed to create collection", zap.Error(err), zap.String("collection_name", collectionName))
		return fmt.Errorf("failed to create collection %s: %w", collectionName, err)
	}

	log.Printf("Collection %s created successfully", collectionName)
	return nil
}

// Define a helper function to create indexes
func createIndex(db *mongo.Database, collectionName string, keys interface{}, options *options.IndexOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.Collection(collectionName)
	model := mongo.IndexModel{
		Keys:    keys,
		Options: options,
	}

	_, err := collection.Indexes().CreateOne(ctx, model)
	if err != nil {
		logger.Error("failed to create index", zap.Error(err), zap.String("collection_name", collectionName))
		return fmt.Errorf("failed to create index on collection %s: %w", collectionName, err)
	}

	log.Printf("Index created successfully on collection %s", collectionName)
	return nil
}
