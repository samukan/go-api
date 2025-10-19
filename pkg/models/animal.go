package models

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type GeoPoint struct {
    Type        string    `bson:"type" json:"type"`
    Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

type Animal struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name      string             `bson:"name" json:"name" validate:"required,min=2,max=100"`
    Species   string             `bson:"species" json:"species" validate:"required"`
    Age       int                `bson:"age" json:"age" validate:"gte=0,lte=120"`
    Adopted   bool               `bson:"adopted" json:"adopted"`
    Image     string             `bson:"image,omitempty" json:"image,omitempty"`
    Owner     string             `bson:"owner,omitempty" json:"owner,omitempty"`
    Location  *GeoPoint          `bson:"location,omitempty" json:"location,omitempty"`
    CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Pagination struct {
    Page  int `form:"page"`
    Limit int `form:"limit"`
}
