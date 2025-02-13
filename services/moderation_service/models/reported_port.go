package models

import "gorm.io/gorm"

type ReportedPort struct {
	gorm.Model
	PostID              uint   `gorm:"not null"`
	ReportedByAccountID uint   `gorm:"not null"`
	Reason              string `gorm:"type:text;not null"`
	ReportResolve       string `gorm:"type:ENUM('report_pending','report_skipped','delete_post');default:'report_pending'"`
	ResolvedByAccountID uint   `gorm:"null"`
}
