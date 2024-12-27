package models

import "gorm.io/gorm"

type AccountAvatarHistory struct {
	gorm.Model
	AccountID    uint    `gorm:"not null"`
	AvatarURL    string  `gorm:"type:TEXT;not null"`
	UploadStatus string  `gorm:"type:ENUM('uploaded','failed');not null"`
	Account      Account `gorm:"foreignkey:AccountID"`
}
