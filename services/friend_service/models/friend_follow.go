package models

import "gorm.io/gorm"

type FriendFollow struct {
	gorm.Model
	FirstAccountID  uint `gorm:"not null"`
	SecondAccountID uint `gorm:"not null"`
	IsFollowed      bool `gorm:"default:true"`
}
