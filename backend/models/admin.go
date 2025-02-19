package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Admin struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username  string             `json:"username" bson:"username" validate:"required"`
	Email     string             `json:"email" bson:"email,omitempty"`
	Password  string             `json:"password" bson:"password" validate:"required"` // store hashed password
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type PasswordPolicy struct {
	MinLength        int  `json:"min_length" bson:"min_length"`
	RequireUppercase bool `json:"require_uppercase" bson:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase" bson:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers" bson:"require_numbers"`
	RequireSpecial   bool `json:"require_special" bson:"require_special"`
}
