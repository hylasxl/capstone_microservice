package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/otp_service"
	"gateway/proto/user_service"
	"log"
	"net/http"
	"strconv"
	"time"
)

func HandlerSendOTPForgetPasswordMessage(userClient user_service.UserServiceClient, otpClient otp_service.OTPServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in SendOTPForgetPasswordMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid payload", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userResp, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(in.AccountID, 10),
		})

		if err != nil || !userResp.IsValid {
			respondWithError(w, http.StatusBadRequest, "invalid userId", err)
			return
		}

		sendOTPResp, err := otpClient.SendForgetPasswordOTP(ctx, &otp_service.SendForgetPasswordOTPRequest{
			AccountID: in.AccountID,
			Email:     in.Email,
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed sending OTP", err)
			return
		}

		var res = &SendOTPForgetPasswordMessageResponse{
			Success: sendOTPResp.Success,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandlerCheckValidOTP(userClient user_service.UserServiceClient, otpClient otp_service.OTPServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in CheckValidOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid payload", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userResp, err := userClient.CheckValidUser(ctx, &user_service.CheckValidUserRequest{
			UserId: strconv.FormatUint(in.AccountID, 10),
		})
		if err != nil || !userResp.IsValid {
			respondWithError(w, http.StatusBadRequest, "invalid userId", err)
			return
		}

		checkResp, err := otpClient.CheckValidOTP(ctx, &otp_service.CheckValidOTPRequest{
			AccountID: in.AccountID,
			OTP:       uint64(in.OTP),
		})

		var res CheckValidOTPResponse

		if checkResp != nil {
			res.Success = checkResp.IsValid
			res.Attempts = int(checkResp.Attempts)
		} else {
			log.Println("CheckValidOTPResponse is nil")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
