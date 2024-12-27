package models

import "gorm.io/gorm"

type AccountAvatar struct {
	gorm.Model
	AvatarURL string `gorm:"type:TEXT"`
	AccountID uint   `gorm:"not null"`
	IsInUsed  bool   `gorm:"default:true"`
	IsDeleted bool   `gorm:"default:false"`
}
