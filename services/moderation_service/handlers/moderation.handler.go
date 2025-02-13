package handlers

import (
	"context"
	"errors"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
	"moderation_service/models"
	ms "moderation_service/proto/moderation_service"
	"regexp"
	"strings"
)

type ModerationService struct {
	ms.ModerationServiceServer
	DB *gorm.DB
}

func (s *ModerationService) IdentifyAndReplaceText(ctx context.Context, in *ms.IdentifyAndReplaceTextRequest) (*ms.IdentifyAndReplaceTextResponse, error) {

	tx := s.DB.Begin()

	if len(strings.TrimSpace(in.Content)) == 0 {
		return &ms.IdentifyAndReplaceTextResponse{
			Error: "Invalid content",
		}, nil
	}

	var bannedWords []string
	if err := tx.Model(&models.BanWord{}).Where("is_deleted = ? ", false).Pluck("content", &bannedWords).Error; err != nil {
		tx.Rollback()
		return &ms.IdentifyAndReplaceTextResponse{
			Error: "Failed to get banned words",
		}, nil
	}

	modifiedContent := in.Content

	normalizedContent := norm.NFD.String(modifiedContent)

	for _, word := range bannedWords {
		normalizedWord := norm.NFD.String(word)
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(normalizedWord) + `\b`)
		replacement := func(match string) string {
			if len(match) > 1 {
				return string(match[0]) + strings.Repeat("*", len(match)-1)
			}
			return match
		}

		normalizedContent = re.ReplaceAllStringFunc(normalizedContent, replacement)
	}

	modifiedContent = normalizedContent

	if err := tx.Commit().Error; err != nil {
		return &ms.IdentifyAndReplaceTextResponse{
			Error: "Failed to commit transaction",
		}, nil
	}

	return &ms.IdentifyAndReplaceTextResponse{
		ReturnedContent: modifiedContent,
	}, nil
}

func (s *ModerationService) HandleReportPost(ctx context.Context, in *ms.ReportPost) (*ms.SingleLineStatusResponse, error) {
	if in.PostID <= 0 || in.ReportedBy <= 0 || len(strings.TrimSpace(in.Reason)) == 0 {
		return nil, errors.New("invalid request")
	}

	var reportData = &models.ReportedPort{
		PostID:              uint(in.PostID),
		Reason:              strings.TrimSpace(in.Reason),
		ReportedByAccountID: uint(in.ReportedBy),
	}

	if err := s.DB.Create(reportData).Error; err != nil {
		return nil, err
	}

	return &ms.SingleLineStatusResponse{Success: true}, nil
}

func (s *ModerationService) HandleReportAccount(ctx context.Context, in *ms.ReportAccount) (*ms.SingleLineStatusResponse, error) {
	if in.AccountID <= 0 || in.ReportedBy <= 0 || len(strings.TrimSpace(in.Reason)) == 0 {
		return nil, errors.New("invalid request")
	}

	var reportData = &models.ReportedUser{
		AccountID:           uint(in.AccountID),
		ReportedByAccountID: uint(in.ReportedBy),
		Reason:              strings.TrimSpace(in.Reason),
	}

	if err := s.DB.Create(reportData).Error; err != nil {
		return nil, err
	}

	return &ms.SingleLineStatusResponse{Success: true}, nil
}

func (s *ModerationService) HandleResolveReportedPost(ctx context.Context, in *ms.ResolveReportedPost) (*ms.SingleLineStatusResponse, error) {
	if in.PostID <= 0 || in.ResolvedBy <= 0 || len(strings.TrimSpace(in.Method)) == 0 {
		return nil, errors.New("invalid request")
	}

	if in.Method != "report_skipped" && in.Method != "delete_post" {
		return nil, errors.New("invalid method")
	}

	if err := s.DB.Model(&models.ReportedPort{}).Where("post_id = ?", uint(in.PostID)).Updates(map[string]interface{}{
		"report_resolve":         in.ResolvedBy,
		"resolved_by_account_id": uint(in.ResolvedBy),
	}).Error; err != nil {
		return nil, err
	}

	return &ms.SingleLineStatusResponse{Success: true}, nil
}

func (s *ModerationService) HandleResolveReportedAccount(ctx context.Context, in *ms.ResolveReportedAccount) (*ms.SingleLineStatusResponse, error) {
	if in.AccountID <= 0 || in.ResolvedBy <= 0 || len(strings.TrimSpace(in.Method)) == 0 {
		return nil, errors.New("invalid request")
	}

	if in.Method != "report_skipped" && in.Method != "delete_user" {
		return nil, errors.New("invalid method")
	}

	if err := s.DB.Model(&models.ReportedUser{}).Where("account_id = ?", in.AccountID).Updates(
		map[string]interface{}{
			"report_resolve":         in.ResolvedBy,
			"resolved_by_account_id": uint(in.ResolvedBy),
		}).Error; err != nil {
		return nil, err
	}

	return &ms.SingleLineStatusResponse{Success: true}, nil
}

func (s *ModerationService) HandleGetReportAccountList(ctx context.Context, in *ms.GetReportedAccountListRequest) (*ms.GetReportedAccountListResponse, error) {
	// Validate input
	if in.Page == 0 || in.PageSize == 0 {
		return nil, errors.New("invalid page or page size")
	}

	// Calculate offset for pagination
	offset := (in.Page - 1) * in.PageSize

	// Query the database for reported users with custom sorting
	var reportedUsers []models.ReportedUser
	result := s.DB.
		Order(`CASE 
			WHEN report_resolve = 'report_pending' THEN 1 
			WHEN report_resolve = 'report_skipped' THEN 2 
			WHEN report_resolve = 'delete_user' THEN 3 
			ELSE 4 
		END`).
		Offset(int(offset)).
		Limit(int(in.PageSize)).
		Find(&reportedUsers)

	if result.Error != nil {
		return nil, result.Error
	}

	// Convert database results to response format without grouping
	var reportData []*ms.ReportAccountData
	for _, user := range reportedUsers {
		reportData = append(reportData, &ms.ReportAccountData{
			AccountID:     uint32(user.AccountID),
			ResolveStatus: user.ReportResolve,
			Reasons:       user.Reason, // Keep individual reports
		})
	}

	// Build the response
	response := &ms.GetReportedAccountListResponse{
		Page:     in.Page,
		PageSize: in.PageSize,
		Data:     reportData,
	}

	return response, nil
}

func (s *ModerationService) GetReportedPosts(ctx context.Context, in *ms.GetReportedPostRequest) (*ms.GetReportedPostResponse, error) {
	if in.Page == 0 || in.PageSize == 0 {
		return nil, errors.New("invalid page or page size")
	}

	// Calculate offset for pagination
	offset := (in.Page - 1) * in.PageSize

	// Query the database for reported users with custom sorting
	var reportedPosts []models.ReportedPort
	result := s.DB.
		Order(`CASE 
			WHEN report_resolve = 'report_pending' THEN 1 
			WHEN report_resolve = 'report_skipped' THEN 2 
			WHEN report_resolve = 'delete_post' THEN 3 
			ELSE 4 
		END`).
		Offset(int(offset)).
		Limit(int(in.PageSize)).
		Find(&reportedPosts)

	if result.Error != nil {
		return nil, result.Error
	}

	// Convert database results to response format without grouping
	var reportData []*ms.ReportedPostData
	for _, post := range reportedPosts {
		reportData = append(reportData, &ms.ReportedPostData{
			ID:            uint32(post.ID),
			PostID:        uint32(post.PostID),
			ResolveStatus: post.ReportResolve,
			Reason:        post.Reason, // Keep individual reports
		})
	}

	// Build the response
	response := &ms.GetReportedPostResponse{
		Page:          in.Page,
		PageSize:      in.PageSize,
		ReportedPosts: reportData,
	}

	return response, nil
}

func (s *ModerationService) GetListBanWords(ctx context.Context, in *ms.GetListBanWordsReq) (*ms.GetListBanWordsRes, error) {
	var banWords []models.BanWord

	// Fetch banned words from database where IsDeleted is false
	if err := s.DB.Where("is_deleted = ?", false).Find(&banWords).Error; err != nil {
		return nil, err
	}

	// Group words by LanguageCode
	languageMap := make(map[string][]*ms.WordData)
	for _, word := range banWords {
		languageMap[word.LanguageCode] = append(languageMap[word.LanguageCode], &ms.WordData{
			ID:   uint32(word.ID),
			Word: word.Content,
		})
	}

	// Convert map to slice of DataSet
	var dataSets []*ms.DataSet
	for lang, words := range languageMap {
		dataSets = append(dataSets, &ms.DataSet{
			LanguageCode: lang,
			Words:        words,
		})
	}

	return &ms.GetListBanWordsRes{Data: dataSets}, nil
}

func (s *ModerationService) EditWord(ctx context.Context, in *ms.EditWordReq) (*ms.EditWordRes, error) {
	var word models.BanWord

	if err := s.DB.First(&word, in.ID).Error; err != nil {
		return nil, err
	}

	word.Content = in.Content

	if err := s.DB.Save(&word).Error; err != nil {
		return nil, err
	}

	return &ms.EditWordRes{Success: true}, nil
}

func (s *ModerationService) DeleteWord(ctx context.Context, in *ms.DeleteWordReq) (*ms.DeleteWordRes, error) {
	if err := s.DB.Model(&models.BanWord{}).Where("id = ?", in.ID).Update("is_deleted", true).Error; err != nil {
		return nil, err
	}

	return &ms.DeleteWordRes{Success: true}, nil
}

func (s *ModerationService) AddWord(ctx context.Context, in *ms.AddWordReq) (*ms.AddWordRes, error) {
	var existingWord models.BanWord
	err := s.DB.Where("content = ? AND language_code = ?", in.Content, in.LanguageCode).First(&existingWord).Error

	if err == nil {
		return &ms.AddWordRes{Success: false}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	word := &models.BanWord{
		Content:            in.Content,
		CreatedByAccountID: uint(in.RequestAccountID),
		LanguageCode:       in.LanguageCode,
	}

	if err := s.DB.Create(word).Error; err != nil {
		return &ms.AddWordRes{Success: false}, nil
	}

	return &ms.AddWordRes{Success: true, ID: uint32(word.ID)}, nil
}
