package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitModerationRoute(router *mux.Router, clients *ServiceClients) {
	moderationRoutes := router.PathPrefix("/api/v1/moderation").Subrouter()

	moderationRoutes.HandleFunc("/report-post", handlers.HandleReportPost(clients.ModerationService)).Methods("POST")
	moderationRoutes.HandleFunc("/report-account", handlers.HandleReportAccount(clients.ModerationService)).Methods("POST")

	moderationRoutes.HandleFunc("/get-reported-account-list", handlers.HandleGetReportedAccount(clients.ModerationService, clients.UserService)).Methods("POST")
	moderationRoutes.HandleFunc("/get-reported-post-list", handlers.HandleGetReportedPosts(clients.ModerationService)).Methods("POST")

	moderationRoutes.HandleFunc("/resolve-reported-post", handlers.HandleResolveReportedPost(clients.ModerationService, clients.PostService)).Methods("POST")
	moderationRoutes.HandleFunc("/resolve-reported-account", handlers.HandleResolveReportedAccount(clients.ModerationService, clients.UserService)).Methods("POST")

	moderationRoutes.HandleFunc("/get-ban-list", handlers.HandlerGetBanWords(clients.ModerationService)).Methods("POST")
	moderationRoutes.HandleFunc("/edit-word", handlers.HandlerEditWord(clients.ModerationService)).Methods("POST")
	moderationRoutes.HandleFunc("/delete-word", handlers.HandlerDeleteWord(clients.ModerationService)).Methods("POST")
	moderationRoutes.HandleFunc("/add-word", handlers.HandlerAddWord(clients.ModerationService)).Methods("POST")
}
