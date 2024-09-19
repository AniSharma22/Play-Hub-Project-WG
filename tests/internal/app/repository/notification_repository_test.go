package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/app/repositories"
	"project2/internal/domain/entities"
	"testing"
	"time"
)

func TestNotificationRepo_CreateNotification(t *testing.T) {
	// Define test cases
	tests := []struct {
		name          string
		notification  *entities.Notification
		mockBehavior  func(sqlmock.Sqlmock, *entities.Notification)
		expectedError error
	}{
		{
			name: "success",
			notification: &entities.Notification{
				UserID:  uuid.New(),
				Message: "Test notification",
				IsRead:  false,
			},
			mockBehavior: func(mock sqlmock.Sqlmock, notification *entities.Notification) {
				mock.ExpectQuery(`INSERT INTO notifications .+`).
					WithArgs(notification.UserID, notification.Message, notification.IsRead).
					WillReturnRows(sqlmock.NewRows([]string{"notification_id"}).AddRow(uuid.New()))
			},
			expectedError: nil,
		},
		{
			name: "query failed",
			notification: &entities.Notification{
				UserID:  uuid.New(),
				Message: "Test notification",
				IsRead:  false,
			},
			mockBehavior: func(mock sqlmock.Sqlmock, notification *entities.Notification) {
				mock.ExpectQuery(`INSERT INTO notifications .+`).
					WithArgs(notification.UserID, notification.Message, notification.IsRead).
					WillReturnError(fmt.Errorf("database errs"))
			},
			expectedError: fmt.Errorf("failed to create notification: database errs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock database and expectations
			db, mock := setup() // You can use the setup function from the previous example
			defer db.Close()

			// Apply the mock behavior for this test case
			tt.mockBehavior(mock, tt.notification)

			// Create a new notification repository
			repo := repositories.NewNotificationRepo(db)

			// Call the CreateNotification function
			ctx := context.Background()
			context.TODO()
			_, err := repo.CreateNotification(ctx, tt.notification)

			// Assert the result
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			// Ensure all expectations are met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFetchNotificationByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	notificationID := uuid.New()
	timeStampz := time.Now()

	query := `SELECT notification_id, user_id, message, is_read, created_at FROM notifications WHERE notification_id = \$1`

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"notification_id", "user_id", "message", "is_read", "created_at"}).
			AddRow(notificationID, uuid.New(), "Test message", false, timeStampz)

		mock.ExpectQuery(query).WithArgs(notificationID).WillReturnRows(rows)

		// Create a new notification repository
		repo := repositories.NewNotificationRepo(db)

		notification, err := repo.FetchNotificationByID(context.Background(), notificationID)

		assert.NoError(t, err)
		assert.NotNil(t, notification)
		assert.Equal(t, notificationID, notification.NotificationID)
	})

	t.Run("no rows", func(t *testing.T) {
		mock.ExpectQuery(query).WithArgs(notificationID).WillReturnError(sql.ErrNoRows)

		// Create a new notification repository
		repo := repositories.NewNotificationRepo(db)

		notification, err := repo.FetchNotificationByID(context.Background(), notificationID)

		assert.NoError(t, err)
		assert.Nil(t, notification)
	})

	t.Run("query errs", func(t *testing.T) {
		mock.ExpectQuery(query).WithArgs(notificationID).WillReturnError(errors.New("query errs"))

		// Create a new notification repository
		repo := repositories.NewNotificationRepo(db)

		_, err := repo.FetchNotificationByID(context.Background(), notificationID)

		assert.Error(t, err)
		assert.Equal(t, "failed to fetch notification by ID: query errs", err.Error())
	})
}

func TestFetchUserNotifications(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewNotificationRepo(db)
	userID := uuid.New()
	timestampz := time.Now()

	query := `SELECT notification_id, user_id, message, is_read, created_at FROM notifications WHERE user_id = \$1 ORDER BY created_at DESC`

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"notification_id", "user_id", "message", "is_read", "created_at"}).
			AddRow(uuid.New(), userID, "Test message 1", false, timestampz).
			AddRow(uuid.New(), userID, "Test message 2", true, timestampz)

		mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)

		notifications, err := repo.FetchUserNotifications(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, notifications, 2)
	})

	t.Run("no notifications", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"notification_id", "user_id", "message", "is_read", "created_at"})

		mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)

		notifications, err := repo.FetchUserNotifications(context.Background(), userID)

		assert.NoError(t, err)
		assert.Empty(t, notifications)
	})

	t.Run("query errs", func(t *testing.T) {
		mock.ExpectQuery(query).WithArgs(userID).WillReturnError(errors.New("query errs"))

		_, err := repo.FetchUserNotifications(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, "failed to fetch user notifications: query errs", err.Error())
	})
}
func TestMarkNotificationAsRead(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewNotificationRepo(db)
	notificationID := uuid.New()

	query := `UPDATE notifications SET is_read = TRUE WHERE notification_id = \$1`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(query).WithArgs(notificationID).WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.MarkNotificationAsRead(context.Background(), notificationID)

		assert.NoError(t, err)
	})

	t.Run("exec errs", func(t *testing.T) {
		mock.ExpectExec(query).WithArgs(notificationID).WillReturnError(errors.New("exec errs"))

		err := repo.MarkNotificationAsRead(context.Background(), notificationID)

		assert.Error(t, err)
		assert.Equal(t, "failed to mark notification as read: exec errs", err.Error())
	})
}

func TestDeleteNotificationByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewNotificationRepo(db)
	notificationID := uuid.New()

	query := `DELETE FROM notifications WHERE notification_id = \$1`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(query).WithArgs(notificationID).WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.DeleteNotificationByID(context.Background(), notificationID)

		assert.NoError(t, err)
	})

	t.Run("exec errs", func(t *testing.T) {
		mock.ExpectExec(query).WithArgs(notificationID).WillReturnError(errors.New("exec errs"))

		err := repo.DeleteNotificationByID(context.Background(), notificationID)

		assert.Error(t, err)
		assert.Equal(t, "failed to delete notification: exec errs", err.Error())
	})
}
