package models

import "gorm.io/gorm"

type Permission struct {
	gorm.Model
	PermissionURL string `gorm:"not null; type:text"`
	Description   string `gorm:"not null; type:text"`
}
