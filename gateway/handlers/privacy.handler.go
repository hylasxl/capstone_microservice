package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/privacy_service"
	"gateway/proto/user_service"
	"net/http"
	"strconv"
	"time"
)

func HandlerSetPrivacy(privacyClient privacy_service.PrivacyServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request SetPrivacyRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid payload", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		checkValidUser, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})

		if err != nil || !checkValidUser.IsValid {
			respondWithError(w, http.StatusBadRequest, "invalid account ID", nil)
			return
		}

		setPrivacyResp, err := privacyClient.SetPrivacy(ctx, &privacy_service.SetPrivacyRequest{
			AccountID:     request.AccountID,
			PrivacyIndex:  uint32(request.PrivacyIndex),
			PrivacyStatus: request.PrivacyStatus,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to set privacy", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(&SingleSuccessResponse{Success: setPrivacyResp.Success}); err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to encode success", err)
		}

	}
}
