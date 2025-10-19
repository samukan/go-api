package controllers

import (
    "errors"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "go-api/pkg/db"
    "go-api/pkg/models"
    "go-api/pkg/utils"
)

var validate = validator.New()

type AnimalController struct {
    Collection *mongo.Collection
}

func NewAnimalController(client *mongo.Client, dbName string) *AnimalController {
    return &AnimalController{Collection: client.Database(dbName).Collection("animals")}
}

// CreateAnimal godoc
// @Summary Create a new animal
// @Tags animals
// @Accept json
// @Produce json
// @Param animal body models.Animal true "Animal"
// @Success 201 {object} models.Animal
// @Failure 400 {object} map[string]string
// @Router /animals [post]
func (ac *AnimalController) CreateAnimal(c *gin.Context) {
    var in models.Animal
    if err := c.ShouldBindJSON(&in); err != nil {
        utils.BadRequest(c, err)
        return
    }
    if err := validate.Struct(in); err != nil {
        utils.BadRequest(c, err)
        return
    }
    in.ID = primitive.NilObjectID
    now := time.Now().UTC()
    in.CreatedAt = now
    in.UpdatedAt = now

    res, err := ac.Collection.InsertOne(db.Ctx, in)
    if err != nil {
        utils.ServerError(c, err)
        return
    }
    in.ID = res.InsertedID.(primitive.ObjectID)
    c.JSON(http.StatusCreated, in)
}

// GetAnimal godoc
// @Summary Get animal by id
// @Tags animals
// @Produce json
// @Param id path string true "Animal ID"
// @Success 200 {object} models.Animal
// @Failure 404 {object} map[string]string
// @Router /animals/{id} [get]
func (ac *AnimalController) GetAnimal(c *gin.Context) {
    id := c.Param("id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        utils.BadRequest(c, errors.New("invalid id"))
        return
    }
    var animal models.Animal
    err = ac.Collection.FindOne(db.Ctx, bson.M{"_id": oid}).Decode(&animal)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            utils.NotFound(c)
            return
        }
        utils.ServerError(c, err)
        return
    }
    c.JSON(http.StatusOK, animal)
}

// ListAnimals godoc
// @Summary List animals with filtering, sorting, pagination
// @Tags animals
// @Produce json
// @Param species query string false "Species (dog, cat, bird, fish, reptile, other)"
// @Param name query string false "Name contains"
// @Param minAge query int false "Minimum age"
// @Param maxAge query int false "Maximum age"
// @Param adopted query bool false "Adopted status"
// @Param sort query string false "Sort field (name, age, createdAt)"
// @Param order query string false "asc or desc"
// @Param page query int false "Page number (1-based)"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /animals [get]
func (ac *AnimalController) ListAnimals(c *gin.Context) {
    filter := bson.M{}

    if species := strings.TrimSpace(c.Query("species")); species != "" {
        filter["species"] = species
    }
    if name := strings.TrimSpace(c.Query("name")); name != "" {
        filter["name"] = bson.M{"$regex": name, "$options": "i"}
    }
    if minAgeStr := c.Query("minAge"); minAgeStr != "" {
        if v, err := strconv.Atoi(minAgeStr); err == nil {
            filter["age"] = bson.M{"$gte": v}
        }
    }
    if maxAgeStr := c.Query("maxAge"); maxAgeStr != "" {
        if v, err := strconv.Atoi(maxAgeStr); err == nil {
            if existing, ok := filter["age"].(bson.M); ok {
                existing["$lte"] = v
                filter["age"] = existing
            } else {
                filter["age"] = bson.M{"$lte": v}
            }
        }
    }
    if adoptedStr := c.Query("adopted"); adoptedStr != "" {
        if adoptedStr == "true" || adoptedStr == "false" {
            filter["adopted"] = adoptedStr == "true"
        }
    }

    // pagination
    page := 1
    limit := 10
    if v, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && v > 0 {
        page = v
    }
    if v, err := strconv.Atoi(c.DefaultQuery("limit", "10")); err == nil && v > 0 && v <= 100 {
        limit = v
    }
    skip := int64((page - 1) * limit)

    // sorting
    sortField := c.DefaultQuery("sort", "createdAt")
    if sortField != "name" && sortField != "age" && sortField != "createdAt" {
        sortField = "createdAt"
    }
    order := c.DefaultQuery("order", "desc")
    sortDir := int32(-1)
    if strings.ToLower(order) == "asc" {
        sortDir = 1
    }

    findOpts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: sortField, Value: sortDir}})

    cur, err := ac.Collection.Find(db.Ctx, filter, findOpts)
    if err != nil {
        utils.ServerError(c, err)
        return
    }
    defer cur.Close(db.Ctx)

    var animals []models.Animal
    if err := cur.All(db.Ctx, &animals); err != nil {
        utils.ServerError(c, err)
        return
    }

    total, err := ac.Collection.CountDocuments(db.Ctx, filter)
    if err != nil {
        utils.ServerError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "items": animals,
        "page":  page,
        "limit": limit,
        "total": total,
    })
}

// UpdateAnimal godoc
// @Summary Update an animal by id
// @Tags animals
// @Accept json
// @Produce json
// @Param id path string true "Animal ID"
// @Param animal body models.Animal true "Animal"
// @Success 200 {object} models.Animal
// @Failure 400 {object} map[string]string
// @Router /animals/{id} [put]
func (ac *AnimalController) UpdateAnimal(c *gin.Context) {
    id := c.Param("id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        utils.BadRequest(c, errors.New("invalid id"))
        return
    }
    var in models.Animal
    if err := c.ShouldBindJSON(&in); err != nil {
        utils.BadRequest(c, err)
        return
    }
    if err := validate.Struct(in); err != nil {
        utils.BadRequest(c, err)
        return
    }
    in.UpdatedAt = time.Now().UTC()

    update := bson.M{
        "$set": bson.M{
            "name":      in.Name,
            "species":   in.Species,
            "age":       in.Age,
            "adopted":   in.Adopted,
            "updatedAt": in.UpdatedAt,
        },
    }

    res := ac.Collection.FindOneAndUpdate(db.Ctx, bson.M{"_id": oid}, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
    var updated models.Animal
    if err := res.Decode(&updated); err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            utils.NotFound(c)
            return
        }
        utils.ServerError(c, err)
        return
    }
    c.JSON(http.StatusOK, updated)
}

// DeleteAnimal godoc
// @Summary Delete an animal by id
// @Tags animals
// @Param id path string true "Animal ID"
// @Success 204 {string} string ""
// @Failure 404 {object} map[string]string
// @Router /animals/{id} [delete]
func (ac *AnimalController) DeleteAnimal(c *gin.Context) {
    id := c.Param("id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        utils.BadRequest(c, errors.New("invalid id"))
        return
    }
    res, err := ac.Collection.DeleteOne(db.Ctx, bson.M{"_id": oid})
    if err != nil {
        utils.ServerError(c, err)
        return
    }
    if res.DeletedCount == 0 {
        utils.NotFound(c)
        return
    }
    c.Status(http.StatusNoContent)
}
