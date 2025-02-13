package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogEntry struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Endpoint  string    `json:"endpoint" bson:"endpoint"`
	AgentID   string    `json:"agent_id,omitempty" bson:"agent_id,omitempty"`
	Status    string    `json:"status" bson:"status"`
	Details   string    `json:"details,omitempty" bson:"details,omitempty"`
}

type LogsResponse struct {
    Logs []LogEntry `json:"logs"`
}