package models

import "time"

type Admin struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username  string    `json:"username" bson:"username"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"password" bson:"password"` // store hashed password
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
