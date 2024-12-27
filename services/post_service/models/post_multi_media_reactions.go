package models

import "gorm.io/gorm"

type PostMultiMediaReaction struct {
	gorm.Model
	PostMediaID  uint           `gorm:"not null"`
	AccountID    uint           `gorm:"not null"`
	IsRecalled   bool           `gorm:"default:false"`
	ReactionType string         `gorm:"type:ENUM('like','dislike','love','hate','cry');default:null"`
	Media        PostMultiMedia `gorm:"foreignkey:PostMediaID"`
}
