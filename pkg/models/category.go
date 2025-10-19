package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name      string             `bson:"name" json:"name" validate:"required,min=2,max=100"`
    CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
