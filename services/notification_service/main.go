package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"notification_service/configs"
	"notification_service/models"
)

func main() {
	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	DB := configs.InitMySQL()
	if err := DB.AutoMigrate(
		&models.Notification{},
		&models.NotificationType{},
	); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	grpcServer := grpc.NewServer()

	log.Println("Notification service started on port 50057")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
