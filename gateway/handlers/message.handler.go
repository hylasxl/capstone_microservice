package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gateway/proto/user_service"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"gateway/proto/message_service"
	"gateway/proto/notification_service"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Redis client
var redisClient = redis.NewClient(&redis.Options{
	Addr: "redis:6379",
})

// Connected WebSocket clients
var clients = make(map[uint32]*websocket.Conn)

// Message struct
type Message struct {
	SenderId   uint32 `json:"sender_id"`
	ReceiverId uint32 `json:"receiver_id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
}

func HandlerWebSocket(notificationClient notification_service.NotificationServiceClient, messageClient message_service.MessageServiceClient) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket Upgrade Error:", err)
			return
		}
		defer conn.Close()

		// Get user_id from query params
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			log.Println("User ID is required")
			conn.Close()
			return
		}

		userIDInt, err := strconv.ParseUint(userID, 10, 32)
		if err != nil {
			log.Println("Invalid User ID:", err)
			conn.Close()
			return
		}
		uID := uint32(userIDInt)

		// Register user connection
		clients[uID] = conn
		log.Printf("User ID %d connected", uID)

		// Deliver offline messages
		deliverOfflineMessages(uID, conn)

		// Start gRPC chat stream
		stream, err := messageClient.ChatStream(context.Background())
		if err != nil {
			return
		}

		// Handle sending and receiving gRPC messages concurrently
		go handleGRPCStream(messageClient, stream)
		handleWebSocketMessages(conn, stream, notificationClient, uID)
	}
}
func handleWebSocketMessages(conn *websocket.Conn, stream message_service.MessageService_ChatStreamClient, notificationClient notification_service.NotificationServiceClient, senderID uint32) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in WebSocket handler: %v", r)
		}
		log.Printf("User %d disconnected", senderID)
		err := conn.Close()
		if err != nil {
			log.Println("Error closing WebSocket:", err)
		}
		delete(clients, senderID) // Remove client on disconnect
	}()

	for {
		var msg map[string]interface{} // Generic map for handling different requests
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("User %d WebSocket connection closed unexpectedly: %v", senderID, err)
			} else {
				log.Println("WebSocket Read Error:", err)
			}
			break // Exit loop on error
		}

		// Check if the message is a request for active users
		if action, ok := msg["action"].(string); ok && action == "get_active_users" {
			activeUsers := GetActiveClients()
			response := map[string]interface{}{
				"action":       "active_users",
				"active_users": activeUsers,
			}
			err := conn.WriteJSON(response)
			if err != nil {
				log.Println("Error sending active users list:", err)
			}
			continue // Skip further processing for this request
		}

		// Handle normal chat messages
		if senderID, ok := msg["sender_id"].(float64); ok {
			if receiverID, ok := msg["receiver_id"].(float64); ok {
				if content, ok := msg["content"].(string); ok {
					timestamp := time.Now().Unix()
					message := Message{
						SenderId:   uint32(senderID),
						ReceiverId: uint32(receiverID),
						Content:    content,
						Timestamp:  timestamp,
					}

					// Send message via gRPC
					err = stream.Send(&message_service.ChatMessage{
						SenderID:   message.SenderId,
						ReceiverID: message.ReceiverId,
						Content:    message.Content,
						Timestamp:  message.Timestamp,
					})
					if err != nil {
						log.Println("Error sending message to gRPC:", err)
						break
					}

					// Deliver to online users
					if receiverConn, found := clients[message.ReceiverId]; found {
						err := receiverConn.WriteJSON(message)
						if err != nil {
							log.Println("Error sending message to receiver:", err)
						}
					} else {
						go sendNotification(notificationClient, message.ReceiverId)
					}
				}
			}
		}
	}
}

func handleGRPCStream(messageClient message_service.MessageServiceClient, stream message_service.MessageService_ChatStreamClient) {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("gRPC stream closed (EOF). Reconnecting...")
			time.Sleep(2 * time.Second)
			newStream, err := messageClient.ChatStream(context.Background())
			if err != nil {
				log.Println("Failed to reconnect gRPC stream:", err)
				time.Sleep(5 * time.Second)
				continue
			}
			stream = newStream
			continue
		}
		if err != nil {
			log.Println("Error receiving from gRPC stream:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Deliver message if user is online
		if receiverConn, found := clients[resp.ReceiverID]; found {
			err = receiverConn.WriteJSON(resp)
			if err != nil {
				log.Println("Error sending message to WebSocket user:", err)
			}
		} else {
			// Save to Redis if user is offline
			ctx := context.Background()
			msgBytes, _ := json.Marshal(resp)
			redisClient.RPush(ctx, "offline:"+strconv.FormatUint(uint64(resp.ReceiverID), 10), string(msgBytes))
		}
	}
}

func deliverOfflineMessages(userID uint32, conn *websocket.Conn) {
	ctx := context.Background()
	messages, err := redisClient.LRange(ctx, "offline:"+strconv.FormatUint(uint64(userID), 10), 0, -1).Result()
	if err != nil {
		log.Println("Error retrieving offline messages:", err)
		return
	}

	for _, msgStr := range messages {
		var msg Message
		if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
			log.Println("Error parsing message:", err)
			continue
		}

		err := conn.WriteJSON(msg)
		if err != nil {
			log.Println("Error sending offline message:", err)
		}
	}

	redisClient.Del(ctx, "offline:"+strconv.FormatUint(uint64(userID), 10))
}

func sendNotification(notificationClient notification_service.NotificationServiceClient, userID uint32) {
	//ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//defer cancel()

}

func HandlerGetChatList(messageClient message_service.MessageServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req GetChatListRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		chatListResp, err := messageClient.GetChatList(ctx, &message_service.GetChatListRequest{
			AccountID: uint32(req.AccountID),
			Page:      req.Page,
			PageSize:  req.PageSize,
		})

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}

		targetUserID := make([]uint64, 0)

		for _, chat := range chatListResp.ChatList {
			targetUserID = append(targetUserID, uint64(chat.TargetAccountID))
		}

		userInfoResp, err := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
			IDs: targetUserID,
		})

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}

		var response = make([]ChatList, 0)

		for index, chat := range chatListResp.ChatList {
			response = append(response, ChatList{
				AccountID:             uint64(chat.AccountID),
				TargetAccountID:       uint64(chat.TargetAccountID),
				DisplayName:           userInfoResp.Infos[index].DisplayName,
				AvatarURL:             userInfoResp.Infos[index].AvatarURL,
				LastMessageTimestamp:  chat.LastMessageTimestamp,
				LastMessageContent:    chat.LastMessageContent,
				UnreadMessageQuantity: chat.UnreadMessageQuantity,
				Page:                  chat.Page,
				PageSize:              chat.PageSize,
				ChatID:                chat.ChatID,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid response", nil)
		}

	}
}

func HandlerGetMessages(messageClient message_service.MessageServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req GetMessageRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		messageResp, err := messageClient.GetMessages(ctx, &message_service.GetMessageRequest{
			ChatID:   req.ChatID,
			Page:     req.Page,
			PageSize: req.PageSize,
		})

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}

		var response GetMessageResponse

		response.ChatID = req.ChatID
		for _, msg := range messageResp.Messages {
			response.Messages = append(response.Messages, MessageData{
				ID:         msg.ID,
				ChatID:     msg.ChatID,
				SenderID:   msg.SenderID,
				ReceiverID: msg.ReceiverID,
				Content:    msg.Content,
				Type:       msg.Type,
				Timestamp:  msg.Timestamp,
				CreatedAt:  msg.CreatedAt,
				UpdatedAt:  msg.UpdatedAt,
				IsRead:     msg.IsRead,
				IsDeleted:  msg.IsDeleted,
				IsRecalled: msg.IsRecalled,
			})
		}
		idArray := make([]uint32, 0)
		listForQuery := make([]uint64, 0)
		idArray = append(idArray, response.Messages[0].ReceiverID, response.Messages[0].SenderID)

		for _, id := range idArray {
			if id != req.RequestAccountID {
				listForQuery = append(listForQuery, uint64(id))
			}
		}

		userInfoResp, _ := userClient.GetListAccountDisplayInfo(ctx, &user_service.GetListAccountDisplayInfoRequest{
			IDs: listForQuery,
		})

		response.PartnerDisplayInfo.DisplayName = userInfoResp.Infos[0].DisplayName
		response.PartnerDisplayInfo.AvatarURL = userInfoResp.Infos[0].AvatarURL
		response.PartnerDisplayInfo.AccountID = uint(userInfoResp.Infos[0].AccountID)
		response.Page = req.Page
		response.PageSize = req.PageSize

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid response", nil)
		}
	}
}

func HandlerActionMessage(messageClient message_service.MessageServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset r.Body for decoding

		var req ActionMessageRequest
		err := json.Unmarshal(body, &req) // Directly using json.Unmarshal for safety
		if err != nil {
			fmt.Println("JSON decode error:", err)
			respondWithError(w, http.StatusBadRequest, "invalid request", err)
			return
		}

		if req.Action != "delete" && req.Action != "recall" {
			respondWithError(w, http.StatusBadRequest, "invalid action request", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		actionResp, err := messageClient.ActionMessage(ctx, &message_service.ActionMessageRequest{
			SenderID:   req.SenderID,
			ReceiverID: req.ReceiverID,
			Timestamp:  req.Timestamp,
			Action:     req.Action,
		})

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}
		var response ActionMessageResponse
		response.Success = actionResp.Success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid response", nil)
		}
	}
}

func HandlerReceiverMarkMessageAsRead(messageClient message_service.MessageServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReceiverMarkMessageAsReadRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		markResp, err := messageClient.ReceiverMarkMessageAsRead(ctx, &message_service.ReceiverMarkMessageAsReadRequest{
			AccountID: req.AccountID,
			ChatID:    req.ChatID,
		})

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}

		var response ReceiverMarkMessageAsReadResponse
		response.Success = markResp.Success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid response", nil)
		}
	}
}

func HandlerGetActiveClients() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeClients := GetActiveClients()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(activeClients)
		if err != nil {
			log.Println("Error encoding active clients response:", err)
			http.Error(w, "Failed to get active clients", http.StatusInternalServerError)
		}
	}
}

func GetActiveClients() []uint32 {
	activeUsers := make([]uint32, 0)

	for userID := range clients {
		activeUsers = append(activeUsers, userID)
	}

	return activeUsers
}
