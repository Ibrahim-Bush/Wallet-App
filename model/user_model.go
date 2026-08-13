package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"-" gorm:"not null"`
	Role     string `json:"role" gorm:"not null"`
}

type Auth_request struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User_claims struct {
	UserID               int    `json:"user_id"`
	Username             string `json:"username"`
	Role                 string `json:"role"`
	jwt.RegisteredClaims        //this struct for the standard fields like: exp and other fields.
}
