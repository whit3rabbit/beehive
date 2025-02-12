package models

import "time"

type BaseModel struct {
	CreatedBy string     `json:"created_by" bson:"created_by"`
	UpdatedBy string     `json:"updated_by" bson:"updated_by"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}
