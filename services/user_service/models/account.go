package models

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	Username               string `gorm:"unique; not null"`
	Password               string `gorm:"not null"`
	AccountRoleID          uint   `gorm:"default:1"`
	AccountCreatedByMethod string `gorm:"type:ENUM('google', 'normal');default:'normal'"`
	IsBanned               bool   `gorm:"default:false"`
	IsRestricted           bool   `gorm:"default:false"`
	IsSelfDeleted          bool   `gorm:"default:false"`
}
