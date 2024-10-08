package repositories

import (
    "context"
    "time"

	"github.com/devops-360-online/go-with-me/config"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(cfg *config.Config) (*mongo.Client, error) {
	print("***********"+cfg.MongoURI)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    clientOptions := options.Client().ApplyURI(cfg.MongoURI)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, err
    }
    // Ping the database to verify connection
    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }
    return client, nil
}
