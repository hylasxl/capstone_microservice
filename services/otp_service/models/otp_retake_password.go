package models

import (
	"gorm.io/gorm"
	"time"
)

type OTPRetakePassword struct {
	gorm.Model
	AccountID uint      `gorm:"not null"`
	OTP       string    `gorm:"not null"`
	Status    string    `gorm:"type:ENUM('valid','expired','invalid');default:'valid';not null"`
	ExpiredAt time.Time `gorm:"not null; type:DATETIME"`
}
