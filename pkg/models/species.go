package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Species struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name      string             `bson:"name" json:"name" validate:"required,min=2,max=200"`
    Category  string             `bson:"category" json:"category" validate:"omitempty"`
    CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
