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
