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
    // Accept both our schema and dataset-style fields
    type animalIn struct {
        Name       string `json:"name"`
        AnimalName string `json:"animal_name"`
        Species    string `json:"species"`
        Birthdate  string `json:"birthdate"`
        Age        *int   `json:"age"`
        Adopted    *bool  `json:"adopted"`
        Image      string           `json:"image"`
        Owner      string           `json:"owner"`
        Location   *models.GeoPoint `json:"location"`
    }
    var body animalIn
    if err := c.ShouldBindJSON(&body); err != nil {
        utils.BadRequest(c, err)
        return
    }

    var in models.Animal
    if body.Name != "" {
        in.Name = body.Name
    } else {
        in.Name = body.AnimalName
    }
    in.Species = body.Species
    if body.Age != nil {
        in.Age = *body.Age
    } else if body.Birthdate != "" {
        if age := utils.AgeFromBirthdate(body.Birthdate); age >= 0 {
            in.Age = age
        }
    }
    if body.Adopted != nil {
        in.Adopted = *body.Adopted
    }
    if body.Image != "" {
        in.Image = body.Image
    }
    if body.Owner != "" {
        in.Owner = body.Owner
    }
    if body.Location != nil {
        // Default to GeoJSON Point if not specified
        gp := *body.Location
        if gp.Type == "" {
            gp.Type = "Point"
        }
        in.Location = &gp
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
    var raw bson.M
    err = ac.Collection.FindOne(db.Ctx, bson.M{"_id": oid}).Decode(&raw)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            utils.NotFound(c)
            return
        }
        utils.ServerError(c, err)
        return
    }
    c.JSON(http.StatusOK, mapAnimal(raw))
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
        // Match species if stored as string or as ObjectID
        ors := []bson.M{{"species": species}}
        if oid, err := primitive.ObjectIDFromHex(species); err == nil {
            ors = append(ors, bson.M{"species": oid})
        }
        filter["$or"] = appendOr(filter["$or"], ors...)
    }
    if name := strings.TrimSpace(c.Query("name")); name != "" {
        // support either name or animal_name
        filter["$or"] = []bson.M{
            {"name": bson.M{"$regex": name, "$options": "i"}},
            {"animal_name": bson.M{"$regex": name, "$options": "i"}},
        }
    }
    // Age may not exist; if birthdate exists in dataset, we can compute ages client-side.
    // We'll not filter by age at DB-level when birthdate is used; keep age filter only if stored.
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
    // allow sorting by birthdate if present in dataset
    allowed := map[string]bool{"name": true, "age": true, "createdAt": true, "birthdate": true, "animal_name": true}
    if !allowed[sortField] {
        sortField = "createdAt"
    }
    order := c.DefaultQuery("order", "desc")
    sortDir := int32(-1)
    if strings.ToLower(order) == "asc" {
        sortDir = 1
    }

    // Build sort spec. If sorting by createdAt, add _id as a secondary sort to approximate creation time for docs missing createdAt.
    sortSpec := bson.D{{Key: sortField, Value: sortDir}}
    if sortField == "createdAt" {
        sortSpec = bson.D{{Key: "createdAt", Value: sortDir}, {Key: "_id", Value: sortDir}}
    }
    findOpts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(sortSpec)

    cur, err := ac.Collection.Find(db.Ctx, filter, findOpts)
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

    items := make([]models.Animal, 0, len(raws))
    for _, r := range raws {
        items = append(items, mapAnimal(r))
    }

    total, err := ac.Collection.CountDocuments(db.Ctx, filter)
    if err != nil {
        utils.ServerError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "items": items,
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
    type animalIn struct {
        Name       string `json:"name"`
        AnimalName string `json:"animal_name"`
        Species    string `json:"species"`
        Birthdate  string `json:"birthdate"`
        Age        *int   `json:"age"`
        Adopted    *bool  `json:"adopted"`
        Image      string           `json:"image"`
        Owner      string           `json:"owner"`
        Location   *models.GeoPoint `json:"location"`
    }
    var body animalIn
    if err := c.ShouldBindJSON(&body); err != nil {
        utils.BadRequest(c, err)
        return
    }

    set := bson.M{}
    if body.Name != "" || body.AnimalName != "" {
        if body.Name != "" {
            set["name"] = body.Name
        } else {
            set["name"] = body.AnimalName
        }
    }
    if body.Species != "" {
        set["species"] = body.Species
    }
    if body.Age != nil {
        set["age"] = *body.Age
    } else if body.Birthdate != "" {
        if age := utils.AgeFromBirthdate(body.Birthdate); age >= 0 {
            set["age"] = age
        }
    }
    if body.Adopted != nil {
        set["adopted"] = *body.Adopted
    }
    if body.Image != "" {
        set["image"] = body.Image
    }
    if body.Owner != "" {
        set["owner"] = body.Owner
    }
    if body.Location != nil {
        gp := *body.Location
        if gp.Type == "" {
            gp.Type = "Point"
        }
        set["location"] = gp
    }
    set["updatedAt"] = time.Now().UTC()

    update := bson.M{"$set": set}

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

// mapAnimal converts a raw bson document (which may come from a different dataset schema)
// to our models.Animal format. It handles aliases like animal_name -> name and computes age from birthdate if needed.
func mapAnimal(raw bson.M) models.Animal {
    var out models.Animal
    if id, ok := raw["_id"].(primitive.ObjectID); ok {
        out.ID = id
    }
    // name from either name or animal_name
    if n, ok := raw["name"].(string); ok && n != "" {
        out.Name = n
    } else if an, ok := raw["animal_name"].(string); ok {
        out.Name = an
    }
    if s, ok := raw["species"].(string); ok {
        out.Species = s
    } else if soid, ok := raw["species"].(primitive.ObjectID); ok {
        out.Species = soid.Hex()
    }
    // Prefer stored age; otherwise derive from birthdate (YYYY-MM-DD)
    if a, ok := raw["age"].(int32); ok {
        out.Age = int(a)
    } else if a64, ok := raw["age"].(int64); ok {
        out.Age = int(a64)
    } else if aF, ok := raw["age"].(float64); ok {
        out.Age = int(aF)
    } else if bd, ok := raw["birthdate"].(string); ok {
        out.Age = utils.AgeFromBirthdate(bd)
        if out.Age < 0 {
            out.Age = 0
        }
    } else if bdt, ok := raw["birthdate"].(time.Time); ok {
        out.Age = utils.AgeFromTime(bdt)
        if out.Age < 0 {
            out.Age = 0
        }
    } else if bddt, ok := raw["birthdate"].(primitive.DateTime); ok {
        out.Age = utils.AgeFromTime(bddt.Time())
        if out.Age < 0 {
            out.Age = 0
        }
    }
    if ad, ok := raw["adopted"].(bool); ok {
        out.Adopted = ad
    }
    if img, ok := raw["image"].(string); ok {
        out.Image = img
    }
    if owner, ok := raw["owner"].(string); ok {
        out.Owner = owner
    }
    // location as GeoJSON
    if loc, ok := raw["location"].(bson.M); ok {
        gp := models.GeoPoint{}
        if t, ok := loc["type"].(string); ok {
            gp.Type = t
        }
        // coordinates might be []interface{} or []float64
        if coords, ok := loc["coordinates"].(primitive.A); ok {
            gp.Coordinates = make([]float64, 0, len(coords))
            for _, v := range coords {
                switch num := v.(type) {
                case float64:
                    gp.Coordinates = append(gp.Coordinates, num)
                case float32:
                    gp.Coordinates = append(gp.Coordinates, float64(num))
                case int32:
                    gp.Coordinates = append(gp.Coordinates, float64(num))
                case int64:
                    gp.Coordinates = append(gp.Coordinates, float64(num))
                case int:
                    gp.Coordinates = append(gp.Coordinates, float64(num))
                }
            }
        } else if coords2, ok := loc["coordinates"].([]interface{}); ok {
            gp.Coordinates = make([]float64, 0, len(coords2))
            for _, v := range coords2 {
                switch num := v.(type) {
                case float64:
                    gp.Coordinates = append(gp.Coordinates, num)
                case float32:
                    gp.Coordinates = append(gp.Coordinates, float64(num))
                case int32:
                    gp.Coordinates = append(gp.Coordinates, float64(num))
                case int64:
                    gp.Coordinates = append(gp.Coordinates, float64(num))
                case int:
                    gp.Coordinates = append(gp.Coordinates, float64(num))
                }
            }
        } else if coordsF64, ok := loc["coordinates"].([]float64); ok {
            gp.Coordinates = coordsF64
        }
        out.Location = &gp
    }
    if ct, ok := raw["createdAt"].(time.Time); ok { out.CreatedAt = ct }
    if ut, ok := raw["updatedAt"].(time.Time); ok { out.UpdatedAt = ut }
    // Fallbacks for legacy docs: derive createdAt from ObjectID timestamp; updatedAt defaults to createdAt
    if out.CreatedAt.IsZero() && out.ID != primitive.NilObjectID {
        out.CreatedAt = out.ID.Timestamp()
    }
    if out.UpdatedAt.IsZero() {
        out.UpdatedAt = out.CreatedAt
    }
    return out
}

// appendOr appends conditions to an existing $or which can be nil, a slice, or missing.
func appendOr(existing any, conds ...bson.M) []bson.M {
    var or []bson.M
    if existing != nil {
        if arr, ok := existing.([]bson.M); ok {
            or = append(or, arr...)
        }
    }
    or = append(or, conds...)
    return or
}
