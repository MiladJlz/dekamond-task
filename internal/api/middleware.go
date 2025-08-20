package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/MiladJlz/dekamond-task/internal/auth"
)

// JWTAuthMiddleware validates JWT tokens and adds user phone to request context
func (h *Handler) JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			JSONError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			JSONError(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		phone, err := auth.ValidateJWT(tokenString, h.jwtSecret)
		if err != nil {
			h.logger.Errorw("jwt validation failed", "error", err)
			JSONError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_phone", phone)
		r = r.WithContext(ctx)

		h.logger.Infow("jwt validated", "phone", phone)
		next.ServeHTTP(w, r)
	}
}

// Optional: Add a helper function to get user phone from context
func GetUserPhoneFromContext(r *http.Request) string {
	if phone, ok := r.Context().Value("user_phone").(string); ok {
		return phone
	}
	return ""
}
