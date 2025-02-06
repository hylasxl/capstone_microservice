package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitMessageRoutes(router *mux.Router, clients *ServiceClients) {
	messageRoute := router.PathPrefix("/api/v1/messages").Subrouter()
	messageRoute.HandleFunc("/get-chat-list", handlers.HandlerGetChatList(clients.MessageService, clients.UserService)).Methods("GET")
}
