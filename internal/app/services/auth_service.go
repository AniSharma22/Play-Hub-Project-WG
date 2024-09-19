package services

import (
	"context"
	"fmt"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/globals"
	"project2/pkg/utils"
)

type AuthService struct {
	userRepo    repository_interfaces.UserRepository
	userService service_interfaces.UserService
}

func NewAuthService(userRepo repository_interfaces.UserRepository, userService service_interfaces.UserService) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		userService: userService,
	}
}

// Signup registers a new user in the system.
func (a *AuthService) Signup(ctx context.Context, user *entities.User) (*entities.User, error) {
	// Check if email is already registered
	exists := a.userService.EmailAlreadyRegistered(ctx, user.Email)

	if exists {
		return nil, fmt.Errorf("email already registered: %w", errs.ErrEmailExists)
	}

	user.Username = utils.GetNameFromEmail(user.Email)

	// Create user
	userId, err := a.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("errs creating user: %w", err)
	}

	user, err = a.userRepo.FetchUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("errs fetching user: %w", err)
	}
	return user, nil
}

// Login authenticates a user with email and password.
func (a *AuthService) Login(ctx context.Context, email string, password []byte) (*entities.User, error) {
	// Fetch the user by email
	user, err := a.userRepo.FetchUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %w", errs.ErrUserNotFound)
	}

	// Compare the stored hashed password with the provided password
	if !utils.VerifyPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid password: %w", errs.ErrInvalidPassword)
	}

	globals.ActiveUser = user.UserID
	return user, nil
}
