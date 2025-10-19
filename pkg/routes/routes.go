package routes

import (
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"

    "go-api/pkg/controllers"
)

func RegisterAnimalRoutes(rg *gin.RouterGroup, client *mongo.Client, dbName string) {
    ctrl := controllers.NewAnimalController(client, dbName)

    g := rg.Group("/animals")
    {
        g.POST("", ctrl.CreateAnimal)
        g.GET("", ctrl.ListAnimals)
        g.GET("/:id", ctrl.GetAnimal)
        g.PUT("/:id", ctrl.UpdateAnimal)
        g.DELETE("/:id", ctrl.DeleteAnimal)
    }
}
