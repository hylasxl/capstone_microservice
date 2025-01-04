package handlers

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"privacy_service/models"
	"privacy_service/proto/privacy_service"
	"strconv"
	"strings"
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
	for i := 1; i <= 6; i++ {
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
func (svc *PrivacyService) GetPrivacy(ctx context.Context, in *privacy_service.GetPrivacyRequest) (*privacy_service.GetPrivacyResponse, error) {
	if in.AccountID <= 0 {
		return nil, errors.New("invalid AccountID")
	}

	var privacyArray = make([]*models.DataPrivacy, 6)

	for i := 1; i <= 6; i++ {
		var privacyRecord models.DataPrivacy
		err := svc.DB.Model(&models.DataPrivacy{}).
			Where("account_id = ? AND data_field_index = ?", in.AccountID, i).
			First(&privacyRecord).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				privacyArray[i-1] = &models.DataPrivacy{
					AccountID:      uint(in.AccountID),
					DataFieldIndex: uint(i),
					PrivacyStatus:  "public",
				}
			} else {
				return nil, err
			}
		} else {
			privacyArray[i-1] = &privacyRecord
		}
	}

	response := &privacy_service.GetPrivacyResponse{
		Privacy: &privacy_service.PrivacyIndices{
			DateOfBirth:    privacyArray[0].PrivacyStatus,
			Gender:         privacyArray[1].PrivacyStatus,
			MaterialStatus: privacyArray[2].PrivacyStatus,
			Phone:          privacyArray[3].PrivacyStatus,
			Email:          privacyArray[4].PrivacyStatus,
			Bio:            privacyArray[5].PrivacyStatus,
		},
	}

	return response, nil
}

func (svc *PrivacyService) SetPrivacy(ctx context.Context, in *privacy_service.SetPrivacyRequest) (*privacy_service.SetPrivacyResponse, error) {
	if in.AccountID <= 0 {
		return nil, errors.New("invalid AccountID")
	}

	if in.PrivacyIndex <= 0 || len(strings.TrimSpace(in.PrivacyStatus)) == 0 {
		return nil, errors.New("invalid PrivacyData")
	}

	tx := svc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var existingPrivacyRecord *models.DataPrivacy
	if err := tx.Model(&models.DataPrivacy{}).
		Where("account_id = ? AND data_field_index = ?", in.AccountID, in.PrivacyIndex).
		First(&existingPrivacyRecord).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			data := &models.DataPrivacy{
				AccountID:      uint(in.AccountID),
				DataFieldIndex: uint(in.PrivacyIndex),
				PrivacyStatus:  in.PrivacyStatus,
			}
			if err := tx.Model(&models.DataPrivacy{}).Create(&data).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			tx.Rollback()
			return nil, err
		}
	} else {
		existingPrivacyRecord.PrivacyStatus = in.PrivacyStatus
		if err := tx.Save(&existingPrivacyRecord).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &privacy_service.SetPrivacyResponse{
		Success: true,
	}, nil
}
