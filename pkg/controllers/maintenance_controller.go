package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"

    "go-api/pkg/db"
)

type MaintenanceController struct {
    DB *mongo.Database
}

func NewMaintenanceController(client *mongo.Client, dbName string) *MaintenanceController {
    return &MaintenanceController{DB: client.Database(dbName)}
}

// BackfillTimestamps sets createdAt from ObjectID timestamp when missing,
// and sets updatedAt to createdAt when missing, for core collections.
func (mc *MaintenanceController) BackfillTimestamps(c *gin.Context) {
    collections := []string{"animals", "categories", "species"}

    // Aggregation pipeline updates (MongoDB 4.2+)
    pipeline := mongoPipelineForBackfill()

    type res struct{ Matched, Modified int64 }
    out := map[string]res{}

    for _, name := range collections {
        coll := mc.DB.Collection(name)
        r, err := coll.UpdateMany(db.Ctx, bson.M{}, pipeline)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"collection": name, "error": err.Error()})
            return
        }
        out[name] = res{Matched: r.MatchedCount, Modified: r.ModifiedCount}
    }

    c.JSON(http.StatusOK, out)
}

func mongoPipelineForBackfill() bson.A {
    return bson.A{
        bson.D{{Key: "$set", Value: bson.M{
            "createdAt": bson.M{"$ifNull": bson.A{"$createdAt", bson.M{"$toDate": "$_id"}}},
        }}},
        bson.D{{Key: "$set", Value: bson.M{
            "updatedAt": bson.M{"$ifNull": bson.A{"$updatedAt", "$createdAt"}},
        }}},
    }
}
