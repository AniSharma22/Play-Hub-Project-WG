package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"project2/internal/config"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"project2/pkg/utils"
	"strings"
	"time"
)

// Define keys for the context
type contextKey string

const UserIdKey = contextKey("userId")
const RoleKey = contextKey("role")
const IpAddrKey = contextKey("ipAddr")

// JwtAuthMiddleware checks the token and validates user access
func JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Getting the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Logger.Warnw("Missing Authorization header", "method", r.Method, "url", r.URL.String(), "time", time.Now())
			w.Header().Set("Content-Type", "application/json")
			//unauthorized(w, "1001", "Missing Authorization header") // Code 1001 for missing auth header
			errs.UnauthorizedError("1001", "Authorization header is missing. Please provide a valid token in the Authorization header.").ToJson2(w)
			return
		}

		// Splitting the token from "Bearer <token>"
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			logger.Logger.Warnw("Missing token in Authorization header", "method", r.Method, "url", r.URL.String(), "time", time.Now())
			w.Header().Set("Content-Type", "application/json")
			//unauthorized(w, "1002", "Missing token in Authorization header") // Code 1002 for missing token
			errs.UnauthorizedError("1002", "Bearer token is missing. Ensure your request contains a valid Bearer token.").ToJson2(w)
			return
		}

		// Parsing and validating the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.Logger.Errorw("Invalid signing method", "method", r.Method, "url", r.URL.String(), "time", time.Now())
				return nil, errors.New("invalid signing method")
			}
			return config.MY_SIGNING_KEY, nil
		})

		// Handling token validation errors
		if err != nil || !token.Valid {
			logger.Logger.Errorw("Invalid token", "error", err, "method", r.Method, "url", r.URL.String(), "time", time.Now())
			w.Header().Set("Content-Type", "application/json")
			//unauthorized(w, "1003", "Invalid or expired token") // Code 1003 for invalid or expired token
			errs.UnauthorizedError("1003", "Your token is invalid or expired. Please log in again to obtain a new token.").ToJson2(w)
			return
		}

		// Extracting the claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			logger.Logger.Warnw("Invalid token claims", "method", r.Method, "url", r.URL.String(), "time", time.Now())
			w.Header().Set("Content-Type", "application/json")
			//unauthorized(w, "1004", "Invalid Token") // Code 1004 for invalid token claims
			errs.UnauthorizedError("1004", "Invalid token: Claims validation failed. Your token contains invalid information.").ToJson2(w)
			return
		}

		// Extracting the userId as a string and converting it to uuid.UUID
		userId, ok := claims["userId"].(string)
		if !ok {
			logger.Logger.Warnw("Invalid Token: Missing userId", "method", r.Method, "url", r.URL.String(), "time", time.Now())
			w.Header().Set("Content-Type", "application/json")
			//unauthorized(w, "1005", "Invalid Token: Missing userId") // Code 1005 for missing userId
			errs.UnauthorizedError("1005", "Invalid token: User ID is missing. Ensure your token contains the necessary claims.").ToJson2(w)
			return
		}

		// Extracting the role from the token claims
		role, ok := claims["role"].(string)
		if !ok {
			logger.Logger.Warnw("Invalid Token: Missing role", "method", r.Method, "url", r.URL.String(), "time", time.Now())
			w.Header().Set("Content-Type", "application/json")
			//unauthorized(w, "1006", "Invalid Token: Missing role") // Code 1006 for missing role
			errs.UnauthorizedError("1006", "Invalid token: User role is missing. A valid token should specify the user role.").ToJson2(w)
			return
		}

		// Extracting the IP address from the request
		ipAddr := utils.GetIP(r)

		// Storing userId, role, and ipAddr in the context separately
		ctx := context.WithValue(r.Context(), UserIdKey, userId)
		ctx = context.WithValue(ctx, RoleKey, role)
		ctx = context.WithValue(ctx, IpAddrKey, ipAddr)
		r = r.WithContext(ctx)

		logger.Logger.Infow("User authenticated successfully", "userId", userId, "role", role, "ipAddr", ipAddr, "method", r.Method, "url", r.URL.String(), "time", time.Now())

		// Proceeding to the next handler
		next.ServeHTTP(w, r)
	})
}

// Helper to return unauthorized error response
func unauthorized(w http.ResponseWriter, errorCode string, message string) {
	w.WriteHeader(http.StatusUnauthorized)
	jsonResponse := map[string]string{
		"error_code": errorCode,
		"message":    message,
	}
	json.NewEncoder(w).Encode(jsonResponse)
}
