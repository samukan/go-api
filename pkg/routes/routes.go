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

    // Categories
    cat := controllers.NewCategoryController(client, dbName)
    cg := rg.Group("/categories")
    {
        cg.POST("", cat.CreateCategory)
        cg.GET("", cat.ListCategories)
        cg.GET("/:id", cat.GetCategory)
        cg.PUT("/:id", cat.UpdateCategory)
        cg.DELETE("/:id", cat.DeleteCategory)
    }

    // Species
    sp := controllers.NewSpeciesController(client, dbName)
    sg := rg.Group("/species")
    {
        sg.POST("", sp.CreateSpecies)
        sg.GET("", sp.ListSpecies)
        sg.GET("/:id", sp.GetSpecies)
        sg.PUT("/:id", sp.UpdateSpecies)
        sg.DELETE("/:id", sp.DeleteSpecies)
    }

    // Maintenance
    mt := controllers.NewMaintenanceController(client, dbName)
    mg := rg.Group("/maintenance")
    {
        mg.POST("/backfill-timestamps", mt.BackfillTimestamps)
    }
}
