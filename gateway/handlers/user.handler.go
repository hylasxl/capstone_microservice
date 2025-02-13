package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"gateway/proto/auth_service"
	"gateway/proto/friend_service"
	"gateway/proto/privacy_service"
	"gateway/proto/user_service"
	"log"
	"net/http"
	"strconv"
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
func HandlerGetProfileInfo(userClient user_service.UserServiceClient, friendClient friend_service.FriendServiceClient, privacyClient privacy_service.PrivacyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetProfileInfoRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		isSelf := request.RequestAccountID == request.TargetAccountID

		// Check Request Account Validity
		checkValidRequestAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.RequestAccountID, 10),
		})
		if err != nil || checkValidRequestAccountID == nil || !checkValidRequestAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "CheckValidUser failed", err)
			return
		}

		// Check Target Account Validity
		checkValidTargetAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.TargetAccountID, 10),
		})
		if err != nil || checkValidTargetAccountID == nil || !checkValidTargetAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "CheckValidUser failed", err)
			return
		}

		// Check Friend Relationship
		relationResp, err := friendClient.CheckIsFriend(ctx, &friend_service.CheckIsFriendRequest{
			FirstAccountID:  request.RequestAccountID,
			SecondAccountID: request.TargetAccountID,
		})
		if err != nil || relationResp == nil {
			respondWithError(w, http.StatusInternalServerError, "CheckIsFriend failed", err)
			return
		}

		// Check Block Relationship
		blockResponse, err := friendClient.CheckIsBlock(ctx, &friend_service.CheckIsBlockedRequest{
			FirstAccountID:  request.RequestAccountID,
			SecondAccountID: request.TargetAccountID,
		})
		if err != nil || blockResponse == nil {
			respondWithError(w, http.StatusInternalServerError, "CheckIsBlock failed or returned nil", err)
			return
		}

		if blockResponse.IsBlocked && !isSelf {
			respondWithError(w, http.StatusForbidden, "User Blocked", nil)
			return
		}

		// Check Follow Status
		followResp, err := friendClient.CheckIsFollow(ctx, &friend_service.CheckIsFollowRequest{
			FromAccountID: uint32(request.RequestAccountID),
			ToAccountID:   uint32(request.TargetAccountID),
		})
		if err != nil || followResp == nil {
			respondWithError(w, http.StatusInternalServerError, "CheckIsFollow failed", err)
			return
		}

		// Fetch Privacy Information
		privacyResp, err := privacyClient.GetPrivacy(ctx, &privacy_service.GetPrivacyRequest{
			AccountID: request.TargetAccountID,
		})
		if err != nil || privacyResp == nil || privacyResp.Privacy == nil {
			respondWithError(w, http.StatusInternalServerError, "GetPrivacy failed or returned nil", err)
			return
		}

		// Fetch Profile Information
		getInfoResp, err := userClient.GetProfileInfo(ctx, &user_service.GetProfileInfoRequest{
			RequestAccountID: uint32(request.RequestAccountID),
			TargetAccountID:  uint32(request.TargetAccountID),
			IsBlocked:        blockResponse.IsBlocked,
			IsFriend:         relationResp.IsFriend,
			IsFollow:         followResp.IsFollow,
			Privacy: &user_service.PrivacyIndices{
				DateOfBirth:    privacyResp.Privacy.DateOfBirth,
				Gender:         privacyResp.Privacy.Gender,
				MaterialStatus: privacyResp.Privacy.MaterialStatus,
				Phone:          privacyResp.Privacy.Phone,
				Email:          privacyResp.Privacy.Email,
				Bio:            privacyResp.Privacy.Bio,
			},
		})
		if err != nil || getInfoResp == nil || getInfoResp.Account == nil || getInfoResp.AccountInfo == nil || getInfoResp.AccountAvatar == nil {
			respondWithError(w, http.StatusInternalServerError, "GetProfileInfo failed or returned nil", err)
			return
		}

		// Construct Response
		response := &GetProfileInfoResponse{
			AccountID: request.TargetAccountID,
			Account: Account{
				Username:      getInfoResp.Account.Username,
				CreateMethod:  getInfoResp.Account.CreateMethod,
				IsBanned:      getInfoResp.Account.IsBanned,
				IsSelfDeleted: getInfoResp.Account.IsSelfDeleted,
				IsRestricted:  getInfoResp.Account.IsRestricted,
				RoleID:        uint64(getInfoResp.Account.RoleID),
			},
			AccountInfo: AccountInfo{
				FirstName:       getInfoResp.AccountInfo.FirstName,
				LastName:        getInfoResp.AccountInfo.LastName,
				DateOfBirth:     getInfoResp.AccountInfo.DateOfBirth,
				Gender:          getInfoResp.AccountInfo.Gender,
				MaterialStatus:  getInfoResp.AccountInfo.MaterialStatus,
				PhoneNumber:     getInfoResp.AccountInfo.PhoneNumber,
				Email:           getInfoResp.AccountInfo.Email,
				NameDisplayType: getInfoResp.AccountInfo.NameDisplayType,
				Bio:             getInfoResp.AccountInfo.Bio,
			},
			AccountAvatar: AccountAvatar{
				AvatarID:  uint64(getInfoResp.AccountAvatar.ID),
				AvatarURL: getInfoResp.AccountAvatar.AvatarURL,
				IsInUse:   getInfoResp.AccountAvatar.IsInUse,
				IsDeleted: getInfoResp.AccountAvatar.IsDeleted,
			},
			PrivacyIndices: PrivacyIndices{
				DateOfBirth:    getInfoResp.Privacy.DateOfBirth,
				Gender:         getInfoResp.Privacy.Gender,
				MaterialStatus: getInfoResp.Privacy.MaterialStatus,
				PhoneNumber:    getInfoResp.Privacy.Phone,
				Email:          getInfoResp.Privacy.Email,
				Bio:            getInfoResp.Privacy.Bio,
			},
			IsFriend:   relationResp.IsFriend,
			IsBlocked:  blockResponse.IsBlocked,
			IsFollowed: followResp.IsFollow,
			Timestamp:  getInfoResp.Timestamp,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandlerChangeUserInfo(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ChangeAccountInfoRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Validate User ID
		checkValidUser, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil || checkValidUser == nil || !checkValidUser.IsValid {
			respondWithError(w, http.StatusInternalServerError, "CheckValidUser failed", err)
			return
		}

		updateResp, err := userClient.ChangeAccountInfo(ctx, &user_service.ChangeAccountDataRequest{
			AccountID:     request.AccountID,
			DataFieldName: request.DataFieldName,
			Data:          request.Data,
		})
		if err != nil || updateResp == nil || !updateResp.Success {
			respondWithError(w, http.StatusInternalServerError, "ChangeUserInfo failed", err)
			return
		}

		response := &SingleSuccessResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandlerChangeAvatar(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ChangeAvatarRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var imgBytes []byte
		if request.Avatar != "" {
			var err error
			imgBytes, err = base64.StdEncoding.DecodeString(request.Avatar)
			if err != nil {
				// Log the error
				log.Printf("Error decoding image: %v", err)
				respondWithError(w, http.StatusBadRequest, "Invalid image file", err)
				return
			}
		}

		updateResp, err := userClient.ChangeAvatar(ctx, &user_service.ChangeAvatarRequest{
			AccountID: request.AccountID,
			Avatar:    imgBytes,
		})

		if err != nil || updateResp == nil {
			respondWithError(w, http.StatusInternalServerError, "ChangeAvatar failed", nil)
			return
		}

		response := &SingleSuccessResponse{
			Success: true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandlerLoginWithGoogle(userClient user_service.UserServiceClient, authClient auth_service.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request LoginWithGoogleRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		loginResp, err := userClient.LoginWithGoogle(ctx, &user_service.LoginWithGoogleRequest{
			Email:       request.Email,
			PhotoURL:    request.PhotoURL,
			AuthCode:    request.AuthCode,
			DisplayName: request.DisplayName,
		})

		if err != nil || loginResp == nil {
			respondWithError(w, http.StatusInternalServerError, "LoginWithGoogle failed", nil)
			return
		}

		permissionResp, err := authClient.GetPermissions(ctx, &auth_service.GetPermissionsRequest{
			RoleId: "1",
		})

		var claims = &auth_service.JWTClaims{
			AccountId:   strconv.FormatUint(loginResp.AccountID, 10),
			Permissions: permissionResp.Url,
			RoleId:      "1",
			Issuer:      "SyncIO",
			Subject:     "Authentication",
			Audience:    "Client SyncIO",
		}

		tokenRes, err := authClient.GenerateTokens(ctx, &auth_service.GenerateTokensRequest{
			Claims: claims,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var response = &LoginWithGoogleResponse{
			AccessToken:  tokenRes.AccessToken,
			RefreshToken: tokenRes.RefreshToken,
			UserID:       strconv.FormatUint(loginResp.AccountID, 10),
			JWTClaims:    claims,
			Success:      true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandlerVerifyUsernameAndEmail(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request VerifyUsernameAndEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		verifyResp, err := userClient.VerifyUsernameAndEmail(ctx, &user_service.VerifyUsernameAndEmailRequest{
			Username: request.Username,
			Email:    request.Email,
		})
		if err != nil || verifyResp == nil {
			respondWithError(w, http.StatusInternalServerError, "VerifyUsernameAndEmail failed", nil)
			return
		}

		var response = &VerifyUsernameAndEmailResponse{
			Success: verifyResp.Success,
			UserID:  uint64(verifyResp.UserID),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandlerChangePassword(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := userClient.ChangePassword(ctx, &user_service.ChangePasswordRequest{
			AccountID:   request.AccountID,
			NewPassword: request.NewPassword,
		})

		if err != nil || resp == nil {
			respondWithError(w, http.StatusInternalServerError, "ChangePassword failed", nil)
			return
		}

		var response = &ChangePasswordResponse{
			Success: resp.Success,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandlerSearchAccounts(userClient user_service.UserServiceClient, friendClient friend_service.FriendServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var in SearchAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		blockList, err := friendClient.GetBlockList(ctx, &friend_service.GetBlockListRequest{
			AccountID: uint32(in.RequestAccountID),
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "GetBlockList failed", err)
			return
		}

		blockedList, err := friendClient.GetBlockedList(ctx, &friend_service.GetBlockedListRequest{
			AccountID: uint32(in.RequestAccountID),
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "GetBlockList failed", err)
			return
		}

		resp, err := userClient.SearchAccount(ctx, &user_service.SearchAccountRequest{
			RequestAccountID: uint32(in.RequestAccountID),
			BlockedList:      blockedList.IDs,
			BlockList:        blockList.IDs,
			Page:             in.Page,
			PageSize:         in.PageSize,
			QueryString:      in.QueryString,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "SearchAccount failed", err)
			return
		}

		var accounts []SingleAccountInfo

		for _, account := range resp.Account {
			accounts = append(accounts, SingleAccountInfo{
				AccountID:   uint(account.AccountID),
				AvatarURL:   account.AvatarURL,
				DisplayName: account.DisplayName,
			})
		}

		var response = &SearchAccountResponse{
			Page:     in.Page,
			PageSize: in.PageSize,
			Accounts: accounts,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerGetNewRegisterationData(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var request GetNewRegisterationDataRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		periodData := request.PeriodYear

		if request.PeriodLabel == "month" {
			periodData = request.PeriodYear*100 + (request.PeriodMonth % 100)
		}

		data, err := userClient.GetNewRegisterationData(ctx, &user_service.GetNewRegisterationDataRequest{
			RequestAccountID: uint32(request.RequestAccountID),
			PeriodLabel:      request.PeriodLabel,
			PeriodData:       periodData,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "GetNewRegisterationData failed", err)
			return
		}

		var dataTerms []DataTerms

		for _, index := range data.Data {
			dataTerms = append(dataTerms, DataTerms{
				Label: index.Label,
				Count: index.Count,
			})
		}

		var response = &GetNewRegisterationDataResponse{
			RequestAccountID: uint64(data.RequestAccountID),
			PeriodLabel:      data.PeriodLabel,
			TotalUsers:       uint32(data.TotalUsers),
			Data:             dataTerms,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerGetUserType(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetUserTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := userClient.CountUserType(ctx, &user_service.CountTypeUserRequest{
			RequestAccountID: uint32(request.RequestAccountID),
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "CountUserType failed", err)
			return
		}

		var response = &GetUserTypeResponse{
			RequestAccountID: uint64(request.RequestAccountID),
			TotalUsers:       uint32(resp.TotalUsers),
			BannedUsers:      uint32(resp.BannedUsers),
			DeletedUsers:     uint32(resp.DeletedUsers),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerGetAccountList(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in GetAccountListRequest

		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := userClient.GetAccountList(ctx, &user_service.GetAccountListRequest{
			RequestID: uint32(in.RequestID),
			Page:      in.Page,
			PageSize:  in.PageSize,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "GetAccountList failed", err)
			return
		}

		var response = &GetAccountListResponse{
			Page:     in.Page,
			PageSize: in.PageSize,
		}

		for _, account := range resp.Accounts {
			response.Accounts = append(response.Accounts, AccountRawInfo{
				AccountID:     uint32(account.AccountID),
				Username:      account.Username,
				IsBanned:      account.IsBanned,
				IsSelfDeleted: account.IsSelfDeleted,
				Method:        account.Method,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerSearchAccountList(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in SearchAccountListRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return

		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := userClient.SearchAccountList(ctx, &user_service.SearchAccountListRequest{
			RequestID:   uint32(in.RequestID),
			Page:        in.Page,
			PageSize:    in.PageSize,
			QueryString: in.QueryString,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "SearchAccountList failed", err)
			return
		}

		var response = &SearchAccountListResponse{
			Page:     in.Page,
			PageSize: in.PageSize,
		}

		for _, account := range resp.Accounts {
			response.Accounts = append(response.Accounts, AccountRawInfo{
				AccountID:     uint32(account.AccountID),
				Username:      account.Username,
				IsBanned:      account.IsBanned,
				IsSelfDeleted: account.IsSelfDeleted,
				Method:        account.Method,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerBanUser(userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ResolveBanUserRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := userClient.ResolveBan(ctx, &user_service.ResolveBanRequest{
			AccountID: uint32(request.AccountID),
			Action:    request.Action,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "ResolveBan failed", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
