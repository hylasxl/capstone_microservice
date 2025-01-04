package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitPrivacyRoute(router *mux.Router, clients *ServiceClients) {
	privacyRoutes := router.PathPrefix("/api/v1/privacy").Subrouter()

	privacyRoutes.HandleFunc("/set-privacy", handlers.HandlerSetPrivacy(clients.PrivacyService, clients.UserService)).Methods("PUT")
}
