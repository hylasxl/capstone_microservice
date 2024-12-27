package models

import "gorm.io/gorm"

type NotificationType struct {
	gorm.Model
	TypeName    string `gorm:"not null; type: text"`
	Description string `gorm:"not null; type: text"`
}
