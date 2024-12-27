package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"gateway/proto/auth_service"
	"gateway/proto/privacy_service"
	"gateway/proto/user_service"
	"log"
	"net/http"
	"strings"
	"time"
)

func HandlerLogin(authClient auth_service.AuthServiceClient, userClient user_service.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
		}

		if request.Username == "" || request.Password == "" {
			http.Error(w, "Username and password are required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		authResp, err := userClient.Login(ctx, &user_service.LoginRequest{
			Username: request.Username,
			Password: request.Password,
		})

		if err != nil || authResp.Error != "" {
			http.Error(w, "Authentication failed: "+authResp.Error, http.StatusUnauthorized)
			return
		}

		permissionResp, err := authClient.GetPermissions(ctx, &auth_service.GetPermissionsRequest{
			RoleId: authResp.RoleId,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var claims = &auth_service.JWTClaims{
			AccountId:   authResp.UserId,
			Permissions: permissionResp.Url,
			RoleId:      authResp.RoleId,
			Issuer:      "SyncIO",
			Subject:     "Authentication",
			Audience:    "Client SyncIO",
		}

		tokenRes, err := authClient.GenerateTokens(ctx, &auth_service.GenerateTokensRequest{
			Claims: claims,
		})

		if err != nil || tokenRes.Error != "" {
			http.Error(w, "Generating token failed: "+tokenRes.Error, http.StatusUnauthorized)
		}

		result := &LoginResponse{
			AccessToken:  tokenRes.AccessToken,
			RefreshToken: tokenRes.RefreshToken,
			UserID:       authResp.UserId,
			Success:      true,
			JWTClaims:    claims,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
func HandlerSignUp(userClient user_service.UserServiceClient, privacyClient privacy_service.PrivacyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			// Log the error
			log.Printf("Error decoding request body: %v", err)
			respondWithError(w, http.StatusBadRequest, "Invalid payload request", err)
			return
		}

		println(request.BirthDate)

		var imgBytes []byte
		if request.Image != "" {
			var err error
			imgBytes, err = base64.StdEncoding.DecodeString(request.Image)
			if err != nil {
				// Log the error
				log.Printf("Error decoding image: %v", err)
				respondWithError(w, http.StatusBadRequest, "Invalid image file", err)
				return
			}
		}

		datePart := strings.Split(request.BirthDate, "T")[0]

		birthDate, err := time.Parse("2006-01-02", datePart)
		if err != nil {
			// Log the error
			log.Printf("Error parsing birth date: %v", err)
			respondWithError(w, http.StatusBadRequest, "Invalid birth date", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		singUpResp, err := userClient.Signup(ctx, &user_service.SignupRequest{
			Username:    request.Username,
			Password:    request.Password,
			FirstName:   request.FirstName,
			LastName:    request.LastName,
			Gender:      request.Gender,
			Email:       request.Email,
			PhoneNumber: request.Phone,
			Birthday:    birthDate.Unix(),
			Avatar:      imgBytes,
		})
		if singUpResp.Error != "" {
			log.Printf("Error during user signup: %v", singUpResp.Error)
			respondWithError(w, http.StatusBadRequest, "Signup failed: "+singUpResp.Error, nil)
			return
		}

		privacyInitResp, err := privacyClient.CreateAccountPrivacyInit(ctx, &privacy_service.CreateAccountPrivacyInitRequest{
			AccountID: singUpResp.AccountId,
		})
		if privacyInitResp.Error != "" {
			log.Printf("Error initializing privacy: %v", err)
			log.Printf("CreateAccountPrivacyInit failed with error: %v", privacyInitResp.Error)
			respondWithError(w, http.StatusBadRequest, "CreateAccountPrivacyInit failed: "+privacyInitResp.Error, nil)
			return
		}

		registerResponse := &SignUpResponse{
			UserID:  singUpResp.AccountId,
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(registerResponse); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
