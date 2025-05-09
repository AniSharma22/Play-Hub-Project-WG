package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
	"project2/internal/domain/entities"
	mocks "project2/tests/mocks/service"
	"testing"
	"time"
)

func TestGetNotificationsHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotificationService := mocks.NewMockNotificationService(ctrl)
	handler := handlers.NewNotificationHandler(mockNotificationService)

	// Create test userId
	userID := uuid.New()
	notifications := []entities.Notification{
		{NotificationID: uuid.New(), Message: "Test notification 1", CreatedAt: time.Now()},
		{NotificationID: uuid.New(), Message: "Test notification 2", CreatedAt: time.Now()},
	}

	// Set up the mock service to return notifications
	mockNotificationService.EXPECT().GetUserNotifications(gomock.Any(), userID).Return(notifications, nil)

	// Create a new request with userId in the context
	req, err := http.NewRequest(http.MethodGet, "/notifications", nil)
	assert.NoError(t, err)
	ctx := context.WithValue(req.Context(), middleware.UserIdKey, userID.String())
	req = req.WithContext(ctx)

	// Create a response recorder to capture the output
	rr := httptest.NewRecorder()

	// Call the handler
	handler.GetNotificationsHandler(rr, req)

	// Assert the response code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Assert the response body
	var responseBody map[string]any
	err = json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Success", responseBody["message"])

	// Check if the notifications array was returned correctly
	responseNotifications, ok := responseBody["notifications"].([]any)
	assert.True(t, ok)
	assert.Equal(t, 2, len(responseNotifications))
}

func TestGetNotificationsHandler_UserIdNotInContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotificationService := mocks.NewMockNotificationService(ctrl)
	handler := handlers.NewNotificationHandler(mockNotificationService)

	// Create a new request without userId in context
	req, err := http.NewRequest(http.MethodGet, "/notifications", nil)
	assert.NoError(t, err)

	// Create a response recorder to capture the output
	rr := httptest.NewRecorder()

	// Call the handler
	handler.GetNotificationsHandler(rr, req)

	// Assert the response code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Assert the response body
	var responseBody map[string]string
	err = json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Could not find the userId", responseBody["message"])
}

func TestGetNotificationsHandler_InvalidUserId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotificationService := mocks.NewMockNotificationService(ctrl)
	handler := handlers.NewNotificationHandler(mockNotificationService)

	// Set up an invalid UUID
	invalidUserID := "invalid-uuid"

	// Create a new request with an invalid userId in the context
	req, err := http.NewRequest(http.MethodGet, "/notifications", nil)
	assert.NoError(t, err)
	ctx := context.WithValue(req.Context(), middleware.UserIdKey, invalidUserID)
	req = req.WithContext(ctx)

	// Create a response recorder to capture the output
	rr := httptest.NewRecorder()

	// Call the handler
	handler.GetNotificationsHandler(rr, req)

	// Assert the response code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Assert the response body
	var responseBody map[string]string
	err = json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Couldn't parse user id", responseBody["message"])
}

func TestGetNotificationsHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotificationService := mocks.NewMockNotificationService(ctrl)
	handler := handlers.NewNotificationHandler(mockNotificationService)

	// Create test userId
	userID := uuid.New()

	// Set up the mock service to return an error
	mockNotificationService.EXPECT().GetUserNotifications(gomock.Any(), userID).Return(nil, errors.New("some error"))

	// Create a new request with userId in the context
	req, err := http.NewRequest(http.MethodGet, "/notifications", nil)
	assert.NoError(t, err)
	ctx := context.WithValue(req.Context(), middleware.UserIdKey, userID.String())
	req = req.WithContext(ctx)

	// Create a response recorder to capture the output
	rr := httptest.NewRecorder()

	// Call the handler
	handler.GetNotificationsHandler(rr, req)

	// Assert the response code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Assert the response body
	var responseBody map[string]string
	err = json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Error fetching notifications", responseBody["message"])
}
