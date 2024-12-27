package models

import (
	"gorm.io/gorm"
	"time"
)

type OnlineHistory struct {
	gorm.Model
	AccountID  uint      `gorm:"not null"`
	LastOnline time.Time `gorm:"type:DATETIME"`
}
