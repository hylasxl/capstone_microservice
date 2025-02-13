package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitMessageRoutes(router *mux.Router, clients *ServiceClients) {
	messageRoute := router.PathPrefix("/api/v1/messages").Subrouter()
	messageRoute.HandleFunc("/get-chat-list", handlers.HandlerGetChatList(clients.MessageService, clients.UserService)).Methods("GET")
	messageRoute.HandleFunc("/get-messages", handlers.HandlerGetMessages(clients.MessageService, clients.UserService)).Methods("GET")

	messageRoute.HandleFunc("/action-message", handlers.HandlerActionMessage(clients.MessageService)).Methods("PATCH")
	messageRoute.HandleFunc("/receiver-mark-message-as-read", handlers.HandlerReceiverMarkMessageAsRead(clients.MessageService)).Methods("PATCH")

}
