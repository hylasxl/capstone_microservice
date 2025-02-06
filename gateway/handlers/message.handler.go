package handlers

import (
	"context"
	"encoding/json"
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

		go func() {
			for {
				time.Sleep(30 * time.Second) // Send ping every 30 seconds
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println("Ping failed, closing connection:", err)
					conn.Close()
					delete(clients, uID)
					break
				}
			}
		}()

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
		log.Printf("User %d disconnected", senderID)
		delete(clients, senderID) // Remove user from clients map
		err := conn.Close()
		if err != nil {
			return
		}
	}()
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			// Check if the WebSocket was closed
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("User %d WebSocket connection closed unexpectedly: %v", senderID, err)
				delete(clients, senderID) // Remove user if connection is lost
				break
			}

			log.Println("WebSocket Read Error:", err)
			continue // Prevent loop from exiting
		}

		msg.Timestamp = time.Now().Unix()

		// Send message to gRPC stream
		err = stream.Send(&message_service.ChatMessage{
			SenderID:   msg.SenderId,
			ReceiverID: msg.ReceiverId,
			Content:    msg.Content,
			Timestamp:  msg.Timestamp,
		})
		if err != nil {
			log.Println("Error sending message to gRPC:", err)
			continue
		}

		// Only forward message if receiver is online
		if receiveConn, found := clients[msg.ReceiverId]; found {
			err := receiveConn.WriteJSON(msg)
			if err != nil {
				log.Println("Error sending message to receiver:", err)
				continue
			}
		} else {
			go sendNotification(notificationClient, msg.ReceiverId)
		}
	}
}

func handleGRPCStream(messageClient message_service.MessageServiceClient, stream message_service.MessageService_ChatStreamClient) {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("gRPC stream closed (EOF). Attempting to reconnect...")
			time.Sleep(2 * time.Second) // Delay before retrying
			newStream, err := messageClient.ChatStream(context.Background())
			if err != nil {
				log.Println("Failed to reconnect gRPC stream:", err)
				time.Sleep(5 * time.Second) // Wait longer before retrying
				continue
			}
			stream = newStream
			continue
		}
		if err != nil {
			log.Println("Error receiving from gRPC stream:", err)
			time.Sleep(2 * time.Second) // Allow a small delay before retrying
			continue
		}

		// Deliver message to the correct WebSocket user
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
