package controllers

import (
    "errors"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "go-api/pkg/db"
    "go-api/pkg/models"
    "go-api/pkg/utils"
)

type CategoryController struct {
    Collection *mongo.Collection
}

func NewCategoryController(client *mongo.Client, dbName string) *CategoryController {
    return &CategoryController{Collection: client.Database(dbName).Collection("categories")}
}

// CreateCategory creates a category (accepts name or category_name)
func (cc *CategoryController) CreateCategory(c *gin.Context) {
    type inBody struct {
        Name         string `json:"name"`
        CategoryName string `json:"category_name"`
    }
    var body inBody
    if err := c.ShouldBindJSON(&body); err != nil {
        utils.BadRequest(c, err)
        return
    }
    var m models.Category
    if body.Name != "" {
        m.Name = strings.TrimSpace(body.Name)
    } else {
        m.Name = strings.TrimSpace(body.CategoryName)
    }
    if m.Name == "" {
        utils.BadRequest(c, errors.New("name is required"))
        return
    }
    now := time.Now().UTC()
    m.CreatedAt = now
    m.UpdatedAt = now
    res, err := cc.Collection.InsertOne(db.Ctx, m)
    if err != nil {
        utils.ServerError(c, err)
        return
    }
    m.ID = res.InsertedID.(primitive.ObjectID)
    c.JSON(http.StatusCreated, m)
}

// GetCategory by id
func (cc *CategoryController) GetCategory(c *gin.Context) {
    oid, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequest(c, errors.New("invalid id"))
        return
    }
    var raw bson.M
    if err := cc.Collection.FindOne(db.Ctx, bson.M{"_id": oid}).Decode(&raw); err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            utils.NotFound(c)
            return
        }
        utils.ServerError(c, err)
        return
    }
    c.JSON(http.StatusOK, mapCategory(raw))
}

// ListCategories with pagination and sorting (name, createdAt)
func (cc *CategoryController) ListCategories(c *gin.Context) {
    filter := bson.M{}
    if name := strings.TrimSpace(c.Query("name")); name != "" {
        filter["name"] = bson.M{"$regex": name, "$options": "i"}
    }
    page := 1
    limit := 10
    if v, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && v > 0 {
        page = v
    }
    if v, err := strconv.Atoi(c.DefaultQuery("limit", "10")); err == nil && v > 0 && v <= 100 {
        limit = v
    }
    skip := int64((page - 1) * limit)

    sortField := c.DefaultQuery("sort", "createdAt")
    if sortField != "name" && sortField != "createdAt" {
        sortField = "createdAt"
    }
    order := c.DefaultQuery("order", "desc")
    sortDir := int32(-1)
    if strings.ToLower(order) == "asc" {
        sortDir = 1
    }
    opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: sortField, Value: sortDir}})

    cur, err := cc.Collection.Find(db.Ctx, filter, opts)
    if err != nil {
        utils.ServerError(c, err)
        return
    }
    defer cur.Close(db.Ctx)
    var raws []bson.M
    if err := cur.All(db.Ctx, &raws); err != nil {
        utils.ServerError(c, err)
        return
    }
    cats := make([]models.Category, 0, len(raws))
    for _, r := range raws { cats = append(cats, mapCategory(r)) }
    total, err := cc.Collection.CountDocuments(db.Ctx, filter)
    if err != nil {
        utils.ServerError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{"items": cats, "page": page, "limit": limit, "total": total})
}

// mapCategory converts raw docs to Category, handling category_name alias
func mapCategory(raw bson.M) models.Category {
    var out models.Category
    if id, ok := raw["_id"].(primitive.ObjectID); ok { out.ID = id }
    if n, ok := raw["name"].(string); ok && n != "" { out.Name = n } else if cn, ok := raw["category_name"].(string); ok { out.Name = cn }
    if ct, ok := raw["createdAt"].(time.Time); ok { out.CreatedAt = ct }
    if ut, ok := raw["updatedAt"].(time.Time); ok { out.UpdatedAt = ut }
    if out.CreatedAt.IsZero() && out.ID != primitive.NilObjectID { out.CreatedAt = out.ID.Timestamp() }
    if out.UpdatedAt.IsZero() { out.UpdatedAt = out.CreatedAt }
    return out
}

// UpdateCategory
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
    oid, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequest(c, errors.New("invalid id"))
        return
    }
    type inBody struct {
        Name         string `json:"name"`
        CategoryName string `json:"category_name"`
    }
    var body inBody
    if err := c.ShouldBindJSON(&body); err != nil {
        utils.BadRequest(c, err)
        return
    }
    set := bson.M{"updatedAt": time.Now().UTC()}
    if n := strings.TrimSpace(body.Name); n != "" {
        set["name"] = n
    } else if n := strings.TrimSpace(body.CategoryName); n != "" {
        set["name"] = n
    }
    res := cc.Collection.FindOneAndUpdate(db.Ctx, bson.M{"_id": oid}, bson.M{"$set": set}, options.FindOneAndUpdate().SetReturnDocument(options.After))
    var out models.Category
    if err := res.Decode(&out); err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            utils.NotFound(c)
            return
        }
        utils.ServerError(c, err)
        return
    }
    c.JSON(http.StatusOK, out)
}

// DeleteCategory
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
    oid, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequest(c, errors.New("invalid id"))
        return
    }
    res, err := cc.Collection.DeleteOne(db.Ctx, bson.M{"_id": oid})
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
