package models

type GetChatListRequest struct {
	AccountID uint64 `json:"account_id"`
	Page      uint32 `json:"page"`
	PageSize  uint32 `json:"page_size"`
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

type GetMessageRequest struct {
	ChatID   string `json:"chat_id"`
	Page     uint32 `json:"page"`
	PageSize uint32 `json:"page_size"`
}
type MessageData struct {
	ID         string `json:"id"`
	ChatID     string `json:"chat_id"`
	SenderID   uint32 `json:"sender_id"`
	ReceiverID uint32 `json:"receiver_id"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	Timestamp  int64  `json:"timestamp"`
	CreatedAt  int    `json:"created_at"`
	UpdatedAt  int    `json:"updated_at"`
	IsDeleted  bool   `json:"is_deleted"`
	IsRecalled bool   `json:"is_recalled"`
	IsRead     bool   `json:"is_read"`
}

type ActionMessageRequest struct {
	SenderID   uint32 `json:"sender_id"`
	ReceiverId uint32 `json:"receiver_id"`
	Timestamp  int64  `json:"timestamp"`
	Action     string `json:"action"`
}

type ReceiverMarkMessageAsReadRequest struct {
	AccountID uint64 `json:"account_id"`
	ChatID    string `json:"chat_id"`
}

type ReceiverMarkMessageAsReadResponse struct {
	Success bool `json:"success"`
}
