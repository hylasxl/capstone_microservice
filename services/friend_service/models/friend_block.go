package models

import "gorm.io/gorm"

type FriendBlock struct {
	gorm.Model
	FirstAccountID  uint `gorm:"not null"`
	SecondAccountID uint `gorm:"not null"`
	IsBlocked       bool `gorm:"default:false"`
}
