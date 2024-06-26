package main

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email             string `gorm:"type:varchar(100);uniqueIndex"`
	Password          string `gorm:"type:varchar(100)"`
	VerificationToken string `gorm:"type:varchar(100)"`
	IsVerified        bool
}
