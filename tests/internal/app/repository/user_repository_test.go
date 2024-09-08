package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/app/repositories"
	"project2/internal/domain/entities"
	"testing"
	"time"
)

func TestUserRepo_CreateUser(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		user          *entities.User
		mockBehavior  func(sqlmock.Sqlmock, *entities.User)
		expectedError error
	}{
		{
			name: "Success",
			user: &entities.User{
				Username:     "testuser",
				Email:        "testuser@example.com",
				Password:     "hashedpassword",
				MobileNumber: "1234567890",
				Gender:       "M",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, user *entities.User) {
				userID := uuid.New()
				mock.ExpectQuery(`INSERT INTO users .+`).
					WithArgs(user.Username, user.Email, user.Password, user.MobileNumber, user.Gender).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(userID))
			},
			expectedError: nil,
		},
		{
			name: "Failure - Database Error",
			user: &entities.User{
				Username:     "testuser",
				Email:        "testuser@example.com",
				Password:     "hashedpassword",
				MobileNumber: "1234567890",
				Gender:       "M",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, user *entities.User) {
				mock.ExpectQuery(`INSERT INTO users .+`).
					WithArgs(user.Username, user.Email, user.Password, user.MobileNumber, user.Gender).
					WillReturnError(errors.New("database error"))
			},
			expectedError: errors.New("failed to insert user into PostgreSQL and retrieve user_id: database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call setup to get db and mock
			db, mock := setup()
			defer db.Close() // Close db after the test finishes

			// Set up mock behavior
			tc.mockBehavior(mock, tc.user)

			// Create a userRepo and call CreateUser
			repo := repositories.NewUserRepo(db)
			ctx := context.Background()
			resultID, err := repo.CreateUser(ctx, tc.user)

			// Assertions
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Equal(t, uuid.Nil, resultID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, resultID)
			}

			// Ensure all expectations are met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_FetchUserByEmail(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewUserRepo(db)
	ctx := context.Background()

	testCases := []struct {
		name          string
		email         string
		mockBehavior  func(sqlmock.Sqlmock, string)
		expectedUser  *entities.User
		expectedError error
	}{
		{
			name:  "Success",
			email: "test@example.com",
			mockBehavior: func(mock sqlmock.Sqlmock, email string) {
				rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "mobile_number", "gender", "role"}).
					AddRow(uuid.New(), "testuser", email, "hashedpassword", "1234567890", "M", "user")
				mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
					WithArgs(email).
					WillReturnRows(rows)
			},
			expectedUser: &entities.User{
				Email: "test@example.com",
				// Other fields will be populated by the mock
			},
			expectedError: nil,
		},
		{
			name:  "User Not Found",
			email: "nonexistent@example.com",
			mockBehavior: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
					WithArgs(email).
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: errors.New("no such user found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mock, tc.email)

			user, err := repo.FetchUserByEmail(ctx, tc.email)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tc.expectedUser.Email, user.Email)
			}
		})
	}
}

func TestUserRepo_FetchUserById(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewUserRepo(db)
	ctx := context.Background()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockBehavior  func(sqlmock.Sqlmock, uuid.UUID)
		expectedUser  *entities.User
		expectedError error
	}{
		{
			name: "Success",
			id:   uuid.New(),
			mockBehavior: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "mobile_number", "gender", "role"}).
					AddRow(id, "testuser", "test@example.com", "hashedpassword", "1234567890", "M", "user")
				mock.ExpectQuery("SELECT (.+) FROM users WHERE user_id = ?").
					WithArgs(id).
					WillReturnRows(rows)
			},
			expectedUser: &entities.User{
				UserID: uuid.New(), // This will be overwritten by the actual ID in the test
			},
			expectedError: nil,
		},
		{
			name: "User Not Found",
			id:   uuid.New(),
			mockBehavior: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE user_id = ?").
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: errors.New("no such user found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mock, tc.id)

			user, err := repo.FetchUserById(ctx, tc.id)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tc.id, user.UserID)
			}
		})
	}
}

func TestUserRepo_FetchAllUsers(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewUserRepo(db)
	ctx := context.Background()

	testCases := []struct {
		name          string
		mockBehavior  func(sqlmock.Sqlmock)
		expectedUsers []entities.User
		expectedError error
	}{
		{
			name: "Success",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "mobile_number", "gender"}).
					AddRow(uuid.New(), "user1", "user1@example.com", "hashedpassword1", "1234567890", "M").
					AddRow(uuid.New(), "user2", "user2@example.com", "hashedpassword2", "0987654321", "F")
				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			expectedUsers: []entities.User{
				{Username: "user1", Email: "user1@example.com"},
				{Username: "user2", Email: "user2@example.com"},
			},
			expectedError: nil,
		},
		{
			name: "No Users",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "mobile_number", "gender"})
				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			expectedUsers: []entities.User{},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mock)

			users, err := repo.FetchAllUsers(ctx)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedUsers), len(users))
				for i, user := range users {
					assert.Equal(t, tc.expectedUsers[i].Username, user.Username)
					assert.Equal(t, tc.expectedUsers[i].Email, user.Email)
				}
			}
		})
	}
}

func TestUserRepo_EmailAlreadyExists(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewUserRepo(db)
	ctx := context.Background()

	testCases := []struct {
		name         string
		email        string
		mockBehavior func(sqlmock.Sqlmock, string)
		expected     bool
	}{
		{
			name:  "Email Exists",
			email: "existing@example.com",
			mockBehavior: func(mock sqlmock.Sqlmock, email string) {
				rows := sqlmock.NewRows([]string{"1"}).AddRow(1)
				mock.ExpectQuery("SELECT 1 FROM users WHERE email = ?").
					WithArgs(email).
					WillReturnRows(rows)
			},
			expected: true,
		},
		{
			name:  "Email Does Not Exist",
			email: "new@example.com",
			mockBehavior: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery("SELECT 1 FROM users WHERE email = ?").
					WithArgs(email).
					WillReturnError(sql.ErrNoRows)
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mock, tc.email)

			exists := repo.EmailAlreadyExists(ctx, tc.email)

			assert.Equal(t, tc.expected, exists)
		})
	}
}

func TestUserRepo_FetchUserByUsername(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewUserRepo(db)
	ctx := context.Background()

	testCases := []struct {
		name          string
		username      string
		mockBehavior  func(sqlmock.Sqlmock, string)
		expectedUser  *entities.User
		expectedError error
	}{
		{
			name:     "Success",
			username: "testuser",
			mockBehavior: func(mock sqlmock.Sqlmock, username string) {
				rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "mobile_number", "gender", "role", "created_at", "updated_at"}).
					AddRow(uuid.New(), username, "test@example.com", "hashedpassword", "1234567890", "M", "user", time.Now(), time.Now())
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
					WithArgs(username).
					WillReturnRows(rows)
			},
			expectedUser: &entities.User{
				Username: "testuser",
				// Other fields will be populated by the mock
			},
			expectedError: nil,
		},
		{
			name:     "User Not Found",
			username: "nonexistent",
			mockBehavior: func(mock sqlmock.Sqlmock, username string) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
					WithArgs(username).
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: errors.New("no such user found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mock, tc.username)

			user, err := repo.FetchUserByUsername(ctx, tc.username)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tc.expectedUser.Username, user.Username)
			}
		})
	}
}
