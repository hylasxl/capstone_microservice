package models

import (
	"gorm.io/gorm"
	"time"
)

type Post struct {
	gorm.Model
	AccountID             uint      `gorm:"not null"`
	Content               string    `gorm:"type:TEXT"`
	IsPublishedLater      bool      `gorm:"default:false"`
	PublishLaterTimestamp time.Time `gorm:"default:null"`
	IsShared              bool      `gorm:"default:false"`
	OriginalPostID        uint      `gorm:"default:null"`
	IsSelfDeleted         bool      `gorm:"default:false"`
	IsDeletedByAdmin      bool      `gorm:"default:false"`
	IsHidden              bool      `gorm:"default:false"`
	IsContentEdited       bool      `gorm:"default:false"`
	IsPublished           bool      `gorm:"default:false"`
	PrivacyStatus         string    `gorm:"type:ENUM('public','private','friend_only');default:'public';not null"`
}
