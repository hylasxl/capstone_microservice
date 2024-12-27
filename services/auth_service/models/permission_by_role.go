package models

import "gorm.io/gorm"

type PermissionByRole struct {
	gorm.Model
	RoleID       uint       `gorm:"not null"`
	PermissionID uint       `gorm:"not null"`
	Permission   Permission `gorm:"foreignKey:PermissionID"`
}
