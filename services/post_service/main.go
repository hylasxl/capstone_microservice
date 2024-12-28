package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"post_service/configs"
	"post_service/handlers"
	"post_service/models"
	"post_service/proto/post_service"
	"time"
)

const (
	maxMessageSize = 256 * 1024 * 1024
)

func main() {
	ctx := context.Background()
	err := os.Setenv("TZ", "Asia/Bangkok")
	if err != nil {
		return
	}
	time.Local = time.FixedZone("UTC+7", 7*3600)
	lis, err := net.Listen("tcp", ":51000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize))
	CLD, err := configs.InitCloudinary(ctx)
	db := configs.InitMySQL()

	post_service.RegisterPostServiceServer(grpcServer, &handlers.PostService{
		CloudinaryClient: (*handlers.CloudinaryService)(CLD),
		DB:               db,
	})

	err = db.AutoMigrate(
		&models.Post{},
		&models.PostComment{},
		&models.PostReaction{},
		&models.PostContentEditHistory{},
		&models.PostCommentEditHistory{},
		&models.PostTagFriend{},
		&models.PostMultiMedia{},
		&models.PostMultiMediaComment{},
		&models.PostMultiMediaReaction{},
		&models.PostMultiMediaEditCommentHistory{},
	)

	if err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	log.Println("Post service started on port 51000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
