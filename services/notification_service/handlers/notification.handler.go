package handlers

import (
	"context"
	"errors"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"gorm.io/gorm"
	"log"
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

func (s *NotificationService) ReceiveFriendRequestNotificationFnc(ctx context.Context, in *ns.ReceiveFriendRequestNotification) (*ns.SingleMessageSentResponse, error) {

	if in.ReceiverAccountID <= 0 || in.SenderAccountID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var token = &models.Device{}

	if err := s.DB.Where("user_id = ?", in.ReceiverAccountID).Order("created_at desc").First(&token).Error; err != nil {
		return nil, err
	}
	if token.Token == "" {
		return nil, errors.New("invalid token")
	}

	client, err := s.FirebaseApp.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	message := &messaging.Message{
		Token: token.Token,
		Notification: &messaging.Notification{
			Title: "New Friend Request",
			Body:  "You have a new friend request from " + in.SenderAccountDisplayName,
		},
		Data: map[string]string{
			"type": "friend_request",
		},
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		log.Printf("failed to send FCM notification: %v", err)
		return nil, err
	}

	var notification = &models.Notification{
		AccountID: uint(in.ReceiverAccountID),
		TypeID:    3,
		IsRead:    false,
		Message:   "You have a new friend request from " + in.SenderAccountDisplayName,
	}

	tx := s.DB.Begin()
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}

	tx.Commit()

	return &ns.SingleMessageSentResponse{
		Success: true,
	}, nil
}

func (s *NotificationService) CommentNotification(ctx context.Context, in *ns.CommentNotificationRequest) (*ns.SingleMessageSentResponse, error) {

	if in.ReceiverAccountID <= 0 || in.SenderAccountID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var token = &models.Device{}

	if err := s.DB.Where("user_id = ?", in.ReceiverAccountID).Order("created_at desc").First(&token).Error; err != nil {
		return nil, err
	}
	if token.Token == "" {
		return nil, errors.New("invalid token")
	}

	client, err := s.FirebaseApp.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	message := &messaging.Message{
		Token: token.Token,
		Notification: &messaging.Notification{
			Title: "Comment",
			Body:  in.SenderAccountDisplayName + " has just commented on your post",
		},
		Data: map[string]string{
			"type": "comment",
		},
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		log.Printf("failed to send FCM notification: %v", err)
		return nil, err
	}

	var notification = &models.Notification{
		AccountID: uint(in.ReceiverAccountID),
		TypeID:    1,
		IsRead:    false,
		Message:   in.SenderAccountDisplayName + " has just commented on your post",
	}

	tx := s.DB.Begin()
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}

	tx.Commit()

	return &ns.SingleMessageSentResponse{
		Success: true,
	}, nil
}

func (s *NotificationService) ReplyCommentNotification(ctx context.Context, in *ns.ReplyCommentNotificationRequest) (*ns.SingleMessageSentResponse, error) {
	if in.ReceiverAccountID <= 0 || in.SenderAccountID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var token = &models.Device{}
	if err := s.DB.Where("user_id = ?", in.ReceiverAccountID).Order("created_at desc").First(&token).Error; err != nil {
		return nil, err
	}

	if strings.TrimSpace(token.Token) == "" {
		return nil, errors.New("invalid token")
	}

	client, err := s.FirebaseApp.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	message := &messaging.Message{
		Token: token.Token,
		Notification: &messaging.Notification{
			Title: "Reply Comment",
			Body:  in.SenderAccountDisplayName + " replied to your comment.",
		},
		Data: map[string]string{
			"type": "reply_comment",
		},
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		log.Printf("failed to send FCM notification: %v", err)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}

	var notification = &models.Notification{
		AccountID: uint(in.ReceiverAccountID),
		TypeID:    2,
		IsRead:    false,
		Message:   in.SenderAccountDisplayName + " replied to your comment.",
	}

	tx := s.DB.Begin()
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}
	tx.Commit()

	return &ns.SingleMessageSentResponse{
		Success: true,
	}, nil
}

func (s *NotificationService) ReactPostNotification(ctx context.Context, in *ns.ReactPostNotificationRequest) (*ns.SingleMessageSentResponse, error) {
	if in.ReceiverAccountID <= 0 || in.SenderAccountID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var token = &models.Device{}
	if err := s.DB.Where("user_id = ?", in.ReceiverAccountID).Order("created_at desc").First(&token).Error; err != nil {
		return nil, err
	}

	if strings.TrimSpace(token.Token) == "" {
		return nil, errors.New("invalid token")
	}

	client, err := s.FirebaseApp.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	message := &messaging.Message{
		Token: token.Token,
		Notification: &messaging.Notification{
			Title: "New Reaction",
			Body:  in.SenderAccountDisplayName + " reacted to your post.",
		},
		Data: map[string]string{
			"type": "react_post",
		},
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		log.Printf("failed to send FCM notification: %v", err)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}

	var notification = &models.Notification{
		AccountID: uint(in.ReceiverAccountID),
		TypeID:    4,
		IsRead:    false,
		Message:   in.SenderAccountDisplayName + " reacted to your post.",
	}

	tx := s.DB.Begin()
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}
	tx.Commit()

	return &ns.SingleMessageSentResponse{
		Success: true,
	}, nil
}

func (s *NotificationService) SharePostNotification(ctx context.Context, in *ns.SharePostNotificationRequest) (*ns.SingleMessageSentResponse, error) {
	if in.ReceiverAccountID <= 0 || in.SenderAccountID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var token = &models.Device{}
	if err := s.DB.Where("user_id = ?", in.ReceiverAccountID).Order("created_at desc").First(&token).Error; err != nil {
		return nil, err
	}

	if strings.TrimSpace(token.Token) == "" {
		return nil, errors.New("invalid token")
	}

	client, err := s.FirebaseApp.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	message := &messaging.Message{
		Token: token.Token,
		Notification: &messaging.Notification{
			Title: "Post Shared",
			Body:  in.SenderAccountDisplayName + " shared your post.",
		},
		Data: map[string]string{
			"type": "share_post",
		},
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		log.Printf("failed to send FCM notification: %v", err)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}

	var notification = &models.Notification{
		AccountID: uint(in.ReceiverAccountID),
		TypeID:    5, // ID for "Share Post" notification
		IsRead:    false,
		Message:   in.SenderAccountDisplayName + " shared your post.",
	}

	tx := s.DB.Begin()
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}
	tx.Commit()

	return &ns.SingleMessageSentResponse{
		Success: true,
	}, nil
}

func (s *NotificationService) FollowNotification(ctx context.Context, in *ns.FollowNotificationRequest) (*ns.SingleMessageSentResponse, error) {
	if in.ReceiverAccountID <= 0 || in.SenderAccountID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var token = &models.Device{}
	if err := s.DB.Where("user_id = ?", in.ReceiverAccountID).Order("created_at desc").First(&token).Error; err != nil {
		return nil, err
	}

	if strings.TrimSpace(token.Token) == "" {
		return nil, errors.New("invalid token")
	}

	client, err := s.FirebaseApp.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	message := &messaging.Message{
		Token: token.Token,
		Notification: &messaging.Notification{
			Title: "New Follower",
			Body:  in.SenderAccountDisplayName + " started following you.",
		},
		Data: map[string]string{
			"type": "follow",
		},
	}

	_, err = client.Send(ctx, message)
	if err != nil {
		log.Printf("failed to send FCM notification: %v", err)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}

	var notification = &models.Notification{
		AccountID: uint(in.ReceiverAccountID),
		TypeID:    7, // ID for "Follow" notification
		IsRead:    false,
		Message:   in.SenderAccountDisplayName + " started following you.",
	}

	tx := s.DB.Begin()
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}
	tx.Commit()

	return &ns.SingleMessageSentResponse{
		Success: true,
	}, nil
}

func (s *NotificationService) MessageNotification(ctx context.Context, in *ns.MessageNotificationRequest) (*ns.SingleMessageSentResponse, error) {
	// Validate input
	if in.SenderAccountID <= 0 || in.ReceiverAccountID <= 0 {
		return nil, errors.New("invalid sender or receiver account ID")
	}

	// Fetch the receiver's device token from the database
	var token = &models.Device{}
	if err := s.DB.Where("user_id = ?", in.ReceiverAccountID).Order("created_at desc").First(&token).Error; err != nil {
		log.Printf("receiver %d not found or token missing: %v", in.ReceiverAccountID, err)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, nil
	}

	if strings.TrimSpace(token.Token) == "" {
		log.Printf("token missing for receiver %d", in.ReceiverAccountID)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, nil
	}

	// Create the Firebase Cloud Messaging (FCM) payload
	client, err := s.FirebaseApp.Messaging(ctx)
	if err != nil {
		log.Printf("failed to create Firebase Messaging client: %v", err)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}

	message := &messaging.Message{
		Token: token.Token,
		Notification: &messaging.Notification{
			Title: "New Message",
			Body:  in.SenderAccountDisplayName + " sent you a message.",
		},
		Data: map[string]string{
			"type": "message",
		},
	}

	// Send the message through Firebase
	_, err = client.Send(ctx, message)
	if err != nil {
		log.Printf("failed to send FCM notification to receiver %d: %v", in.ReceiverAccountID, err)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}

	// Save the notification in the database
	var notification = &models.Notification{
		AccountID: uint(in.ReceiverAccountID),
		TypeID:    9, // Type ID for "Message" notification
		IsRead:    false,
		Message:   in.SenderAccountDisplayName + " sent you a message.",
	}

	tx := s.DB.Begin()
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		log.Printf("failed to save notification for receiver %d: %v", in.ReceiverAccountID, err)
		return &ns.SingleMessageSentResponse{
			Success: false,
		}, err
	}
	tx.Commit()

	// Return success response
	return &ns.SingleMessageSentResponse{
		Success: true,
	}, nil
}

func (s *NotificationService) GetNotification(ctx context.Context, in *ns.GetNotificationRequest) (*ns.GetNotificationResponse, error) {
	if in.AccountID <= 0 {
		return nil, errors.New("invalid account ID")
	}

	if in.Page < 1 {
		in.Page = 1
	}

	if in.PageSize < 1 {
		in.PageSize = 20
	}

	var notifications []*models.Notification

	// Calculate the offset for pagination
	offset := (in.Page - 1) * in.PageSize

	if err := s.DB.Where("account_id = ?", in.AccountID).
		Order("created_at desc").
		Offset(int(offset)).
		Limit(int(in.PageSize)).
		Find(&notifications).Error; err != nil {
		return nil, err
	}

	var response []*ns.NotificationContent

	for _, notification := range notifications {
		response = append(response, &ns.NotificationContent{
			ID:       uint64(notification.ID),
			Content:  notification.Message,
			DateTime: notification.CreatedAt.Unix(),
			IsRead:   notification.IsRead,
		})
	}

	return &ns.GetNotificationResponse{
		Account:       in.AccountID,
		Notifications: response,
		Page:          in.Page,
		PageSize:      in.PageSize,
	}, nil
}

func (s *NotificationService) MarkAsReadNoti(ctx context.Context, req *ns.MarkAsReadNotiRequest) (*ns.MarkAsReadNotiResponse, error) {

	if req.AccountID <= 0 {
		return nil, errors.New("invalid account ID")
	}

	tx := s.DB.Begin()
	result := tx.Model(&models.Notification{}).Where("account_id = ?", req.AccountID).Update("is_read", true)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	rowsAffected := result.RowsAffected

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &ns.MarkAsReadNotiResponse{
		Success:  true,
		Quantity: uint32(rowsAffected),
	}, nil
}

func (s *NotificationService) CountUnReadNoti(ctx context.Context, req *ns.CountUnreadNotiRequest) (*ns.CountUnreadNotiResponse, error) {
	if req.AccountID <= 0 {
		return nil, errors.New("invalid account ID")
	}

	var count int64
	if err := s.DB.Model(&models.Notification{}).
		Where("account_id = ? AND is_read = ?", req.AccountID, false).
		Count(&count).Error; err != nil {
		return nil, err
	}

	return &ns.CountUnreadNotiResponse{
		Quantity: uint32(count),
	}, nil
}
