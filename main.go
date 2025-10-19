package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    _ "go-api/docs" // swagger docs
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    "github.com/joho/godotenv"

    "go-api/pkg/config"
    "go-api/pkg/db"
    "go-api/pkg/routes"
)

// @title Animals API
// @version 1.0
// @description A simple REST API for managing animals.
// @BasePath /api/v1
// @schemes http

func main() {
    // Load .env if present
    _ = godotenv.Load()

    // Load configuration
    cfg := config.Load()

    // Initialize DB
    client, err := db.Connect(cfg.MongoURI)
    if err != nil {
        log.Fatalf("failed to connect to MongoDB: %v", err)
    }
    defer client.Disconnect(db.Ctx)

    r := gin.Default()

    // health
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // Swagger UI + static doc
    r.GET("/openapi/doc.json", func(c *gin.Context) {
        // support local dev and container path
        candidates := []string{"./docs/swagger.json", "/docs/swagger.json"}
        for _, p := range candidates {
            if _, err := os.Stat(p); err == nil {
                b, err := os.ReadFile(p)
                if err != nil {
                    c.String(http.StatusInternalServerError, err.Error())
                    return
                }
                c.Data(http.StatusOK, "application/json", b)
                return
            }
        }
        c.JSON(http.StatusNotFound, gin.H{"error": "swagger.json not found"})
    })
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi/doc.json")))

    api := r.Group("/api/v1")
    routes.RegisterAnimalRoutes(api, client, cfg.DatabaseName)

    port := cfg.Port
    if p := os.Getenv("PORT"); p != "" {
        port = p
    }
    if port == "" {
        port = "8080"
    }
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
