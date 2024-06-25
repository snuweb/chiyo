package main

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email             string `gorm:"uniqueIndex"`
	Password          string
	VerificationToken string
	IsVerified        bool
}
