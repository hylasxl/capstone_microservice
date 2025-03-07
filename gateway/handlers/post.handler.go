package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/proto/friend_service"
	"gateway/proto/moderation_service"
	"gateway/proto/notification_service"
	"gateway/proto/post_service"
	"gateway/proto/user_service"
	"github.com/gorilla/mux"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HandlerCreatePost(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("DEBUG: Received request to create a post")

		// Parse the multipart form with a size limit
		err := r.ParseMultipartForm(256 * 1024 * 1024)
		if err != nil {
			log.Println("ERROR: Failed to parse multipart form:", err)
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		content := r.FormValue("content")
		privacyStatus := r.FormValue("privacy_status")
		isPublishedLater := r.FormValue("is_published_later") == "true"

		log.Println("DEBUG: Parsed form values -> content:", content, "privacy_status:", privacyStatus, "is_published_later:", isPublishedLater)

		// Handle optional timestamp parsing
		var timestampInt int64
		timestamp := r.FormValue("published_later_timestamp")
		if isPublishedLater {
			if timestamp == "" {
				log.Println("ERROR: Missing timestamp for scheduled post")
				respondWithError(w, http.StatusBadRequest, "Missing timestamp for scheduled post", nil)
				return
			}
			timestampInt, err = strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				log.Println("ERROR: Invalid timestamp:", timestamp, "Error:", err)
				respondWithError(w, http.StatusBadRequest, "Invalid timestamp", err)
				return
			}
		}
		log.Println("DEBUG: Parsed timestamp:", timestampInt)

		// Parse account ID
		accountID := r.FormValue("account_id")
		accID, err := strconv.ParseInt(accountID, 10, 64)
		if err != nil {
			log.Println("ERROR: Invalid account ID:", accountID, "Error:", err)
			respondWithError(w, http.StatusBadRequest, "Invalid account ID", err)
			return
		}
		log.Println("DEBUG: Parsed account ID:", accID)

		// Parse tagged accounts
		tagAccountIDStr := r.FormValue("tag_account_ids")
		var tagAccountIDs []string
		if tagAccountIDStr != "" {
			tagAccountIDs = strings.Split(tagAccountIDStr, ",")
		}
		log.Println("DEBUG: Parsed tagged accounts:", tagAccountIDs)

		// Context for post creation
		postCtx, postCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer postCancel()

		// Prepare the request
		createPostRequest := &post_service.CreatePostRequest{
			AccountID:              uint64(accID),
			Content:                content,
			IsPublishedLater:       isPublishedLater,
			PublishedLateTimestamp: timestampInt,
			PrivacyStatus:          privacyStatus,
			TagAccountIDs:          tagAccountIDs,
		}

		log.Println("DEBUG: Sending request to CreateNewPost service")
		createPostResp, err := postClient.CreateNewPost(postCtx, createPostRequest)
		if err != nil {
			log.Println("ERROR: Cannot create post:", err)
			respondWithError(w, http.StatusInternalServerError, "Cannot create post", err)
			return
		}
		log.Println("DEBUG: Post created successfully with PostID:", createPostResp.PostID)

		// Context for image upload
		imageCtx, imageCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer imageCancel()

		// Process uploaded images
		var mediaMessages []*MultiMediaMessage
		files := r.MultipartForm.File["medias"]
		log.Println("DEBUG: Number of uploaded files:", len(files))
		if len(files) > 0 {
			for _, fileHeader := range files {
				log.Println("DEBUG: Processing file:", fileHeader.Filename, "Size:", fileHeader.Size, "Type:", fileHeader.Header.Get("Content-Type"))

				file, err := fileHeader.Open()
				if err != nil {
					log.Println("ERROR: Failed to open file:", fileHeader.Filename, "Error:", err)
					respondWithError(w, http.StatusInternalServerError, "Failed to open file", err)
					return
				}
				defer file.Close()

				fileContent, err := io.ReadAll(file)
				if err != nil {
					log.Println("ERROR: Failed to read file content:", fileHeader.Filename, "Error:", err)
					respondWithError(w, http.StatusInternalServerError, "Failed to read file content", err)
					return
				}

				mediaMessages = append(mediaMessages, &MultiMediaMessage{
					Type:         fileHeader.Header.Get("Content-Type"),
					UploadStatus: "uploaded",
					Content:      fileHeader.Filename,
					Media:        fileContent,
				})
			}
		}

		// Prepare image upload request
		var requestMM []*post_service.MultiMediaMessage
		for _, mediaMessage := range mediaMessages {
			requestMM = append(requestMM, &post_service.MultiMediaMessage{
				Media:        mediaMessage.Media,
				Content:      mediaMessage.Content,
				MediaType:    "picture",
				UploadStatus: mediaMessage.UploadStatus,
			})
		}

		// Upload images if there are any
		var mediaURLs []string
		if len(requestMM) > 0 {
			log.Println("DEBUG: Uploading", len(requestMM), "images to UploadPostImage service")

			mediaRequest := &post_service.UploadImageRequest{
				PostID: createPostResp.PostID,
				Medias: requestMM,
			}

			mediaResp, err := postClient.UploadPostImage(imageCtx, mediaRequest)
			if err != nil {
				log.Println("ERROR: Cannot upload images:", err)
				respondWithError(w, http.StatusInternalServerError, "Cannot upload image", err)
				return
			}
			mediaURLs = mediaResp.MediaURLs
			log.Println("DEBUG: Uploaded images successfully, URLs:", mediaURLs)
		}

		// Send response
		response := &CreatePostResponse{
			PostID:    strconv.FormatUint(createPostResp.PostID, 10),
			MediaURLs: mediaURLs,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Println("ERROR: Cannot encode response:", err)
			respondWithError(w, http.StatusInternalServerError, "Cannot encode response", err)
			return
		}

		log.Println("DEBUG: Post created successfully and response sent")
	}
}

func readFileToBytes(file *multipart.FileHeader) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
		}
	}(f)

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return fileBytes, nil
}

func HandlerSharePost(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request SharePostRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: request.AccountID,
		})

		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusBadRequest, "Invalid Account ID", err)
			return
		}

		if !request.IsShared {
			http.Error(w, "IsShared field must be true", http.StatusUnauthorized)
			return
		}

		if strings.TrimSpace(request.Content) != "" {
			modifiedContent, err := moderationClient.IdentifyAndReplaceText(ctx, &moderation_service.IdentifyAndReplaceTextRequest{
				Content: request.Content,
			})
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "Failed to check text", err)
				return
			}
			request.Content = modifiedContent.ReturnedContent
		}

		AccID, err := strconv.ParseUint(request.AccountID, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid AccountID", err)
			return
		}

		OriginPostID, err := strconv.ParseUint(request.OriginalPostID, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid Origin PostID", err)
		}

		sharePostResp, err := postClient.SharePost(ctx, &post_service.SharePostRequest{
			AccountID:      AccID,
			Content:        request.Content,
			IsShared:       request.IsShared,
			OriginalPostID: OriginPostID,
			PrivacyStatus:  request.PrivacyStatus,
			TagAccountIDs:  request.TagAccountIDs,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to share post", err)
			return
		}

		if sharePostResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "failed to share post", errors.New(sharePostResp.Error))
			return
		}

		var response = &SharePostResponse{
			PostID: strconv.FormatUint(sharePostResp.PostID, 10),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}
	}
}

func HandlerCommentPost(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, moderationClient moderation_service.ModerationServiceClient, notificationClient notification_service.NotificationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CommentPostRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Request", err)
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
		if request.PostID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid PostID", nil)
			return
		}
		if strings.TrimSpace(request.Content) != "" {
			modifiedContent, err := moderationClient.IdentifyAndReplaceText(ctx, &moderation_service.IdentifyAndReplaceTextRequest{
				Content: request.Content,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "failed to check text", err)
				return
			}
			request.Content = modifiedContent.ReturnedContent
		}

		commentResp, err := postClient.CommentPost(ctx, &post_service.CommentPostRequest{
			PostID:    request.PostID,
			Content:   request.Content,
			AccountID: request.AccountID,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "", errors.New("failed to comment post"))
			return
		}

		if commentResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "", errors.New(commentResp.Error))
			return
		}

		userDataResp, err := userClient.GetAccountInfo(ctx, &user_service.GetAccountInfoRequest{
			AccountID: uint32(request.AccountID),
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "", errors.New("failed to get account info"))
			return
		}

		_, err = notificationClient.CommentNotification(ctx, &notification_service.CommentNotificationRequest{
			ReceiverAccountID:        int64(commentResp.PostAccountID),
			SenderAccountID:          int64(request.AccountID),
			SenderAccountDisplayName: userDataResp.AccountInfo.FirstName + " " + userDataResp.AccountInfo.LastName,
		})

		if err != nil {
			return
		}

		response := &CommentPostResponse{}

		response.Success = true
		response.CommentID = commentResp.CommentID
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to create comment", http.StatusInternalServerError)
			return
		}
	}
}
func HandlerReplyComment(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, moderationClient moderation_service.ModerationServiceClient, notificationClient notification_service.NotificationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var response ReplyCommentResponse

		var request ReplyCommentRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusUnauthorized, "Invalid account ID", err)
			return
		}

		if request.PostID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid post ID", nil)
			return
		}
		if request.OriginalCommentID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid original comment ID", nil)
			return
		}

		if strings.TrimSpace(request.Content) != "" {
			modifiedContent, err := moderationClient.IdentifyAndReplaceText(ctx, &moderation_service.IdentifyAndReplaceTextRequest{
				Content: request.Content,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to check text", err)
				return
			}
			request.Content = modifiedContent.ReturnedContent
		}

		commentResp, err := postClient.ReplyComment(ctx, &post_service.ReplyCommentRequest{
			PostID:            request.PostID,
			AccountID:         uint32(request.AccountID),
			OriginalCommentID: request.OriginalCommentID,
			ReplyContent:      request.Content,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to reply to comment", errors.New(commentResp.Error))
			return
		}

		if commentResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to reply to comment", fmt.Errorf(commentResp.Error))
			return
		}

		userDataResp, err := userClient.GetAccountInfo(ctx, &user_service.GetAccountInfoRequest{
			AccountID: uint32(request.AccountID),
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "", errors.New("failed to get account info"))
			return
		}

		_, err = notificationClient.ReplyCommentNotification(ctx, &notification_service.ReplyCommentNotificationRequest{
			ReceiverAccountID:        int64(commentResp.PostCommentID),
			SenderAccountID:          int64(request.AccountID),
			SenderAccountDisplayName: userDataResp.AccountInfo.FirstName + " " + userDataResp.AccountInfo.LastName,
		})

		if err != nil {
			return
		}

		response.Success = true
		response.CommentID = int64(commentResp.ReplyCommentID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
		}
	}
}

func HandlerGetSinglePost(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetSinglePostRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if request.PostID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid post ID", errors.New("postID is required and must be greater than zero"))
			return
		}

		postResp, err := postClient.GetSinglePost(ctx, &post_service.GetSinglePostRequest{
			PostID: request.PostID,
		})

		if err != nil || postResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to get post", nil)
			return
		}

		var response = &GetSinglePostResponse{
			Success:             true,
			PostID:              postResp.PostID,
			Content:             postResp.Content,
			PrivacyStatus:       postResp.PrivacyStatus,
			TotalCommentNumber:  postResp.TotalCommentNumber,
			TotalReactionNumber: postResp.TotalReactionNumber,
			TotalShareNumber:    postResp.TotalShareNumber,
		}

		if len(postResp.Medias) > 0 {
			for _, media := range postResp.Medias {
				var data = &post_service.MediaDisplay{
					MediaID:             media.MediaID,
					URL:                 media.URL,
					Content:             media.Content,
					TotalCommentNumber:  media.TotalCommentNumber,
					TotalReactionNumber: media.TotalReactionNumber,
				}

				response.Medias = append(response.Medias, data)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerDeletePost(postClient post_service.PostServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeletePostRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if request.PostID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid post ID", nil)
			return
		}

		deletePostResp, err := postClient.DeletePost(ctx, &post_service.DeletePostRequest{
			PostID: request.PostID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete post", err)
			return
		}
		if deletePostResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete post", fmt.Errorf(deletePostResp.Error))
			return
		}

		response := &DeletePostResponse{
			Success: true,
			PostID:  deletePostResp.PostID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerEditComment(postClient post_service.PostServiceClient, moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request EditPostCommentRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if request.CommentID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid comment ID", nil)
			return
		}

		if len(strings.TrimSpace(request.Content)) > 0 {
			modifiedResp, err := moderationClient.IdentifyAndReplaceText(ctx,
				&moderation_service.IdentifyAndReplaceTextRequest{
					Content: request.Content,
				})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to modify comment", err)
				return
			}
			if modifiedResp.Error != "" {
				respondWithError(w, http.StatusInternalServerError, "Failed to modify comment", fmt.Errorf(modifiedResp.Error))
				return
			}

			request.Content = modifiedResp.ReturnedContent
		} else {
			respondWithError(w, http.StatusBadRequest, "Invalid content", nil)
			return
		}

		editResp, err := postClient.EditComment(ctx, &post_service.EditCommentRequest{
			CommentID: request.CommentID,
			Content:   request.Content,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to edit comment", err)
			return
		}
		if editResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to edit comment", fmt.Errorf(editResp.Error))
			return
		}

		response := &EditPostCommentResponse{
			Success:   true,
			CommentID: editResp.CommentID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}

	}
}

func HandlerDeletePostComment(postClient post_service.PostServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var vars = mux.Vars(r)
		id := vars["id"]
		commentID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid comment ID", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var request = &DeletePostCommentRequest{
			CommentID: commentID,
		}
		if request.CommentID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid comment ID", nil)
			return
		}

		deleteResp, err := postClient.DeleteComment(ctx, &post_service.DeleteCommentRequest{
			CommentID: request.CommentID,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete comment", err)
			return
		}

		if deleteResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete comment", fmt.Errorf(deleteResp.Error))
			return
		}

		response := &DeletePostCommentResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerDeletePostImage(postClient post_service.PostServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeletePostImageRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		if request.PostID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid post ID", nil)
			return
		}

		if request.MediaID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid media ID", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		deleteResp, err := postClient.DeletePostImage(ctx, &post_service.DeletePostImageRequest{
			PostID:  request.PostID,
			MediaID: request.MediaID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete post image", err)
			return
		}
		if deleteResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete post image", fmt.Errorf(deleteResp.Error))
			return
		}

		response := &DeletePostImageResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerReactPost(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ReactPostRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		if request.PostID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid post ID", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})

		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		reactPostResp, err := postClient.ReactPost(ctx, &post_service.ReactPostRequest{
			PostID:    request.PostID,
			AccountID: request.AccountID,
			ReactType: request.ReactType,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to react post", err)
			return
		}

		if reactPostResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to react post", fmt.Errorf(reactPostResp.Error))
			return
		}

		response := &ReactPostResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerRemoveReactPost(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request RemoveReactPostRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		if request.PostID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid post ID", nil)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		removeResp, err := postClient.RemoveReactPost(ctx, &post_service.RemoveReactPostRequest{
			PostID:    request.PostID,
			AccountID: request.AccountID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to remove post", err)
			return
		}
		if removeResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to remove post", fmt.Errorf(removeResp.Error))
			return
		}
		response := &RemoveReactPostResponse{
			Success: true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerReactImage(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ReactImageRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		if request.MediaID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid media ID", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		reactImageResp, err := postClient.ReactImage(ctx, &post_service.ReactImageRequest{
			AccountID: request.AccountID,
			MediaID:   request.MediaID,
			ReactType: request.ReactType,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to react image", err)
			return
		}
		if reactImageResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to react image", fmt.Errorf(reactImageResp.Error))
			return
		}
		response := &ReactImageResponse{
			Success: true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
		}
	}
}

func HandlerRemoveReactImage(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request RemoveReactImageRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}
		if request.MediaID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid media ID", nil)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		removeResp, err := postClient.RemoveReactImage(ctx, &post_service.RemoveReactImageRequest{
			AccountID: request.AccountID,
			MediaID:   request.MediaID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to remove post", err)
			return
		}
		if removeResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to remove post", fmt.Errorf(removeResp.Error))
			return
		}
		response := &RemoveReactImageResponse{
			Success: true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerCommentImage(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CommentImageRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		if request.MediaID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid media ID", nil)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		if len(strings.TrimSpace(request.Content)) > 0 {
			modifiedResp, err := moderationClient.IdentifyAndReplaceText(
				ctx, &moderation_service.IdentifyAndReplaceTextRequest{
					Content: request.Content,
				})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to replace text", err)
				return
			}
			if modifiedResp.Error != "" {
				respondWithError(w, http.StatusInternalServerError, "Failed to replace text", fmt.Errorf(modifiedResp.Error))
				return
			}
			request.Content = modifiedResp.ReturnedContent
		} else {
			respondWithError(w, http.StatusInternalServerError, "Content is empty", nil)
			return
		}

		commentResp, err := postClient.CommentImage(ctx, &post_service.CommentImageRequest{
			AccountID: request.AccountID,
			MediaID:   request.MediaID,
			Content:   request.Content,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to comment image", err)
			return
		}

		if commentResp.Error != "" {
			respondWithError(w, http.StatusInternalServerError, "Failed to comment image", fmt.Errorf(commentResp.Error))
			return
		}

		response := &CommentImageResponse{
			CommentID: commentResp.CommentID,
			Success:   true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerReplyCommentImage(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ReplyCommentImageRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}
		if request.MediaID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid media ID", nil)
			return
		}
		if request.OriginalCommentID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid origin comment ID", nil)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		if len(strings.TrimSpace(request.Content)) > 0 {
			modifiedResp, err := moderationClient.IdentifyAndReplaceText(
				ctx, &moderation_service.IdentifyAndReplaceTextRequest{
					Content: request.Content,
				})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to replace text", err)
				return
			}
			if modifiedResp.Error != "" {
				respondWithError(w, http.StatusInternalServerError, "Failed to replace text", fmt.Errorf(modifiedResp.Error))
				return
			}
			request.Content = modifiedResp.ReturnedContent
		} else {
			respondWithError(w, http.StatusInternalServerError, "Content is empty", nil)
			return
		}

		_, err = postClient.ReplyCommentImage(ctx, &post_service.ReplyCommentImageRequest{
			AccountID:         request.AccountID,
			MediaID:           request.MediaID,
			Content:           request.Content,
			OriginalCommentID: request.OriginalCommentID,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to reply image", err)
		}

		var response = &ReplyCommentImageResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerEditCommentImage(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, moderationClient moderation_service.ModerationServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request EditCommentImageRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}
		if request.CommentID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid media ID", nil)
			return
		}

		if request.AccountID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid account ID", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})

		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		if len(strings.TrimSpace(request.Content)) > 0 {
			modifiedContent, err := moderationClient.IdentifyAndReplaceText(
				ctx, &moderation_service.IdentifyAndReplaceTextRequest{
					Content: request.Content,
				})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to replace text", err)
				return
			}
			request.Content = modifiedContent.ReturnedContent
		} else {
			respondWithError(w, http.StatusInternalServerError, "Content is empty", nil)
			return
		}

		_, err = postClient.EditCommentImage(ctx, &post_service.EditCommentImageRequest{
			AccountID: request.AccountID,
			Content:   request.Content,
			CommentID: request.CommentID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to edit image", err)
		}

		var response = &EditCommentImageResponse{
			Success: true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerDeleteCommentImage(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request DeleteCommentImageRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		if request.CommentID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid media ID", nil)
			return
		}
		if request.MediaID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid media ID", nil)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		_, err = postClient.DeleteCommentImage(ctx, &post_service.DeleteCommentImageRequest{
			AccountID: request.AccountID,
			MediaID:   request.MediaID,
			CommentID: request.CommentID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete image", err)
			return
		}
		var response = &DeleteCommentImageResponse{
			Success: true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerGetPostComments(postClient post_service.PostServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetPostCommentsRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}
		if request.PostID <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid post ID", nil)
			return
		}
		if request.Page <= 0 {
			request.Page = 1
		}
		if request.PageSize <= 0 {
			request.PageSize = 10
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := postClient.GetPostComment(ctx, &post_service.GetPostCommentRequest{
			PostID:   request.PostID,
			Page:     uint32(request.Page),
			PageSize: uint32(request.PageSize),
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get post comments", err)
			return
		}

		var response = &GetPostCommentsResponse{
			Success:            true,
			PostID:             request.PostID,
			TotalCommentNumber: uint64(resp.TotalComments),
			Comments:           resp.Comments,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}

	}
}

func HandlerGetWallPost(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, friendClient friend_service.FriendServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetWallPostListRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkRequestAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.RequestAccountID, 10),
		})
		if err != nil || !checkRequestAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		checkTargetAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.TargetAccountID, 10),
		})

		if err != nil || !checkTargetAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid account ID", err)
			return
		}

		isFriendCheck, err := friendClient.CheckIsFriend(ctx, &friend_service.CheckIsFriendRequest{
			FirstAccountID:  request.TargetAccountID,
			SecondAccountID: request.RequestAccountID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to check friend check", err)
			return
		}

		isBlockCheck, err := friendClient.CheckIsBlock(ctx, &friend_service.CheckIsBlockedRequest{
			FirstAccountID:  request.TargetAccountID,
			SecondAccountID: request.RequestAccountID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to check friend check", err)
			return
		}

		postCResp, err := postClient.GetWallPostList(ctx, &post_service.GetWallPostListRequest{
			TargetAccountID:           request.TargetAccountID,
			RequestAccountID:          request.RequestAccountID,
			Page:                      uint32(request.Page),
			PageSize:                  uint32(request.PageSize),
			IsFriend:                  isFriendCheck.IsFriend,
			IsAccountBlockedEachOther: isBlockCheck.IsBlocked,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get post list", err)
			return
		}

		var returnedPostData = make([]DisplayPost, 0)

		for _, post := range postCResp.Posts {

			var accountList []uint64
			accountList = append(accountList, post.AccountID)
			accountInfoResp, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
				IDs: accountList,
			})

			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to get account list", err)
				return
			}

			var mediaArr = make([]PostShareMediaDisplay, 0)
			for _, media := range post.Medias {
				mediaArr = append(mediaArr, PostShareMediaDisplay{
					URL:     media.URL,
					Content: media.Content,
					MediaID: media.MediaID,
				})
			}

			var sharePostData = &SharePostDataDisplay{}
			if post.IsShared {

				var accountListShare []uint64
				accountListShare = append(accountListShare, post.SharePostData.AccountID)

				accountShareInfo, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
					IDs: accountListShare,
				})

				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Failed to get account list", err)
					return
				}

				var sharePostMedias = make([]PostShareMediaDisplay, 0)
				if post.SharePostData != nil {
					for _, media := range post.SharePostData.Medias {
						sharePostMedias = append(sharePostMedias, PostShareMediaDisplay{
							URL:     media.URL,
							Content: media.Content,
							MediaID: media.MediaID,
						})
					}
					sharePostData.PostID = post.SharePostData.PostID
					sharePostData.Content = post.SharePostData.Content
					sharePostData.IsContentEdited = post.SharePostData.IsContentEdited
					sharePostData.PrivacyStatus = post.SharePostData.PrivacyStatus
					sharePostData.CreatedAt = post.SharePostData.CreatedAt
					sharePostData.IsPublished = post.SharePostData.IsPublished
					sharePostData.Medias = sharePostMedias
					sharePostData.Account = SingleAccountInfo{
						AccountID:   uint(accountShareInfo.Infos[0].AccountID),
						AvatarURL:   accountShareInfo.Infos[0].AvatarURL,
						DisplayName: accountShareInfo.Infos[0].DisplayName,
					}
				}
			} else {
				post.SharePostData = nil
			}

			var postReaction PostReactionDisplay

			postReaction.TotalQuantity = uint64(post.Reactions.Count)
			postReaction.Reactions = make([]PostReactionData, 0, post.Reactions.Count)

			var accountReactionIDs []uint64
			for _, reaction := range post.Reactions.DisplayData {
				accountReactionIDs = append(accountReactionIDs, reaction.AccountID)
			}

			if len(accountReactionIDs) > 0 {
				userInfoDisplayList, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
					IDs: accountReactionIDs,
				})
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Failed to get account display info", err)
					return
				}

				for index, userInfoDisplay := range userInfoDisplayList.Infos {
					if index >= len(post.Reactions.DisplayData) {
						break
					}
					var accountInfo = SingleAccountInfo{
						AccountID:   uint(userInfoDisplay.AccountID),
						AvatarURL:   userInfoDisplay.AvatarURL,
						DisplayName: userInfoDisplay.DisplayName,
					}
					var postReactionData = PostReactionData{
						ReactionType: post.Reactions.DisplayData[index].ReactionType,
						Account:      accountInfo,
					}
					postReaction.Reactions = append(postReaction.Reactions, postReactionData)
				}
			}

			var postCommentDisplay PostCommentDisplay
			postCommentDisplay.TotalQuantity = uint64(post.Comments.Count)

			var postShareDisplay PostShareDisplay
			postShareDisplay.TotalQuantity = uint64(post.Shares.Count)
			var accountShareIDs = make([]uint64, post.Shares.Count)
			for _, share := range post.Shares.DisplayData {
				accountShareIDs = append(accountShareIDs, share.AccountID)
			}

			if len(accountShareIDs) > 0 {
				userDisplayList, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
					IDs: accountShareIDs,
				})
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Failed to get account display info", err)
					return
				}
				for index, userInfoDisplay := range userDisplayList.Infos {
					var accountInfo = SingleAccountInfo{
						AccountID:   uint(userInfoDisplay.AccountID),
						AvatarURL:   userInfoDisplay.AvatarURL,
						DisplayName: userInfoDisplay.DisplayName,
					}
					var shareData = PostShareData{
						CreatedAt: post.Shares.DisplayData[index].CreatedAt,
						Account:   accountInfo,
					}

					postShareDisplay.Shares = append(postShareDisplay.Shares, shareData)
				}
			}

			var displayPost = &DisplayPost{
				CreatedAt:               post.CreatedAt,
				PostID:                  post.PostID,
				Content:                 post.Content,
				SharePostID:             post.SharePostID,
				IsHidden:                post.IsHidden,
				IsContentEdited:         post.IsContentEdited,
				IsShared:                post.IsShared,
				IsPublished:             post.IsPublished,
				IsPublishedLater:        post.IsPublishedLater,
				PublishedLaterTimestamp: post.PublishedLaterTimestamp,
				PrivacyStatus:           post.PrivacyStatus,
				InteractionType:         post.InteractionType,
				Medias:                  mediaArr,
				SharePostData:           *sharePostData,
				Reactions:               postReaction,
				CommentQuantity:         postCommentDisplay,
				Shares:                  postShareDisplay,
				Account: SingleAccountInfo{
					AccountID:   uint(accountInfoResp.Infos[0].AccountID),
					AvatarURL:   accountInfoResp.Infos[0].AvatarURL,
					DisplayName: accountInfoResp.Infos[0].DisplayName,
				},
			}

			returnedPostData = append(returnedPostData, *displayPost)
		}

		var response = &GetWallPostListResponse{
			TargetAccountID: request.TargetAccountID,
			Page:            request.Page,
			PageSize:        request.PageSize,
			Posts:           returnedPostData,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerGetNewFeeds(postClient post_service.PostServiceClient, userClient user_service.UserServiceClient, friendClient friend_service.FriendServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetNewFeedsRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		checkAccountID, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(request.AccountID, 10),
		})

		if err != nil || !checkAccountID.IsValid {
			respondWithError(w, http.StatusInternalServerError, "Invalid AccountID", err)
			return
		}

		listFriendResp, err := friendClient.GetListFriend(ctx, &friend_service.GetListFriendRequest{
			AccountID: strconv.FormatUint(request.AccountID, 10),
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get friend list", err)
			return
		}

		listIDInt64 := make([]uint64, len(listFriendResp.ListFriendIDs))
		for _, id := range listFriendResp.ListFriendIDs {
			idInt64, _ := strconv.ParseUint(id, 10, 64)
			listIDInt64 = append(listIDInt64, idInt64)
		}

		userInteraction, err := friendClient.GetUserInteraction(ctx, &friend_service.GetUserInteractionRequest{
			AccountID: request.AccountID,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get user interaction", err)
			return
		}

		interactions := make([]*post_service.InteractionScore, len(userInteraction.Interactions))
		for i, interaction := range userInteraction.Interactions {
			interactions[i] = &post_service.InteractionScore{
				AccountID: interaction.AccountID,
				Score:     interaction.Score,
			}
		}

		newFeedsResp, err := postClient.GetNewFeeds(ctx, &post_service.GetNewFeedsRequest{
			AccountID:     request.AccountID,
			Page:          request.Page,
			PageSize:      request.PageSize,
			SeenPostIDs:   request.SeenPostID,
			ListFriendIDs: listIDInt64,
			Interactions:  interactions,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get new feed list", err)
			return
		}

		var returnedPostData = make([]DisplayPost, 0)

		for _, post := range newFeedsResp.Posts {

			var accountList []uint64
			accountList = append(accountList, post.AccountID)
			accountInfoResp, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
				IDs: accountList,
			})

			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to get account list", err)
				return
			}

			var mediaArr = make([]PostShareMediaDisplay, 0)
			for _, media := range post.Medias {
				mediaArr = append(mediaArr, PostShareMediaDisplay{
					URL:     media.URL,
					Content: media.Content,
					MediaID: media.MediaID,
				})
			}

			var sharePostData = &SharePostDataDisplay{}
			if post.IsShared {

				var accountListShare []uint64
				accountListShare = append(accountListShare, post.SharePostData.AccountID)

				accountShareInfo, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
					IDs: accountListShare,
				})

				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Failed to get account list", err)
					return
				}

				var sharePostMedias = make([]PostShareMediaDisplay, 0)
				if post.SharePostData != nil {
					for _, media := range post.SharePostData.Medias {
						sharePostMedias = append(sharePostMedias, PostShareMediaDisplay{
							URL:     media.URL,
							Content: media.Content,
							MediaID: media.MediaID,
						})
					}
					sharePostData.PostID = post.SharePostData.PostID
					sharePostData.Content = post.SharePostData.Content
					sharePostData.IsContentEdited = post.SharePostData.IsContentEdited
					sharePostData.PrivacyStatus = post.SharePostData.PrivacyStatus
					sharePostData.CreatedAt = post.SharePostData.CreatedAt
					sharePostData.IsPublished = post.SharePostData.IsPublished
					sharePostData.Medias = sharePostMedias
					sharePostData.Account = SingleAccountInfo{
						AccountID:   uint(accountShareInfo.Infos[0].AccountID),
						AvatarURL:   accountShareInfo.Infos[0].AvatarURL,
						DisplayName: accountShareInfo.Infos[0].DisplayName,
					}
				}
			} else {
				post.SharePostData = nil
			}

			var postReaction PostReactionDisplay

			postReaction.TotalQuantity = uint64(post.Reactions.Count)
			postReaction.Reactions = make([]PostReactionData, 0, post.Reactions.Count)

			var accountReactionIDs []uint64
			for _, reaction := range post.Reactions.DisplayData {
				accountReactionIDs = append(accountReactionIDs, reaction.AccountID)
			}

			if len(accountReactionIDs) > 0 {
				userInfoDisplayList, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
					IDs: accountReactionIDs,
				})
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Failed to get account display info", err)
					return
				}

				for index, userInfoDisplay := range userInfoDisplayList.Infos {
					if index >= len(post.Reactions.DisplayData) {
						break
					}
					var accountInfo = SingleAccountInfo{
						AccountID:   uint(userInfoDisplay.AccountID),
						AvatarURL:   userInfoDisplay.AvatarURL,
						DisplayName: userInfoDisplay.DisplayName,
					}
					var postReactionData = PostReactionData{
						ReactionType: post.Reactions.DisplayData[index].ReactionType,
						Account:      accountInfo,
					}
					postReaction.Reactions = append(postReaction.Reactions, postReactionData)
				}
			}

			var postCommentDisplay PostCommentDisplay
			postCommentDisplay.TotalQuantity = uint64(post.Comments.Count)

			var postShareDisplay PostShareDisplay
			postShareDisplay.TotalQuantity = uint64(post.Shares.Count)
			var accountShareIDs = make([]uint64, post.Shares.Count)
			for _, share := range post.Shares.DisplayData {
				accountShareIDs = append(accountShareIDs, share.AccountID)
			}

			if len(accountShareIDs) > 0 {
				userDisplayList, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
					IDs: accountShareIDs,
				})
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Failed to get account display info", err)
					return
				}
				for index, userInfoDisplay := range userDisplayList.Infos {
					var accountInfo = SingleAccountInfo{
						AccountID:   uint(userInfoDisplay.AccountID),
						AvatarURL:   userInfoDisplay.AvatarURL,
						DisplayName: userInfoDisplay.DisplayName,
					}
					var shareData = PostShareData{
						CreatedAt: post.Shares.DisplayData[index].CreatedAt,
						Account:   accountInfo,
					}

					postShareDisplay.Shares = append(postShareDisplay.Shares, shareData)
				}
			}

			var displayPost = &DisplayPost{
				CreatedAt:               post.CreatedAt,
				PostID:                  post.PostID,
				Content:                 post.Content,
				SharePostID:             post.SharePostID,
				IsHidden:                post.IsHidden,
				IsContentEdited:         post.IsContentEdited,
				IsShared:                post.IsShared,
				IsPublished:             post.IsPublished,
				IsPublishedLater:        post.IsPublishedLater,
				PublishedLaterTimestamp: post.PublishedLaterTimestamp,
				PrivacyStatus:           post.PrivacyStatus,
				InteractionType:         post.InteractionType,
				Medias:                  mediaArr,
				SharePostData:           *sharePostData,
				Reactions:               postReaction,
				CommentQuantity:         postCommentDisplay,
				Shares:                  postShareDisplay,
				Account: SingleAccountInfo{
					AccountID:   uint(accountInfoResp.Infos[0].AccountID),
					AvatarURL:   accountInfoResp.Infos[0].AvatarURL,
					DisplayName: accountInfoResp.Infos[0].DisplayName,
				},
			}

			returnedPostData = append(returnedPostData, *displayPost)
		}

		var response = &GetNewFeedsResponse{
			AccountID: request.AccountID,
			Page:      uint64(request.Page),
			PageSize:  uint64(request.PageSize),
			Posts:     returnedPostData,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerGetNewPostStatisticData(postClient post_service.PostServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetNewPostStatisticDataRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		periodData := request.PeriodYear

		if request.PeriodLabel == "month" {
			periodData = request.PeriodYear*100 + (request.PeriodMonth % 100)
		}

		resp, err := postClient.GetNewPostStatisticData(ctx, &post_service.GetNewPostStatisticDataRequest{
			PeriodLabel: request.PeriodLabel,
			PeriodData:  periodData,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get post statistic data", err)
			return
		}

		var data []DataTerms
		for _, post := range resp.Data {
			data = append(data, DataTerms{
				Label: post.Label,
				Count: post.Count,
			})
		}

		var response = &GetNewPostStatisticDataResponse{
			RequestAccountID: request.RequestAccountID,
			Data:             data,
			PeriodLabel:      request.PeriodLabel,
			TotalPosts:       uint32(resp.TotalPosts),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
			return
		}
	}
}

func HandlerGetMediaStatistic(postClient post_service.PostServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetMediaStatisticRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		periodData := request.PeriodYear

		if request.PeriodLabel == "month" {
			periodData = request.PeriodYear*100 + (request.PeriodMonth % 100)
		}
		resp, err := postClient.GetMediaStatistic(ctx, &post_service.GetMediaStatisticRequest{
			PeriodLabel: request.PeriodLabel,
			PeriodData:  periodData,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get post media statistic data", err)
			return
		}

		var data []DataTerms
		for _, post := range resp.Data {
			data = append(data, DataTerms{
				Label: post.Label,
				Count: post.Count,
			})
		}

		var response = &GetMediaStatisticResponse{
			RequestAccountID: request.RequestAccountID,
			Data:             data,
			PeriodLabel:      request.PeriodLabel,
			TotalMedias:      uint32(resp.TotalMedias),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
		}

	}
}

func HandlerGetPostWMediaStatistic(postClient post_service.PostServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request GetPostWMediaStatisticRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		periodData := request.PeriodYear

		if request.PeriodLabel == "month" {
			periodData = request.PeriodYear*100 + (request.PeriodMonth % 100)
		}

		resp, err := postClient.GetPostWMediaStatistic(ctx, &post_service.GetPostWMediaStatisticRequest{
			PeriodLabel: request.PeriodLabel,
			PeriodData:  periodData,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get post media statistic data", err)
			return
		}

		var response = &GetPostWMediaStatisticResponse{
			RequestAccountID: request.RequestAccountID,
			TotalPosts:       resp.TotalPost,
			TotalPostWMedias: resp.TotalPostWMedia,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
		}
	}
}
