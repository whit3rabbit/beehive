package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
    ID        primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
    AgentID   string                `json:"agent_id" bson:"agent_id" validate:"required"`
    Type      string                `json:"type" bson:"type" validate:"required,oneof=command_shell file_operation ui_automation browser_automation"`
    Parameters map[string]interface{} `json:"parameters" bson:"parameters" validate:"required"`
    Status    string                `json:"status" bson:"status" validate:"required,oneof=queued running completed failed cancelled timeout"`
    Output    *Output               `json:"output,omitempty" bson:"output,omitempty"`
    CreatedAt time.Time             `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time             `json:"updated_at" bson:"updated_at"`
    Timeout   int                   `json:"timeout" bson:"timeout"`
    StartedAt time.Time             `json:"started_at,omitempty" bson:"started_at,omitempty"`
}

type TaskCreationResponse struct {
    TaskID    string    `json:"task_id"`
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
}

type TaskCancelResponse struct {
    TaskID    string    `json:"task_id"`
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
}

type AgentRegistrationResponse struct {
    APIKey    string    `json:"api_key"`
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
}

type Output struct {
    Logs  string `json:"logs,omitempty" bson:"logs,omitempty"`
    Error string `json:"error,omitempty" bson:"error,omitempty"`
}