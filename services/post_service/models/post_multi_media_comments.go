package models

import "gorm.io/gorm"

type PostMultiMediaComment struct {
	gorm.Model
	PostMediaID      uint                    `gorm:"not null"`
	AccountID        uint                    `gorm:"not null"`
	Content          string                  `gorm:"type:TEXT"`
	IsSelfDeleted    bool                    `gorm:"default:false"`
	IsDeletedByAdmin bool                    `gorm:"default:false"`
	IsReply          bool                    `gorm:"default:false"`
	IsEdited         bool                    `gorm:"default:false"`
	Level            uint                    `gorm:"default:1"`
	ReplyFromID      uint                    `gorm:"default: null"`
	ReplyFrom        *PostMultiMediaComment  `gorm:"foreignkey:ReplyFromID"`
	Media            PostMultiMedia          `gorm:"foreignkey:PostMediaID"`
	Replies          []PostMultiMediaComment `gorm:"foreignkey:ReplyFromID"`
}
