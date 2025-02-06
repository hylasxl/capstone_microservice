package middlewares

import (
	"context"
	"encoding/json"
	"gateway/proto/auth_service"
	"log"
	"net/http"
	"strings"
	"time"
)

func AuthMiddleware(authService auth_service.AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "no auth header provided", nil)
				return
			}

			println(authHeader)

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				respondWithError(w, http.StatusUnauthorized, "no token provided", nil)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			res, err := authService.ValidateToken(ctx, &auth_service.ValidateTokenRequest{Token: token})
			if err != nil || !res.Valid {
				respondWithError(w, http.StatusUnauthorized, res.ErrorMessage, nil)
				return
			}

			isAuthorized, err := hasRequiredPermission(r.URL.Path, res.Permissions)
			log.Println(r.URL.Path)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "permission check failed", nil)
				return
			}
			if !isAuthorized {
				respondWithError(w, http.StatusUnauthorized, "user is unauthorized with this url permission", nil)
				return
			}

			ctx = context.WithValue(ctx, "userId", res.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func hasRequiredPermission(requestURL string, permissions []string) (bool, error) {
	for _, permission := range permissions {
		if permission == requestURL {
			return true, nil
		}
	}
	return false, nil
}

func respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	if err != nil {
		errorResponse["details"] = err.Error()
	}

	err = json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		return
	}
}
