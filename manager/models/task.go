package models

import "time"

type Task struct {
	TaskID     string                 `json:"task_id" bson:"task_id"`
	AgentID    string                 `json:"agent_id" bson:"agent_id"`
	Type       string                 `json:"type" bson:"type"`
	Parameters map[string]interface{} `json:"parameters" bson:"parameters"`
	Status     string                 `json:"status" bson:"status"`
	Output     map[string]interface{} `json:"output,omitempty" bson:"output,omitempty"`
	CreatedAt  time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" bson:"updated_at"`
}
