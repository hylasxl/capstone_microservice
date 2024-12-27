package models

import "gorm.io/gorm"

type AccountChangeNameHistory struct {
	gorm.Model
	AccountID     uint    `gorm:"not null"`
	ChangingType  string  `gorm:"type:ENUM('display_name','user_name');default:'display_name';not null"`
	FromFirstName string  `gorm:"null"`
	FromLastName  string  `gorm:"default: null"`
	FromUsername  string  `gorm:"default: null"`
	ToFirstName   string  `gorm:"default: null"`
	ToLastName    string  `gorm:"default: null"`
	ToUsername    string  `gorm:"default: null"`
	Account       Account `gorm:"foreignkey:AccountID"`
}
