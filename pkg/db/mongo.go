package db

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Ctx = context.Background()

func Connect(uri string) (*mongo.Client, error) {
    ctx, cancel := context.WithTimeout(Ctx, 10*time.Second)
    defer cancel()
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        return nil, err
    }
    return client, nil
}
