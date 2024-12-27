package models

import "gorm.io/gorm"

type PostMultiMediaEditCommentHistory struct {
	gorm.Model
	MediaID       uint                  `gorm:"not null"`
	CommentID     uint                  `gorm:"not null"`
	BeforeContent string                `gorm:"not null;type:TEXT"`
	AfterContent  string                `gorm:"not null;type:TEXT"`
	Media         PostMultiMedia        `gorm:"foreignkey:MediaID"`
	Comment       PostMultiMediaComment `gorm:"foreignkey:CommentID"`
}
