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

	ctrl := gomock.NewController(t)

	// Create mock userService
	mockUserService := mocks.NewMockUserService(ctrl)

	// Create an instance of UserHandler with the mocked service
	userHandler := handlers.NewUserHandler(mockUserService)

	// Test case where the user ID is "me"
	t.Run("success with 'me' as userID", func(t *testing.T) {
		userIdUuid := uuid.New()
		userID := userIdUuid.String()
		expectedUser := &entities.User{
			UserID:   userIdUuid,
			Email:    "test@example.com",
			Username: "Test User",
		}

		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userID)
		mockUserService.EXPECT().GetUserByID(gomock.Any(), gomock.Eq(userIdUuid)).Return(expectedUser, nil).Times(1)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		r := mux.NewRouter()
		r.HandleFunc("/users/{id}", userHandler.GetUserProfileHandler)
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Success", responseBody["message"])
		//assert.Equal(t, userID, responseBody["user"].(map[string]interface{})["id"])
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
		userIdUuid := uuid.New()
		userID := userIdUuid.String()

		// Simulate context with userID for 'me'
		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userID)

		// Mock GetUserByID to return an error
		mockUserService.EXPECT().GetUserByID(gomock.Any(), gomock.Eq(userIdUuid)).Return(nil, errors.New("user not found")).Times(1)

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

func TestGetAllUsersHandler(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock the userService
	mockUserService := mocks.NewMockUserService(ctrl)

	// Create an instance of UserHandler with the mocked userService
	userHandler := handlers.NewUserHandler(mockUserService)

	// Define some mock users
	mockUsers := []entities.User{
		{
			UserID:   uuid.New(),
			Email:    "user1@example.com",
			Username: "user1",
		},
		{
			UserID:   uuid.New(),
			Email:    "user2@example.com",
			Username: "user2",
		},
	}

	// Test successful retrieval of all users
	t.Run("success retrieving all users", func(t *testing.T) {
		// Mock the GetAllUsers function to return the mock users
		mockUserService.EXPECT().GetAllUsers(gomock.Any()).Return(mockUsers, nil)

		// Create a request and response recorder
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rr := httptest.NewRecorder()

		// Call the handler function
		http.HandlerFunc(userHandler.GetAllUsersHandler).ServeHTTP(rr, req)

		// Check if the response code is 200 OK
		assert.Equal(t, http.StatusOK, rr.Code)

		// Parse the response body
		var responseBody map[string]any
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)

		// Convert the "users" key from responseBody to the expected structure
		usersData, ok := responseBody["users"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, usersData, len(mockUsers))

		for i, userData := range usersData {
			userMap := userData.(map[string]interface{})
			assert.Equal(t, mockUsers[i].Email, userMap["email"])
			assert.Equal(t, mockUsers[i].Username, userMap["username"])
			assert.Equal(t, mockUsers[i].UserID.String(), userMap["user_id"])
		}

		// Check the response message and code
		assert.Equal(t, "Success", responseBody["message"])
		assert.Equal(t, float64(http.StatusOK), responseBody["code"])
	})

	// Test case where GetAllUsers fails
	t.Run("failure to get users", func(t *testing.T) {
		// Mock the GetAllUsers function to return an error
		mockUserService.EXPECT().GetAllUsers(gomock.Any()).Return(nil, errors.New("some error"))

		// Create a request and response recorder
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rr := httptest.NewRecorder()

		// Call the handler function
		http.HandlerFunc(userHandler.GetAllUsersHandler).ServeHTTP(rr, req)

		// Check if the response code is 500 Internal Server Error
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		// Parse the response body
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)

		// Check the error message
		assert.Equal(t, "Couldn't get users", responseBody["message"])
	})

	t.Run("failure to encode response", func(t *testing.T) {
		// Mock the GetAllUsers function to return users
		mockUserService.EXPECT().GetAllUsers(gomock.Any()).Return(mockUsers, nil)

		// Create a request
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		// Custom ResponseWriter that simulates an error when writing the response
		failingWriter := &failingResponseWriter{httptest.NewRecorder()}

		// Call the handler function with the failing writer
		http.HandlerFunc(userHandler.GetAllUsersHandler).ServeHTTP(failingWriter, req)

		// Check if the response code is 500 Internal Server Error
		assert.Equal(t, http.StatusInternalServerError, failingWriter.Code)

		// Check the error message in the response body
		//responseBody := failingWriter.Body.String()
		//fmt.Print(responseBody)
		//assert.Contains(t, responseBody, "Error occurred while encoding the response to json")
	})
}
