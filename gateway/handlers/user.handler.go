package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/user_service"
	"net/http"
	"time"
)

func HandlerCheckDuplicate(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CheckDuplicateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid payload request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var isDuplicated bool = false

		switch request.DataType {
		case "username":
			userServiceResp, err := userClient.CheckExistingUsername(
				ctx, &user_service.CheckExistingUsernameRequest{
					Username: request.Data,
				})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "CheckExistingUsername failed", err)
				return
			}
			isDuplicated = userServiceResp.IsExisting
			break
		case "email":
			userServiceResp, err := userClient.CheckExistingEmail(
				ctx, &user_service.CheckExistingEmailRequest{
					Email: request.Data,
				})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "CheckExistingEmail failed", err)
				return
			}
			isDuplicated = userServiceResp.IsExisting
			break
		case "phone":
			userServiceResp, err := userClient.CheckExistingPhone(
				ctx, &user_service.CheckExistingPhoneRequest{
					Phone: request.Data,
				})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "CheckExistingPhone failed", err)
				return
			}
			isDuplicated = userServiceResp.IsExisting
			break
		default:
			respondWithError(w, http.StatusInternalServerError, "Invalid request type", nil)
			return
		}

		response := &CheckDuplicateResponse{
			IsDuplicate: isDuplicated,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerCheckValidUser(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CheckValidUserRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid payload request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userResp, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.AccountID,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "CheckValidUser failed", err)
			return
		}

		var response = &CheckValidUserResponse{
			IsValid: userResp.IsValid,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}

	}
}

func HandlerGetAccountInfo(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetAccountInfoRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userResp, err := userClient.GetAccountInfo(ctx, &user_service.GetAccountInfoRequest{
			AccountID: uint32(request.AccountID),
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to get account info", err)
			return
		}

		var accountData = &Account{
			Username:      userResp.Account.Username,
			RoleID:        uint64(userResp.Account.RoleID),
			CreateMethod:  userResp.Account.CreateMethod,
			IsBanned:      userResp.Account.IsBanned,
			IsSelfDeleted: userResp.Account.IsSelfDeleted,
			IsRestricted:  userResp.Account.IsRestricted,
		}

		var accountInfo = &AccountInfo{
			FirstName:       userResp.AccountInfo.FirstName,
			LastName:        userResp.AccountInfo.LastName,
			DateOfBirth:     userResp.AccountInfo.DateOfBirth,
			Gender:          userResp.AccountInfo.Gender,
			MaterialStatus:  userResp.AccountInfo.MaterialStatus,
			PhoneNumber:     userResp.AccountInfo.PhoneNumber,
			Email:           userResp.AccountInfo.Email,
			NameDisplayType: userResp.AccountInfo.NameDisplayType,
			Bio:             userResp.AccountInfo.Bio,
		}

		var accountAvatar = &AccountAvatar{
			AvatarID:  uint64(userResp.AccountAvatar.ID),
			AvatarURL: userResp.AccountAvatar.AvatarURL,
			IsInUse:   userResp.AccountAvatar.IsInUse,
			IsDeleted: userResp.AccountAvatar.IsDeleted,
		}

		var response = &GetAccountInfoResponse{
			AccountID:     uint64(userResp.AccountID),
			Account:       *accountData,
			AccountInfo:   *accountInfo,
			AccountAvatar: *accountAvatar,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
			return
		}

	}
}
