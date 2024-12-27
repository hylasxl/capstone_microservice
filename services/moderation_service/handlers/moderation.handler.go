package handlers

import (
	"context"
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
