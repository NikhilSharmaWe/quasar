package model

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getDBClient() (*mongo.Client, error) {
	mongoURL := os.Getenv("MONGO_URL")
	clientOptions := options.Client().ApplyURI(mongoURL)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getCollection(client *mongo.Client, name string) *mongo.Collection {
	dbName := os.Getenv("DATABASE_NAME")
	collection := client.Database(dbName).Collection(name)
	return collection
}
