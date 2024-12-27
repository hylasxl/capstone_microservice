package models

import "gorm.io/gorm"

type PostReaction struct {
	gorm.Model
	PostID       uint   `gorm:"not null"`
	AccountID    uint   `gorm:"not null"`
	IsRecalled   bool   `gorm:"default:false"`
	ReactionType string `gorm:"type:ENUM('like','dislike','love','hate','cry');default:null"`
	Post         Post   `gorm:"foreignkey:PostID"`
}
