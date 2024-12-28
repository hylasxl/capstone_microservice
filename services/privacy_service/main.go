package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"privacy_service/configs"
	"privacy_service/handlers"
	"privacy_service/models"
	"privacy_service/proto/privacy_service"
)

func main() {
	lis, err := net.Listen("tcp", ":50071")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	db := configs.InitMySQL()

	privacy_service.RegisterPrivacyServiceServer(grpcServer, &handlers.PrivacyService{
		DB: db,
	})

	err = db.AutoMigrate(
		&models.DataPrivacy{},
		&models.DataPrivacyIndex{},
	)
	if err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	log.Println("Privacy service started on port 50071")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
