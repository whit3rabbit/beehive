package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Agent struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UUID      string             `json:"uuid" bson:"uuid"`
	Hostname  string             `json:"hostname" bson:"hostname"`
	MacHash   string             `json:"mac_hash" bson:"mac_hash"`
	Nickname  string             `json:"nickname" bson:"nickname"`
	Role      string             `json:"role" bson:"role"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
