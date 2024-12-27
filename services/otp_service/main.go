package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"otp_service/configs"
	"otp_service/models"
)

func main() {
	lis, err := net.Listen("tcp", ":50059")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	db := configs.InitMySQL()

	err = db.AutoMigrate(
		&models.OTPInputs{},
		models.OTPRetakePassword{})
	if err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	log.Println("OTP service started on port 50059")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
