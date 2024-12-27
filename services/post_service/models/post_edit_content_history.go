package models

import "gorm.io/gorm"

type PostContentEditHistory struct {
	gorm.Model
	PostID        uint   `gorm:"not null"`
	BeforeContent string `gorm:"type:text"`
	AfterContent  string `gorm:"type:text"`
	Post          Post   `gorm:"foreignkey:PostID"`
}
