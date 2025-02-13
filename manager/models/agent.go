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