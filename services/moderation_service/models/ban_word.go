package models

import "gorm.io/gorm"

type BanWord struct {
	gorm.Model
	CreatedByAccountID uint   `gorm:"not null"`
	Content            string `gorm:"type:TEXT;not null"`
	LanguageCode       string `gorm:"type:TEXT;not null"`
	IsDeleted          bool   `gorm:"default:false"`
}
