package models

import "gorm.io/gorm"

type PostTagFriend struct {
	gorm.Model
	PostID       uint `gorm:"not null"`
	TagAccountID uint `gorm:"not null"`
	IsDeleted    bool `gorm:"default:false"`
	Order        uint `gorm:"not null"`
}
