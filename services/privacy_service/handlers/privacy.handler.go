package handlers

import (
	"context"
	"gorm.io/gorm"
	"privacy_service/models"
	"privacy_service/proto/privacy_service"
	"strconv"
)

type PrivacyService struct {
	privacy_service.UnimplementedPrivacyServiceServer
	DB *gorm.DB
}

func (svc *PrivacyService) CreateAccountPrivacyInit(ctx context.Context, in *privacy_service.CreateAccountPrivacyInitRequest) (*privacy_service.CreateAccountPrivacyInitResponse, error) {
	tx := svc.DB.Begin()
	accountIDUint, err := strconv.ParseUint(in.AccountID, 10, 32)
	if err != nil {
		tx.Rollback()
		return &privacy_service.CreateAccountPrivacyInitResponse{
			Error: "Invalid AccountID format",
		}, nil
	}
	var listItems []models.DataPrivacy
	for i := 1; i <= 5; i++ {
		listItems = append(listItems, models.DataPrivacy{
			AccountID:      uint(accountIDUint),
			DataFieldIndex: uint(i),
		})
	}

	if err := tx.Create(&listItems).Error; err != nil {
		tx.Rollback()
		return &privacy_service.CreateAccountPrivacyInitResponse{
			Error: "Failed to initialize account privacy settings",
		}, nil
	}

	tx.Commit()
	return &privacy_service.CreateAccountPrivacyInitResponse{}, nil
}
