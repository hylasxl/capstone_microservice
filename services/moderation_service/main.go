package main

import (
	"google.golang.org/grpc"
	"log"
	"moderation_service/configs"
	"moderation_service/handlers"
	"moderation_service/proto/moderation_service"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", ":50056")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db := configs.InitMySQL()

	grpcServer := grpc.NewServer()
	moderation_service.RegisterModerationServiceServer(grpcServer, &handlers.ModerationService{
		DB: db,
	})

	log.Println("Moderation service started on port 50056")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
