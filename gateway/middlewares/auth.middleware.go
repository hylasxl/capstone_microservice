package middlewares

import (
	"context"
	"gateway/proto/auth_service"
	"net/http"
	"strings"
	"time"
)

func AuthMiddleware(authService auth_service.AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			res, err := authService.ValidateToken(ctx, &auth_service.ValidateTokenRequest{Token: token})
			if err != nil || !res.Valid {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			isAuthorized, err := hasRequiredPermission(r.URL.Path, res.Permissions)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
			if !isAuthorized {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
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
