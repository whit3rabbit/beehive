package mongodb

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client is the MongoDB client instance.
var (
	Client     *mongo.Client
	clientOnce sync.Once
	connectErr error
)

// Connect initializes and verifies a connection to MongoDB using the given URI.
// It uses a singleton pattern to ensure only one client is created.
func Connect(uri string) error {
	clientOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI(uri)

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			connectErr = err
			return
		}

		// Verify connection
		if err := client.Ping(ctx, nil); err != nil {
			connectErr = err
			return
		}

		Client = client
		log.Println("Connected to MongoDB!")
	})

	return connectErr
}

// Disconnect closes the connection to MongoDB.
func Disconnect() error {
	if Client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := Client.Disconnect(ctx)
	return err
}
