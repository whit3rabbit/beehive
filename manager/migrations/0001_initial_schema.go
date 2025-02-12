package migrations

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Migration 0001: Initial schema setup
var Migration0001 = Migration{
	Version:     1,
	Description: "Create initial collections and indexes",
	Up: func(db *mongo.Database) error {
		// Create admins collection
		err := createCollection(db, "admins", nil)
		if err != nil {
			return err
		}

		// Create unique index on username in admins collection
		indexOptions := options.Index().SetUnique(true)
		keys := bson.M{"username": 1}
		err = createIndex(db, "admins", keys, indexOptions)
		if err != nil {
			return err
		}

		// Create agents collection
		err = createCollection(db, "agents", nil)
		if err != nil {
			return err
		}

		// Create unique index on uuid in agents collection
		keys = bson.M{"uuid": 1}
		err = createIndex(db, "agents", keys, options.Index().SetUnique(true))
		if err != nil {
			return err
		}

		// Create roles collection
		err = createCollection(db, "roles", nil)
		if err != nil {
			return err
		}

		// Create tasks collection
		err = createCollection(db, "tasks", nil)
		if err != nil {
			return err
		}

		// Create logs collection
		err = createCollection(db, "logs", nil)
		if err != nil {
			return err
		}

		log.Println("Migration 0001 Up executed successfully")
		return nil
	},
	Down: func(db *mongo.Database) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Drop collections (for rollback purposes)
		err := db.Collection("admins").Drop(ctx)
		if err != nil {
			return err
		}
		err = db.Collection("agents").Drop(ctx)
		if err != nil {
			return err
		}
		err = db.Collection("roles").Drop(ctx)
		if err != nil {
			return err
		}
		err = db.Collection("tasks").Drop(ctx)
		if err != nil {
			return err
		}
		err = db.Collection("logs").Drop(ctx)
		if err != nil {
			return err
		}

		log.Println("Migration 0001 Down executed successfully")
		return nil
	},
}
