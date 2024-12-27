package main

import (
	"context"
	"database/sql"
	"google.golang.org/grpc"
	"log"
	"net"
	"user_service/configs"
	"user_service/handlers"
	"user_service/models"
	"user_service/proto/user_service"
)

func main() {
	ctx := context.Background()
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	DB := configs.InitMySQL()
	DBConn, _ := DB.DB()
	defer func(DBConn *sql.DB) {
		err := DBConn.Close()
		if err != nil {

		}
	}(DBConn)

	Cld, err := configs.InitCloudinary(ctx)
	if err != nil {
		log.Fatalf("failed to init cloudinary: %v", err)
	}

	grpcServer := grpc.NewServer()
	user_service.RegisterUserServiceServer(grpcServer, &handlers.UserService{
		CloudinaryClient: (*handlers.CloudinaryService)(Cld),
		DB:               DB,
	})

	err = DB.AutoMigrate(
		&models.Account{},
		&models.AccountAvatar{},
		&models.AccountChangeNameHistory{},
		&models.AccountInfo{},
		&models.AccountRole{},
		&models.AccountAvatarHistory{},
	)
	if err != nil {
		return
	} else {
		log.Println("User database migrate successfully")
	}

	log.Println("User service started on port 50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
