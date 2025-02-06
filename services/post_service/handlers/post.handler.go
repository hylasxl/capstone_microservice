package handlers

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"post_service/models"
	ps "post_service/proto/post_service"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PostService struct {
	ps.UnimplementedPostServiceServer
	DB               *gorm.DB
	CloudinaryClient *CloudinaryService
}

var validReactions = []string{"like", "dislike", "love", "hate", "cry"}

func (s *PostService) CreateNewPost(ctx context.Context, in *ps.CreatePostRequest) (*ps.CreatePostResponse, error) {

	if in.PrivacyStatus != "public" && in.PrivacyStatus != "private" && in.PrivacyStatus != "friend_only" {
		return &ps.CreatePostResponse{
			Error: "Invalid privacy status",
		}, nil
	}

	tx := s.DB.Begin()

	if in.IsPublishedLater {
		exeTime := time.Unix(in.PublishedLateTimestamp, 0)
		duration := time.Until(exeTime)
		if duration < 0 {
			tx.Rollback()
			errMsg := fmt.Sprintf("Cannot schedule task: time %v has already passed", exeTime)
			log.Println(errMsg)
			return &ps.CreatePostResponse{Error: errMsg}, errors.New(errMsg)
		}
	}

	post := &models.Post{
		AccountID:             uint(in.AccountID),
		Content:               strings.TrimSpace(in.Content),
		IsPublishedLater:      in.IsPublishedLater,
		PublishLaterTimestamp: time.Unix(in.PublishedLateTimestamp, 0),
		IsShared:              false,
		PrivacyStatus:         in.PrivacyStatus,
		IsPublished:           !in.IsPublishedLater,
	}

	if err := tx.Create(post).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to create post: %v", err)
		log.Println(errMsg)
		return &ps.CreatePostResponse{Error: errMsg}, err
	}

	for _, id := range in.TagAccountIDs {
		ID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to parse tag account ID: %v", err)
			log.Println(errMsg)
			return &ps.CreatePostResponse{Error: errMsg}, err
		}

		var order *uint

		if err := tx.Model(&models.PostTagFriend{}).Where(
			"post_id = ?", post.ID).Select("MAX(`order`)").Scan(&order).Error; err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to get post tag order: %v", err)
			log.Println(errMsg)
			return &ps.CreatePostResponse{Error: errMsg}, err
		}
		if order == nil {
			defaultOrder := uint(1)
			order = &defaultOrder
		} else {
			*order = *order + 1
		}

		data := &models.PostTagFriend{
			PostID:       post.ID,
			TagAccountID: uint(ID),
			Order:        *order,
		}
		if err := tx.Create(data).Error; err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to create tag post: %v", err)
			log.Println(errMsg)
			return &ps.CreatePostResponse{Error: errMsg}, err
		}
	}

	if in.IsPublishedLater == true {
		err := s.schedulePost(post.ID, in.PublishedLateTimestamp)
		if err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to schedule post: %v", err)
			log.Println(errMsg)
			return &ps.CreatePostResponse{Error: errMsg}, err
		}
	}

	tx.Commit()

	return &ps.CreatePostResponse{PostID: uint64(post.ID)}, nil
}
func (s *PostService) schedulePost(postID uint, timestamp int64) error {
	exeTime := time.Unix(timestamp, 0)
	if time.Now().After(exeTime) {
		errMsg := fmt.Sprintf("Cannot schedule post: time %v has already passed", exeTime)
		log.Println(errMsg)
		return fmt.Errorf(errMsg)
	}

	delay := time.Until(exeTime)

	go func() {
		time.Sleep(delay)
		if err := s.DB.Model(models.Post{}).Where("id = ?", postID).Update(
			"is_published", true).Error; err != nil {
			log.Println(err)
		}
	}()

	return nil
}

func (s *PostService) UploadPostImage(ctx context.Context, in *ps.UploadImageRequest) (*ps.UploadImageResponse, error) {
	fmt.Println("Connected to iuploaf post image")
	fmt.Println(len(in.Medias))
	fmt.Println(in.PostID)
	if len(in.Medias) == 0 {
		return &ps.UploadImageResponse{Error: "No media files provided"}, nil
	}
	if in.PostID <= 0 {
		return &ps.UploadImageResponse{Error: "Invalid PostID"}, nil
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Transaction rolled back due to panic:", r)
		}
	}()

	var (
		URLs     []string
		UrlMutex sync.Mutex
		wg       sync.WaitGroup
		errChan  = make(chan error, len(in.Medias))
	)

	for i, data := range in.Medias {
		wg.Add(1)
		go func(index int, mediaData *ps.MultiMediaMessage) {
			fmt.Printf("Uploading image %v\n", index)
			defer wg.Done()
			uploadedUrl, err := s.CloudinaryClient.UploadImage(mediaData.Media)
			statusUpload := "uploaded"
			if err != nil {
				statusUpload = "failed"
				log.Printf("Failed to upload image at index %d: %v", index, err)
				errChan <- fmt.Errorf("image upload failed at index %d: %v", index, err)
				return
			}

			imageData := &models.PostMultiMedia{
				PostID:       uint(in.PostID),
				URL:          uploadedUrl,
				Content:      mediaData.Content,
				MediaType:    mediaData.MediaType,
				UploadStatus: statusUpload,
			}

			if err := tx.Create(imageData).Error; err != nil {
				errChan <- fmt.Errorf("failed to record upload result for image at index %d: %v", index, err)
				return
			}
			if statusUpload == "uploaded" {
				UrlMutex.Lock()
				URLs = append(URLs, uploadedUrl)
				UrlMutex.Unlock()
			}
		}(i, data)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		tx.Rollback()
		for err := range errChan {
			log.Println(err)
		}
		return &ps.UploadImageResponse{Error: "One or more image uploads failed."}, nil
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.UploadImageResponse{Error: errMsg}, err
	}

	return &ps.UploadImageResponse{MediaURLs: URLs}, nil
}

func (s *PostService) SharePost(ctx context.Context, in *ps.SharePostRequest) (*ps.SharePostResponse, error) {

	if in.AccountID <= 0 {
		return &ps.SharePostResponse{Error: "Invalid AccountID"}, nil
	}

	if in.IsShared == false {
		return &ps.SharePostResponse{Error: "You can't share a post"}, nil
	}

	if in.OriginalPostID <= 0 {
		return &ps.SharePostResponse{Error: "Invalid OriginalPostID"}, nil
	}

	if in.PrivacyStatus != "public" && in.PrivacyStatus != "private" && in.PrivacyStatus != "friend_only" {
		return &ps.SharePostResponse{
			Error: "Invalid privacy status",
		}, nil
	}

	tx := s.DB.Begin()

	postData := &models.Post{
		AccountID:        uint(in.AccountID),
		Content:          in.Content,
		IsPublishedLater: false,
		IsShared:         true,
		OriginalPostID:   uint(in.OriginalPostID),
		PrivacyStatus:    in.PrivacyStatus,
	}

	if err := tx.Create(postData).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to create post: %v", err)
		log.Println(errMsg)
		return &ps.SharePostResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.SharePostResponse{Error: errMsg}, err
	}

	return &ps.SharePostResponse{
		PostID: uint64(postData.ID),
	}, nil
}

func (s *PostService) CommentPost(ctx context.Context, in *ps.CommentPostRequest) (*ps.CommentPostResponse, error) {

	if in.AccountID <= 0 {
		return &ps.CommentPostResponse{Error: "Invalid AccountID"}, nil
	}
	if in.PostID <= 0 {
		return &ps.CommentPostResponse{Error: "Invalid PostID"}, nil
	}
	if len(strings.TrimSpace(in.Content)) == 0 {
		return &ps.CommentPostResponse{Error: "Invalid Content"}, nil
	}

	tx := s.DB.Begin()

	var existingPost *models.Post

	if err := tx.Model(models.Post{}).Where("id = ? AND is_self_deleted = false AND is_deleted_by_admin = false", uint(in.PostID)).First(&existingPost).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find post with ID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.CommentPostResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find post with ID %d: %v", in.PostID, err)
			log.Println(errMsg)
		}
	}

	commentData := &models.PostComment{
		PostID:    uint(in.PostID),
		Content:   strings.TrimSpace(in.Content),
		AccountID: uint(in.AccountID),
	}

	if err := tx.Create(commentData).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to create post comment: %v", err)
		log.Println(errMsg)
		return &ps.CommentPostResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.CommentPostResponse{Error: errMsg}, err
	}

	return &ps.CommentPostResponse{
		CommentID:     int64(commentData.ID),
		PostAccountID: uint64(existingPost.AccountID),
	}, nil
}

func (s *PostService) DeletePost(ctx context.Context, in *ps.DeletePostRequest) (*ps.DeletePostResponse, error) {
	if in.PostID <= 0 {
		return &ps.DeletePostResponse{Error: "Invalid PostID"}, nil
	}

	tx := s.DB.Begin()

	var existingPost *models.Post
	if err := tx.Where("id = ?", uint(in.PostID)).First(&existingPost).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find post with ID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.DeletePostResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find post with ID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.DeletePostResponse{Error: errMsg}, err
		}
	}

	if err := tx.Delete(existingPost).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to delete post with ID %d: %v", in.PostID, err)
		log.Println(errMsg)
		return &ps.DeletePostResponse{Error: errMsg}, err
	}

	if err := tx.Model(&models.Post{}).Where("id = ?", in.PostID).Update("is_self_deleted", true).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to delete post with ID %d: %v", in.PostID, err)
		log.Println(errMsg)
		return &ps.DeletePostResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.DeletePostResponse{Error: errMsg}, err
	}

	return &ps.DeletePostResponse{
		PostID: uint64(existingPost.ID),
	}, nil
}

func (s *PostService) ReplyComment(ctx context.Context, in *ps.ReplyCommentRequest) (*ps.ReplyCommentResponse, error) {
	if in.PostID <= 0 {
		return &ps.ReplyCommentResponse{Error: "Invalid PostID"}, nil
	}

	if in.OriginalCommentID < 1 {
		return &ps.ReplyCommentResponse{Error: "Invalid OriginalCommentID"}, nil
	}

	if len(strings.TrimSpace(in.ReplyContent)) == 0 {
		return &ps.ReplyCommentResponse{Error: "Invalid ReplyContent"}, nil
	}

	tx := s.DB.Begin()

	var existingPost *models.Post

	if err := tx.Model(models.Post{}).Where("id = ? AND is_self_deleted = false AND is_deleted_by_admin = false", uint(in.PostID)).First(&existingPost).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find post with ID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.ReplyCommentResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find post with ID %d: %v", in.PostID, err)
			log.Println(errMsg)
		}
	}

	var originalComment models.PostComment
	if err := tx.Model(&originalComment).Where("id = ? AND post_id = ? ", in.OriginalCommentID, in.PostID).First(&originalComment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find original comment with ID %d: %v", in.OriginalCommentID, err)
			log.Println(errMsg)
			return &ps.ReplyCommentResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find original comment with ID %d: %v", in.OriginalCommentID, err)
			log.Println(errMsg)
			return &ps.ReplyCommentResponse{Error: errMsg}, err
		}

	}

	commentData := &models.PostComment{
		PostID:      uint(in.PostID),
		AccountID:   uint(in.AccountID),
		Content:     strings.TrimSpace(in.ReplyContent),
		ReplyFromID: uint(in.OriginalCommentID),
		Level:       originalComment.Level + 1,
		IsReply:     true,
	}

	if err := tx.Create(commentData).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to create post comment: %v", err)
		log.Println(errMsg)
		return &ps.ReplyCommentResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.ReplyCommentResponse{Error: errMsg}, err
	}

	return &ps.ReplyCommentResponse{
		ReplyCommentID: uint64(commentData.ID),
		PostCommentID:  uint64(existingPost.AccountID),
	}, nil
}

func (s *PostService) EditComment(ctx context.Context, in *ps.EditCommentRequest) (*ps.EditCommentResponse, error) {

	if in.CommentID <= 0 {
		return &ps.EditCommentResponse{Error: "Invalid CommentID"}, nil
	}

	if strings.TrimSpace(in.Content) == "" {
		return &ps.EditCommentResponse{Error: "Invalid Content"}, nil
	}

	tx := s.DB.Begin()

	var existingComment *models.PostComment

	if err := tx.Where("id = ?", uint(in.CommentID)).First(&existingComment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find comment with ID %d: %v", in.CommentID, err)
			log.Println(errMsg)
			return &ps.EditCommentResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find comment with ID %d: %v", in.CommentID, err)
			log.Println(errMsg)
			return &ps.EditCommentResponse{Error: errMsg}, err
		}
	}

	var originComment = existingComment.Content

	if err := tx.Model(existingComment).Where("id = ?", in.CommentID).Update("content", strings.TrimSpace(in.Content)).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to edit comment with ID %d: %v", in.CommentID, err)
		log.Println(errMsg)
		return &ps.EditCommentResponse{Error: errMsg}, err
	}

	if err := tx.Model(existingComment).Where("id = ?", in.CommentID).Update("is_edited", true).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to edit comment with ID %d: %v", in.CommentID, err)
		log.Println(errMsg)
		return &ps.EditCommentResponse{Error: errMsg}, err
	}

	var commentHis = &models.PostCommentEditHistory{
		PostID:        uint(existingComment.PostID),
		CommentID:     uint(in.CommentID),
		BeforeContent: originComment,
		AfterContent:  strings.TrimSpace(in.Content),
	}

	if err := tx.Create(&commentHis).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to edit comment with ID %d: %v", in.CommentID, err)
		log.Println(errMsg)
		return &ps.EditCommentResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.EditCommentResponse{Error: errMsg}, err
	}

	return &ps.EditCommentResponse{
		CommentID: in.CommentID,
	}, nil
}

func (s *PostService) DeleteComment(ctx context.Context, in *ps.DeleteCommentRequest) (*ps.DeleteCommentResponse, error) {
	if in.CommentID <= 0 {
		return &ps.DeleteCommentResponse{Error: "Invalid CommentID"}, nil
	}
	tx := s.DB.Begin()

	var mainComment *models.PostComment

	if err := tx.Where("id = ?", uint(in.CommentID)).First(&mainComment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find comment with ID %d: %v", in.CommentID, err)
			log.Println(errMsg)
			return &ps.DeleteCommentResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find comment with ID %d: %v", in.CommentID, err)
			log.Println(errMsg)
			return &ps.DeleteCommentResponse{Error: errMsg}, err
		}
	}

	if err := markAsDeletedNestedReplies(tx, uint(in.CommentID)); err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to delete comments: %v", err)
		log.Println(errMsg)
		return &ps.DeleteCommentResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.DeleteCommentResponse{Error: errMsg}, err
	}

	return &ps.DeleteCommentResponse{}, nil
}

func markAsDeletedNestedReplies(tx *gorm.DB, commentID uint) error {

	var replies []models.PostComment
	if err := tx.Where("reply_from_id = ?", commentID).Find(&replies).Error; err != nil {
		return err
	}

	for _, reply := range replies {
		if err := markAsDeletedNestedReplies(tx, reply.ID); err != nil {
			return err
		}
	}

	if err := tx.Model(&models.PostComment{}).
		Where("id = ? OR reply_from_id = ?", commentID, commentID).
		Updates(map[string]interface{}{"is_self_deleted": true}).Error; err != nil {
		return err
	}

	return nil
}

func (s *PostService) DeletePostImage(ctx context.Context, in *ps.DeletePostImageRequest) (*ps.DeletePostImageResponse, error) {
	if in.PostID <= 0 {
		return &ps.DeletePostImageResponse{Error: "Invalid PostID"}, nil
	}

	if in.MediaID <= 0 {
		return &ps.DeletePostImageResponse{Error: "Invalid MediaID"}, nil
	}
	tx := s.DB.Begin()

	var existingMedia *models.PostMultiMedia
	if err := tx.Where("id = ?", uint(in.MediaID)).First(&existingMedia).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find media with ID %d: %v", in.MediaID, err)
			log.Println(errMsg)
			return &ps.DeletePostImageResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find media with ID %d: %v", in.MediaID, err)
			log.Println(errMsg)
			return &ps.DeletePostImageResponse{Error: errMsg}, err
		}
	}

	if err := tx.Model(existingMedia).Where("id = ?", in.MediaID).Update("is_self_deleted", true).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to edit post image: %v", err)
		log.Println(errMsg)
		return &ps.DeletePostImageResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.DeletePostImageResponse{Error: errMsg}, err
	}

	return &ps.DeletePostImageResponse{}, nil
}
func (s *PostService) ReactPost(ctx context.Context, in *ps.ReactPostRequest) (*ps.ReactPostResponse, error) {
	if in.PostID <= 0 {
		return &ps.ReactPostResponse{Error: "Invalid PostID"}, nil
	}
	if in.AccountID <= 0 {
		return &ps.ReactPostResponse{Error: "Invalid AccountID"}, nil
	}
	isValidReaction := false
	for _, reaction := range validReactions {
		if in.ReactType == reaction {
			isValidReaction = true
			break
		}
	}

	if !isValidReaction {
		return &ps.ReactPostResponse{Error: "Invalid Reaction"}, nil
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	var existingReact models.PostReaction

	err := tx.Model(&models.PostReaction{}).
		Where("post_id = ? AND account_id = ?", in.PostID, in.AccountID).
		First(&existingReact).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create a new reaction
		newReaction := models.PostReaction{
			PostID:       uint(in.PostID),
			AccountID:    uint(in.AccountID),
			ReactionType: in.ReactType,
		}
		if err := tx.Create(&newReaction).Error; err != nil {
			log.Printf("Failed to create new reaction: %v", err)
			return &ps.ReactPostResponse{Error: err.Error()}, nil
		}
	} else if err != nil {
		log.Printf("Failed to fetch existing reaction: %v", err)
		return &ps.ReactPostResponse{Error: err.Error()}, nil
	} else {
		// Update the existing reaction
		updates := map[string]interface{}{"reaction_type": in.ReactType}
		if existingReact.IsRecalled {
			updates["is_recalled"] = false
		}
		if err := tx.Model(&existingReact).
			Where("id = ?", existingReact.ID).
			Updates(updates).Error; err != nil {
			log.Printf("Failed to update existing reaction: %v", err)
			return &ps.ReactPostResponse{Error: err.Error()}, nil
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Transaction commit failed: %v", err)
		return &ps.ReactPostResponse{Error: "Transaction commit failed"}, nil
	}

	return &ps.ReactPostResponse{}, nil
}

func (s *PostService) RemoveReactPost(ctx context.Context, in *ps.RemoveReactPostRequest) (*ps.RemoveReactPostResponse, error) {
	if in.PostID <= 0 {
		return &ps.RemoveReactPostResponse{Error: "Invalid PostID"}, nil
	}
	if in.AccountID <= 0 {
		return &ps.RemoveReactPostResponse{Error: "Invalid AccountID"}, nil
	}
	tx := s.DB.Begin()

	var existingReaction *models.PostReaction
	if err := tx.Where("post_id = ? AND account_id = ?", in.PostID, in.AccountID).First(&existingReaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find reaction with ID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.RemoveReactPostResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find reaction with ID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.RemoveReactPostResponse{Error: errMsg}, err
		}
	}

	if err := tx.Model(existingReaction).Where("id = ?", existingReaction.ID).Updates(map[string]interface{}{"is_recalled": true}).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to recall reactions: %v", err)
		log.Println(errMsg)
		return &ps.RemoveReactPostResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.RemoveReactPostResponse{Error: errMsg}, err
	}

	return &ps.RemoveReactPostResponse{}, nil
}

func (s *PostService) ReactImage(ctx context.Context, in *ps.ReactImageRequest) (*ps.ReactImageResponse, error) {
	if in.MediaID <= 0 {
		return &ps.ReactImageResponse{Error: "Invalid MediaID"}, nil
	}

	if in.AccountID <= 0 {
		return &ps.ReactImageResponse{Error: "Invalid AccountID"}, nil
	}

	isValidReaction := false
	for _, reaction := range validReactions {
		if in.ReactType == reaction {
			isValidReaction = true
			break
		}
	}

	if !isValidReaction {
		return &ps.ReactImageResponse{Error: "Invalid Reaction"}, nil
	}

	tx := s.DB.Begin()

	var reactionData = &models.PostMultiMediaReaction{
		PostMediaID:  uint(in.MediaID),
		AccountID:    uint(in.AccountID),
		ReactionType: in.ReactType,
	}

	if err := tx.Create(&reactionData).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to create reaction: %v", err)
		log.Println(errMsg)
		return &ps.ReactImageResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.ReactImageResponse{Error: errMsg}, err
	}

	return &ps.ReactImageResponse{}, nil
}

func (s *PostService) RemoveReactImage(ctx context.Context, in *ps.RemoveReactImageRequest) (*ps.RemoveReactImageResponse, error) {
	if in.MediaID <= 0 {
		return &ps.RemoveReactImageResponse{Error: "Invalid MediaID"}, nil
	}

	if in.AccountID <= 0 {
		return &ps.RemoveReactImageResponse{Error: "Invalid AccountID"}, nil
	}

	tx := s.DB.Begin()

	var existingReaction *models.PostMultiMediaReaction
	if err := tx.Where("media_id = ? AND account_id = ?", in.MediaID, in.AccountID).First(&existingReaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find reaction with ID %d: %v", in.MediaID, err)
			log.Println(errMsg)
			return &ps.RemoveReactImageResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find reaction with ID %d: %v", in.MediaID, err)
			log.Println(errMsg)
			return &ps.RemoveReactImageResponse{Error: errMsg}, err
		}
	}

	if err := tx.Model(existingReaction).Where("id = ?", existingReaction.ID).Updates(map[string]interface{}{"is_recalled": true}).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to recall reactions: %v", err)
		log.Println(errMsg)
		return &ps.RemoveReactImageResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.RemoveReactImageResponse{Error: errMsg}, err
	}

	return &ps.RemoveReactImageResponse{}, nil
}

func (s *PostService) CommentImage(ctx context.Context, in *ps.CommentImageRequest) (*ps.CommentImageResponse, error) {
	if in.MediaID <= 0 {
		return &ps.CommentImageResponse{Error: "Invalid MediaID"}, nil
	}
	if in.AccountID <= 0 {
		return &ps.CommentImageResponse{Error: "Invalid AccountID"}, nil
	}
	if strings.TrimSpace(in.Content) == "" {
		return &ps.CommentImageResponse{Error: "Invalid Content"}, nil
	}

	tx := s.DB.Begin()

	var commentData = &models.PostMultiMediaComment{
		PostMediaID: uint(in.MediaID),
		AccountID:   uint(in.AccountID),
		Content:     strings.TrimSpace(in.Content),
	}

	if err := tx.Create(&commentData).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to create comment: %v", err)
		log.Println(errMsg)
		return &ps.CommentImageResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.CommentImageResponse{Error: errMsg}, err
	}

	return &ps.CommentImageResponse{}, nil
}

func (s *PostService) ReplyCommentImage(ctx context.Context, in *ps.ReplyCommentImageRequest) (*ps.ReplyCommentImageResponse, error) {
	if in.MediaID <= 0 {
		return &ps.ReplyCommentImageResponse{Error: "Invalid MediaID"}, nil
	}
	if in.AccountID <= 0 {
		return &ps.ReplyCommentImageResponse{Error: "Invalid AccountID"}, nil
	}
	if strings.TrimSpace(in.Content) == "" {
		return &ps.ReplyCommentImageResponse{Error: "Invalid Content"}, nil
	}
	tx := s.DB.Begin()

	var replyData = &models.PostMultiMediaComment{
		PostMediaID: uint(in.MediaID),
		AccountID:   uint(in.AccountID),
		Content:     strings.TrimSpace(in.Content),
		IsReply:     true,
		Level:       uint(in.CommentLevel),
		ReplyFromID: uint(in.OriginalCommentID),
	}

	if err := tx.Create(&replyData).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to create comment: %v", err)
		log.Println(errMsg)
		return &ps.ReplyCommentImageResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.ReplyCommentImageResponse{Error: errMsg}, err
	}

	return &ps.ReplyCommentImageResponse{}, nil
}

func (s *PostService) EditCommentImage(ctx context.Context, in *ps.EditCommentImageRequest) (*ps.EditCommentImageResponse, error) {
	if in.CommentID <= 0 {
		return &ps.EditCommentImageResponse{Error: "Invalid CommentID"}, nil
	}
	if in.AccountID <= 0 {
		return &ps.EditCommentImageResponse{Error: "Invalid AccountID"}, nil
	}
	if strings.TrimSpace(in.Content) == "" {
		return &ps.EditCommentImageResponse{Error: "Invalid Content"}, nil
	}
	tx := s.DB.Begin()

	var existingComment *models.PostMultiMediaComment

	if err := tx.Where("id = ?", in.CommentID).First(&existingComment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find comment with ID %d: %v", in.CommentID, err)
			log.Println(errMsg)
			return &ps.EditCommentImageResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find comment with ID %d: %v", in.CommentID, err)
			log.Println(errMsg)
			return &ps.EditCommentImageResponse{Error: errMsg}, err
		}
	}

	var originComment = existingComment.Content

	if err := tx.Model(existingComment).
		Where("id = ?", existingComment.ID).
		Updates(map[string]interface{}{"is_edited": true, "content": in.Content}).
		Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to edit comment: %v", err)
		log.Println(errMsg)
		return &ps.EditCommentImageResponse{Error: errMsg}, err
	}

	var commentHis = &models.PostMultiMediaEditCommentHistory{
		MediaID:       uint(existingComment.PostMediaID),
		CommentID:     uint(in.CommentID),
		BeforeContent: originComment,
		AfterContent:  in.Content,
	}

	if err := tx.Create(&commentHis).Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to create comment history: %v", err)
		log.Println(errMsg)
		return &ps.EditCommentImageResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.EditCommentImageResponse{Error: errMsg}, err
	}

	return &ps.EditCommentImageResponse{}, nil
}

func (s *PostService) DeleteCommentImage(ctx context.Context, in *ps.DeleteCommentImageRequest) (*ps.DeleteCommentImageResponse, error) {
	if in.CommentID <= 0 {
		return &ps.DeleteCommentImageResponse{Error: "Invalid CommentID"}, nil
	}
	if in.AccountID <= 0 {
		return &ps.DeleteCommentImageResponse{Error: "Invalid AccountID"}, nil
	}
	if in.MediaID <= 0 {
		return &ps.DeleteCommentImageResponse{Error: "Invalid MediaID"}, nil
	}
	tx := s.DB.Begin()

	var existingComment *models.PostMultiMediaComment
	if err := tx.Where("id = ?", in.CommentID).First(&existingComment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find comment with ID %d: %v", in.CommentID, err)
			log.Println(errMsg)
			return &ps.DeleteCommentImageResponse{Error: errMsg}, err
		} else {
			tx.Rollback()
			errMsg := fmt.Sprintf("Failed to find comment with ID %d: %v", in.CommentID, err)
			log.Println(errMsg)
			return &ps.DeleteCommentImageResponse{Error: errMsg}, err
		}
	}

	if err := markCommentAsRecalled(tx, existingComment.ID); err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to mark recalled comment: %v", err)
		log.Println(errMsg)
		return &ps.DeleteCommentImageResponse{Error: errMsg}, err
	}

	if err := tx.Commit().Error; err != nil {
		errMsg := fmt.Sprintf("Transaction commit failed: %v", err)
		log.Println(errMsg)
		return &ps.DeleteCommentImageResponse{Error: errMsg}, err
	}
	return &ps.DeleteCommentImageResponse{}, nil
}

func markCommentAsRecalled(tx *gorm.DB, commentID uint) error {
	var replies []models.PostMultiMediaComment
	if err := tx.Where("reply_from_id = ?", commentID).Find(&replies).Error; err != nil {
		return err
	}

	for _, reply := range replies {
		if err := markAsDeletedNestedReplies(tx, reply.ID); err != nil {
			return err
		}
	}

	if err := tx.Model(&models.PostMultiMediaComment{}).
		Where("id = ? OR reply_from_id = ? ", commentID, commentID).
		Updates(map[string]interface{}{"is_self_deleted": true}).Error; err != nil {
		return err
	}
	return nil
}

func (s *PostService) CountPostComment(ctx context.Context, in *ps.CountPostCommentRequest) (*ps.CountPostCommentResponse, error) {

	if in.GetPostID() <= 0 {
		return &ps.CountPostCommentResponse{Error: "Invalid PostID"}, nil
	}

	var count int64

	if err := s.DB.Model(&models.PostComment{}).
		Where("post_id = ? AND is_self_deleted = ? AND is_deleted_by_admin = ?", in.PostID, false, false).
		Count(&count).Error; err != nil {
		errMsg := fmt.Sprintf("Failed to count comments for PostID %d: %v", in.PostID, err)
		log.Println(errMsg)
		return &ps.CountPostCommentResponse{Error: errMsg}, err
	}

	return &ps.CountPostCommentResponse{
		CommentQuantity: count,
	}, nil
}

func (s *PostService) CountPostReaction(ctx context.Context, in *ps.CountPostReactionRequest) (*ps.CountPostReactionResponse, error) {
	if in.GetPostID() <= 0 {
		return &ps.CountPostReactionResponse{Error: "Invalid PostID"}, nil
	}
	var count int64

	if err := s.DB.Model(&models.PostReaction{}).
		Where("post_id = ? AND is_recalled = ?", in.PostID, false).
		Count(&count).Error; err != nil {
		errMsg := fmt.Sprintf("Failed to count reactions for PostID %d: %v", in.PostID, err)
		log.Println(errMsg)
		return &ps.CountPostReactionResponse{Error: errMsg}, err
	}

	return &ps.CountPostReactionResponse{
		ReactionQuantity: count,
	}, nil
}

func (s *PostService) CountPostShare(ctx context.Context, in *ps.CountPostShareRequest) (*ps.CountPostShareResponse, error) {
	if in.GetPostID() <= 0 {
		return &ps.CountPostShareResponse{Error: "Invalid PostID"}, nil
	}
	var count int64
	if err := s.DB.Model(&models.Post{}).
		Where("original_post_id = ? AND is_self_deleted = ? AND is_shared = ?", in.PostID, false, true).
		Count(&count).Error; err != nil {
		errMsg := fmt.Sprintf("Failed to count shares for PostID %d: %v", in.PostID, err)
		log.Println(errMsg)
		return &ps.CountPostShareResponse{Error: errMsg}, err
	}

	return &ps.CountPostShareResponse{
		ShareQuantity: count,
	}, nil
}
func (s *PostService) GetPostComment(ctx context.Context, in *ps.GetPostCommentRequest) (*ps.GetPostCommentResponse, error) {
	if in.GetPostID() <= 0 {
		return &ps.GetPostCommentResponse{Error: "Invalid PostID"}, nil
	}

	var comments []models.PostComment
	if err := s.DB.
		Where("post_id = ? AND reply_from_id IS NULL AND is_self_deleted = false AND is_deleted_by_admin = false",
			in.PostID).
		Order("created_at DESC").
		Offset(int((in.Page - 1) * in.PageSize)).
		Limit(int(in.PageSize)).
		Find(&comments).Error; err != nil {
		errMsg := fmt.Sprintf("Failed to get comments for PostID %d: %v", in.PostID, err)
		log.Println(errMsg)
		return &ps.GetPostCommentResponse{Error: errMsg}, err
	}

	var responseComment []*ps.Comment
	for _, comment := range comments {
		mappedComment, err := mappedWithReplies(s.DB, comment)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to get comment for PostID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.GetPostCommentResponse{Error: errMsg}, err
		}
		responseComment = append(responseComment, mappedComment)
	}

	var totalComment, _ = s.CountPostComment(ctx, &ps.CountPostCommentRequest{PostID: in.PostID})

	return &ps.GetPostCommentResponse{
		PostID:        uint64(in.PostID),
		Comments:      responseComment,
		TotalComments: uint32(totalComment.CommentQuantity),
	}, nil
}

func mappedWithReplies(db *gorm.DB, comment models.PostComment) (*ps.Comment, error) {
	mappedComment := &ps.Comment{
		CommentID:   uint64(comment.ID),
		AccountID:   uint64(comment.AccountID),
		Content:     comment.Content,
		IsEdited:    comment.IsEdited,
		ReplyFromID: uint64(comment.ReplyFromID),
		Level:       uint32(comment.Level),
		Replies:     []*ps.Comment{},
	}

	var replies []models.PostComment
	if err := db.Where("reply_from_id = ? AND is_self_deleted = false AND is_deleted_by_admin = false",
		comment.ID).
		Order("created_at ASC").
		Find(&replies).Error; err != nil {
		return nil, err
	}

	for _, reply := range replies {
		mappedReply, err := mappedWithReplies(db, reply)
		if err != nil {
			return nil, err
		}
		mappedComment.Replies = append(mappedComment.Replies, mappedReply)
	}

	return mappedComment, nil
}

func (s *PostService) GetPostReaction(ctx context.Context, in *ps.GetPostReactionRequest) (*ps.GetPostReactionResponse, error) {
	if in.GetPostID() <= 0 {
		return &ps.GetPostReactionResponse{Error: "Invalid PostID"}, nil
	}

	var postReactions []models.PostReaction
	if err := s.DB.Model(&models.PostReaction{}).
		Where("id = ? AND is_recalled = ?", in.PostID, false).
		Find(&postReactions).Error; err != nil {
		errMsg := fmt.Sprintf("Failed to get post reactions for PostID %d: %v", in.PostID, err)
		log.Println(errMsg)
		return &ps.GetPostReactionResponse{Error: errMsg}, err
	}

	var returnedData []*ps.ReactionDisplay
	for _, postReaction := range postReactions {
		var data = &ps.ReactionDisplay{
			ReactionType: postReaction.ReactionType,
			AccountID:    uint64(postReaction.AccountID),
		}
		returnedData = append(returnedData, data)
	}

	return &ps.GetPostReactionResponse{
		Reactions: returnedData,
	}, nil
}

func (s *PostService) GetPostMediaComment(ctx context.Context, in *ps.GetPostMediaCommentRequest) (*ps.GetPostMediaCommentResponse, error) {
	if in.GetPostMediaID() <= 0 {
		return &ps.GetPostMediaCommentResponse{Error: "Invalid PostMediaID"}, nil
	}

	var comments []models.PostMultiMediaComment
	if err := s.DB.
		Where("post_media_id = ? AND reply_from_id IS NULL AND is_self_deleted = false AND is_deleted_by_admin = false",
			in.PostMediaID).
		Order("created_at DESC").
		Offset(int((in.Page - 1) * in.PageSize)).
		Limit(int(in.PageSize)).
		Find(&comments).Error; err != nil {
		errMsg := fmt.Sprintf("Failed to get comments for PostMediaID %d: %v", in.PostMediaID, err)
		log.Println(errMsg)
		return &ps.GetPostMediaCommentResponse{Error: errMsg}, err
	}

	var responseComments []*ps.MediaComment
	for _, comment := range comments {
		mappedComment, err := mapMediaCommentWithReplies(s.DB, comment)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to get comment for PostMediaID %d: %v", in.PostMediaID, err)
			log.Println(errMsg)
			return &ps.GetPostMediaCommentResponse{Error: errMsg}, err
		}
		responseComments = append(responseComments, mappedComment)
	}

	var totalComments int64
	if err := s.DB.Model(&models.PostMultiMediaComment{}).
		Where("post_media_id = ? AND is_self_deleted = false AND is_deleted_by_admin = false", in.PostMediaID).
		Count(&totalComments).Error; err != nil {
		errMsg := fmt.Sprintf("Failed to count comments for PostMediaID %d: %v", in.PostMediaID, err)
		log.Println(errMsg)
		return &ps.GetPostMediaCommentResponse{Error: errMsg}, err
	}

	return &ps.GetPostMediaCommentResponse{
		PostMediaID:   uint64(in.PostMediaID),
		Comments:      responseComments,
		TotalComments: uint32(totalComments),
	}, nil
}

func mapMediaCommentWithReplies(db *gorm.DB, comment models.PostMultiMediaComment) (*ps.MediaComment, error) {
	mappedComment := &ps.MediaComment{
		CommentID:   uint64(comment.ID),
		AccountID:   uint64(comment.AccountID),
		Content:     comment.Content,
		IsEdited:    comment.IsEdited,
		ReplyFromID: uint64(comment.ReplyFromID),
		Level:       uint32(comment.Level),
		Replies:     []*ps.MediaComment{},
	}

	var replies []models.PostMultiMediaComment
	if err := db.Where("reply_from_id = ? AND is_self_deleted = false AND is_deleted_by_admin = false", comment.ID).
		Order("created_at ASC").
		Find(&replies).Error; err != nil {
		return nil, err
	}

	for _, reply := range replies {
		mappedReply, err := mapMediaCommentWithReplies(db, reply)
		if err != nil {
			return nil, err
		}
		mappedComment.Replies = append(mappedComment.Replies, mappedReply)
	}

	return mappedComment, nil
}

func (s *PostService) GetSinglePost(ctx context.Context, in *ps.GetSinglePostRequest) (*ps.GetSinglePostResponse, error) {
	if in.PostID <= 0 {
		return &ps.GetSinglePostResponse{Error: "Invalid PostID"}, nil
	}

	var postData models.Post
	if err := s.DB.Where("id = ?", in.PostID).First(&postData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errMsg := fmt.Sprintf("Failed to get post for PostID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.GetSinglePostResponse{Error: errMsg}, err
		} else {
			errMsg := fmt.Sprintf("Failed to get post for PostID %d: %v", in.PostID, err)
			log.Println(errMsg)
			return &ps.GetSinglePostResponse{Error: errMsg}, err
		}
	}

	var medias []*ps.MediaDisplay

	var postMedia []models.PostMultiMedia

	if err := s.DB.Model(postMedia).
		Where("id = ? AND is_self_deleted = ? AND is_deleted_by_admin = ?", in.PostID, false, false).
		Find(&postMedia).Error; err != nil {
		errMsg := fmt.Sprintf("Failed to get post media for PostID %d: %v", in.PostID, err)
		log.Println(errMsg)
		return &ps.GetSinglePostResponse{Error: errMsg}, err
	}

	for _, media := range postMedia {
		var totalComments int64
		if err := s.DB.Model(&models.PostMultiMediaComment{}).
			Where("post_media_id = ? AND is_self_deleted = false AND is_deleted_by_admin = false", media.ID).
			Count(&totalComments).Error; err != nil {
			errMsg := fmt.Sprintf("Failed to count comments for PostMediaID %d: %v", media.ID, err)
			log.Println(errMsg)
			return &ps.GetSinglePostResponse{Error: errMsg}, err
		}
		var totalReactions int64
		if err := s.DB.Model(&models.PostMultiMediaReaction{}).
			Where("post_media_id = ? AND is_recalled = false").
			Count(&totalReactions).Error; err != nil {
			errMsg := fmt.Sprintf("Failed to count reactions for PostMediaID %d: %v", media.ID, err)
			log.Println(errMsg)
			return &ps.GetSinglePostResponse{Error: errMsg}, err
		}
		var data = &ps.MediaDisplay{
			URL:                 media.URL,
			Content:             media.Content,
			MediaID:             uint64(media.ID),
			TotalReactionNumber: uint64(totalReactions),
			TotalCommentNumber:  uint64(totalComments),
		}

		medias = append(medias, data)
	}

	var totalComments, _ = s.CountPostComment(ctx, &ps.CountPostCommentRequest{PostID: in.PostID})
	var totalShares, _ = s.CountPostShare(ctx, &ps.CountPostShareRequest{PostID: in.PostID})
	var totalReactions, _ = s.CountPostReaction(ctx, &ps.CountPostReactionRequest{PostID: in.PostID})

	return &ps.GetSinglePostResponse{
		PostID:              uint64(in.PostID),
		Content:             postData.Content,
		PrivacyStatus:       postData.PrivacyStatus,
		TotalCommentNumber:  uint64(totalComments.CommentQuantity),
		TotalShareNumber:    uint64(totalShares.ShareQuantity),
		TotalReactionNumber: uint64(totalReactions.ReactionQuantity),
		Medias:              medias,
	}, nil
}

func (s *PostService) GetWallPostList(ctx context.Context, in *ps.GetWallPostListRequest) (*ps.GetWallPostListResponse, error) {

	if in.IsAccountBlockedEachOther {
		return &ps.GetWallPostListResponse{Error: "AccountBlockedEachOther"}, errors.New("AccountBlockedEachOther")
	}

	if in.Page < 0 {
		in.Page = 1
	}

	if in.PageSize < 0 {
		in.PageSize = 10
	}

	var isSelf = false

	if in.TargetAccountID == in.RequestAccountID {
		if in.IsFriend {
			return &ps.GetWallPostListResponse{Error: "Invalid Friend"}, errors.New("invalid Friend")
		} else {
			isSelf = true
		}
	}

	var posts []models.Post
	var returnedPost = make([]*ps.DisplayPost, 0)
	tx := s.DB.Begin()
	offset := (in.Page - 1) * in.PageSize

	switch isSelf {
	case true:
		if err := tx.Model(models.Post{}).
			Where(map[string]interface{}{
				"is_self_deleted":     false,
				"is_deleted_by_admin": false,
				"is_hidden":           false,
				"account_id":          in.TargetAccountID,
			}).
			Order("created_at DESC").
			Limit(int(in.PageSize)).
			Offset(int(offset)).
			Find(&posts).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				posts = make([]models.Post, 0)
			} else {
				return &ps.GetWallPostListResponse{Error: err.Error()}, err
			}
		}

		if len(posts) > 0 {
			for _, post := range posts {
				var displayPost = ps.DisplayPost{
					PostID:                  uint64(post.ID),
					Content:                 strings.TrimSpace(post.Content),
					IsShared:                post.IsShared,
					SharePostID:             uint64(post.OriginalPostID),
					IsSelfDeleted:           post.IsSelfDeleted,
					IsDeletedByAdmin:        post.IsDeletedByAdmin,
					IsHidden:                post.IsHidden,
					IsContentEdited:         post.IsContentEdited,
					PrivacyStatus:           post.PrivacyStatus,
					CreatedAt:               post.CreatedAt.Unix(),
					IsPublishedLater:        post.IsPublishedLater,
					PublishedLaterTimestamp: post.PublishLaterTimestamp.Unix(),
					IsPublished:             post.IsPublished,
					AccountID:               uint64(post.AccountID),
				}

				if post.IsShared {
					var originPost models.Post
					var ShareData *ps.DisplayPost
					if err := tx.Model(models.Post{}).
						Where(map[string]interface{}{
							"is_self_deleted":     false,
							"is_deleted_by_admin": false,
							"is_hidden":           false,
							"id":                  post.OriginalPostID,
							"privacy_status":      "public",
						}, post.OriginalPostID).First(&originPost).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							ShareData = &ps.DisplayPost{
								Error: "Post not found",
							}
						} else {
							ShareData = &ps.DisplayPost{
								Error: err.Error(),
							}
						}
					} else {

						var medias []models.PostMultiMedia
						var mediasDisplay = make([]*ps.MediaDisplay, 0)

						if err := tx.Model(models.PostMultiMedia{}).Where(
							map[string]interface{}{
								"is_self_deleted":     false,
								"is_deleted_by_admin": false,
								"upload_status":       "uploaded",
								"post_id":             uint64(post.OriginalPostID),
							}).Find(&medias).Error; err != nil {
							if errors.Is(err, gorm.ErrRecordNotFound) {
								medias = make([]models.PostMultiMedia, 0)
							} else {
							}
						}

						for _, media := range medias {
							mediasDisplay = append(mediasDisplay, &ps.MediaDisplay{
								URL:     media.URL,
								MediaID: uint64(media.ID),
							})
						}

						ShareData = &ps.DisplayPost{
							PostID:          uint64(originPost.ID),
							Content:         strings.TrimSpace(originPost.Content),
							IsContentEdited: originPost.IsContentEdited,
							PrivacyStatus:   originPost.PrivacyStatus,
							CreatedAt:       originPost.CreatedAt.Unix(),
							IsPublished:     originPost.IsPublished,
							Medias:          mediasDisplay,
							AccountID:       uint64(originPost.AccountID),
						}
					}
					displayPost.SharePostData = ShareData
				} else {
					displayPost.SharePostData = nil
				}

				if !post.IsShared {
					displayPost.SharePostData = nil
				}

				var Media []models.PostMultiMedia
				var MediaDisplay []*ps.MediaDisplay

				if err := tx.Model(models.PostMultiMedia{}).Where(
					map[string]interface{}{
						"is_self_deleted":     false,
						"is_deleted_by_admin": false,
						"upload_status":       "uploaded",
						"post_id":             uint64(post.ID),
					}).Find(&Media).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						MediaDisplay = make([]*ps.MediaDisplay, 0)
						Media = make([]models.PostMultiMedia, 0)
					} else {
					}
				}

				for _, media := range Media {
					MediaDisplay = append(MediaDisplay, &ps.MediaDisplay{
						URL:     media.URL,
						MediaID: uint64(media.ID),
					})
				}

				var reactions []models.PostReaction
				postReaction := &ps.PostReactions{}

				if err := tx.Model(models.PostReaction{}).
					Where(map[string]interface{}{
						"is_recalled": false,
						"post_id":     post.ID,
					}).Find(&reactions).Error; err != nil {
				}

				if uint32(len(reactions)) > 0 {
					postReaction.DisplayData = make([]*ps.ReactionDisplay, 0, len(reactions))
					for _, reaction := range reactions {
						postReaction.DisplayData = append(postReaction.DisplayData, &ps.ReactionDisplay{
							ReactionType: reaction.ReactionType,
							AccountID:    uint64(reaction.AccountID),
						})
					}
					postReaction.Count = uint32(len(reactions))

				}

				var postShares []models.Post
				postShareData := &ps.PostShares{}

				if err := tx.Model(models.Post{}).Where(
					map[string]interface{}{
						"is_self_deleted":     false,
						"is_deleted_by_admin": false,
						"is_published":        true,
						"privacy_status":      "public",
						"original_post_id":    post.ID,
						"is_hidden":           false,
					}).Find(&postShares).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						postShares = make([]models.Post, 0)
					} else {
					}
				}

				if len(postShares) > 0 {
					for _, postShare := range postShares {
						postShareData.DisplayData = append(postShareData.DisplayData, &ps.ShareDisplay{
							AccountID: uint64(postShare.AccountID),
							CreatedAt: postShare.CreatedAt.Unix(),
						})
					}
					postShareData.Count = uint32(len(postShares))
				}

				var countComment int64
				if err := tx.Model(&models.PostComment{}).
					Where("post_id = ? AND is_self_deleted = ? AND is_deleted_by_admin = ?", post.ID, false, false).
					Count(&countComment).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						countComment = 0
					} else {
						countComment = 0
					}
				}

				if countComment < 0 {
					countComment = 0
				}

				postComment := &ps.PostComments{
					Count: uint32(countComment),
				}

				var interactionType = ""
				var reactModel models.PostReaction
				if err := tx.Model(models.PostReaction{}).
					Where(map[string]interface{}{
						"is_recalled": false,
						"post_id":     post.ID,
						"account_id":  in.RequestAccountID,
					}).Find(&reactModel).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						interactionType = ""
					} else {
					}
				} else {
					interactionType = reactModel.ReactionType
				}

				displayPost.Reactions = postReaction
				displayPost.Medias = MediaDisplay
				displayPost.Shares = postShareData
				displayPost.Comments = postComment
				displayPost.InteractionType = interactionType

				returnedPost = append(returnedPost, &displayPost)
			}
		}

		break
	case false:
		if in.IsFriend == true {
			if err := tx.Model(models.Post{}).
				Where("is_self_deleted = ? AND is_deleted_by_admin = ? AND is_hidden = ? AND account_id = ? AND (privacy_status = ? OR privacy_status = ?)",
					false, false, false, in.TargetAccountID, "public", "friend_only").
				Order("created_at DESC").
				Limit(int(in.PageSize)).
				Offset(int(offset)).
				Find(&posts).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					posts = make([]models.Post, 0)
				} else {
					return &ps.GetWallPostListResponse{Error: err.Error()}, err
				}
			}

			if len(posts) > 0 {
				for _, post := range posts {
					var displayPost = ps.DisplayPost{
						PostID:                  uint64(post.ID),
						Content:                 strings.TrimSpace(post.Content),
						IsShared:                post.IsShared,
						SharePostID:             uint64(post.OriginalPostID),
						IsSelfDeleted:           post.IsSelfDeleted,
						IsDeletedByAdmin:        post.IsDeletedByAdmin,
						IsHidden:                post.IsHidden,
						IsContentEdited:         post.IsContentEdited,
						PrivacyStatus:           post.PrivacyStatus,
						CreatedAt:               post.CreatedAt.Unix(),
						IsPublishedLater:        post.IsPublishedLater,
						PublishedLaterTimestamp: post.PublishLaterTimestamp.Unix(),
						IsPublished:             post.IsPublished,
						AccountID:               uint64(post.AccountID),
					}

					if post.IsShared {
						var originPost models.Post
						var ShareData *ps.DisplayPost
						if err := tx.Model(models.Post{}).
							Where(map[string]interface{}{
								"is_self_deleted":     false,
								"is_deleted_by_admin": false,
								"is_hidden":           false,
								"id":                  post.OriginalPostID,
								"privacy_status":      "public",
							}, post.OriginalPostID).First(&originPost).Error; err != nil {
							if errors.Is(err, gorm.ErrRecordNotFound) {
								ShareData = &ps.DisplayPost{
									Error: "Post not found",
								}
							} else {
								ShareData = &ps.DisplayPost{
									Error: err.Error(),
								}
							}
						} else {

							var medias []models.PostMultiMedia
							var mediasDisplay = make([]*ps.MediaDisplay, 0)

							if err := tx.Model(models.PostMultiMedia{}).Where(
								map[string]interface{}{
									"is_self_deleted":     false,
									"is_deleted_by_admin": false,
									"upload_status":       "uploaded",
									"post_id":             post.OriginalPostID,
								}).Find(&medias).Error; err != nil {
								if errors.Is(err, gorm.ErrRecordNotFound) {
									medias = make([]models.PostMultiMedia, 0)
								} else {
								}
							}

							for _, media := range medias {
								mediasDisplay = append(mediasDisplay, &ps.MediaDisplay{
									URL:     media.URL,
									MediaID: uint64(media.ID),
								})
							}

							ShareData = &ps.DisplayPost{
								PostID:          uint64(originPost.ID),
								Content:         strings.TrimSpace(originPost.Content),
								IsContentEdited: originPost.IsContentEdited,
								PrivacyStatus:   originPost.PrivacyStatus,
								CreatedAt:       originPost.CreatedAt.Unix(),
								IsPublished:     originPost.IsPublished,
								Medias:          mediasDisplay,
								AccountID:       uint64(originPost.AccountID),
							}
						}
						displayPost.SharePostData = ShareData
					} else {
						displayPost.SharePostData = nil
					}

					if !post.IsShared {
						displayPost.SharePostData = nil
					}

					var Media []models.PostMultiMedia
					var MediaDisplay []*ps.MediaDisplay

					if err := tx.Model(models.PostMultiMedia{}).Where(
						map[string]interface{}{
							"is_self_deleted":     false,
							"is_deleted_by_admin": false,
							"upload_status":       "uploaded",
							"post_id":             post.ID,
						}).Find(&Media).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							MediaDisplay = make([]*ps.MediaDisplay, 0)
							Media = make([]models.PostMultiMedia, 0)
						} else {
						}
					}

					for _, media := range Media {
						MediaDisplay = append(MediaDisplay, &ps.MediaDisplay{
							URL:     media.URL,
							MediaID: uint64(media.ID),
						})
					}

					var reactions []models.PostReaction
					postReaction := &ps.PostReactions{}

					if err := tx.Model(models.PostReaction{}).
						Where(map[string]interface{}{
							"is_recalled": false,
							"post_id":     post.ID,
						}).Find(&reactions).Error; err != nil {
					}

					if uint32(len(reactions)) > 0 {
						postReaction.DisplayData = make([]*ps.ReactionDisplay, 0, len(reactions))
						for _, reaction := range reactions {
							postReaction.DisplayData = append(postReaction.DisplayData, &ps.ReactionDisplay{
								ReactionType: reaction.ReactionType,
								AccountID:    uint64(reaction.AccountID),
							})
						}
						postReaction.Count = uint32(len(reactions))
					}

					var postShares []models.Post
					postShareData := &ps.PostShares{}

					if err := tx.Model(models.Post{}).Where(
						map[string]interface{}{
							"is_self_deleted":     false,
							"is_deleted_by_admin": false,
							"is_published":        true,
							"privacy_status":      "public",
							"original_post_id":    post.ID,
							"is_hidden":           false,
						}).Find(&postShares).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							postShares = make([]models.Post, 0)
						} else {
						}
					}

					if len(postShares) > 0 {
						for _, postShare := range postShares {
							postShareData.DisplayData = append(postShareData.DisplayData, &ps.ShareDisplay{
								AccountID: uint64(postShare.AccountID),
								CreatedAt: postShare.CreatedAt.Unix(),
							})
						}
						postShareData.Count = uint32(len(postShares))
					}

					var countComment int64
					if err := tx.Model(&models.PostComment{}).
						Where("post_id = ? AND is_self_deleted = ? AND is_deleted_by_admin = ?", post.ID, false, false).
						Count(&countComment).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							countComment = 0
						} else {
							countComment = 0
						}
					}

					if countComment < 0 {
						countComment = 0
					}

					postComment := &ps.PostComments{
						Count: uint32(countComment),
					}

					var interactionType = ""
					var reactModel models.PostReaction
					if err := tx.Model(models.PostReaction{}).
						Where(map[string]interface{}{
							"is_recalled": false,
							"post_id":     post.ID,
							"account_id":  in.RequestAccountID,
						}).Find(&reactModel).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							interactionType = ""
						} else {
						}
					} else {
						interactionType = reactModel.ReactionType
					}

					displayPost.Reactions = postReaction
					displayPost.Medias = MediaDisplay
					displayPost.Shares = postShareData
					displayPost.Comments = postComment
					displayPost.InteractionType = interactionType

					returnedPost = append(returnedPost, &displayPost)
				}
			}
		} else if in.IsFriend == false {
			if err := tx.Model(models.Post{}).
				Where(map[string]interface{}{
					"is_self_deleted":     false,
					"is_deleted_by_admin": false,
					"is_hidden":           false,
					"account_id":          in.TargetAccountID,
					"privacy_status":      "public",
				}).
				Order("created_at DESC").
				Limit(int(in.PageSize)).
				Offset(int(offset)).
				Find(&posts).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					posts = make([]models.Post, 0)
				} else {
					return &ps.GetWallPostListResponse{Error: err.Error()}, err
				}
			}

			if len(posts) > 0 {
				for _, post := range posts {
					var displayPost = ps.DisplayPost{
						PostID:                  uint64(post.ID),
						Content:                 strings.TrimSpace(post.Content),
						IsShared:                post.IsShared,
						SharePostID:             uint64(post.OriginalPostID),
						IsSelfDeleted:           post.IsSelfDeleted,
						IsDeletedByAdmin:        post.IsDeletedByAdmin,
						IsHidden:                post.IsHidden,
						IsContentEdited:         post.IsContentEdited,
						PrivacyStatus:           post.PrivacyStatus,
						CreatedAt:               post.CreatedAt.Unix(),
						IsPublishedLater:        post.IsPublishedLater,
						PublishedLaterTimestamp: post.PublishLaterTimestamp.Unix(),
						IsPublished:             post.IsPublished,
						AccountID:               uint64(post.AccountID),
					}

					if post.IsShared {
						var originPost models.Post
						var ShareData *ps.DisplayPost
						if err := tx.Model(models.Post{}).
							Where(map[string]interface{}{
								"is_self_deleted":     false,
								"is_deleted_by_admin": false,
								"is_hidden":           false,
								"id":                  post.OriginalPostID,
								"privacy_status":      "public",
							}, post.OriginalPostID).First(&originPost).Error; err != nil {
							if errors.Is(err, gorm.ErrRecordNotFound) {
								ShareData = &ps.DisplayPost{
									Error: "Post not found",
								}
							} else {
								ShareData = &ps.DisplayPost{
									Error: err.Error(),
								}
							}
						} else {

							var medias []models.PostMultiMedia
							var mediasDisplay = make([]*ps.MediaDisplay, 0)

							if err := tx.Model(models.PostMultiMedia{}).Where(
								map[string]interface{}{
									"is_self_deleted":     false,
									"is_deleted_by_admin": false,
									"upload_status":       "uploaded",
									"post_id":             post.OriginalPostID,
								}).Find(&medias).Error; err != nil {
								if errors.Is(err, gorm.ErrRecordNotFound) {
									medias = make([]models.PostMultiMedia, 0)
								} else {
								}
							}

							for _, media := range medias {
								mediasDisplay = append(mediasDisplay, &ps.MediaDisplay{
									URL:     media.URL,
									MediaID: uint64(media.ID),
								})
							}

							ShareData = &ps.DisplayPost{
								PostID:          uint64(originPost.ID),
								Content:         strings.TrimSpace(originPost.Content),
								IsContentEdited: originPost.IsContentEdited,
								PrivacyStatus:   originPost.PrivacyStatus,
								CreatedAt:       originPost.CreatedAt.Unix(),
								IsPublished:     originPost.IsPublished,
								Medias:          mediasDisplay,
								AccountID:       uint64(post.AccountID),
							}
						}
						displayPost.SharePostData = ShareData
					} else {
						displayPost.SharePostData = nil
					}

					if !post.IsShared {
						displayPost.SharePostData = nil
					}

					var Media []models.PostMultiMedia
					var MediaDisplay []*ps.MediaDisplay

					if err := tx.Model(models.PostMultiMedia{}).Where(
						map[string]interface{}{
							"is_self_deleted":     false,
							"is_deleted_by_admin": false,
							"upload_status":       "uploaded",
							"post_id":             post.ID,
						}).Find(&Media).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							MediaDisplay = make([]*ps.MediaDisplay, 0)
							Media = make([]models.PostMultiMedia, 0)
						} else {
						}
					}

					for _, media := range Media {
						MediaDisplay = append(MediaDisplay, &ps.MediaDisplay{
							URL:     media.URL,
							MediaID: uint64(media.ID),
						})
					}

					var reactions []models.PostReaction
					postReaction := &ps.PostReactions{}

					if err := tx.Model(models.PostReaction{}).
						Where(map[string]interface{}{
							"is_recalled": false,
							"post_id":     post.ID,
						}).Find(&reactions).Error; err != nil {
					}

					if uint32(len(reactions)) > 0 {
						postReaction.DisplayData = make([]*ps.ReactionDisplay, 0, len(reactions))
						for _, reaction := range reactions {
							postReaction.DisplayData = append(postReaction.DisplayData, &ps.ReactionDisplay{
								ReactionType: reaction.ReactionType,
								AccountID:    uint64(reaction.AccountID),
							})
						}
						postReaction.Count = uint32(len(reactions))
					}

					var postShares []models.Post
					postShareData := &ps.PostShares{}

					if err := tx.Model(models.Post{}).Where(
						map[string]interface{}{
							"is_self_deleted":     false,
							"is_deleted_by_admin": false,
							"is_published":        true,
							"privacy_status":      "public",
							"original_post_id":    post.ID,
							"is_hidden":           false,
						}).Find(&postShares).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							postShares = make([]models.Post, 0)
						} else {
						}
					}

					if len(postShares) > 0 {
						for _, postShare := range postShares {
							postShareData.DisplayData = append(postShareData.DisplayData, &ps.ShareDisplay{
								AccountID: uint64(postShare.AccountID),
								CreatedAt: postShare.CreatedAt.Unix(),
							})
						}
						postShareData.Count = uint32(len(postShares))
					}

					var countComment int64
					if err := tx.Model(&models.PostComment{}).
						Where("post_id = ? AND is_self_deleted = ? AND is_deleted_by_admin = ?", post.ID, false, false).
						Count(&countComment).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							countComment = 0
						} else {
							countComment = 0
						}
					}

					if countComment < 0 {
						countComment = 0
					}

					postComment := &ps.PostComments{
						Count: uint32(countComment),
					}

					var interactionType = ""
					var reactModel models.PostReaction
					if err := tx.Model(models.PostReaction{}).
						Where(map[string]interface{}{
							"is_recalled": false,
							"post_id":     post.ID,
							"account_id":  in.RequestAccountID,
						}).Find(&reactModel).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							interactionType = ""
						} else {
						}
					} else {
						interactionType = reactModel.ReactionType
					}

					displayPost.Reactions = postReaction
					displayPost.Medias = MediaDisplay
					displayPost.Shares = postShareData
					displayPost.Comments = postComment
					displayPost.InteractionType = interactionType

					returnedPost = append(returnedPost, &displayPost)
				}
			}
		}
		break
	default:
	}

	if tx.Commit().Error != nil {
		return &ps.GetWallPostListResponse{Error: tx.Commit().Error.Error()}, tx.Commit().Error
	}

	return &ps.GetWallPostListResponse{
		TargetAccountID: in.TargetAccountID,
		Page:            in.Page,
		Posts:           returnedPost,
	}, nil
}

// rankPostInteractions will rank posts based on their interactions (likes, comments, shares, etc.)
func (s *PostService) rankPostInteractions(ctx context.Context) map[int]int {
	// Step 1: Query all published posts
	var posts []models.Post
	err := s.DB.WithContext(ctx).
		Where("is_deleted_by_admin = ? AND is_hidden = ? AND is_published = ?", false, false, true).
		Find(&posts).Error
	if err != nil {
		fmt.Println("Error fetching posts:", err)
		return nil
	}

	// Step 2: Prepare a map to hold scores
	postScores := make(map[int]int) // Map[PostID]Score

	// Step 3: Calculate scores for each post
	for _, post := range posts {
		score := 0

		// Step 3.1: Calculate reactions score
		var reactionsCount int64
		err = s.DB.WithContext(ctx).
			Model(&models.PostReaction{}).
			Where("post_id = ? AND is_recalled = ?", post.ID, false).
			Count(&reactionsCount).Error
		if err != nil {
			fmt.Println("Error counting reactions:", err)
			continue
		}
		score += int(reactionsCount) * 10

		// Step 3.2: Calculate comments score
		var commentsCount int64
		err = s.DB.WithContext(ctx).
			Model(&models.PostComment{}).
			Where("post_id = ? AND is_self_deleted = ?", post.ID, false).
			Count(&commentsCount).Error
		if err != nil {
			fmt.Println("Error counting comments:", err)
			continue
		}
		score += int(commentsCount) * 15

		// Step 3.3: Calculate time decay factor
		timeSincePublished := time.Now().Sub(post.CreatedAt).Seconds()
		if timeSincePublished < 0 {
			timeSincePublished = 0 // Clamp negative values
		}
		timeFactor := 10000 / (1 + timeSincePublished)
		score += int(timeFactor)

		// Step 3.4: Shared post bonus
		if post.IsShared {
			score += 10
		}

		// Add score to the map
		postScores[int(post.ID)] = score
	}

	// Step 4: Return the map of post scores
	return postScores
}

func (s *PostService) GetNewFeeds(ctx context.Context, in *ps.GetNewFeedsRequest) (*ps.GetNewFeedsResponse, error) {
	if in.AccountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	seenPostMap := make(map[uint64]struct{}, len(in.SeenPostIDs))
	for _, id := range in.SeenPostIDs {
		seenPostMap[id] = struct{}{}
	}

	allRankingPost := s.rankPostInteractions(ctx)
	interactions := in.Interactions
	listFriendIds := in.ListFriendIDs

	log.Printf("Original Array: %v", allRankingPost)

	type postScore struct {
		PostID uint
		Score  int
	}

	var availablePosts []postScore
	for postID, score := range allRankingPost {
		if _, seen := seenPostMap[uint64(postID)]; !seen {
			availablePosts = append(availablePosts, postScore{PostID: uint(postID), Score: score})
		}
	}

	log.Printf("Sort Array By Seen: %v", availablePosts)

	for i := 0; i < len(availablePosts); {
		postID := availablePosts[i].PostID
		var postData models.Post
		if err := s.DB.Model(&models.Post{}).Where("id = ?", postID).First(&postData).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				availablePosts = append(availablePosts[:i], availablePosts[i+1:]...)
				continue
			} else {
				return nil, err
			}
		}
		score, found := getScoreByAccountID(interactions, uint64(postData.AccountID))
		if found {
			availablePosts[i].Score *= int(score)
		}
		i++
	}

	sort.Slice(availablePosts, func(i, j int) bool {
		return availablePosts[i].Score > availablePosts[j].Score
	})

	log.Printf("Sort Array by DESC: %v", availablePosts)

	if uint32(len(availablePosts)) > in.PageSize {
		availablePosts = availablePosts[:in.PageSize]
	}

	log.Printf("Sort Array by DESC and get 10 items: %v", availablePosts)

	var responseDisplayPost []*ps.DisplayPost

	for _, post := range availablePosts {
		postID := post.PostID
		var postData models.Post
		if err := s.DB.Model(&models.Post{}).Where("id = ?", postID).First(&postData).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			} else {
				return nil, err
			}
		}

		if postData.PrivacyStatus == "private" && uint64(postData.AccountID) != in.AccountID {
			continue
		}

		if (postData.PrivacyStatus == "friend_only") && (uint64(postData.AccountID) != in.AccountID || !checkIsFriend(listFriendIds, uint64(postData.AccountID))) {
			continue
		}

		var displayPost = ps.DisplayPost{
			PostID:                  uint64(postData.ID),
			Content:                 strings.TrimSpace(postData.Content),
			IsShared:                postData.IsShared,
			SharePostID:             uint64(postData.OriginalPostID),
			IsSelfDeleted:           postData.IsSelfDeleted,
			IsDeletedByAdmin:        postData.IsDeletedByAdmin,
			IsHidden:                postData.IsHidden,
			IsContentEdited:         postData.IsContentEdited,
			PrivacyStatus:           postData.PrivacyStatus,
			CreatedAt:               postData.CreatedAt.Unix(),
			IsPublishedLater:        postData.IsPublishedLater,
			PublishedLaterTimestamp: postData.PublishLaterTimestamp.Unix(),
			IsPublished:             postData.IsPublished,
			AccountID:               uint64(postData.AccountID),
		}

		if postData.IsShared {
			var originPost models.Post
			var ShareData *ps.DisplayPost
			if err := s.DB.Model(models.Post{}).
				Where(map[string]interface{}{
					"id":             postData.OriginalPostID,
					"privacy_status": "public",
				}, postData.OriginalPostID).First(&originPost).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					ShareData = &ps.DisplayPost{
						Error: "Post not found",
					}
				} else {
					ShareData = &ps.DisplayPost{
						Error: err.Error(),
					}
				}
			} else {

				var medias []models.PostMultiMedia
				var mediasDisplay = make([]*ps.MediaDisplay, 0)

				if err := s.DB.Model(models.PostMultiMedia{}).Where(
					map[string]interface{}{
						"is_self_deleted":     false,
						"is_deleted_by_admin": false,
						"upload_status":       "uploaded",
						"post_id":             uint64(postData.OriginalPostID),
					}).Find(&medias).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						medias = make([]models.PostMultiMedia, 0)
					} else {
					}
				}

				for _, media := range medias {
					mediasDisplay = append(mediasDisplay, &ps.MediaDisplay{
						URL:     media.URL,
						MediaID: uint64(media.ID),
					})
				}

				ShareData = &ps.DisplayPost{
					PostID:          uint64(originPost.ID),
					Content:         strings.TrimSpace(originPost.Content),
					IsContentEdited: originPost.IsContentEdited,
					PrivacyStatus:   originPost.PrivacyStatus,
					CreatedAt:       originPost.CreatedAt.Unix(),
					IsPublished:     originPost.IsPublished,
					Medias:          mediasDisplay,
					AccountID:       uint64(originPost.AccountID),
				}
			}
			displayPost.SharePostData = ShareData
		} else {
			displayPost.SharePostData = nil
		}

		if !postData.IsShared {
			displayPost.SharePostData = nil
		}

		var Media []models.PostMultiMedia
		var MediaDisplay []*ps.MediaDisplay

		if err := s.DB.Model(models.PostMultiMedia{}).Where(
			map[string]interface{}{
				"is_self_deleted":     false,
				"is_deleted_by_admin": false,
				"upload_status":       "uploaded",
				"post_id":             uint64(postData.ID),
			}).Find(&Media).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				MediaDisplay = make([]*ps.MediaDisplay, 0)
				Media = make([]models.PostMultiMedia, 0)
			} else {
			}
		}

		for _, media := range Media {
			MediaDisplay = append(MediaDisplay, &ps.MediaDisplay{
				URL:     media.URL,
				MediaID: uint64(media.ID),
			})
		}

		var reactions []models.PostReaction
		postReaction := &ps.PostReactions{}

		if err := s.DB.Model(models.PostReaction{}).
			Where(map[string]interface{}{
				"is_recalled": false,
				"post_id":     postData.ID,
			}).Find(&reactions).Error; err != nil {
		}

		if uint32(len(reactions)) > 0 {
			postReaction.DisplayData = make([]*ps.ReactionDisplay, 0, len(reactions))
			for _, reaction := range reactions {
				postReaction.DisplayData = append(postReaction.DisplayData, &ps.ReactionDisplay{
					ReactionType: reaction.ReactionType,
					AccountID:    uint64(reaction.AccountID),
				})
			}
			postReaction.Count = uint32(len(reactions))
			fmt.Printf("Post Reaction Data: Count=%d, DisplayData=%+v\n",
				postReaction.Count, postReaction.DisplayData)
		}

		var postShares []models.Post
		postShareData := &ps.PostShares{}

		if err := s.DB.Model(models.Post{}).Where(
			map[string]interface{}{
				"is_self_deleted":     false,
				"is_deleted_by_admin": false,
				"is_published":        true,
				"privacy_status":      "public",
				"original_post_id":    postData.ID,
				"is_hidden":           false,
			}).Find(&postShares).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				postShares = make([]models.Post, 0)
			} else {
			}
		}

		if len(postShares) > 0 {
			for _, postShare := range postShares {
				postShareData.DisplayData = append(postShareData.DisplayData, &ps.ShareDisplay{
					AccountID: uint64(postShare.AccountID),
					CreatedAt: postShare.CreatedAt.Unix(),
				})
			}
			postShareData.Count = uint32(len(postShares))
		}

		var countComment int64
		if err := s.DB.Model(&models.PostComment{}).
			Where("post_id = ? AND is_self_deleted = ? AND is_deleted_by_admin = ?", postData.ID, false, false).
			Count(&countComment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				countComment = 0
			} else {
				countComment = 0
			}
		}

		if countComment < 0 {
			countComment = 0
		}

		postComment := &ps.PostComments{
			Count: uint32(countComment),
		}

		var interactionType = ""
		var reactModel models.PostReaction
		if err := s.DB.Model(models.PostReaction{}).
			Where(map[string]interface{}{
				"is_recalled": false,
				"post_id":     postData.ID,
				"account_id":  in.AccountID,
			}).Find(&reactModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				interactionType = ""
			} else {
			}
		} else {
			interactionType = reactModel.ReactionType
		}

		displayPost.Reactions = postReaction
		displayPost.Medias = MediaDisplay
		displayPost.Shares = postShareData
		displayPost.Comments = postComment
		displayPost.InteractionType = interactionType

		responseDisplayPost = append(responseDisplayPost, &displayPost)
	}

	var response = &ps.GetNewFeedsResponse{
		AccountID: in.AccountID,
		Page:      in.Page,
		PageSize:  in.PageSize,
		Posts:     responseDisplayPost,
	}

	return response, nil
}

func getScoreByAccountID(interactions []*ps.InteractionScore, accountID uint64) (uint64, bool) {
	for _, interaction := range interactions {
		if interaction.AccountID == accountID {
			return interaction.Score, true
		}
	}
	return 0, false
}

func checkIsFriend(friendListIDs []uint64, accountID uint64) bool {
	for _, id := range friendListIDs {
		if id == accountID {
			return true
		}
	}
	return false
}
