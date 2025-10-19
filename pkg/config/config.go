package config

import (
    "os"
)

type Config struct {
    Port         string
    MongoURI     string
    DatabaseName string
}

func Load() Config {
    cfg := Config{
        Port:         getenv("PORT", "8080"),
        MongoURI:     getenv("MONGO_URI", "mongodb://localhost:27017"),
        DatabaseName: getenv("MONGO_DB", "goapi"),
    }
    return cfg
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
