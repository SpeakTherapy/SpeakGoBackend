package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	FirstName     *string            `json:"first_name" bson:"first_name" validate:"required,min=2,max=100"`
	LastName      *string            `json:"last_name" bson:"last_name" validate:"required,min=2,max=100"`
	Email         *string            `json:"email" validate:"email,required"`
	Password      *string            `json:"password" validate:"required,min=6"`
	Role          string             `json:"role"`
	ReferenceCode string 		   	 `json:"reference_code" bson:"reference_code"`
	ProfileImage  string             `json:"profile_image" bson:"profile_image"`
	Token         *string            `json:"token"`
	RefreshToken  *string            `json:"refresh_token" bson:"refresh_token"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
	UserID        string             `json:"user_id" bson:"user_id"`
}
