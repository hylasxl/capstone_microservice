package handlers

import (
	"context"
	"errors"
	firebase "firebase.google.com/go/v4"
	"gorm.io/gorm"
	"notification_service/models"
	ns "notification_service/proto/notification_service"
	"strings"
)

type NotificationService struct {
	ns.UnimplementedNotificationServiceServer
	DB          *gorm.DB
	FirebaseApp *firebase.App
}

func (s *NotificationService) RegisterDevice(ctx context.Context, in *ns.RegisterDeviceRequest) (*ns.RegisterDeviceResponse, error) {
	if in.UserID <= 0 {
		return nil, errors.New("invalid user id")
	}
	if len(strings.TrimSpace(in.FCMToken)) == 0 {
		return nil, errors.New("invalid FCM token")
	}

	var device = &models.Device{
		UserID: uint(in.UserID),
		Token:  in.FCMToken,
	}
	tx := s.DB.Begin()

	if err := tx.Create(&device).Error; err != nil {
		tx.Rollback()
		return &ns.RegisterDeviceResponse{
			Success: false,
		}, err
	}

	tx.Commit()

	return &ns.RegisterDeviceResponse{
		Success: true,
	}, nil

}
