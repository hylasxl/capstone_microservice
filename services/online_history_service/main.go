package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"online_history_service/configs"
	"online_history_service/models"
)

func main() {
	lis, err := net.Listen("tcp", ":50058")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	db := configs.InitMySQL()
	err = db.AutoMigrate(
		&models.OnlineHistory{},
	)
	if err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	log.Println("Online history service started on port 50058")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
