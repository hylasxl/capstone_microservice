package routes

import (
	"fmt"
	authpb "gateway/proto/auth_service"
	friendpb "gateway/proto/friend_service"
	moderationpb "gateway/proto/moderation_service"
	postpb "gateway/proto/post_service"
	privacypb "gateway/proto/privacy_service"
	userpb "gateway/proto/user_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

const (
	maxMessageSize = 256 * 1024 * 1024 //
)

type ServiceClients struct {
	AuthService       authpb.AuthServiceClient
	UserService       userpb.UserServiceClient
	PrivacyService    privacypb.PrivacyServiceClient
	FriendService     friendpb.FriendServiceClient
	PostService       postpb.PostServiceClient
	ModerationService moderationpb.ModerationServiceClient
	Connections       []*grpc.ClientConn
}

func InitializeServiceClients() (*ServiceClients, error) {
	services := map[string]string{
		"authService":          getServiceAddr("authService"),
		"userService":          getServiceAddr("userService"),
		"privacyService":       getServiceAddr("privacyService"),
		"friendService":        getServiceAddr("friendService"),
		"postService":          getServiceAddr("postService"),
		"moderationService":    getServiceAddr("moderationService"),
		"messageService":       getServiceAddr("messageService"),
		"notificationService":  getServiceAddr("notificationService"),
		"onlineHistoryService": getServiceAddr("onlineHistoryService"),
		"otpService":           getServiceAddr("otpService"),
	}

	clients := &ServiceClients{}
	var connections []*grpc.ClientConn

	for serviceName, serviceAddr := range services {
		conn, err := grpc.Dial(serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to %s at %s: %w", serviceName, serviceAddr, err)
		}

		log.Printf("Connected to %s at %s", serviceName, serviceAddr)
		connections = append(connections, conn)

		switch serviceName {
		case "authService":
			clients.AuthService = authpb.NewAuthServiceClient(conn)
		case "userService":
			clients.UserService = userpb.NewUserServiceClient(conn)
		case "privacyService":
			clients.PrivacyService = privacypb.NewPrivacyServiceClient(conn)
		case "friendService":
			clients.FriendService = friendpb.NewFriendServiceClient(conn)
		case "postService":
			clients.PostService = postpb.NewPostServiceClient(conn)
		case "moderationService":
			clients.ModerationService = moderationpb.NewModerationServiceClient(conn)
		}
	}

	clients.Connections = connections
	return clients, nil
}

func (c *ServiceClients) CloseConnections() {
	for _, conn := range c.Connections {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
}

func getServiceAddr(serviceName string) string {
	defaultServices := map[string]string{
		"authService":          "auth_service:50051",
		"friendService":        "friend_service:50054",
		"messageService":       "message_service:50055",
		"moderationService":    "moderation_service:50056",
		"notificationService":  "notification_service:50057",
		"onlineHistoryService": "online_history_service:50058",
		"otpService":           "otp_service:50059",
		"postService":          "post_service:50060",
		"privacyService":       "privacy_service:50061",
		"userService":          "user_service:50052",
	}

	defaultAddr := defaultServices[serviceName]

	return defaultAddr
}
