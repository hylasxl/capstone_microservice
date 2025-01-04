package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitializeUserRoutes(router *mux.Router, clients *ServiceClients) {
	userRoutes := router.PathPrefix("/api/v1/users").Subrouter()

	userRoutes.HandleFunc("/get-infos", handlers.HandlerGetAccountInfo(clients.UserService)).Methods("GET")
	userRoutes.HandleFunc("/get-profile-info", handlers.HandlerGetProfileInfo(clients.UserService, clients.FriendService, clients.PrivacyService)).Methods("GET")

	userRoutes.HandleFunc("/check-duplicate", handlers.HandlerCheckDuplicate(clients.UserService)).Methods("POST")
	userRoutes.HandleFunc("/check-valid-user", handlers.HandlerCheckValidUser(clients.UserService)).Methods("POST")
	userRoutes.HandleFunc("/change-avatar", handlers.HandlerChangeAvatar(clients.UserService)).Methods("POST")

	userRoutes.HandleFunc("/change-user-info", handlers.HandlerChangeUserInfo(clients.UserService)).Methods("PATCH")
}
