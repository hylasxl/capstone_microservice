package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitWebsocketRoute(router *mux.Router, clients *ServiceClients) {

	router.HandleFunc("/ws", handlers.HandlerWebSocket(clients.NotificationService, clients.MessageService))
}
