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

type SpeciesController struct {
    Collection *mongo.Collection
}

func NewSpeciesController(client *mongo.Client, dbName string) *SpeciesController {
    return &SpeciesController{Collection: client.Database(dbName).Collection("species")}
}

func (sc *SpeciesController) CreateSpecies(c *gin.Context) {
    type inBody struct {
        Name         string `json:"name"`
        SpeciesName  string `json:"species_name"`
        Category     string `json:"category"` // id or string
    }
    var body inBody
    if err := c.ShouldBindJSON(&body); err != nil {
        utils.BadRequest(c, err)
        return
    }
    var m models.Species
    if body.Name != "" {
        m.Name = strings.TrimSpace(body.Name)
    } else {
        m.Name = strings.TrimSpace(body.SpeciesName)
    }
    if m.Name == "" {
        utils.BadRequest(c, errors.New("name is required"))
        return
    }
    if body.Category != "" {
        if oid, err := primitive.ObjectIDFromHex(body.Category); err == nil {
            m.Category = oid.Hex()
        } else {
            m.Category = body.Category
        }
    }
    now := time.Now().UTC()
    m.CreatedAt = now
    m.UpdatedAt = now
    res, err := sc.Collection.InsertOne(db.Ctx, m)
    if err != nil {
        utils.ServerError(c, err)
        return
    }
    m.ID = res.InsertedID.(primitive.ObjectID)
    c.JSON(http.StatusCreated, m)
}

func (sc *SpeciesController) GetSpecies(c *gin.Context) {
    oid, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        utils.BadRequest(c, errors.New("invalid id"))
        return
    }
    var raw bson.M
    if err := sc.Collection.FindOne(db.Ctx, bson.M{"_id": oid}).Decode(&raw); err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            utils.NotFound(c)
            return
        }
        utils.ServerError(c, err)
        return
    }
    c.JSON(http.StatusOK, mapSpecies(raw))
}

func (sc *SpeciesController) ListSpecies(c *gin.Context) {
    filter := bson.M{}
    if name := strings.TrimSpace(c.Query("name")); name != "" {
        filter["$or"] = []bson.M{
            {"name": bson.M{"$regex": name, "$options": "i"}},
            {"species_name": bson.M{"$regex": name, "$options": "i"}},
        }
    }
    if category := strings.TrimSpace(c.Query("category")); category != "" {
        ors := []bson.M{{"category": category}}
        if oid, err := primitive.ObjectIDFromHex(category); err == nil {
            ors = append(ors, bson.M{"category": oid})
        }
        filter["$or"] = appendOr(filter["$or"], ors...)
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
    allowed := map[string]bool{"name": true, "createdAt": true, "species_name": true}
    if !allowed[sortField] { sortField = "createdAt" }
    order := c.DefaultQuery("order", "desc")
    sortDir := int32(-1); if strings.ToLower(order)=="asc" { sortDir=1 }
    opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: sortField, Value: sortDir}})
    cur, err := sc.Collection.Find(db.Ctx, filter, opts)
    if err != nil { utils.ServerError(c, err); return }
    defer cur.Close(db.Ctx)
    var raws []bson.M
    if err := cur.All(db.Ctx, &raws); err != nil { utils.ServerError(c, err); return }
    items := make([]models.Species, 0, len(raws))
    for _, r := range raws { items = append(items, mapSpecies(r)) }
    total, err := sc.Collection.CountDocuments(db.Ctx, filter)
    if err != nil { utils.ServerError(c, err); return }
    c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "limit": limit, "total": total})
}

func (sc *SpeciesController) UpdateSpecies(c *gin.Context) {
    oid, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil { utils.BadRequest(c, errors.New("invalid id")); return }
    type inBody struct {
        Name        string `json:"name"`
        SpeciesName string `json:"species_name"`
        Category    string `json:"category"`
    }
    var body inBody
    if err := c.ShouldBindJSON(&body); err != nil { utils.BadRequest(c, err); return }
    set := bson.M{"updatedAt": time.Now().UTC()}
    if n := strings.TrimSpace(body.Name); n != "" { set["name"] = n } else if n := strings.TrimSpace(body.SpeciesName); n != "" { set["name"] = n }
    if body.Category != "" {
        if oid, err := primitive.ObjectIDFromHex(body.Category); err == nil { set["category"] = oid } else { set["category"] = body.Category }
    }
    res := sc.Collection.FindOneAndUpdate(db.Ctx, bson.M{"_id": oid}, bson.M{"$set": set}, options.FindOneAndUpdate().SetReturnDocument(options.After))
    var out models.Species
    if err := res.Decode(&out); err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) { utils.NotFound(c); return }
        utils.ServerError(c, err); return
    }
    c.JSON(http.StatusOK, out)
}

func (sc *SpeciesController) DeleteSpecies(c *gin.Context) {
    oid, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil { utils.BadRequest(c, errors.New("invalid id")); return }
    res, err := sc.Collection.DeleteOne(db.Ctx, bson.M{"_id": oid})
    if err != nil { utils.ServerError(c, err); return }
    if res.DeletedCount == 0 { utils.NotFound(c); return }
    c.Status(http.StatusNoContent)
}

func mapSpecies(raw bson.M) models.Species {
    var out models.Species
    if id, ok := raw["_id"].(primitive.ObjectID); ok { out.ID = id }
    if n, ok := raw["name"].(string); ok && n != "" { out.Name = n } else if sn, ok := raw["species_name"].(string); ok { out.Name = sn }
    if c, ok := raw["category"].(string); ok { out.Category = c } else if coid, ok := raw["category"].(primitive.ObjectID); ok { out.Category = coid.Hex() }
    if ct, ok := raw["createdAt"].(time.Time); ok { out.CreatedAt = ct }
    if ut, ok := raw["updatedAt"].(time.Time); ok { out.UpdatedAt = ut }
    if out.CreatedAt.IsZero() && out.ID != primitive.NilObjectID { out.CreatedAt = out.ID.Timestamp() }
    if out.UpdatedAt.IsZero() { out.UpdatedAt = out.CreatedAt }
    return out
}
