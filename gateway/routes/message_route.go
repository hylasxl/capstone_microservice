package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitMessageRoutes(router *mux.Router, clients *ServiceClients) {
	messageRoute := router.PathPrefix("/api/v1/messages").Subrouter()
	messageRoute.HandleFunc("/get-chat-list", handlers.HandlerGetChatList(clients.MessageService, clients.UserService)).Methods("POST")
	messageRoute.HandleFunc("/get-messages", handlers.HandlerGetMessages(clients.MessageService, clients.UserService)).Methods("POST")

	messageRoute.HandleFunc("/action-message", handlers.HandlerActionMessage(clients.MessageService)).Methods("PATCH")
	messageRoute.HandleFunc("/receiver-mark-message-as-read", handlers.HandlerReceiverMarkMessageAsRead(clients.MessageService)).Methods("PATCH")

	messageRoute.HandleFunc("/create-new-chat", handlers.HandlerCreateNewChat(clients.MessageService)).Methods("POST")
	messageRoute.HandleFunc("/delete-chat", handlers.HandlerDeleteChat(clients.MessageService)).Methods("DELETE")

}
