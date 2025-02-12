package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AgentID    string             `json:"agent_id" bson:"agent_id"`
	Type       string             `json:"type" bson:"type"`
	Parameters map[string]interface{} `json:"parameters" bson:"parameters"`
	Status     string             `json:"status" bson:"status"`
	Output     map[string]interface{} `json:"output,omitempty" bson:"output,omitempty"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
	Timeout    int                `json:"timeout" bson:"timeout"`
	StartedAt  time.Time          `json:"started_at,omitempty" bson:"started_at,omitempty"`
}
