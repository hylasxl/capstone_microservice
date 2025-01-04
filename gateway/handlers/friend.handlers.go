package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/friend_service"
	"gateway/proto/user_service"
	"net/http"
	"strconv"
	"time"
)

func HandlerSendFriendRequest(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request SendFriendRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkFirstAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.FromAccountID,
		})

		if err != nil || !checkFirstAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.FromAccountID, http.StatusUnauthorized)
			return
		}

		checkSecondAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.ToAccountID,
		})
		if err != nil || !checkSecondAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.ToAccountID, http.StatusUnauthorized)
			return
		}

		friendServiceResp, err := friendClient.SendFriend(ctx, &friend_service.SendFriendRequest{
			FromAccountID: request.FromAccountID,
			ToAccountID:   request.ToAccountID,
		})

		if err != nil {
			http.Error(w, "Send friend request error: "+friendServiceResp.Error, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(friendServiceResp); err != nil {
			http.Error(w, "Send friend response error: "+friendServiceResp.Error, http.StatusInternalServerError)
			return
		}
	}
}

func HandlerRecallRequest(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request RecallRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkSenderAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.SenderAccountID,
		})
		if err != nil || !checkSenderAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.SenderAccountID, http.StatusUnauthorized)
			return
		}

		senderID, err := strconv.ParseInt(request.SenderAccountID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid account ID "+request.SenderAccountID, http.StatusUnauthorized)
			return
		}

		requestID, err := strconv.ParseInt(request.RequestID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid request ID "+request.RequestID, http.StatusUnauthorized)
			return
		}

		recallResp, err := friendClient.RecallFriendRequest(ctx, &friend_service.RecallRequest{
			SenderID:  uint64(senderID),
			RequestID: uint64(requestID),
		})

		if err != nil {
			http.Error(w, "Recall friend request error: "+err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(recallResp); err != nil {
			http.Error(w, "Recall friend response error: "+recallResp.Error, http.StatusInternalServerError)
		}

	}
}

func HandlerResolveFriendRequest(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ResolveFriendRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkReceiverAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.ReceiverAccountID,
		})

		if err != nil || !checkReceiverAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.ReceiverAccountID, http.StatusUnauthorized)
			return
		}

		receiverID, err := strconv.ParseInt(request.ReceiverAccountID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid account ID "+request.ReceiverAccountID, http.StatusUnauthorized)
			return
		}

		requestID, err := strconv.ParseInt(request.RequestID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid request ID "+request.RequestID, http.StatusUnauthorized)
			return
		}

		_, err = friendClient.ResolveFriendRequest(ctx, &friend_service.FriendListResolveRequest{
			ReceiverID: uint64(receiverID),
			RequestID:  uint64(requestID),
			Action:     request.Action,
		})

		if err != nil {
			http.Error(w, "Resolve friend request error: "+err.Error(), http.StatusUnauthorized)
			return
		}

		response := &ResolveFriendResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Resolve friend response error: "+response.Error, http.StatusInternalServerError)
		}
	}
}

func HandlerUnfriend(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request UnfriendRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkFromAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.FromAccountID,
		})
		if err != nil || !checkFromAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.FromAccountID, http.StatusUnauthorized)
			return
		}

		checkToAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.ToAccountID,
		})
		if err != nil || !checkToAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.ToAccountID, http.StatusUnauthorized)
			return
		}

		unfriendResp, err := friendClient.Unfriend(ctx, &friend_service.UnfriendRequest{
			FromAccountID: request.FromAccountID,
			ToAccountID:   request.ToAccountID,
		})

		if err != nil {
			http.Error(w, "Unfriend request error: "+err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(unfriendResp); err != nil {
			http.Error(w, "Unfriend response error: "+unfriendResp.Error, http.StatusInternalServerError)
			return
		}
	}
}

func HandlerResolveFollow(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request FollowRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkFromAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.FromAccountID,
		})
		if err != nil || !checkFromAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.FromAccountID, http.StatusUnauthorized)
			return
		}

		checkToAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.ToAccountID,
		})
		if err != nil || !checkToAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.ToAccountID, http.StatusUnauthorized)
			return
		}

		followResp, err := friendClient.ResolveFriendFollow(ctx, &friend_service.FriendFollowResolveRequest{
			FromAccountID: request.FromAccountID,
			ToAccountID:   request.ToAccountID,
			Action:        request.Action,
		})

		if err != nil {
			http.Error(w, "Resolve friend request error: "+err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(followResp); err != nil {
			http.Error(w, "Resolve friend response error: "+followResp.Error, http.StatusInternalServerError)
			return
		}
	}
}

func HandlerResolveBlock(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request BlockRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkFromAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.FromAccountID,
		})
		if err != nil || !checkFromAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.FromAccountID, http.StatusUnauthorized)
			return
		}

		checkToAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.ToAccountID,
		})
		if err != nil || !checkToAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.ToAccountID, http.StatusUnauthorized)
			return
		}

		blockResp, err := friendClient.ResolveFriendBlock(ctx, &friend_service.FriendBlockResolveRequest{
			FromAccountID: request.FromAccountID,
			ToAccountID:   request.ToAccountID,
			Action:        request.Action,
		})

		if err != nil {
			http.Error(w, "Resolve friend request error: "+err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(blockResp); err != nil {
			http.Error(w, "Resolve friend response error: "+blockResp.Error, http.StatusInternalServerError)
			return
		}
	}
}

func HandlerGetPendingList(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetPendingListRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.AccountID,
		})
		if err != nil || !checkAccountID.IsValid {
			http.Error(w, "Invalid account ID "+request.AccountID, http.StatusUnauthorized)
			return
		}

		getPendingListResp, err := friendClient.GetPendingList(ctx, &friend_service.GetPendingListRequest{
			AccountID: request.AccountID,
			Page:      int64(request.Page),
		})
		if err != nil {
			http.Error(w, "Get pending list error: "+err.Error(), http.StatusUnauthorized)
			return
		}

		var listIDs = make([]uint64, len(getPendingListResp.GetListPending()))
		for i, listID := range getPendingListResp.ListPending {
			listIDs[i] = listID.AccountID
		}

		getAccountInfosResp, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
			IDs: listIDs,
		})

		var response GetPendingListResponse
		response.Page = int(getPendingListResp.Page)

		returnedData := make([]GetPendingListReturnSingleLine, len(listIDs))
		for i, data := range getAccountInfosResp.Infos {
			var accData SingleAccountInfo
			accData.AvatarURL = data.AvatarURL
			accData.DisplayName = data.DisplayName
			accData.AccountID = uint(data.AccountID)

			returnedData[i] = GetPendingListReturnSingleLine{
				AccountInfo:   accData,
				RequestID:     strconv.FormatUint(getPendingListResp.ListPending[i].RequestID, 10),
				CreatedAt:     getPendingListResp.ListPending[i].CreatedAt,
				MutualFriends: getPendingListResp.ListPending[i].MutualFriends,
			}
		}

		response.Data = returnedData

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Get pending list response error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandlerGetListFriend(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var request GetListFriendIDs
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.AccountID,
		})

		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusBadRequest, "Invalid account ID "+request.AccountID, err)
			return
		}

		getListFriendResp, err := friendClient.GetListFriend(ctx, &friend_service.GetListFriendRequest{
			AccountID: request.AccountID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Get List Friend Err", err)
			return
		}

		var listIDs = make([]uint64, len(getListFriendResp.ListFriendIDs))
		for i, listID := range getListFriendResp.ListFriendIDs {
			id, err := strconv.ParseUint(listID, 10, 64)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "Invalid List ID "+listID, err)
				return
			}
			listIDs[i] = id
		}

		getAccountInfoResp, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
			IDs: listIDs,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Get List Friend Err", err)
			return
		}

		response := map[string]interface{}{
			"Infos":   getAccountInfoResp.Infos,
			"success": true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Get friend list response error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandlerCountPendingFriendRequest(friendClient friend_service.FriendServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CountPendingFriendRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})

		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusBadRequest, "Invalid AccountID", nil)
			return
		}

		countResp, err := friendClient.CountPending(ctx, &friend_service.CountPendingRequest{
			AccountID: uint32(request.AccountID),
		})
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Count friend request error", err)
		}

		response := CountPendingFriendResponse{
			Quantity: uint64(countResp.Quantity),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Count friend response error", err)
		}
	}
}

func HandlerCheckExistingFriendRequest(friendClient friend_service.FriendServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CheckExistingFriendRequestRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkResp, err := friendClient.CheckExistingRequest(ctx, &friend_service.CheckExistingRequestRequest{
			FromAccountID: request.FromAccountID,
			ToAccountID:   request.ToAccountID,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Check existing friend request error", err)
			return
		}

		response := CheckExistingFriendRequestResponse{
			IsExisting: checkResp.IsExisting,
			RequestID:  checkResp.RequestID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Check existing friend response error", err)
		}
	}
}
