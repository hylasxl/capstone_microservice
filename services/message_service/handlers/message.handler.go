package handlers

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"message_service/models"
	ms "message_service/proto/message_service"
	"strconv"
	"sync"
	"time"
)

type MessageService struct {
	ms.UnimplementedMessageServiceServer
	MongoClient *mongo.Client
	RedisClient *redis.Client
	mu          sync.Mutex
	streams     map[uint32]ms.MessageService_ChatStreamServer
}

func NewMessageService(mongoClient *mongo.Client, redisClient *redis.Client) *MessageService {
	return &MessageService{
		MongoClient: mongoClient,
		RedisClient: redisClient,
		streams:     make(map[uint32]ms.MessageService_ChatStreamServer),
	}
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

type GetChatListRequest struct {
	AccountID uint64 `json:"account_id"`
	Page      uint32 `json:"page"`
	PageSize  uint32 `json:"page_size"`
}

func (s *MessageService) ChatStream(stream ms.MessageService_ChatStreamServer) error {
	var userID uint32

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		userID = req.SenderID

		s.mu.Lock()
		s.streams[userID] = stream
		s.mu.Unlock()

		go func() {
			err := s.storeMessage(req)
			if err != nil {
				log.Println(err)
			}
		}()

		s.mu.Lock()
		receiverStream, exists := s.streams[req.ReceiverID]
		s.mu.Unlock()
		if exists {
			err := receiverStream.Send(&ms.ChatMessageReturn{
				Timestamp: req.Timestamp,
				Success:   true,
			})
			if err != nil {
				log.Println("Error sending message to receiver:", err)
			}
		} else {
			go s.storeOfflineMessage(req)
		}
	}
}
func (s *MessageService) storeMessage(msg *ms.ChatMessage) error {
	ctx := context.TODO()
	db := s.MongoClient.Database("admin")

	userIDs := make([]uint32, 0)
	userIDs = append(userIDs, msg.SenderID, msg.ReceiverID)

	retrievedChat, err := models.FindChatByParticipants(ctx, db, userIDs)
	if err != nil {
		return err
	}

	// Convert ChatMessage to Message model
	message := models.Message{
		ID:          primitive.NewObjectID(),
		ChatID:      retrievedChat.ID, // Find or create chat
		SenderID:    uint(msg.SenderID),
		ReceiverID:  uint(msg.ReceiverID),
		Content:     msg.Content,
		MessageType: "text",
		Timestamp:   msg.Timestamp,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsDeleted:   false,
		IsRecalled:  false,
		IsRead:      false,
	}

	// Insert message into MongoDB
	_, err = models.InsertMessage(context.Background(), db, &message)
	if err != nil {
		return err
	}

	// Update or create a chat record
	err = updateOrCreateChat(context.Background(), db, message)
	if err != nil {
		return err
	}

	return nil
}

func (s *MessageService) storeOfflineMessage(msg *ms.ChatMessage) {
	ctx := context.Background()
	messageJSON := bson.M{
		"sender_id":   msg.SenderID,
		"receiver_id": msg.ReceiverID,
		"content":     msg.Content,
		"timestamp":   msg.Timestamp,
	}
	data, err := bson.Marshal(messageJSON)
	if err != nil {
		log.Println("Error marshalling offline message:", err)
		return
	}

	s.RedisClient.RPush(ctx, "offline:"+strconv.Itoa(int(msg.ReceiverID)), data)
}

func updateOrCreateChat(ctx context.Context, db *mongo.Database, message models.Message) error {
	collection := db.Collection(models.Chat{}.CollectionName())

	// Find existing chat
	filter := bson.M{
		"participants": bson.M{"$all": []uint{message.SenderID, message.ReceiverID}},
	}

	update := bson.M{
		"$set": bson.M{
			"last_message":    message.Content,
			"last_message_at": time.Now(),
			"updated_at":      time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// If no chat exists, create a new one
	if result.MatchedCount == 0 {
		chat := models.Chat{
			Participants:  []uint{message.SenderID, message.ReceiverID},
			LastMessage:   message.Content,
			LastMessageAt: time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_, err := models.CreateChat(ctx, db, &chat)
		return err
	}

	return nil
}

func (s *MessageService) GetChatList(ctx context.Context, req *ms.GetChatListRequest) (*ms.GetChatListResponse, error) {
	chatListResp, err := models.GetChatList(ctx, s.MongoClient.Database("admin"), models.GetChatListRequest{
		AccountID: uint64(req.AccountID),
		Page:      req.Page,
		PageSize:  req.PageSize,
	})

	if err != nil {
		return nil, err
	}

	response := &ms.GetChatListResponse{}

	for _, chat := range chatListResp {
		response.ChatList = append(response.ChatList, &ms.ChatList{
			ChatID:                chat.ChatID,
			AccountID:             uint32(chat.AccountID),
			TargetAccountID:       uint32(chat.TargetAccountID),
			DisplayName:           chat.DisplayName,
			AvatarURL:             chat.AvatarURL,
			LastMessageTimestamp:  chat.LastMessageTimestamp,
			LastMessageContent:    chat.LastMessageContent,
			UnreadMessageQuantity: chat.UnreadMessageQuantity,
			Page:                  chat.Page,
			PageSize:              chat.PageSize,
		})
	}
	return response, nil
}

func (s *MessageService) GetMessages(ctx context.Context, req *ms.GetMessageRequest) (*ms.GetMessageResponse, error) {

	messageResp, err := models.GetMessages(ctx, s.MongoClient.Database("admin"), models.GetMessageRequest{
		ChatID:   req.ChatID,
		Page:     req.Page,
		PageSize: req.PageSize,
	})

	if err != nil {
		return nil, err
	}

	response := &ms.GetMessageResponse{}

	for _, message := range messageResp {
		response.Messages = append(response.Messages, &ms.MessageData{
			ID:         message.ID,
			ChatID:     message.ChatID,
			SenderID:   message.SenderID,
			ReceiverID: message.ReceiverID,
			Content:    message.Content,
			Type:       "text",
			Timestamp:  message.Timestamp,
			CreatedAt:  int64(message.CreatedAt),
			UpdatedAt:  int64(message.UpdatedAt),
			IsDeleted:  message.IsDeleted,
			IsRecalled: message.IsRecalled,
			IsRead:     message.IsRead,
		})
	}

	return response, nil
}

func (s *MessageService) ActionMessage(ctx context.Context, req *ms.ActionMessageRequest) (*ms.ActionMessageResponse, error) {

	switch req.Action {
	case "delete":
		{
			err := models.DeleteMessage(ctx, s.MongoClient.Database("admin"), models.ActionMessageRequest{
				SenderID:   req.SenderID,
				ReceiverId: req.ReceiverID,
				Timestamp:  req.Timestamp,
			})
			if err != nil {
				return nil, err
			}
			break
		}
	case "recall":
		{
			err := models.RecallMessage(ctx, s.MongoClient.Database("admin"), models.ActionMessageRequest{
				SenderID:   req.SenderID,
				ReceiverId: req.ReceiverID,
				Timestamp:  req.Timestamp,
			})
			if err != nil {
				return nil, err
			}
			break
		}
	default:
		return &ms.ActionMessageResponse{
			Success: false,
		}, errors.New("invalid action")
	}

	return &ms.ActionMessageResponse{
		Success: true,
	}, nil
}

func (s *MessageService) ReceiverMarkMessageAsRead(ctx context.Context, req *ms.ReceiverMarkMessageAsReadRequest) (*ms.ReceiverMarkMessageAsReadResponse, error) {

	err := models.ReceiverMarkMessageAsRead(ctx, s.MongoClient.Database("admin"), models.ReceiverMarkMessageAsReadRequest{
		AccountID: uint64(req.AccountID),
		ChatID:    req.ChatID,
	})

	if err != nil {
		return nil, err
	}

	return &ms.ReceiverMarkMessageAsReadResponse{
		Success: true,
	}, nil
}
