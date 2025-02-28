package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"privacy_service/configs"
	"privacy_service/handlers"
	"privacy_service/proto/privacy_service"
)

func main() {
	lis, err := net.Listen("tcp", ":52000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	db := configs.InitMySQL()

	privacy_service.RegisterPrivacyServiceServer(grpcServer, &handlers.PrivacyService{
		DB: db,
	})

	log.Println("Privacy service started on port 50900")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
