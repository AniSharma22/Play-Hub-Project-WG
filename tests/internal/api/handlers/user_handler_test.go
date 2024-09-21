package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
	"project2/internal/domain/entities"
	mocks "project2/tests/mocks/service"
	"testing"
)

func TestGetUserProfileHandler(t *testing.T) {
	// Create mock controller and defer its finish
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock userService
	mockUserService := mocks.NewMockUserService(ctrl)

	// Create an instance of UserHandler with the mocked service
	userHandler := handlers.NewUserHandler(mockUserService)

	// Test case where the user ID is "me"
	t.Run("success with 'me' as userID", func(t *testing.T) {
		// Create a mock user object
		userID, _ := uuid.NewUUID()
		expectedUser := &entities.User{
			UserID:   userID,
			Email:    "test@example.com",
			Username: "Test User",
		}

		// Simulate context with userID for 'me'
		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userID)

		// Mock the GetUserByID call
		mockUserService.EXPECT().GetUserByID(gomock.Any(), gomock.Eq(userID)).Return(expectedUser, nil)

		// Create request and recorder
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		// Set up router and invoke the handler
		r := mux.NewRouter()
		r.HandleFunc("/users/{id}", userHandler.GetUserProfileHandler)
		r.ServeHTTP(rr, req)

		// Assertions
		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Success", responseBody["message"])
		assert.Equal(t, expectedUser.UserID, responseBody["user"].(map[string]interface{})["id"])
	})

	// Test case where userID is invalid UUID
	t.Run("failure due to invalid UUID", func(t *testing.T) {
		// Create a request with invalid UUID
		req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
		rr := httptest.NewRecorder()

		// Set up router and invoke the handler
		r := mux.NewRouter()
		r.HandleFunc("/users/{id}", userHandler.GetUserProfileHandler)
		r.ServeHTTP(rr, req)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Couldn't parse user id", responseBody["message"])
	})

	// Test case where user is not found
	t.Run("failure when user not found", func(t *testing.T) {
		userID := "sample-uuid"

		// Simulate context with userID for 'me'
		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userID)

		// Mock GetUserByID to return an error
		mockUserService.EXPECT().GetUserByID(gomock.Any(), gomock.Eq(userID)).Return(nil, errors.New("user not found"))

		// Create request and recorder
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		// Set up router and invoke the handler
		r := mux.NewRouter()
		r.HandleFunc("/users/{id}", userHandler.GetUserProfileHandler)
		r.ServeHTTP(rr, req)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Couldn't get user", responseBody["message"])
	})
}
