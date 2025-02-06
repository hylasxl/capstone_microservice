package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitializeFriendRoutes(router *mux.Router, clients *ServiceClients) {
	friendRoutes := router.PathPrefix("/api/v1/friends").Subrouter()
	friendRoutes.HandleFunc("/send-request", handlers.HandlerSendFriendRequest(clients.FriendService, clients.UserService, clients.NotificationService)).Methods("POST")
	friendRoutes.HandleFunc("/recall-request", handlers.HandlerRecallRequest(clients.FriendService, clients.UserService)).Methods("POST")
	friendRoutes.HandleFunc("/resolve-request", handlers.HandlerResolveFriendRequest(clients.FriendService, clients.UserService)).Methods("POST")
	friendRoutes.HandleFunc("/unfriend", handlers.HandlerUnfriend(clients.FriendService, clients.UserService)).Methods("POST")
	friendRoutes.HandleFunc("/resolve-follow", handlers.HandlerResolveFollow(clients.FriendService, clients.UserService, clients.NotificationService)).Methods("POST")
	friendRoutes.HandleFunc("/resolve-block", handlers.HandlerResolveBlock(clients.FriendService, clients.UserService)).Methods("POST")

	friendRoutes.HandleFunc("/get-pending-list", handlers.HandlerGetPendingList(clients.FriendService, clients.UserService)).Methods("GET")
	friendRoutes.HandleFunc("/get-list-friend", handlers.HandlerGetListFriend(clients.FriendService, clients.UserService)).Methods("GET")
	friendRoutes.HandleFunc("/count-pending-request", handlers.HandlerCountPendingFriendRequest(clients.FriendService, clients.UserService)).Methods("GET")
	friendRoutes.HandleFunc("/check-existing-request", handlers.HandlerCheckExistingFriendRequest(clients.FriendService)).Methods("GET")
	friendRoutes.HandleFunc("/check-is-follow", handlers.HandlerCheckIsFollow(clients.FriendService)).Methods("GET")
	friendRoutes.HandleFunc("/check-is-block", handlers.HandlerCheckIsBlock(clients.FriendService)).Methods("GET")
}
