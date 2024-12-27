package models

import "gorm.io/gorm"

type PostCommentEditHistory struct {
	gorm.Model
	PostID        uint        `gorm:"not null"`
	CommentID     uint        `gorm:"not null"`
	BeforeContent string      `gorm:"not null;type:TEXT"`
	AfterContent  string      `gorm:"not null;type:TEXT"`
	Post          Post        `gorm:"foreignkey:PostID"`
	Comment       PostComment `gorm:"foreignkey:CommentID"`
}
