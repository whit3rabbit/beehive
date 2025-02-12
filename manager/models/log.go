package models

import "time"

type LogEntry struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Endpoint  string    `json:"endpoint" bson:"endpoint"`
	AgentID   string    `json:"agent_id,omitempty" bson:"agent_id,omitempty"`
	Status    string    `json:"status" bson:"status"`
	Details   string    `json:"details,omitempty" bson:"details,omitempty"`
}
