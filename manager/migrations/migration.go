package migrations

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"time"

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
func RunMigrations(dbURI, dbName string, migrations []Migration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	db := client.Database(dbName)

	// Create a collection to track migrations
	migrationCollectionName := "migrations"
	migrationCollection := db.Collection(migrationCollectionName)

	// Ensure the migration collection exists
	names, err := db.ListCollectionNames(ctx, &mongo.ListCollectionsOptions{})
	if err != nil {
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
			return fmt.Errorf("failed to create migration collection: %w", err)
		}
	}

	// Get the last applied migration version
	var lastAppliedMigration struct {
		Version int `bson:"version"`
	}
	err = migrationCollection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})).Decode(&lastAppliedMigration)
	if err != nil && err != mongo.ErrNoDocuments {
		return fmt.Errorf("failed to get last applied migration: %w", err)
	}

	lastVersion := lastAppliedMigration.Version

	// Apply migrations
	for _, migration := range migrations {
		if migration.Version > lastVersion {
			log.Printf("Applying migration version %d: %s", migration.Version, migration.Description)
			if err := migration.Up(db); err != nil {
				return fmt.Errorf("failed to apply migration version %d: %w", migration.Version, err)
			}

			// Record the migration
			_, err = migrationCollection.InsertOne(ctx, bson.M{"version": migration.Version, "description": migration.Description, "applied_at": time.Now()})
			if err != nil {
				return fmt.Errorf("failed to record migration version %d: %w", migration.Version, err)
			}
			log.Printf("Migration version %d applied successfully", migration.Version)
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}

// Define a helper function to get the MongoDB database
func getDatabase(dbURI, dbName string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	return client.Database(dbName), nil
}

// Define a helper function to create a collection with options
func createCollection(db *mongo.Database, collectionName string, options *options.CreateCollectionOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.CreateCollection(ctx, collectionName, options)
	if err != nil {
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
		return fmt.Errorf("failed to create index on collection %s: %w", collectionName, err)
	}

	log.Printf("Index created successfully on collection %s", collectionName)
	return nil
}
