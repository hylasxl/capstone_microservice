package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitializeAuthRoutes(router *mux.Router, clients *ServiceClients) {
	authRoutes := router.PathPrefix("/api/v1/authentication").Subrouter()
	authRoutes.HandleFunc("/login", handlers.HandlerLogin(clients.AuthService, clients.UserService)).Methods("POST")
	authRoutes.HandleFunc("/register", handlers.HandlerSignUp(clients.UserService, clients.PrivacyService)).Methods("POST")
}
