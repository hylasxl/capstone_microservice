package models

import "gorm.io/gorm"

type Device struct {
	gorm.Model
	UserID uint   `gorm:"not null"`
	Token  string `gorm:"not null; type: text"`
}
