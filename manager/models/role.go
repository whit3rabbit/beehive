package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name" validate:"required"`
	Description  string             `json:"description,omitempty" bson:"description,omitempty"`
	Applications []string             `json:"applications,omitempty" bson:"applications,omitempty"`
	DefaultTasks []string             `json:"default_tasks,omitempty" bson:"default_tasks,omitempty"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	BaseModel `json:"base" bson:"base"`
}
