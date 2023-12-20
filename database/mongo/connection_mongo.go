/*
Package mongo provides a streamlined interface for connecting to a MongoDB database
within the github.com/a-dev-mobile/app-update-api project. Key functionalities include initializing MongoDB
client instances and constructing database connection strings.

Key Functions:
  - GetDB: Establishes and returns a MongoDB client based on given configuration and context.
  - buildConnString: Generates a connection string for MongoDB using environment variables
    and application configuration for secure credential management.

This package simplifies database interaction by abstracting connection details,
offering a straightforward approach to establish database connections for various
operations within the application.

Example Usage:

	func main() {
		ctx := context.Background()
		cfg := config.LoadConfig()
		dbClient, err := mongo.GetDB(ctx, cfg)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}
		// Use dbClient to interact with the database
	}
*/
package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetDB initializes and returns a new MongoDB client using provided context and configuration.
// It ensures the establishment of a connection to the MongoDB database, returning an error
// if the connection setup fails. This function is key for accessing the database in various parts of the application.
func GetDB(ctx context.Context, dbUser string, dbPassword string, dbHost string, dbPort string) (*mongo.Client, error) {

	connString := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/",
		dbUser, dbPassword, dbHost, dbPort)

	clientOptions := options.Client().ApplyURI(connString)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Проверка подключения
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB using connection string %s: %w", connString, err)
	}

	return client, nil
}
