package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/notification_service"
	"log"
	"net/http"
	"time"
)

func HandlerRegisterDevice(notificationClient notification_service.NotificationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request RegisterDeviceRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid payload", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		registerResp, err := notificationClient.RegisterDevice(ctx, &notification_service.RegisterDeviceRequest{
			UserID:   uint32(request.UserID),
			FCMToken: request.Token,
		})

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "failed to register device", err)
		}

		var response = &RegisterDeviceResponse{
			Success: registerResp.Success,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerGetNotifications(notificationClient notification_service.NotificationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetNotificationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid payload", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		getNotifiResp, err := notificationClient.GetNotification(ctx, &notification_service.GetNotificationRequest{
			AccountID: request.AccountID,
			Page:      uint32(request.Page),
			PageSize:  uint32(request.PageSize),
		})

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "failed to get notifications", err)
			return
		}

		var noti []NotificationContent

		for _, notification := range getNotifiResp.Notifications {
			noti = append(noti, NotificationContent{
				ID:       notification.ID,
				Content:  notification.Content,
				DateTime: notification.DateTime,
				IsRead:   notification.IsRead,
			})
		}

		var response = &GetNotificationResponse{
			AccountID:     request.AccountID,
			Page:          request.Page,
			PageSize:      request.PageSize,
			Notifications: noti,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerCountUnreadNotifications(notificationClient notification_service.NotificationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CountUnreadNotiRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid payload", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		countResp, err := notificationClient.CountUnReadNoti(ctx, &notification_service.CountUnreadNotiRequest{
			AccountID: uint32(request.AccountID),
		})

		if err != nil || countResp == nil {
			respondWithError(w, http.StatusInternalServerError, "failed to count unread notifications", nil)
			return
		}

		var response = &CountUnreadNotiResponse{
			Quantity: uint64(countResp.Quantity),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerMarkAsReadNoti(notificationClient notification_service.NotificationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request MarkAsReadNotiRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid payload", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		markResp, err := notificationClient.MarkAsReadNoti(ctx, &notification_service.MarkAsReadNotiRequest{
			AccountID: uint32(request.AccountID),
		})

		if err != nil || markResp == nil {
			respondWithError(w, http.StatusInternalServerError, "failed to mark notification as read", nil)
		}
		var response = &MarkAsReadNotiResponse{
			Quantity: uint64(markResp.Quantity),
			Success:  true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
