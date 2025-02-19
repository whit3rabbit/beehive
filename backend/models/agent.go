package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Agent struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UUID      string             `json:"uuid" bson:"uuid" validate:"required"`
	Hostname  string             `json:"hostname" bson:"hostname" validate:"required"`
	MacHash   string             `json:"mac_hash" bson:"mac_hash" validate:"required"`
	Nickname  string             `json:"nickname" bson:"nickname"`
	Role      string             `json:"role" bson:"role"`
	APIKey    string             `json:"-" bson:"api_key"`           // Store API key but don't expose in JSON
	APISecret string             `json:"-" bson:"api_secret"`        // Store API secret but don't expose in JSON
	Status    string             `json:"status" bson:"status"`       // "active", "inactive", "disconnected"
	LastSeen  time.Time          `json:"last_seen" bson:"last_seen"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type AgentSummary struct {
    UUID     string    `json:"uuid"`
    Nickname string    `json:"nickname"`
    Hostname string    `json:"hostname"`
    LastSeen time.Time `json:"last_seen"`
    Role     string    `json:"role"`
}

type HeartbeatRequest struct {
    UUID      string    `json:"uuid" validate:"required"`
    Timestamp time.Time `json:"timestamp" validate:"required"`  // Added timestamp per spec
}

type HeartbeatResponse struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
}

// ToSummary converts an Agent to an AgentSummary.
func (a *Agent) ToSummary() AgentSummary {
    return AgentSummary{
        UUID:     a.UUID,
        Nickname: a.Nickname,
        Hostname: a.Hostname,
        LastSeen: a.LastSeen,
        Role:     a.Role,
    }
}
