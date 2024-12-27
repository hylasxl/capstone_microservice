package models

import "gorm.io/gorm"

type PostComment struct {
	gorm.Model
	PostID           uint          `gorm:"not null"`
	AccountID        uint          `gorm:"not null"`
	Content          string        `gorm:"type:TEXT"`
	IsSelfDeleted    bool          `gorm:"default:false"`
	IsDeletedByAdmin bool          `gorm:"default:false"`
	IsReply          bool          `gorm:"default:false"`
	IsEdited         bool          `gorm:"default:false"`
	Level            uint          `gorm:"default:1"`
	ReplyFromID      uint          `gorm:"default: null"`
	Post             Post          `gorm:"foreignkey:PostID"`
	ReplyFrom        *PostComment  `gorm:"foreignkey:ReplyFromID"`
	Replies          []PostComment `gorm:"foreignkey:ReplyFromID"`
}
