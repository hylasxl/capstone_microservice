package handlers

import (
	"gateway/proto/auth_service"
	"gateway/proto/post_service"
)

type CreatePostRequest struct {
	AccountID               string              `json:"account_id"`
	Content                 string              `json:"content"`
	IsPublishedLater        bool                `json:"is_published_later"`
	PublishedLaterTimestamp int64               `json:"published_later_timestamp"`
	PrivacyStatus           string              `json:"privacy_status"`
	TagAccountIDs           []string            `json:"tag_account_ids"`
	Medias                  []MultiMediaMessage `json:"medias"`
}

type MultiMediaMessage struct {
	Type         string `json:"type"`
	UploadStatus string `json:"upload_status"`
	Content      string `json:"content"`
	Media        []byte `json:"media"`
}

type CreatePostResponse struct {
	PostID    string   `json:"post_id"`
	MediaURLs []string `json:"media_urls"`
}

type SharePostRequest struct {
	AccountID      string   `json:"account_id"`
	Content        string   `json:"content"`
	IsShared       bool     `json:"is_shared"`
	OriginalPostID string   `json:"original_post_id"`
	PrivacyStatus  string   `json:"privacy_status"`
	TagAccountIDs  []string `json:"tag_account_ids"`
}

type SharePostResponse struct {
	PostID string `json:"post_id"`
}

type SendFriendRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
}

type SendFriendResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type ResolveFriendRequest struct {
	ReceiverAccountID string `json:"receiver_account_id"`
	RequestID         string `json:"request_id"`
	Action            string `json:"action"`
}

type ResolveFriendResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type RecallRequest struct {
	SenderAccountID string `json:"sender_account_id"`
	RequestID       string `json:"request_id"`
}

type RecallResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type UnfriendRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
}

type UnfriendResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type FollowRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Action        string `json:"action"`
}

type FollowResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type BlockRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Action        string `json:"action"`
}

type BlockResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type GetPendingListRequest struct {
	AccountID string `json:"account_id"`
	Page      int    `json:"page"`
}

type GetPendingListResponse struct {
	Page  int                              `json:"page"`
	Data  []GetPendingListReturnSingleLine `json:"data"`
	Error string                           `json:"error"`
}

type GetPendingListReturnSingleLine struct {
	AccountInfo   SingleAccountInfo `json:"account_info"`
	RequestID     string            `json:"request_id"`
	CreatedAt     int64             `json:"created_at"`
	MutualFriends int64             `json:"mutual_friends"`
}

type SingleAccountInfo struct {
	AccountID   uint   `json:"account_id"`
	AvatarURL   string `json:"avatar_url"`
	DisplayName string `json:"display_name"`
}

type GetListFriendIDs struct {
	AccountID string `json:"account_id"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string                  `json:"access_token"`
	RefreshToken string                  `json:"refresh_token"`
	UserID       string                  `json:"user_id"`
	JWTClaims    *auth_service.JWTClaims `json:"jwt_claims"`
	Success      bool                    `json:"success"`
}

type SignUpRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Image     string `json:"image"`
}

type SignUpResponse struct {
	UserID  string `json:"user_id"`
	Success bool   `json:"success"`
}

type CheckDuplicateRequest struct {
	Data     string `json:"data"`
	DataType string `json:"data_type"`
}

type CheckDuplicateResponse struct {
	IsDuplicate bool `json:"is_duplicate"`
}

type CheckValidUserRequest struct {
	AccountID string `json:"account_id"`
}

type CheckValidUserResponse struct {
	IsValid bool `json:"is_valid"`
}

type CommentPostRequest struct {
	AccountID uint64 `json:"account_id"`
	PostID    uint64 `json:"post_id"`
	Content   string `json:"content"`
}

type CommentPostResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type ReplyCommentRequest struct {
	AccountID         uint64 `json:"account_id"`
	Content           string `json:"content"`
	OriginalCommentID uint64 `json:"original_comment_id"`
	PostID            uint64 `json:"post_id"`
}

type ReplyCommentResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type GetSinglePostRequest struct {
	PostID uint64 `json:"post_id"`
}

type GetSinglePostResponse struct {
	PostID              uint64                       `json:"post_id"`
	Content             string                       `json:"content"`
	PrivacyStatus       string                       `json:"privacy_status"`
	Medias              []*post_service.MediaDisplay `json:"medias"`
	TotalCommentNumber  uint64                       `json:"total_comment_number"`
	TotalReactionNumber uint64                       `json:"total_reaction_number"`
	TotalShareNumber    uint64                       `json:"total_share_number"`
	Error               string                       `json:"error"`
	Success             bool                         `json:"success"`
}

type DeletePostRequest struct {
	PostID uint64 `json:"post_id"`
}

type DeletePostResponse struct {
	PostID  uint64 `json:"post_id"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type EditPostCommentRequest struct {
	CommentID uint64 `json:"comment_id"`
	Content   string `json:"content"`
}

type EditPostCommentResponse struct {
	CommentID uint64 `json:"comment_id"`
	Error     string `json:"error"`
	Success   bool   `json:"success"`
}

type DeletePostCommentRequest struct {
	CommentID uint64 `json:"comment_id"`
}
type DeletePostCommentResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type DeletePostImageRequest struct {
	PostID  uint64 `json:"post_id"`
	MediaID uint64 `json:"media_id"`
}

type DeletePostImageResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type ReactPostRequest struct {
	PostID    uint64 `json:"post_id"`
	AccountID uint64 `json:"account_id"`
	ReactType string `json:"react_type"`
}

type ReactPostResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type RemoveReactPostRequest struct {
	PostID    uint64 `json:"post_id"`
	AccountID uint64 `json:"account_id"`
}

type RemoveReactPostResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type ReactImageRequest struct {
	MediaID   uint64 `json:"media_id"`
	AccountID uint64 `json:"account_id"`
	ReactType string `json:"react_type"`
}

type ReactImageResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type RemoveReactImageRequest struct {
	MediaID   uint64 `json:"media_id"`
	AccountID uint64 `json:"account_id"`
}

type RemoveReactImageResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type CommentImageRequest struct {
	MediaID   uint64 `json:"media_id"`
	AccountID uint64 `json:"account_id"`
	Content   string `json:"content"`
}

type CommentImageResponse struct {
	CommentID uint64 `json:"comment_id"`
	Error     string `json:"error"`
	Success   bool   `json:"success"`
}

type ReplyCommentImageRequest struct {
	MediaID           uint64 `json:"media_id"`
	AccountID         uint64 `json:"account_id"`
	Content           string `json:"content"`
	OriginalCommentID uint64 `json:"original_comment_id"`
}

type ReplyCommentImageResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type EditCommentImageRequest struct {
	CommentID uint64 `json:"comment_id"`
	Content   string `json:"content"`
	AccountID uint64 `json:"account_id"`
}

type EditCommentImageResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type DeleteCommentImageRequest struct {
	CommentID uint64 `json:"comment_id"`
	AccountID uint64 `json:"account_id"`
	MediaID   uint64 `json:"media_id"`
}

type DeleteCommentImageResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type GetPostCommentsRequest struct {
	PostID   uint64 `json:"post_id"`
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
}

type GetPostCommentsResponse struct {
	Error              string                  `json:"error"`
	Success            bool                    `json:"success"`
	PostID             uint64                  `json:"post_id"`
	TotalCommentNumber uint64                  `json:"total_comment_number"`
	Comments           []*post_service.Comment `json:"comments"`
}

type GetAccountInfoRequest struct {
	AccountID uint64 `json:"account_id"`
}

type GetAccountInfoResponse struct {
	AccountID     uint64        `json:"account_id"`
	Account       Account       `json:"account"`
	AccountInfo   AccountInfo   `json:"account_info"`
	AccountAvatar AccountAvatar `json:"account_avatar"`
}

type Account struct {
	Username      string `json:"username"`
	RoleID        uint64 `json:"role_id"`
	CreateMethod  string `json:"create_method"`
	IsBanned      bool   `json:"is_banned"`
	IsRestricted  bool   `json:"is_restricted"`
	IsSelfDeleted bool   `json:"is_self_deleted"`
}

type AccountInfo struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	DateOfBirth     int64  `json:"date_of_birth"`
	Gender          string `json:"gender"`
	MaterialStatus  string `json:"material_status"`
	PhoneNumber     string `json:"phone_number"`
	Email           string `json:"email"`
	NameDisplayType string `json:"name_display_type"`
}

type AccountAvatar struct {
	AvatarID  uint64 `json:"avatar_id"`
	AvatarURL string `json:"avatar_url"`
	IsInUse   bool   `json:"is_in_use"`
	IsDeleted bool   `json:"is_deleted"`
}

type CountPendingFriendRequest struct {
	AccountID uint64 `json:"account_id"`
}

type CountPendingFriendResponse struct {
	Quantity uint64 `json:"quantity"`
	Error    string `json:"error"`
}

type GetWallPostListRequest struct {
	TargetAccountID  uint64 `json:"target_account_id"`
	RequestAccountID uint64 `json:"request_account_id"`
	Page             uint64 `json:"page"`
	PageSize         uint64 `json:"page_size"`
}

type GetWallPostListResponse struct {
	TargetAccountID uint64        `json:"target_account_id"`
	Page            uint64        `json:"page"`
	PageSize        uint64        `json:"page_size"`
	Posts           []DisplayPost `json:"posts"`
	Error           string        `json:"error"`
}

type DisplayPost struct {
	PostID                  uint64                  `json:"post_id"`
	Content                 string                  `json:"content"`
	IsShared                bool                    `json:"is_shared"`
	SharePostID             uint64                  `json:"share_post_id"`
	SharePostData           SharePostDataDisplay    `json:"share_post_data"`
	IsHidden                bool                    `json:"is_hidden"`
	IsContentEdited         bool                    `json:"is_content_edited"`
	PrivacyStatus           string                  `json:"privacy_status"`
	InteractionType         string                  `json:"interaction_type"`
	Medias                  []PostShareMediaDisplay `json:"medias"`
	CreatedAt               int64                   `json:"created_at"`
	IsPublishedLater        bool                    `json:"is_published_later"`
	PublishedLaterTimestamp int64                   `json:"published_later_timestamp"`
	IsPublished             bool                    `json:"is_published"`
	Reactions               PostReactionDisplay     `json:"reactions"`
	CommentQuantity         PostCommentDisplay      `json:"comment_quantity"`
	Shares                  PostShareDisplay        `json:"shares"`
	Error                   string                  `json:"error"`
	Account                 SingleAccountInfo       `json:"account"`
}

type SharePostDataDisplay struct {
	PostID          uint64                  `json:"post_id"`
	Content         string                  `json:"content"`
	IsContentEdited bool                    `json:"is_content_edited"`
	PrivacyStatus   string                  `json:"privacy_status"`
	CreatedAt       int64                   `json:"created_at"`
	IsPublished     bool                    `json:"is_published"`
	Medias          []PostShareMediaDisplay `json:"medias"`
	Account         SingleAccountInfo       `json:"account"`
}

type PostShareMediaDisplay struct {
	URL     string `json:"url"`
	Content string `json:"content"`
	MediaID uint64 `json:"media_id"`
}

type PostReactionDisplay struct {
	TotalQuantity uint64             `json:"total_quantity"`
	Reactions     []PostReactionData `json:"reactions"`
}

type PostReactionData struct {
	ReactionType string            `json:"reaction_type"`
	Account      SingleAccountInfo `json:"account"`
}

type PostCommentDisplay struct {
	TotalQuantity uint64 `json:"total_quantity"`
}

type PostShareDisplay struct {
	TotalQuantity uint64          `json:"total_quantity"`
	Shares        []PostShareData `json:"shares"`
}

type PostShareData struct {
	Account   SingleAccountInfo `json:"account"`
	CreatedAt int64             `json:"created_at"`
}
