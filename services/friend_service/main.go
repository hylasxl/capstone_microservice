package main

import (
	"friend_service/configs"
	"friend_service/handlers"
	"friend_service/models"
	"friend_service/proto/friend_service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db := configs.InitMySQL()

	grpcServer := grpc.NewServer()
	friend_service.RegisterFriendServiceServer(grpcServer, &handlers.FriendService{
		DB: db,
	})

	err = db.AutoMigrate(
		&models.FriendBlock{},
		&models.FriendFollow{},
		&models.FriendList{},
		&models.FriendListRequest{},
	)
	if err != nil {
		log.Fatalf("failed to auto migrate friend database: %v", err)
	}

	log.Println("Friend service started on port 50054")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
