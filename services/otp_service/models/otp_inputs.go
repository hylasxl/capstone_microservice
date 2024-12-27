package models

import "gorm.io/gorm"

type OTPInputs struct {
	gorm.Model
	AccountID  uint   `gorm:"not null"`
	OTP        string `gorm:"not null"`
	Status     string `gorm:"type:ENUM('approved','rejected','pending');default:'approved';not null"`
	OTPSection string `gorm:"not null; default:'reset password'"`
}
