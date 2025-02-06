package main

import (
	"context"
	"firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"log"
	"net"
	"notification_service/configs"
	"notification_service/handlers"
	"notification_service/models"
	"notification_service/proto/notification_service"
	"os"
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
		&models.Device{},
	); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	firebaseApp, err := initFirebase()
	if err != nil {
		log.Fatalf("failed to init firebase: %v", err)
		return
	}

	grpcServer := grpc.NewServer()
	notification_service.RegisterNotificationServiceServer(grpcServer, &handlers.NotificationService{
		DB:          DB,
		FirebaseApp: firebaseApp,
	})

	log.Println("Notification service started on port 50057")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initFirebase() (*firebase.App, error) {
	credentialsPath := os.Getenv("FIREBASE_CREDENTIALS")
	if credentialsPath == "" {
		log.Fatalf("FIREBASE_CREDENTIALS environment variable is not set")
	}

	opt := option.WithCredentialsFile(credentialsPath)
	conf := &firebase.Config{ProjectID: "syncio-7a920"}
	app, err := firebase.NewApp(context.Background(), conf, opt)
	if err != nil {
		return nil, err
	}
	return app, nil
}
