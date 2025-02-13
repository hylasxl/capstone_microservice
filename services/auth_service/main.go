package main

import (
	"auth_service/configs"
	"auth_service/handlers"
	"auth_service/proto/auth_service"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	err := os.Setenv("TZ", "Asia/Bangkok")
	if err != nil {
		return
	}
	time.Local = time.FixedZone("UTC+7", 7*3600)
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	DB := configs.InitMySQL()
	grpcServer := grpc.NewServer()
	auth_service.RegisterAuthServiceServer(grpcServer, &handlers.AuthService{
		DB: DB,
	})

	log.Println("Auth service is running on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
