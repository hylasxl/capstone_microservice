package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitNotificationRoutes(router *mux.Router, clients *ServiceClients) {
	notificationRoutes := router.PathPrefix("/api/v1/notifications").Subrouter()

	notificationRoutes.HandleFunc("/register-device", handlers.HandlerRegisterDevice(clients.NotificationService)).Methods("POST")
	notificationRoutes.HandleFunc("/mark-read", handlers.HandlerMarkAsReadNoti(clients.NotificationService)).Methods("POST")

	notificationRoutes.HandleFunc("/get-notifications", handlers.HandlerGetNotifications(clients.NotificationService)).Methods("POST")
	notificationRoutes.HandleFunc("/count-unread", handlers.HandlerCountUnreadNotifications(clients.NotificationService)).Methods("POST")
}
