package models

type GetChatListRequest struct {
	AccountID uint64 `json:"account_id"`
	Page      uint32 `json:"page"`
	PageSize  uint32 `json:"page_size"`
}

type ChatList struct {
	AccountID             uint64 `json:"account_id"`
	TargetAccountID       uint64 `json:"target_account_id"`
	DisplayName           string `json:"display_name"`
	AvatarURL             string `json:"avatar_url"`
	LastMessageTimestamp  int64  `json:"last_message_timestamp"`
	LastMessageContent    string `json:"last_message_content"`
	UnreadMessageQuantity uint64 `json:"unread_message_quantity"`
	Page                  uint32 `json:"page"`
	PageSize              uint32 `json:"page_size"`
}
