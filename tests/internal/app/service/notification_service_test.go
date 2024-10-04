package service_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/domain/entities"
	"testing"
)

func TestNotificationService_GetUserNotifications(t *testing.T) {
	// Test data
	userId := uuid.New()
	ctx := context.TODO()

	notifications := []entities.Notification{{
		NotificationID: userId,
		UserID:         userId,
		Message:        "Test Notification",
	}}

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError bool
		mockOutput    []entities.Notification
	}{
		{
			name: "success",
			mockSetup: func() {
				mockNotificationRepo.EXPECT().
					FetchUserNotifications(ctx, userId).
					Return(notifications, nil)
			},
			mockOutput:    notifications,
			expectedError: false,
		},
		{
			name: "failure",
			mockSetup: func() {
				mockNotificationRepo.EXPECT().
					FetchUserNotifications(ctx, userId).
					Return(nil, errors.New("test errs"))
			},
			mockOutput:    nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()
			// Run the mock setup for each test case
			tt.mockSetup()

			// Call the function under test
			result, err := notificationService.GetUserNotifications(ctx, userId)

			// Assert the results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockOutput, result)
			}
		})
	}
}
