package models

import "gorm.io/gorm"

type AccountRole struct {
	gorm.Model
	Role        string `gorm:"type: ENUM('user','admin');default:'user';not null"`
	Description string `gorm:"type: TEXT"`
}
