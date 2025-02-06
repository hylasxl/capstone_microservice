package main

import (
	"google.golang.org/grpc"
	"log"
	"message_service/configs"
	"message_service/handlers"
	"message_service/proto/message_service"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	configs.ConnectMongoDB()
	redisClient := configs.ConnectRedis()

	grpcServer := grpc.NewServer()
	
	message_service.RegisterMessageServiceServer(grpcServer, handlers.NewMessageService(configs.Client, redisClient))

	log.Println("Message service started on port 50055")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
