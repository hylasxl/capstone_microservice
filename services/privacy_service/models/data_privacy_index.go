package models

import "gorm.io/gorm"

type DataPrivacyIndex struct {
	gorm.Model
	DataFieldName string `gorm:"not null"`
}
