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
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	RequestID int    `json:"request_id"`
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
	Error     string `json:"error"`
	CommentID int64  `json:"comment_id"`
	Success   bool   `json:"success"`
}

type ReplyCommentRequest struct {
	AccountID         uint64 `json:"account_id"`
	Content           string `json:"content"`
	OriginalCommentID uint64 `json:"original_comment_id"`
	PostID            uint64 `json:"post_id"`
}

type ReplyCommentResponse struct {
	CommentID int64  `json:"comment_id"`
	Error     string `json:"error"`
	Success   bool   `json:"success"`
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
	Bio             string `json:"bio"`
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

type PrivacyIndices struct {
	DateOfBirth    string `json:"date_of_birth"`
	Gender         string `json:"gender"`
	MaterialStatus string `json:"material_status"`
	PhoneNumber    string `json:"phone_number"`
	Email          string `json:"email"`
	Bio            string `json:"bio"`
}

type GetProfileInfoRequest struct {
	TargetAccountID  uint64 `json:"target_account_id"`
	RequestAccountID uint64 `json:"request_account_id"`
}
type GetProfileInfoResponse struct {
	AccountID      uint64         `json:"account_id"`
	Account        Account        `json:"account"`
	AccountInfo    AccountInfo    `json:"account_info"`
	AccountAvatar  AccountAvatar  `json:"account_avatar"`
	PrivacyIndices PrivacyIndices `json:"privacy_indices"`
	IsFriend       bool           `json:"is_friend"`
	IsBlocked      bool           `json:"is_blocked"`
	IsFollowed     bool           `json:"is_followed"`
	Timestamp      int64          `json:"timestamp"`
}

type GetNewFeedsRequest struct {
	AccountID  uint64   `json:"account_id"`
	SeenPostID []uint64 `json:"seen_post_id"`
	Page       uint32   `json:"page"`
	PageSize   uint32   `json:"page_size"`
}

type GetNewFeedsResponse struct {
	AccountID uint64        `json:"account_id"`
	Page      uint64        `json:"page"`
	PageSize  uint64        `json:"page_size"`
	Posts     []DisplayPost `json:"posts"`
}

type CheckExistingFriendRequestRequest struct {
	FromAccountID uint64 `json:"from_account_id"`
	ToAccountID   uint64 `json:"to_account_id"`
}

type CheckExistingFriendRequestResponse struct {
	IsExisting bool   `json:"is_existing"`
	RequestID  uint64 `json:"request_id"`
}

type RegisterDeviceRequest struct {
	UserID uint64 `json:"user_id"`
	Token  string `json:"token"`
}

type RegisterDeviceResponse struct {
	Success bool `json:"success"`
}

type SingleSuccessResponse struct {
	Success bool `json:"success"`
}

type SetPrivacyRequest struct {
	AccountID     uint64 `json:"account_id"`
	PrivacyIndex  uint64 `json:"privacy_index"`
	PrivacyStatus string `json:"privacy_status"`
}

type ChangeAccountInfoRequest struct {
	AccountID     uint64 `json:"account_id"`
	Data          string `json:"data"`
	DataFieldName string `json:"data_field_name"`
}

type ChangeAvatarRequest struct {
	AccountID uint64 `json:"account_id"`
	Avatar    string `json:"avatar"`
}

type GetNotificationRequest struct {
	AccountID uint64 `json:"account_id"`
	Page      uint64 `json:"page"`
	PageSize  uint64 `json:"page_size"`
}

type GetNotificationResponse struct {
	AccountID     uint64                `json:"account_id"`
	Page          uint64                `json:"page"`
	PageSize      uint64                `json:"page_size"`
	Notifications []NotificationContent `json:"notifications"`
}

type NotificationContent struct {
	ID       uint64 `json:"id"`
	Content  string `json:"content"`
	DateTime int64  `json:"date_time"`
	IsRead   bool   `json:"is_read"`
}

type LoginWithGoogleRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	PhotoURL    string `json:"photo_url"`
	AuthCode    string `json:"auth_code"`
}

type LoginWithGoogleResponse struct {
	AccessToken  string                  `json:"access_token"`
	RefreshToken string                  `json:"refresh_token"`
	UserID       string                  `json:"user_id"`
	JWTClaims    *auth_service.JWTClaims `json:"jwt_claims"`
	Success      bool                    `json:"success"`
}

type VerifyUsernameAndEmailRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type VerifyUsernameAndEmailResponse struct {
	Success bool   `json:"success"`
	UserID  uint64 `json:"user_id"`
}

type SendOTPForgetPasswordMessageRequest struct {
	AccountID uint64 `json:"account_id"`
	Email     string `json:"email"`
}

type SendOTPForgetPasswordMessageResponse struct {
	Success bool `json:"success"`
}

type CheckValidOTPRequest struct {
	AccountID uint64 `json:"account_id"`
	OTP       int    `json:"otp"`
}
type CheckValidOTPResponse struct {
	Success  bool `json:"success"`
	Attempts int  `json:"attempts"`
}

type ChangePasswordRequest struct {
	AccountID   uint64 `json:"account_id"`
	NewPassword string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Success bool `json:"success"`
}

type CheckIsFollowRequest struct {
	FromAccountID uint64 `json:"from_account_id"`
	ToAccountID   uint64 `json:"to_account_id"`
}

type CheckIsFollowResponse struct {
	IsFollowed bool `json:"is_follow"`
}

type CheckIsBlockedRequest struct {
	FromAccountID uint64 `json:"from_account_id"`
	ToAccountID   uint64 `json:"to_account_id"`
}

type CheckIsBlockedResponse struct {
	IsBlocked bool `json:"is_block"`
}

type CountUnreadNotiRequest struct {
	AccountID uint64 `json:"account_id"`
}

type CountUnreadNotiResponse struct {
	Quantity uint64 `json:"quantity"`
}

type MarkAsReadNotiRequest struct {
	AccountID uint64 `json:"account_id"`
}

type MarkAsReadNotiResponse struct {
	Quantity uint64 `json:"quantity"`
	Success  bool   `json:"success"`
}

type ChatList struct {
	ChatID                string   `json:"chat_id"`
	AccountID             uint64   `json:"account_id"`
	TargetAccountID       uint64   `json:"target_account_id"`
	DisplayName           string   `json:"display_name"`
	AvatarURL             string   `json:"avatar_url"`
	LastMessageTimestamp  int64    `json:"last_message_timestamp"`
	LastMessageContent    string   `json:"last_message_content"`
	UnreadMessageQuantity uint64   `json:"unread_message_quantity"`
	Page                  uint32   `json:"page"`
	PageSize              uint32   `json:"page_size"`
	Participants          []uint32 `json:"participants"`
}

type GetChatListRequest struct {
	AccountID uint64 `json:"account_id"`
	Page      uint32 `json:"page"`
	PageSize  uint32 `json:"page_size"`
}

type GetMessageRequest struct {
	ChatID           string `json:"chat_id"`
	Page             uint32 `json:"page"`
	PageSize         uint32 `json:"page_size"`
	RequestAccountID uint32 `json:"request_account_id"`
}

type GetMessageResponse struct {
	ChatID             string            `json:"chat_id"`
	Messages           []MessageData     `json:"messages"`
	PartnerDisplayInfo SingleAccountInfo `json:"partner_display_info"`
	Page               uint32            `json:"page"`
	PageSize           uint32            `json:"page_size"`
}
type MessageData struct {
	ID         string `json:"id"`
	ChatID     string `json:"chat_id"`
	SenderID   uint32 `json:"sender_id"`
	ReceiverID uint32 `json:"receiver_id"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	Timestamp  int64  `json:"timestamp"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
	IsDeleted  bool   `json:"is_deleted"`
	IsRecalled bool   `json:"is_recalled"`
	IsRead     bool   `json:"is_read"`
}

type ActionMessageRequest struct {
	SenderID   uint32 `json:"sender_id"`
	ReceiverID uint32 `json:"receiver_id"`
	Timestamp  int64  `json:"timestamp"`
	Action     string `json:"action"`
}

type ActionMessageResponse struct {
	Success bool `json:"success"`
}

type ReceiverMarkMessageAsReadRequest struct {
	AccountID uint32 `json:"account_id"`
	ChatID    string `json:"chat_id"`
}

type ReceiverMarkMessageAsReadResponse struct {
	Success bool `json:"success"`
}

type ReportPost struct {
	PostID     uint32 `json:"post_id"`
	ReportedBy uint32 `json:"reported_by"`
	Reason     string `json:"reason"`
}

type ReportUser struct {
	AccountID  uint32 `json:"account_id"`
	ReportedBy uint32 `json:"reported_by"`
	Reason     string `json:"reason"`
}

type ResolveReportedPost struct {
	PostID     uint32 `json:"post_id"`
	ResolvedBy uint32 `json:"resolved_by"`
	Method     string `json:"method"`
}

type ResolveReportedAccount struct {
	AccountID  uint32 `json:"account_id"`
	ResolvedBy uint32 `json:"resolved_by"`
	Method     string `json:"method"`
}

type SingleStatusResponse struct {
	Success bool `json:"success"`
}

type SearchAccountRequest struct {
	RequestAccountID uint64 `json:"request_account_id"`
	QueryString      string `json:"query_string"`
	Page             uint32 `json:"page"`
	PageSize         uint32 `json:"page_size"`
}

type SearchAccountResponse struct {
	Accounts []SingleAccountInfo `json:"accounts"`
	Page     uint32              `json:"page"`
	PageSize uint32              `json:"page_size"`
}

type GetNewRegisterationDataRequest struct {
	RequestAccountID uint64 `json:"request_account_id"`
	PeriodLabel      string `json:"period_label"`
	PeriodYear       uint32 `json:"period_year"`
	PeriodMonth      uint32 `json:"period_month"`
}

type GetNewRegisterationDataResponse struct {
	RequestAccountID uint64      `json:"request_account_id"`
	PeriodLabel      string      `json:"period_label"`
	TotalUsers       uint32      `json:"total_users"`
	Data             []DataTerms `json:"data"`
}

type DataTerms struct {
	Label string `json:"label"`
	Count uint64 `json:"count"`
}

type GetUserTypeRequest struct {
	RequestAccountID uint64 `json:"request_account_id"`
}

type GetUserTypeResponse struct {
	RequestAccountID uint64 `json:"request_account_id"`
	TotalUsers       uint32 `json:"total_users"`
	BannedUsers      uint32 `json:"banned_users"`
	DeletedUsers     uint32 `json:"deleted_users"`
}
type GetAccountListRequest struct {
	RequestID uint32 `json:"request_id"`
	Page      uint32 `json:"page"`
	PageSize  uint32 `json:"page_size"`
}

type AccountRawInfo struct {
	AccountID     uint32 `json:"account_id"`
	Username      string `json:"username"`
	IsBanned      bool   `json:"is_banned"`
	Method        string `json:"method"`
	IsSelfDeleted bool   `json:"is_self_deleted"`
}

type GetAccountListResponse struct {
	Accounts []AccountRawInfo `json:"accounts"`
	Page     uint32           `json:"page"`
	PageSize uint32           `json:"page_size"`
}

type SearchAccountListRequest struct {
	RequestID   uint32 `json:"request_id"`
	QueryString string `json:"query_string"`
	Page        uint32 `json:"page"`
	PageSize    uint32 `json:"page_size"`
}

type SearchAccountListResponse struct {
	Accounts []AccountRawInfo `json:"accounts"`
	Page     uint32           `json:"page"`
	PageSize uint32           `json:"page_size"`
}

type ResolveBanUserRequest struct {
	AccountID uint32 `json:"account_id"`
	Action    string `json:"action"`
}

type ResolveBanUserResponse struct {
	Success bool `json:"success"`
}

type ReportAccountData struct {
	AccountID     uint32 `json:"account_id"`
	Reason        string `json:"reason"`
	Username      string `json:"username"`
	ResolveStatus string `json:"resolve_status"`
}

type GetReportedAccountListRequest struct {
	RequestAccountID uint64 `json:"request_account_id"`
	Page             uint32 `json:"page"`
	PageSize         uint32 `json:"page_size"`
}

type GetReportedAccountListResponse struct {
	Accounts []ReportAccountData `json:"accounts"`
	Page     uint32              `json:"page"`
	PageSize uint32              `json:"page_size"`
}

type GetNewPostStatisticDataRequest struct {
	RequestAccountID uint64 `json:"request_account_id"`
	PeriodLabel      string `json:"period_label"`
	PeriodYear       uint32 `json:"period_year"`
	PeriodMonth      uint32 `json:"period_month"`
}

type GetNewPostStatisticDataResponse struct {
	RequestAccountID uint64      `json:"request_account_id"`
	PeriodLabel      string      `json:"period_label"`
	Data             []DataTerms `json:"data"`
	TotalPosts       uint32      `json:"total_posts"`
}

type GetMediaStatisticRequest struct {
	RequestAccountID uint64 `json:"request_account_id"`
	PeriodLabel      string `json:"period_label"`
	PeriodYear       uint32 `json:"period_year"`
	PeriodMonth      uint32 `json:"period_month"`
}

type GetMediaStatisticResponse struct {
	RequestAccountID uint64      `json:"request_account_id"`
	PeriodLabel      string      `json:"period_label"`
	Data             []DataTerms `json:"data"`
	TotalMedias      uint32      `json:"total_medias"`
}

type GetPostWMediaStatisticRequest struct {
	RequestAccountID uint64 `json:"request_account_id"`
	PeriodLabel      string `json:"period_label"`
	PeriodYear       uint32 `json:"period_year"`
	PeriodMonth      uint32 `json:"period_month"`
}

type GetPostWMediaStatisticResponse struct {
	RequestAccountID uint64 `json:"request_account_id"`
	TotalPosts       uint32 `json:"total_posts"`
	TotalPostWMedias uint32 `json:"total_post_w_medias"`
}

type GetReportedPostRequest struct {
	RequestAccountID uint64 `json:"request_account_id"`
	Page             uint32 `json:"page"`
	PageSize         uint32 `json:"page_size"`
}
type GetReportedPostResponse struct {
	Page          uint32             `json:"page"`
	PageSize      uint32             `json:"page_size"`
	ReportedPosts []ReportedPostData `json:"reported_posts"`
}

type ReportedPostData struct {
	ID            uint32 `json:"id"`
	PostID        uint32 `json:"post_id"`
	Reason        string `json:"reason"`
	ResolveStatus string `json:"resolve_status"`
}

type GetListBanWordsReq struct {
	RequestAccountID uint32 `json:"request_account_id"`
}

type WordData struct {
	ID   uint32 `json:"id"`
	Word string `json:"word"`
}

type DataSet struct {
	LanguageCode string     `json:"language_code"`
	Words        []WordData `json:"words"`
}

type GetListBanWordsRes struct {
	Data []DataSet `json:"data"`
}

type EditWordReq struct {
	ID      uint32 `json:"id"`
	Content string `json:"content"`
}

type EditWordRes struct {
	Success bool `json:"success"`
}

type DeleteWordReq struct {
	ID uint32 `json:"id"`
}

type DeleteWordRes struct {
	Success bool `json:"success"`
}

type AddWordReq struct {
	RequestAccountID uint32 `json:"request_account_id"`
	Content          string `json:"content"`
	LanguageCode     string `json:"language_code"`
}

type GetBlockListByAccountRequest struct {
	AccountID uint32 `json:"account_id"`
}

type GetBlockListByAccountResponse struct {
	Accounts []SingleAccountInfo `json:"accounts"`
	Success  bool                `json:"success"`
}

type GetBlockAndBlockedByAccountRequest struct {
	AccountID uint32 `json:"account_id"`
}

type GetBlockAndBlockedByAccountResponse struct {
	Accounts []SingleAccountInfo `json:"accounts"`
	Success  bool                `json:"success"`
}

type CreateChatRequest struct {
	FirstAccountID  uint32 `json:"first_account_id"`
	SecondAccountID uint32 `json:"second_account_id"`
}

type CreateChatResponse struct {
	Success bool   `json:"success"`
	ChatID  string `json:"chat_id"`
}

type DeleteChatRequest struct {
	ChatID string `json:"chat_id"`
}

type DeleteChatResponse struct {
	Success bool `json:"success"`
}
