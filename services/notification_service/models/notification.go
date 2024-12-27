package models

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	AccountID uint             `gorm:"not null"`
	TypeID    uint             `gorm:"not null"`
	Message   string           `gorm:"not null; type:text"`
	IsRead    bool             `gorm:"not null; default:false"`
	Type      NotificationType `gorm:"not null; foreignKey: TypeID"`
}
