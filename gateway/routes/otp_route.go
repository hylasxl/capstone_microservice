package routes

import (
	"gateway/handlers"
	"github.com/gorilla/mux"
)

func InitializeOTPRoutes(router *mux.Router, clients *ServiceClients) {
	otpRoutes := router.PathPrefix("/api/v1/otps").Subrouter()
	otpRoutes.HandleFunc("/send-forget-password", handlers.HandlerSendOTPForgetPasswordMessage(clients.UserService, clients.OTPService)).Methods("POST")
	otpRoutes.HandleFunc("/check-valid-otp", handlers.HandlerCheckValidOTP(clients.UserService, clients.OTPService)).Methods("POST")
}
