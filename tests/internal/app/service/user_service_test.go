package service_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/domain/entities"
	"project2/pkg/globals"
	"project2/pkg/utils"
	"testing"
)

func TestUserService_Signup(t *testing.T) {
	userId, _ := uuid.NewUUID()
	tests := []struct {
		name                       string
		newUser                    *entities.User
		mockCreateUserRepo         error
		mockEmailAlreadyRegistered bool // Changed from errs to bool
		expectedError              bool
		CreateUserCalled           int
	}{
		{
			name: "Successful Signup",
			newUser: &entities.User{
				Email:        "test.test@watchguard.com",
				Password:     "TestPassword",
				MobileNumber: "8989898989",
				Gender:       "male",
			},
			mockCreateUserRepo:         nil,
			mockEmailAlreadyRegistered: false, // Email does not exist
			expectedError:              false,
			CreateUserCalled:           1,
		},
		{
			name: "Signup Failure",
			newUser: &entities.User{
				Email:        "test.test@watchguard.com",
				Password:     "TestPassword",
				MobileNumber: "8989898989",
				Gender:       "male",
			},
			mockCreateUserRepo:         errors.New("mock repository errs"),
			mockEmailAlreadyRegistered: false,
			expectedError:              true,
			CreateUserCalled:           1,
		},
		{
			name: "Signup Failure - Email Already Registered",
			newUser: &entities.User{
				Email:        "test.test@watchguard.com",
				Password:     "TestPassword",
				MobileNumber: "8989898989",
				Gender:       "male",
			},
			mockCreateUserRepo:         nil,
			mockEmailAlreadyRegistered: true, // Email exists
			expectedError:              true,
			CreateUserCalled:           0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			// Prepare the context
			ctx := context.TODO()

			// Mock the repository call with the expected behavior
			mockUserRepo.EXPECT().
				CreateUser(ctx, tt.newUser).
				Return(userId, tt.mockCreateUserRepo).
				Times(tt.CreateUserCalled)

			// Mock the repository call with the expected behaviour for EmailAlreadyExists
			mockUserRepo.EXPECT().
				EmailAlreadyExists(ctx, tt.newUser.Email).
				Return(tt.mockEmailAlreadyRegistered).
				Times(1)

			// Call the Signup method
			err := userService.Signup(ctx, tt.newUser)

			// Assert the expected errs outcome
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify if the global variable is set
			if !tt.expectedError {
				assert.Equal(t, userId, globals.ActiveUser)
			}
		})
	}
}

func TestUserService_EmailAlreadyExists(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		mockRepoResult bool
		mockRepoError  error
		expected       bool
	}{
		{
			name:           "Email Exists",
			email:          "test@example.com",
			mockRepoResult: true,
			mockRepoError:  nil,
			expected:       true,
		},
		{
			name:           "Error While Scanning in Repository",
			email:          "test@example.com",
			mockRepoResult: true,
			mockRepoError:  errors.New("sql errs"),
			expected:       true,
		},
		{
			name:           "Email Does Not Exist",
			email:          "test@example.com",
			mockRepoResult: false,
			mockRepoError:  nil,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			// Prepare the context
			ctx := context.TODO()

			mockUserRepo.EXPECT().
				EmailAlreadyExists(ctx, tt.email).
				Return(tt.mockRepoResult).
				Times(1)

			exists := userService.EmailAlreadyRegistered(ctx, tt.email)

			assert.Equal(t, tt.expected, exists)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	userId, _ := uuid.NewUUID()
	hashedPass, _ := utils.GetHashedPassword([]byte("ValidPassword"))

	// Sample user data
	validUser := &entities.User{
		UserID:   userId,
		Email:    "test@example.com",
		Password: hashedPass,
	}

	tests := []struct {
		name               string
		email              string
		password           []byte
		mockFetchUserError error
		mockFetchedUser    *entities.User
		expectedError      bool
		expectedErrorMsg   string
		expectedUserID     uuid.UUID
	}{
		{
			name:               "Successful Login",
			email:              "test@example.com",
			password:           []byte("ValidPassword"),
			mockFetchUserError: nil,
			mockFetchedUser:    validUser,
			expectedError:      false,
			expectedUserID:     userId,
		},
		{
			name:               "User Not Found",
			email:              "notfound@example.com",
			password:           []byte("AnyPassword"),
			mockFetchUserError: nil,
			mockFetchedUser:    nil,
			expectedError:      true,
			expectedErrorMsg:   "user not found",
			expectedUserID:     uuid.Nil,
		},
		{
			name:               "Invalid Password",
			email:              "test@example.com",
			password:           []byte("InvalidPassword"),
			mockFetchUserError: nil,
			mockFetchedUser:    validUser,
			expectedError:      true,
			expectedErrorMsg:   "invalid password",
			expectedUserID:     uuid.Nil,
		},
		{
			name:               "Fetch User Error",
			email:              "test@example.com",
			password:           []byte("ValidPassword"),
			mockFetchUserError: errors.New("database errs"),
			mockFetchedUser:    nil,
			expectedError:      true,
			expectedErrorMsg:   "failed to fetch user by email: database errs",
			expectedUserID:     uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			// Prepare the context
			ctx := context.TODO()

			// Mock the repository call for FetchUserByEmail
			mockUserRepo.EXPECT().
				FetchUserByEmail(ctx, tt.email).
				Return(tt.mockFetchedUser, tt.mockFetchUserError).
				Times(1)

			// Call the Login method
			user, err := userService.Login(ctx, tt.email, tt.password)

			// Assert the expected errs outcome
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserID, user.UserID)
			}

			// Verify if the global variable is set correctly on successful login
			if !tt.expectedError {
				assert.Equal(t, tt.expectedUserID, globals.ActiveUser)
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	userId, _ := uuid.NewUUID()
	// Sample user data
	validUser := &entities.User{
		UserID: userId,
		Email:  "test@example.com",
	}

	tests := []struct {
		name             string
		userID           uuid.UUID
		mockFetchUser    *entities.User
		mockFetchError   error
		expectedUser     *entities.User
		expectedErrorMsg string
		expectedError    bool
	}{
		{
			name:           "User Found",
			userID:         userId,
			mockFetchUser:  validUser,
			mockFetchError: nil,
			expectedUser:   validUser,
			expectedError:  false,
		},
		{
			name:             "User Not Found",
			userID:           uuid.New(),
			mockFetchUser:    nil,
			mockFetchError:   errors.New("user not found"),
			expectedUser:     nil,
			expectedError:    true,
			expectedErrorMsg: "user not found",
		},
		{
			name:             "Repository Error",
			userID:           uuid.New(),
			mockFetchUser:    nil,
			mockFetchError:   errors.New("database errs"),
			expectedUser:     nil,
			expectedError:    true,
			expectedErrorMsg: "database errs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			// Prepare the context
			ctx := context.TODO()

			// Mock the repository call for FetchUserById
			mockUserRepo.EXPECT().
				FetchUserById(ctx, tt.userID).
				Return(tt.mockFetchUser, tt.mockFetchError).
				Times(1)

			// Call the GetUserByID method
			user, err := userService.GetUserByID(ctx, tt.userID)

			// Assert the expected outcome
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}

func TestUserService_GetUserByEmail(t *testing.T) {
	userId, _ := uuid.NewUUID()
	// Sample user data
	validUser := &entities.User{
		UserID: userId,
		Email:  "test@example.com",
	}

	tests := []struct {
		name             string
		email            string
		mockFetchUser    *entities.User
		mockFetchError   error
		expectedUser     *entities.User
		expectedErrorMsg string
		expectedError    bool
	}{
		{
			name:           "User Found",
			email:          "test@example.com",
			mockFetchUser:  validUser,
			mockFetchError: nil,
			expectedUser:   validUser,
			expectedError:  false,
		},
		{
			name:             "User Not Found",
			email:            "notfound@example.com",
			mockFetchUser:    nil,
			mockFetchError:   errors.New("user not found"),
			expectedUser:     nil,
			expectedError:    true,
			expectedErrorMsg: "user not found",
		},
		{
			name:             "Repository Error",
			email:            "test@example.com",
			mockFetchUser:    nil,
			mockFetchError:   errors.New("database errs"),
			expectedUser:     nil,
			expectedError:    true,
			expectedErrorMsg: "database errs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			// Prepare the context
			ctx := context.TODO()

			// Mock the repository call for FetchUserByEmail
			mockUserRepo.EXPECT().
				FetchUserByEmail(ctx, tt.email).
				Return(tt.mockFetchUser, tt.mockFetchError).
				Times(1)

			// Call the GetUserByEmail method
			user, err := userService.GetUserByEmail(ctx, tt.email)

			// Assert the expected outcome
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}

func TestUserService_GetUserByUsername(t *testing.T) {
	userId, _ := uuid.NewUUID()
	// Sample user data
	validUser := &entities.User{
		UserID:   userId,
		Username: "testuser",
	}

	tests := []struct {
		name             string
		username         string
		mockFetchUser    *entities.User
		mockFetchError   error
		expectedUser     *entities.User
		expectedErrorMsg string
		expectedError    bool
	}{
		{
			name:           "User Found",
			username:       "testuser",
			mockFetchUser:  validUser,
			mockFetchError: nil,
			expectedUser:   validUser,
			expectedError:  false,
		},
		{
			name:             "User Not Found",
			username:         "notfounduser",
			mockFetchUser:    nil,
			mockFetchError:   errors.New("user not found"),
			expectedUser:     nil,
			expectedError:    true,
			expectedErrorMsg: "user not found",
		},
		{
			name:             "Repository Error",
			username:         "testuser",
			mockFetchUser:    nil,
			mockFetchError:   errors.New("database errs"),
			expectedUser:     nil,
			expectedError:    true,
			expectedErrorMsg: "database errs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			// Prepare the context
			ctx := context.TODO()

			// Mock the repository call for FetchUserByUsername
			mockUserRepo.EXPECT().
				FetchUserByUsername(ctx, tt.username).
				Return(tt.mockFetchUser, tt.mockFetchError).
				Times(1)

			// Call the GetUserByUsername method
			user, err := userService.GetUserByUsername(ctx, tt.username)

			// Assert the expected outcome
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}
