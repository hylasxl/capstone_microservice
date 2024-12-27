package models

import "gorm.io/gorm"

type DataPrivacy struct {
	gorm.Model
	AccountID      uint   `gorm:"not null"`
	DataFieldIndex uint   `gorm:"not null"`
	PrivacyStatus  string `gorm:"type:ENUM('public','private','friend_only');default:'public';not null"`
}
