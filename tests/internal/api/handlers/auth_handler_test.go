package handlers_test

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"project2/internal/api/handlers"
	"project2/internal/domain/entities"
	"project2/pkg/errs"
	mocks2 "project2/tests/mocks/service"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSignupHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mocking the authService
	authService := mocks2.NewMockAuthService(ctrl)
	// Test request body
	requestBody := map[string]string{
		"email":    "test.test@watchguard.com",
		"password": "StrongPass123!",
		"phone_no": "8888888888",
		"gender":   "male",
	}
	body, _ := json.Marshal(requestBody)

	// Create a new AuthHandler with mocked service
	authHandler := handlers.NewAuthHandler(authService)

	// Mock the signup response
	authService.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(&entities.User{
		UserID: uuid.New(),
		Role:   "user",
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authHandler.SignupHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jsonResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jsonResponse)

	assert.Equal(t, "user", jsonResponse["role"])
	assert.NotEmpty(t, jsonResponse["token"])
}

func TestSignupHandler_EmailConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks2.NewMockAuthService(ctrl)
	authHandler := handlers.NewAuthHandler(authService)

	// Mock conflict error for signup
	authService.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(nil, errs.ErrEmailExists)

	// Test request body
	requestBody := map[string]string{
		"email":    "test.test@watchguard.com",
		"password": "StrongPass123!",
		"phone_no": "8899776655",
		"gender":   "male",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authHandler.SignupHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var jsonResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jsonResponse)

	assert.Equal(t, "Email already exists", jsonResponse["message"])
}

func TestLoginHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks2.NewMockAuthService(ctrl)
	authHandler := handlers.NewAuthHandler(authService)

	// Test request body
	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "StrongPass123!",
	}
	body, _ := json.Marshal(requestBody)

	// Mock login success
	authService.EXPECT().Login(gomock.Any(), "test@example.com", gomock.Any()).Return(&entities.User{
		UserID: uuid.New(),
		Role:   "user",
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authHandler.LoginHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jsonResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jsonResponse)

	assert.Equal(t, "user", jsonResponse["role"])
	assert.NotEmpty(t, jsonResponse["token"])
}

func TestLoginHandler_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks2.NewMockAuthService(ctrl)
	authHandler := handlers.NewAuthHandler(authService)

	// Mock login failure due to invalid password
	authService.EXPECT().Login(gomock.Any(), "test.test@watchguard.com", gomock.Any()).Return(nil, errs.ErrInvalidPassword)

	// Test request body
	requestBody := map[string]string{
		"email":    "test.test@watchguard.com",
		"password": "WrongPassword123!",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authHandler.LoginHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var jsonResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jsonResponse)

	assert.Equal(t, "Invalid username or password", jsonResponse["message"])
}

func TestLogoutHandler_Success(t *testing.T) {
	authHandler := handlers.NewAuthHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	authHandler.LogoutHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jsonResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jsonResponse)

	assert.Equal(t, "Logout Successful", jsonResponse["message"])
}
