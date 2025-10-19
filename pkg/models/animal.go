package models

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Animal struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name      string             `bson:"name" json:"name" validate:"required,min=2,max=100"`
    Species   string             `bson:"species" json:"species" validate:"required,oneof=dog cat bird fish reptile other"`
    Age       int                `bson:"age" json:"age" validate:"gte=0,lte=120"`
    Adopted   bool               `bson:"adopted" json:"adopted"`
    CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Pagination struct {
    Page  int `form:"page"`
    Limit int `form:"limit"`
}
