package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"project2/internal/api/middleware"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"project2/internal/config"
)

func TestJwtAuthMiddleware(t *testing.T) {
	// Mock next handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value(middleware.UserIdKey)
		role := r.Context().Value(middleware.RoleKey)
		assert.NotNil(t, userId)
		assert.NotNil(t, role)
		w.WriteHeader(http.StatusOK)
	})

	// Create a router and register the middleware
	router := mux.NewRouter()
	router.Use(middleware.JwtAuthMiddleware)
	router.Handle("/protected", nextHandler).Methods("GET")

	// Test cases
	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name:         "Missing Authorization header",
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]string{"error_code": "1001", "message": "Missing Authorization header"},
		},
		{
			name:         "Missing token in Authorization header",
			authHeader:   "Bearer ",
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]string{"error_code": "1002", "message": "Missing token in Authorization header"},
		},
		{
			name:         "Invalid token",
			authHeader:   "Bearer invalid.token.here",
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]string{"error_code": "1003", "message": "Invalid or expired token"},
		},
		{
			name:         "Valid token",
			authHeader:   createValidToken(t, "userIdValue", "userRoleValue"),
			expectedCode: http.StatusOK,
			expectedBody: nil, // no specific body expected for success
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedBody != nil {
				var responseBody map[string]string
				err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, responseBody)
			}
		})
	}
}

// Helper function to create a valid JWT token
func createValidToken(t *testing.T, userId, role string) string {
	claims := jwt.MapClaims{
		"userId": userId,
		"role":   role,
		"exp":    jwt.TimeFunc().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(config.MY_SIGNING_KEY)
	assert.NoError(t, err)
	return signedToken
}
