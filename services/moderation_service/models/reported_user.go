package models

import "gorm.io/gorm"

type ReportedUser struct {
	gorm.Model
	AccountID           uint   `gorm:"not null"`
	ReportedByAccountID uint   `gorm:"not null"`
	Reason              string `gorm:"type:text;not null"`
	ReportResolve       string `gorm:"type:ENUM('report_pending','report_skipped','delete_user');default:'report_pending'"`
	ResolvedByAccountID uint   `gorm:"not null"`
}
