package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"project2/internal/config"
	"project2/pkg/logger"
	"strings"
	"time"
)

// Define keys for the context
type contextKey string

const UserIdKey = contextKey("userId")
const RoleKey = contextKey("role")

// JwtAuthMiddleware check the token and
func JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			unauthorized(w, "Missing Authorization header")
			return
		}

		// Split the token from "Bearer <token>"
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			w.Header().Set("Content-Type", "application/json")
			unauthorized(w, "Missing token in Authorization header")
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return config.MY_SIGNING_KEY, nil
		})

		// Handle token validation errors
		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			logger.Logger.Errorw("Invalid token", "error", err, "time", time.Now())
			unauthorized(w, "Invalid or expired token")
			return
		}

		// Extract the claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			unauthorized(w, "Invalid Token")
			return
		}

		// Extract the userId as a string and convert it to uuid.UUID
		userId, ok := claims["userId"].(string)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			unauthorized(w, "Invalid Token")
			return
		}

		// Extract the role from the token claims
		role, ok := claims["role"].(string)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			unauthorized(w, "Invalid Token: Missing role")
			return
		}

		// Store userId and role in the context separately
		ctx := context.WithValue(r.Context(), UserIdKey, userId)
		ctx = context.WithValue(ctx, RoleKey, role)
		r = r.WithContext(ctx)

		// Proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// Helper to return unauthorized error response
func unauthorized(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusUnauthorized)
	jsonResponse := map[string]string{
		"code":    "401",
		"message": message,
	}
	json.NewEncoder(w).Encode(jsonResponse)
}
