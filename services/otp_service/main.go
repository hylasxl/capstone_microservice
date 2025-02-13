package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"otp_service/configs"
	"otp_service/handlers"
	"otp_service/proto/otp_service"
	"time"
)

func main() {
	err := os.Setenv("TZ", "Asia/Bangkok")
	if err != nil {
		return
	}
	time.Local = time.FixedZone("UTC+7", 7*3600)
	lis, err := net.Listen("tcp", ":50059")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db := configs.InitMySQL()
	grpcServer := grpc.NewServer()
	otp_service.RegisterOTPServiceServer(grpcServer, &handlers.OTPService{
		DB: db,
	})

	log.Println("OTP service started on port 50059")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
