package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/moderation_service"
	"gateway/proto/post_service"
	"gateway/proto/user_service"
	"log"
	"net/http"
	"time"
)

func HandleReportPost(moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ReportPost
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.HandleReportPost(ctx, &moderation_service.ReportPost{
			PostID:     request.PostID,
			Reason:     request.Reason,
			ReportedBy: request.ReportedBy,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := SingleStatusResponse{Success: resp.Success}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandleReportAccount(moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ReportUser
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.HandleReportAccount(ctx, &moderation_service.ReportAccount{
			AccountID:  request.AccountID,
			ReportedBy: request.ReportedBy,
			Reason:     request.Reason,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := SingleStatusResponse{Success: resp.Success}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandleResolveReportedPost(moderationClient moderation_service.ModerationServiceClient, postClient post_service.PostServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ResolveReportedPost
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.HandleResolveReportedPost(ctx, &moderation_service.ResolveReportedPost{
			PostID:     request.PostID,
			ResolvedBy: request.ResolvedBy,
			Method:     request.Method,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp.Success {
			_, err := postClient.DeletePostByAdmin(ctx, &post_service.AdminDeletePostRequest{
				PostID:     request.PostID,
				DeleteType: 1,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		response := SingleStatusResponse{Success: resp.Success}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandleResolveReportedAccount(moderationClient moderation_service.ModerationServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ResolveReportedAccount
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.HandleResolveReportedAccount(ctx, &moderation_service.ResolveReportedAccount{
			Method:     request.Method,
			AccountID:  request.AccountID,
			ResolvedBy: request.ResolvedBy,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp.Success {
			_, err := userClient.CustomDeleteAccount(ctx, &user_service.CustomDeleteAccountRequest{
				AccountID: request.AccountID,
				Method:    "admin",
			})

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		response := SingleStatusResponse{Success: resp.Success}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandleGetReportedAccount(moderationClient moderation_service.ModerationServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var request GetReportedAccountListRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.HandleGetReportAccountList(ctx, &moderation_service.GetReportedAccountListRequest{
			Page:     request.Page,
			PageSize: request.PageSize,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var response GetReportedAccountListResponse
		response.Page = resp.Page
		response.PageSize = resp.PageSize

		// Cache to store previously fetched usernames
		usernameCache := make(map[uint32]string)

		for _, acc := range resp.Data {
			var username string

			// Check if username is already cached
			if cachedUsername, found := usernameCache[acc.AccountID]; found {
				username = cachedUsername
			} else {
				// Fetch username if not cached
				usnResp, err := userClient.GetUsername(ctx, &user_service.GetUsernameRequest{
					AccountID: acc.AccountID,
				})

				if err != nil {
					username = "Unknown" // Handle error gracefully
				} else {
					username = usnResp.Username
					// Store in cache
					usernameCache[acc.AccountID] = username
				}
			}

			// Append account data with the retrieved username
			response.Accounts = append(response.Accounts, ReportAccountData{
				Username:      username,
				AccountID:     acc.AccountID,
				Reason:        acc.Reasons,
				ResolveStatus: acc.ResolveStatus,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandleGetReportedPosts(moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetReportedPostRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.GetReportedPosts(ctx, &moderation_service.GetReportedPostRequest{
			Page:     request.Page,
			PageSize: request.PageSize,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := &GetReportedPostResponse{
			Page:     resp.Page,
			PageSize: resp.PageSize,
		}

		for _, report := range resp.ReportedPosts {
			response.ReportedPosts = append(response.ReportedPosts, ReportedPostData{
				ID:            report.ID,
				PostID:        report.PostID,
				Reason:        report.Reason,
				ResolveStatus: report.ResolveStatus,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerGetBanWords(moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetListBanWordsReq
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.GetListBanWords(ctx, &moderation_service.GetListBanWordsReq{
			RequestAccountID: request.RequestAccountID,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerEditWord(moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request EditWordReq
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.EditWord(ctx, &moderation_service.EditWordReq{
			ID:      request.ID,
			Content: request.Content,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerDeleteWord(moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeleteWordReq
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.DeleteWord(ctx, &moderation_service.DeleteWordReq{
			ID: request.ID,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerAddWord(moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request AddWordReq
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := moderationClient.AddWord(ctx, &moderation_service.AddWordReq{
			Content:          request.Content,
			RequestAccountID: request.RequestAccountID,
			LanguageCode:     request.LanguageCode,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
