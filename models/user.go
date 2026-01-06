package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// =======================
// MongoDB User Document
// =======================
type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	FirstName       string             `bson:"first_name" json:"first_name"`
	LastName        string             `bson:"last_name" json:"last_name"`
	Email           string             `bson:"email" json:"email"`
	Password        string             `bson:"password" json:"password"` // hashed password
	Role            string             `bson:"role" json:"role"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	Token           string             `bson:"token,omitempty" json:"token,omitempty"`
	RefreshToken    string             `bson:"refresh_token,omitempty" json:"refresh_token,omitempty"`
	FavouriteGenres []Genre            `bson:"favourite_genres" json:"favourite_genres"`
}

// =======================
// User Registration Input
// =======================
type UserRegister struct {
	FirstName       string  `json:"first_name" validate:"required,min=2,max=100"`
	LastName        string  `json:"last_name" validate:"required,min=2,max=100"`
	Email           string  `json:"email" validate:"required,email"`
	Password        string  `json:"password" validate:"required,min=6"`
	FavouriteGenres []Genre `json:"favourite_genres" validate:"dive"`
}

// =======================
// User Login Input
// =======================
type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// =======================
// API Response
// =======================
type UserResponse struct {
	UserID          string  `json:"user_id"`
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	Email           string  `json:"email"`
	Role            string  `json:"role"`
	Token           string  `json:"token"`
	RefreshToken    string  `json:"refresh_token"`
	FavouriteGenres []Genre `json:"favourite_genres"`
}
