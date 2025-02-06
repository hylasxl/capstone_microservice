package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitializePostRoutes(router *mux.Router, clients *ServiceClients) {
	postRoutes := router.PathPrefix("/api/v1/posts").Subrouter()

	postRoutes.HandleFunc("/get-post-by-id", handlers.HandlerGetSinglePost(clients.PostService, clients.UserService)).Methods("GET")
	postRoutes.HandleFunc("/get-post-comments", handlers.HandlerGetPostComments(clients.PostService)).Methods("GET")
	postRoutes.HandleFunc("/get-wall-posts", handlers.HandlerGetWallPost(clients.PostService, clients.UserService, clients.FriendService)).Methods("GET")
	postRoutes.HandleFunc("/get-new-feeds", handlers.HandlerGetNewFeeds(clients.PostService, clients.UserService, clients.FriendService)).Methods("GET")

	postRoutes.HandleFunc("/create-new-post", handlers.HandlerCreatePost(clients.PostService, clients.UserService, clients.ModerationService)).Methods("POST")
	postRoutes.HandleFunc("/share-post", handlers.HandlerSharePost(clients.PostService, clients.UserService, clients.ModerationService)).Methods("POST")
	postRoutes.HandleFunc("/comment-post", handlers.HandlerCommentPost(clients.PostService, clients.UserService, clients.ModerationService, clients.NotificationService)).Methods("POST")
	postRoutes.HandleFunc("/reply-comment-post", handlers.HandlerReplyComment(clients.PostService, clients.UserService, clients.ModerationService, clients.NotificationService)).Methods("POST")
	postRoutes.HandleFunc("/react-post", handlers.HandlerReactPost(clients.PostService, clients.UserService)).Methods("POST")
	postRoutes.HandleFunc("/react-image", handlers.HandlerReactImage(clients.PostService, clients.UserService)).Methods("POST")
	postRoutes.HandleFunc("/comment-image", handlers.HandlerCommentImage(clients.PostService, clients.UserService, clients.ModerationService)).Methods("POST")
	postRoutes.HandleFunc("/reply-comment-image", handlers.HandlerReplyCommentImage(clients.PostService, clients.UserService, clients.ModerationService)).Methods("POST")

	postRoutes.HandleFunc("/edit-comment-post", handlers.HandlerEditComment(clients.PostService, clients.ModerationService)).Methods("PUT")
	postRoutes.HandleFunc("/edit-comment-image", handlers.HandlerEditCommentImage(clients.PostService, clients.UserService, clients.ModerationService)).Methods("PUT")

	postRoutes.HandleFunc("/remove-react-post", handlers.HandlerRemoveReactPost(clients.PostService, clients.UserService)).Methods("DELETE")
	postRoutes.HandleFunc("/delete-post/{id}", handlers.HandlerDeletePost(clients.PostService)).Methods("DELETE")
	postRoutes.HandleFunc("/delete-post-comment", handlers.HandlerDeletePostComment(clients.PostService)).Methods("DELETE")
	postRoutes.HandleFunc("/delete-post-image", handlers.HandlerDeletePostImage(clients.PostService)).Methods("DELETE")
	postRoutes.HandleFunc("/remove-react-image", handlers.HandlerRemoveReactImage(clients.PostService, clients.UserService)).Methods("DELETE")
	postRoutes.HandleFunc("/delete-comment-image", handlers.HandlerDeleteCommentImage(clients.PostService, clients.UserService)).Methods("DELETE")
}
