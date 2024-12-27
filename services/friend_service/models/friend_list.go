package models

import "gorm.io/gorm"

type FriendList struct {
	gorm.Model
	FirstAccountID  uint `gorm:"not null"`
	SecondAccountID uint `gorm:"not null"`
	IsValid         bool `gorm:"default:true"`
}
